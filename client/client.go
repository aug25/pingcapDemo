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
	"sync"
)

const clientport = "3302"

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var text string
	for text != "exit" { // break the loop if text == "exit"
		fmt.Print("Enter your test opeartion: <node> <operation> <args>")
		scanner.Scan()
		text = scanner.Text()
		batchtext := strings.Split(text, ";")
		wg := &sync.WaitGroup{}
		for _, nodetext := range batchtext {
			// only node name provided, do nothing
			nodetext = strings.TrimSpace(nodetext)
			testarr := strings.Split(nodetext, " ")
			if len(testarr) < 2 {
				fmt.Println("please provide a operation for node " + testarr[0])
				continue
			}
			// perform test method on testhost
			var testhost = testarr[0]
			var testmethod = testarr[1:]
			wg.Add(1)
			go testOnNode(testhost, testmethod, wg)
		}
		wg.Wait()
	}

}

func testOnNode(testhost string, method []string, wg *sync.WaitGroup) {
	cliconn, err := jsonrpc.Dial("tcp", testhost+":"+clientport)
	fmt.Print("connected to " + testhost + ", ")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Connect failed, please enter valid node name")
		wg.Done()
		return
	}
	req := nodeServer.TestRequest{Echostr: "hello, node server"}
	var res nodeServer.TestResponse
	var remoteMethod string
	switch method[0] {
	case "kill":
		switch testhost[:2] {
		case "pd":
			req.Serverbin = "pd-server"
			//testOnPD(cliconn,testmethod)
		case "ti":
			switch testhost[2:4] {
			case "kv":
				req.Serverbin = "tikv-server"
				//testOnTiKV(cliconn,testmethod)
			case "db":
				req.Serverbin = "tidb-server"
				//testOnTiDB(cliconn,testmethod)
			}
		}
		remoteMethod = "KillServer"
	case "cpu":
		fmt.Printf("Stress Cpu for %s seconds\n", method[1])
		remoteMethod = "StressCPU"
		req.Cpubusytime, _ = strconv.ParseInt(method[1], 10, 64)
	case "io":
		fmt.Printf("Stress IO for %sMB\n", method[1])
		remoteMethod = "StressIO"
		req.IO1Mcount, _ = strconv.ParseInt(method[1], 10, 64)
	default:
		remoteMethod = "GetCurrentIP"
	}
	err = cliconn.Call("NodeServer."+remoteMethod, req, &res)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("respone from %s is %s\n", testhost, res.Respstr)
	wg.Done()
}

// Deprecated: currently just provide common test like echo,get IP,kill, all move to abstract Node
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

// Deprecated: currently just provide common test like echo,get IP,kill, all move to abstract Node
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

// Deprecated: currently just provide common test like echo,get IP,kill, all move to abstract Node
func testOnTiDB(cliconn *rpc.Client, method []string) {
	tidbreq := nodeServer.TiDBTestRequest{Echostr: "hello, TiDB server"}
	var tidbres nodeServer.TiDBTestResponse
	var remoteMethod string
	switch method[0] {
	case "kill":
		remoteMethod = "KillTiDBServer"
	case "cpu":
		fmt.Printf("Stress TiDB Cpu for %s seconds\n", method[1])
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
