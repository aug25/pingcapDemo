package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"pingcapDemo/nodeServer"
	"strings"
)

const clientport = "3302"

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var text string
	for text != "exit" { // break the loop if text == "exit"
		fmt.Print("Enter your test opeartion: <node> <operation> ")
		scanner.Scan()
		text = scanner.Text()
		testarr:= strings.Split(text," ")
		if len(testarr)<2{
			continue
		}
		var testhost = testarr[0]
		var testmethod = testarr[1]
		fmt.Println("connected to "+testhost)
		cliconn, err := jsonrpc.Dial("tcp", testhost+":"+clientport)
		if err != nil {
			fmt.Println(err)
		}
		switch testhost[:2] {
		case "pd":
			testOnPD(cliconn,testmethod)
		case "ti":
			switch testhost[2:4] {
			case "kv":
				testOnTiKV(cliconn,testmethod)
			case "db":
				testOnTiDB(cliconn,testmethod)
			}
		}
	}

}

func testOnPD(cliconn *rpc.Client,method string) {
	pdreq := nodeServer.PdTestRequest{"hello, pd server"}
	var pdres nodeServer.PdTestResponse
	var remoteMethod string
	switch method {
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

func testOnTiKV(cliconn *rpc.Client,method string) {
	tikvreq := nodeServer.TiKVTestRequest{"hello, TiKV server"}
	var tikvres nodeServer.TiKVTestResponse
	var remoteMethod string
	switch method {
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
func testOnTiDB(cliconn *rpc.Client,method string) {
	tidbreq := nodeServer.TiDBTestRequest{"hello, TiDB server"}
	var tidbres nodeServer.TiDBTestResponse
	var remoteMethod string
	switch method {
	case "kill":
		remoteMethod = "KillTiDBServer"
	default:
		remoteMethod = "GetCurrentIP"
	}
	err := cliconn.Call("TiDBNodeServer."+remoteMethod, tidbreq, &tidbres)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("respone is %s\n", tidbres.Respstr)
}
