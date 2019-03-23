package nodeServer

import (
	"os/exec"
	"runtime"
	"time"
)
import "pingcapDemo/util"

type NodeServer struct{}

type TestRequest struct {
	Serverbin   string
	Echostr     string
	Cpubusytime int64
}

type TestResponse struct {
	Respstr string
}

func (p *NodeServer) Echostr(req TestRequest, res *TestResponse) error {
	res.Respstr = " RemoteServer get: [ " + req.Echostr + " ]\n"
	return nil
}

func (p *NodeServer) GetCurrentIP(req TestRequest, res *TestResponse) error {
	res.Respstr = util.GetLocalIp()
	return nil
}

func (p *NodeServer) KillServer(req TestRequest, res *TestResponse) error {
	_, err := exec.Command("sh", "-c", "pkill -SIGINT "+req.Serverbin).Output()
	if err != nil {
		return err
	} else {
		res.Respstr = "OK"
		return nil
	}
}

func (p *NodeServer) StressCPU(req TestRequest, res *TestResponse) error {
	done := make(chan int)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}
			}
		}()
	}
	time.Sleep(time.Second * time.Duration(req.Cpubusytime))
	close(done)
	res.Respstr = "Stress CPU finished"
	return nil
}
