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
	"pingcapDemo/util"
	"strconv"
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
var pdclusterstr, titkvpdstr, tidbpdbstr, networkid string
var pdnodes, tikvnodes, tidbnodes, allnodes []*node.Node
var haproxy, control node.Node

func main() {
	flag.Parse()
	pdnodes = make([]*node.Node, *pdno)
	tikvnodes = make([]*node.Node, *tikvno)
	tidbnodes = make([]*node.Node, *tidbno)

	var a, b, c bytes.Buffer
	a.Grow(100)
	b.Grow(100)
	c.Grow(100)
	a.WriteString("--initial-cluster=")
	b.WriteString("--pd=")
	c.WriteString("--path=")
	for i := 0; i < *pdno; i++ {

		pdName := fmt.Sprintf("pd%d", i)
		a.WriteString(pdName)
		a.WriteString("=http://")
		a.WriteString(pdName)
		a.WriteString(":2380,")

		b.WriteString(pdName)
		b.WriteString(":2379,")

		c.WriteString(pdName)
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
	fmt.Println("Time to limit resource for docker container or go to 'control' node from another terminal and run client test")
	fmt.Println("CMD: <nodename> <resource> <args> or enter exit to shutdown and clean cluster")
	for text != "exit" { // break the loop if text == "exit"
		scanner.Scan()
		text = scanner.Text()
		node, err := util.ParseCMD(allnodes, text)
		if err != nil {
			log(err)
		} else {
			limitContainerResources(cli, node)
		}
	}
	shutdownTiDBTestCluster(cli)
}

func limitContainerResources(cli *client.Client, node *node.Node) {
	switch node.TestMethod {
	case "cpu":
		var cpushare, cpuperiod, cpuquota int64
		if len(node.TestArgs) > 0 {
			cpushare, _ = strconv.ParseInt(node.TestArgs[0], 10, 64)
		}
		if len(node.TestArgs) > 1 {
			cpuperiod, _ = strconv.ParseInt(node.TestArgs[1], 10, 64)
		}
		if len(node.TestArgs) > 2 {
			cpuquota, _ = strconv.ParseInt(node.TestArgs[2], 10, 64)
		}
		setCPU(cli, node, cpushare, cpuperiod, cpuquota)
		return
	case "memory":
		var memory, memoryswap int64
		if len(node.TestArgs) > 0 {
			memory, _ = strconv.ParseInt(node.TestArgs[0], 10, 64)
		}
		if len(node.TestArgs) > 1 {
			memoryswap, _ = strconv.ParseInt(node.TestArgs[1], 10, 64)
		}

		setMemory(cli, node, memory, memoryswap)
		return
	case "io":
		return
	case "network":
		return
	default:
		fmt.Println("Invalid input, currently support cpu/memory limitation")
		return
	}
}

func setCPU(cli *client.Client, node *node.Node, cpushare int64, cpuperiod int64, cpuquota int64) {
	fmt.Printf("Set %s CPU share to %d, compare to other nodes default 1024, cpu period to %d, cpu quota to %d \n", node.Name, cpushare, cpuperiod, cpuquota)
	var updateConfig = container.UpdateConfig{
		Resources: container.Resources{
			CPUShares: cpushare,
			CPUPeriod: cpuperiod,
			CPUQuota:  cpuquota,
		},
	}
	cli.ContainerUpdate(context.Background(), node.ContainerID, updateConfig)
}

func setMemory(cli *client.Client, node *node.Node, memory int64, memoryswap int64) {
	if memoryswap < memory {
		fmt.Println("MemorySwap(memory+swap) must larger than Memory")
		return
	}
	fmt.Printf("Set %s Memory to %s bytes, Memory+Swap to %s bytes \n", node.Name, strconv.FormatInt(memory, 10), strconv.FormatInt(memoryswap, 10))
	var updateConfig = container.UpdateConfig{
		Resources: container.Resources{
			Memory:     memory,
			MemorySwap: memoryswap,
		},
	}
	cli.ContainerUpdate(context.Background(), node.ContainerID, updateConfig)
}

func startupTiDBTestCluster(cli *client.Client) {
	fmt.Printf("Create network %s for cluster\n", mynetwork)
	networkresp, err := cli.NetworkCreate(context.Background(), mynetwork, types.NetworkCreate{})
	networkid = networkresp.ID
	log(err)

	fmt.Println("Create and startup PD cluster")
	for i, n := range pdnodes {
		n = new(node.Node)
		pdnodes[i] = n
		n.Name = fmt.Sprintf("pd%d", i)
		n.Type = node.NodePD
		n.ContainerID = createContainer(cli, n)
		err = cli.NetworkConnect(context.Background(), networkid, n.ContainerID, nil)
		log(err)
		go startContainer(n.ContainerID, cli)
	}
	time.Sleep(time.Second * 20)

	fmt.Println("Create and startup TiKV cluster")
	for i, n := range tikvnodes {
		n = new(node.Node)
		tikvnodes[i] = n
		n.Name = fmt.Sprintf("tikv%d", i)
		n.Type = node.NodeTiKV
		n.ContainerID = createContainer(cli, n)
		err = cli.NetworkConnect(context.Background(), networkid, n.ContainerID, nil)
		log(err)
		go startContainer(n.ContainerID, cli)
	}
	allnodes = append(pdnodes, tikvnodes...)
	time.Sleep(time.Second * 20)

	fmt.Println("Create and startup TiDB cluster")
	for i, n := range tidbnodes {
		n = new(node.Node)
		tidbnodes[i] = n
		n.Name = fmt.Sprintf("tidb%d", i)
		n.Type = node.NodeTiDB
		n.ContainerID = createContainer(cli, n)
		err = cli.NetworkConnect(context.Background(), networkid, n.ContainerID, nil)
		log(err)
		go startContainer(n.ContainerID, cli)
	}
	allnodes = append(allnodes, tidbnodes...)
	time.Sleep(time.Second * 20)

	fmt.Println("Create and startup Haproxy")
	haproxy = node.Node{
		Name: "haproxy",
		Type: node.NodeProxy,
	}

	haproxy.ContainerID = createContainer(cli, &haproxy)
	cli.NetworkConnect(context.Background(), networkid, haproxy.ContainerID, nil)
	startContainer(haproxy.ContainerID, cli)
	allnodes = append(allnodes, &haproxy)

	fmt.Println("Create and startup Control Node")
	control = node.Node{
		Name: "control",
		Type: node.NodeControl,
	}
	control.ContainerID = createContainer(cli, &control)
	cli.NetworkConnect(context.Background(), networkid, control.ContainerID, nil)
	startContainer(control.ContainerID, cli)
	allnodes = append(allnodes, &control)

}

func shutdownTiDBTestCluster(cli *client.Client) {
	fmt.Println("Stop and remove entire Cluster")
	stopContainer(haproxy.ContainerID, cli)
	removeContainer(haproxy.ContainerID, cli)
	stopContainer(control.ContainerID, cli)
	removeContainer(control.ContainerID, cli)
	for _, tidbnode := range tidbnodes {
		stopContainer(tidbnode.ContainerID, cli)
		removeContainer(tidbnode.ContainerID, cli)
	}
	for _, tikvnode := range tikvnodes {
		stopContainer(tikvnode.ContainerID, cli)
		removeContainer(tikvnode.ContainerID, cli)
	}
	for _, pdnode := range pdnodes {
		stopContainer(pdnode.ContainerID, cli)
		removeContainer(pdnode.ContainerID, cli)
	}
	time.Sleep(time.Second * 3)
	fmt.Printf("Remove network %s\n", mynetwork)
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
	var nodemount = []mount.Mount{
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
			Source:   rootPath + "/log",
			Target:   "/log",
			ReadOnly: false,
		}}
	hostConfig = &container.HostConfig{
		Mounts: nodemount,
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
		hostConfig = &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   rootPath + "/config",
					Target:   "/usr/local/etc/haproxy",
					ReadOnly: false,
				},
			},
			PortBindings: nat.PortMap{
				mysqlport:   []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: mysqlport}},
				haproxyport: []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: haproxyport}},
			},
		}
	case node.NodeControl:
		config = &container.Config{
			Image:    "alpine",
			Hostname: cnode.Name,
			Cmd:      []string{"/usr/bin/top"},
		}
		hostConfig = &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   rootPath + "/crossplatform/client",
					Target:   "/testbin/client",
					ReadOnly: false,
				},
			},
		}
	}
	body, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, cnode.Name)
	log(err)
	if err == nil {
		fmt.Printf("Container: %s create SUCCESS\n", body.ID[:12])
	}
	return body.ID
}

// Startup container
func startContainer(containerID string, cli *client.Client) {
	err := cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	log(err)
	if err == nil {
		fmt.Printf("Container: %s startup SUCCESS\n", containerID[:12])
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
	timeout := time.Second * 2
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
		fmt.Printf("Container: %s removed\n", containerID[:12])
	}
}

func log(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
