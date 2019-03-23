package node

import (
	"fmt"
)

type NodeType string

const (
	NodeTiDB    NodeType = "TiDB"
	NodePD      NodeType = "PD"
	NodeTiKV    NodeType = "TiKV"
	NodeProxy   NodeType = "Haproxy"
	NodeControl NodeType = "Control"
)

type Node struct {
	Name        string
	Type        NodeType
	ContainerID string
	TestMethod  string
	TestArgs    []string
}

func (node *Node) Startup() {
	fmt.Println("Startup Node:" + node.Name + " Type:" + string(node.Type))
}

func (node *Node) Shutdown() {
	fmt.Println("Shutdown Node:" + node.Name)
}
