package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"pingcapDemo/nodeServer"
	"strconv"
	"strings"
)

const clientport = "3302"

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var text string
	for text != "exit" { // break the loop if text == "exit"
		fmt.Print("Enter your test opeartion: <node> <operation> <args>")
		scanner.Scan()
		text = scanner.Text()
		// only node name provided, do nothing
		testarr := strings.Split(text, " ")
		if len(testarr) < 2 {
			continue
		}
		// perform test method on testhost
		var testhost = testarr[0]
		var testmethod = testarr[1:]
		fmt.Println("connected to " + testhost)
		cliconn, err := jsonrpc.Dial("tcp", testhost+":"+clientport)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Connect failed, please enter valid node name")
			continue
		}
		var serverbin string
		switch testhost[:2] {
		case "pd":
			serverbin = "pd-server"
			//testOnPD(cliconn,testmethod)
		case "ti":
			switch testhost[2:4] {
			case "kv":
				serverbin = "tikv-server"
				//testOnTiKV(cliconn,testmethod)
			case "db":
				serverbin = "tidb-server"
				//testOnTiDB(cliconn,testmethod)
			}
		}
		testOnNode(cliconn, testmethod, serverbin)
	}

}

func testOnNode(cliconn *rpc.Client, method []string, serverbin string) {
	req := nodeServer.TestRequest{Echostr: "hello, node server", Serverbin: serverbin}
	var res nodeServer.TestResponse
	var remoteMethod string
	switch method[0] {
	case "kill":
		remoteMethod = "KillServer"
	case "cpu":
		fmt.Printf("Stree Cpu for %s seconds\n", method[1])
		remoteMethod = "StressCPU"
		req.Cpubusytime, _ = strconv.ParseInt(method[1], 10, 64)
	default:
		remoteMethod = "GetCurrentIP"
	}
	err := cliconn.Call("NodeServer."+remoteMethod, req, &res)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("respone is %s\n", res.Respstr)
}

// Deprecated: currently just provide common test like echo,get IP,kill, all move to abstract NodeType
func testOnPD(cliconn *rpc.Client, method []string) {
	pdreq := nodeServer.PdTestRequest{"hello, PD server"}
	var pdres nodeServer.PdTestResponse
	var remoteMethod string
	switch method[0] {
	case "kill":
		remoteMethod = "KillPdServer"
	default:
		remoteMethod = "GetCurrentIP"
	}
	err := cliconn.Call("PdNodeServer."+remoteMethod, pdreq, &pdres)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("respone is %s\n", pdres.Respstr)
}

// Deprecated: currently just provide common test like echo,get IP,kill, all move to abstract NodeType
func testOnTiKV(cliconn *rpc.Client, method []string) {
	tikvreq := nodeServer.TiKVTestRequest{"hello, TiKV server"}
	var tikvres nodeServer.TiKVTestResponse
	var remoteMethod string
	switch method[0] {
	case "kill":
		remoteMethod = "KillTiKVServer"
	default:
		remoteMethod = "GetCurrentIP"
	}
	err := cliconn.Call("TiKVNodeServer."+remoteMethod, tikvreq, &tikvres)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("respone is %s\n", tikvres.Respstr)
}

// Deprecated: currently just provide common test like echo,get IP,kill, all move to abstract NodeType
func testOnTiDB(cliconn *rpc.Client, method []string) {
	tidbreq := nodeServer.TiDBTestRequest{Echostr: "hello, TiDB server"}
	var tidbres nodeServer.TiDBTestResponse
	var remoteMethod string
	switch method[0] {
	case "kill":
		remoteMethod = "KillTiDBServer"
	case "cpu":
		fmt.Printf("Stree TiDB Cpu for %s seconds\n", method[1])
		remoteMethod = "StressCPU"
		tidbreq.Cpubusytime, _ = strconv.ParseInt(method[1], 10, 64)
	default:
		remoteMethod = "GetCurrentIP"
	}
	err := cliconn.Call("TiDBNodeServer."+remoteMethod, tidbreq, &tidbres)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("respone is %s\n", tidbres.Respstr)
}
