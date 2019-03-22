package node

import (
	"fmt"
)

type nodeType string

const (
	NodeTiDB  nodeType = "TiDB"
	NodePD    nodeType = "PD"
	NodeTiKV  nodeType = "TiKV"
	NodeProxy nodeType = "Haproxy"
	NodeControl nodeType = "Control"
)

type Node struct{
	Name string
	Type nodeType
}

func (node *Node) Startup(){
	fmt.Println("Startup Node:"+ node.Name +" Type:" + string(node.Type)  )
}

func (node *Node) Shutdown(){
	fmt.Println("Shutdown Node:"+ node.Name)
}


