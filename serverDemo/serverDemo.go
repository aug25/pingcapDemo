package main

/*
This is the entry point to startup the program
provide a json config file to setup the TiDB cluster
*/
import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"pingcapDemo/nodeServer"
)

func main() {
	rpc.Register(new(nodeServer.NodeServer))

	listener, err := net.Listen("tcp", "0.0.0.0:3302")
	fmt.Println("Start listening on tcp:3302")
	if err != nil {
		fmt.Println(err)
	}

	for {
		conn, err := listener.Accept() // accept request
		if err != nil {
			continue
		}

		go func(conn net.Conn) { // handle request
			fmt.Fprintf(os.Stdout, "new client in coming: %s\n", conn.RemoteAddr())
			fmt.Fprintf(os.Stdout, "I am %s\n", conn.LocalAddr())
			jsonrpc.ServeConn(conn)
		}(conn)
	}

}
