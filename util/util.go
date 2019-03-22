package util

import (
	"fmt"
	"os/exec"
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

