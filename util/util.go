package util

import (
	"errors"
	"fmt"
	"os/exec"
	"pingcapDemo/node"
	"strings"
)

//make shell cmd call
func exeSysCommand(cmdStr string) string {
	cmd := exec.Command("sh", "-c", cmdStr)
	opBytes, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(opBytes)
}

func GetLocalIp() string {
	tmp := exeSysCommand("ifconfig eth0 | grep 'inet addr' | cut -d : -f 2 | cut -d ' ' -f 1")
	if len(tmp) == 0 {
		fmt.Println("GetLocalIp Failed")
		return ""
	}
	fmt.Println("this is from which node")
	localip := strings.Trim(tmp, "\n")
	return localip
}

func ParseCMD(allnodes []*node.Node, text string) (*node.Node, error) {
	var textarr = strings.Fields(text)
	if len(textarr) < 2 {
		return &node.Node{}, errors.New("cmd not supported, please enter '<node> <operation> <args>' or 'exit'")
	} else {
		//test on node
		testhost := textarr[0]
		testmethod := textarr[1]
		var testargs []string
		if len(textarr) > 2 {
			testargs = textarr[2:]
		}
		for _, n := range allnodes {
			if n.Name == testhost {
				n.TestMethod = testmethod
				n.TestArgs = testargs
				return n, nil
			}
		}
		return &node.Node{}, errors.New("Invalid node name!")
	}

}
