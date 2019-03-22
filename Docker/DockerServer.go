package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"os"
	"path/filepath"
	"pingcapDemo/node"
	"time"
)

const (
	mynetwork    = "mypingcapnetwork"
	pdclientport = "2379"
	pdpeerport   = "2380"
	tikvport     = "20160"
	tidbport     = "4000"
	tidbhttpport = "10080"
	mysqlport    = "3690"
	haproxyport  = "8080"
)

var pdno = flag.Int("pdno", 3, "number of PD server to run")
var tikvno = flag.Int("tikvno", 3, "number of TiKV server to run")
var tidbno = flag.Int("tidbno", 2, "number of TiDB server to run")

// Just a demo, not a good practice
// Be careful about global variable, := in func will create local var and mask global var
var pdclusterstr, titkvpdstr, tidbpdbstr, haproxyid,controlid, networkid string
var pdids, tikvids, tidbids []string
var pdnodes, tikvnodes, tidbnodes []node.Node
var haproxy,control node.Node

func main() {
	flag.Parse()
	pdids = make([]string, *pdno)
	tikvids = make([]string, *tikvno)
	tidbids = make([]string, *tidbno)

	pdnodes = make([]node.Node, *pdno)
	tikvnodes = make([]node.Node, *tikvno)
	tidbnodes = make([]node.Node, *tidbno)

	var a, b, c bytes.Buffer
	a.Grow(100)
	b.Grow(100)
	c.Grow(100)
	a.WriteString("--initial-cluster=")
	b.WriteString("--pd=")
	c.WriteString("--path=")
	for i, n := range pdnodes {
		n.Name = fmt.Sprintf("pd%d", i)
		a.WriteString(n.Name)
		a.WriteString("=http://")
		a.WriteString(n.Name)
		a.WriteString(":2380,")

		b.WriteString(n.Name)
		b.WriteString(":2379,")

		c.WriteString(n.Name)
		c.WriteString(":2379,")

	}
	pdclusterstr = a.String()[:a.Len()-1]
	titkvpdstr = b.String()[:b.Len()-1]
	tidbpdbstr = c.String()[:c.Len()-1]

	//change to lower version depend on your docker env
	var cli, err = client.NewClientWithOpts(client.WithVersion("1.39"))
	log(err)
	defer cli.Close()

	//listImage(cli)
	startupTiDBTestCluster(cli)
	scanner := bufio.NewScanner(os.Stdin)
	var text string
	fmt.Println("Time to do some test from 'control' node, enter exit to shutdown and clean cluster")
	for text != "exit" { // break the loop if text == "exit"
		scanner.Scan()
		text = scanner.Text()
	}
	shutdownTiDBTestCluster(cli)
}

func startupTiDBTestCluster(cli *client.Client) {
	networkresp, err := cli.NetworkCreate(context.Background(), mynetwork, types.NetworkCreate{})
	networkid = networkresp.ID
	log(err)

	fmt.Println("Create and startup PD cluster")
	for i, n := range pdnodes {
		n.Name = fmt.Sprintf("pd%d", i)
		n.Type = node.NodePD
		pdids[i] = createContainer(cli, &n)
		err = cli.NetworkConnect(context.Background(), networkid, pdids[i], nil)
		log(err)
		go startContainer(pdids[i], cli)
	}

	time.Sleep(time.Second * 20)

	fmt.Println("Create and startup TiKV cluster")
	for i, n := range tikvnodes {
		n.Name = fmt.Sprintf("tikv%d", i)
		n.Type = node.NodeTiKV
		tikvids[i] = createContainer(cli, &n)
		err = cli.NetworkConnect(context.Background(), networkid, tikvids[i], nil)
		log(err)
		go startContainer(tikvids[i], cli)
	}
	time.Sleep(time.Second * 20)

	fmt.Println("Create and startup TiDB cluster")
	for i, n := range tidbnodes {
		n.Name = fmt.Sprintf("tidb%d", i)
		n.Type = node.NodeTiDB
		tidbids[i] = createContainer(cli, &n)
		err = cli.NetworkConnect(context.Background(), networkid, tidbids[i], nil)
		log(err)
		go startContainer(tidbids[i], cli)
	}
	time.Sleep(time.Second * 20)

	fmt.Println("Create and startup Haproxy")
	haproxy = node.Node{
		Name: "haproxy",
		Type: node.NodeProxy,
	}

	haproxyid = createContainer(cli, &haproxy)
	cli.NetworkConnect(context.Background(), networkid, haproxyid, nil)
	go startContainer(haproxyid, cli)

	fmt.Println("Create and startup Control Node")
	control = node.Node{
		Name: "control",
		Type: node.NodeControl,
	}
	controlid = createContainer(cli, &control)
	cli.NetworkConnect(context.Background(), networkid, controlid, nil)
	go startContainer(controlid, cli)
}

func shutdownTiDBTestCluster(cli *client.Client) {
	fmt.Println("Stop and remove entire Cluster")
	stopContainer(haproxyid, cli)
	removeContainer(haproxyid, cli)
	stopContainer(controlid, cli)
	removeContainer(controlid, cli)
	for _, tidbid := range tidbids {
		stopContainer(tidbid, cli)
		removeContainer(tidbid, cli)
	}
	for _, tikvid := range tikvids {
		stopContainer(tikvid, cli)
		removeContainer(tikvid, cli)
	}
	for _, pdid := range pdids {
		stopContainer(pdid, cli)
		removeContainer(pdid, cli)
	}
	time.Sleep(time.Second * 3)

	cli.NetworkRemove(context.Background(), networkid)
}

// List images
func listImage(cli *client.Client) {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	log(err)
	for _, image := range images {
		fmt.Println(image)
		fmt.Println()
	}
}

// Create container
func createContainer(cli *client.Client, cnode *node.Node) string {

	wd, _ := os.Getwd()
	rootPath := filepath.Dir(wd)
	var config *container.Config
	var hostConfig *container.HostConfig
	var mymount = []mount.Mount{
		{
			Type:     mount.TypeBind,
			Source:   rootPath + "/data",
			Target:   "/data",
			ReadOnly: false,
		},
		{
			Type:     mount.TypeBind,
			Source:   rootPath + "/output",
			Target:   "/output",
			ReadOnly: false,
		},
		{
			Type:     mount.TypeBind,
			Source:   rootPath + "/crossplatform/serverDemo",
			Target:   "/testbin/serverDemo",
			ReadOnly: false,
		},
		{
			Type:     mount.TypeBind,
			Source:   rootPath + "/crossplatform/client",
			Target:   "/testbin/client",
			ReadOnly: false,
		},
		{
			Type:     mount.TypeBind,
			Source:   rootPath + "/log",
			Target:   "/log",
			ReadOnly: false,
		}}
	hostConfig = &container.HostConfig{
		Mounts: mymount,
	}
	switch cnode.Type {
	case node.NodePD:
		config = &container.Config{
			Image:    "pingcap/pd:latest",
			Hostname: cnode.Name,
			ExposedPorts: nat.PortSet{
				pdclientport: struct{}{},
				pdpeerport:   struct{}{},
			},
			Cmd: []string{"--name=" + cnode.Name,
				"--data-dir=/data/" + cnode.Name,
				"--log-file=/log/" + cnode.Name + ".log",
				"--client-urls=http://0.0.0.0:" + pdclientport,
				"--advertise-client-urls=http://" + cnode.Name + ":" + pdclientport,
				"--peer-urls=http://0.0.0.0:" + pdpeerport,
				"--advertise-peer-urls=http://" + cnode.Name + ":" + pdpeerport,
				pdclusterstr,
			},
		}
	case node.NodeTiKV:
		config = &container.Config{
			Image: "pingcap/tikv:latest",
			ExposedPorts: nat.PortSet{
				tikvport: struct{}{},
			},
			Hostname: cnode.Name,
			Cmd: []string{
				"--data-dir=/data/" + cnode.Name,
				"--log-file=/log/" + cnode.Name + ".log",
				"--addr=0.0.0.0:" + tikvport,
				"--advertise-addr=" + cnode.Name + ":" + tikvport,
				titkvpdstr,
			},
		}
	case node.NodeTiDB:
		config = &container.Config{
			Image: "pingcap/tidb:latest",
			ExposedPorts: nat.PortSet{
				tidbport:     struct{}{},
				tidbhttpport: struct{}{},
			},
			Hostname: cnode.Name,
			Cmd: []string{
				tidbpdbstr,
				"--store=tikv",
				"--log-file=/log/" + cnode.Name + ".log",
			},
		}
	case node.NodeProxy:
		config = &container.Config{
			Image: "haproxy:latest",
			ExposedPorts: nat.PortSet{
				mysqlport:   struct{}{},
				haproxyport: struct{}{},
			},
			Hostname: cnode.Name,
			Cmd:      []string{},
		}
		hostConfig = &container.HostConfig{Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   rootPath + "/config",
				Target:   "/usr/local/etc/haproxy",
				ReadOnly: false,
			},
		},
			PortBindings: nat.PortMap{
				mysqlport: []nat.PortBinding{{HostIP:   "0.0.0.0", HostPort: mysqlport,},},
				haproxyport: []nat.PortBinding{{HostIP:   "0.0.0.0", HostPort: haproxyport,},},
			},
		}
	case node.NodeControl:
		config = &container.Config{
			Image: "alpine",
			Hostname: cnode.Name,
			Cmd:[]string{"/usr/bin/top"},
		}
	}
	body, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, cnode.Name)
	log(err)
	if err == nil {
		fmt.Printf("Container: %s create SUCCESS\n",body.ID[:12])
	}
	return body.ID
}

// Startup container
func startContainer(containerID string, cli *client.Client) {
	err := cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	log(err)
	if err == nil {
		fmt.Printf("Container: %s startup SUCCESS\n",containerID[:12])
	}
	//startup serverDemo on the node and listening for test request
	execConfig := types.ExecConfig{
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		Cmd:          []string{"/testbin/serverDemo"},
		Tty:          true,
		Detach:       false,
	}
	exec, err := cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	execAttachConfig := types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	}
	containerConn, err := cli.ContainerExecAttach(context.Background(), exec.ID, execAttachConfig)
	_, _, err = containerConn.Reader.ReadLine()

}

// Stop container
func stopContainer(containerID string, cli *client.Client) {
	timeout := time.Second * 10
	err := cli.ContainerStop(context.Background(), containerID, &timeout)
	log(err)
	if err == nil {
		fmt.Printf("Container: %s stopped\n", containerID[:12])
	}
}

// Remove container
func removeContainer(containerID string, cli *client.Client) {
	err := cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
	log(err)
	if err == nil {
		fmt.Printf("Container: %s removed\n",containerID[:12])
	}
}

func log(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
