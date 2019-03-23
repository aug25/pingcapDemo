package nodeServer

import (
	"os/exec"
	"pingcapDemo/util"
	"runtime"
	"time"
)

// Deprecated: move to commond NodeServer
type TiDBNodeServer struct{}

type TiDBTestRequest struct {
	Echostr     string
	Cpubusytime int64
}

type TiDBTestResponse struct {
	Respstr string
}

func (p *TiDBNodeServer) Echostr(req TiDBTestRequest, res *TiDBTestResponse) error {
	res.Respstr = " TiDB Server get: [ " + req.Echostr + " ]\n"
	return nil
}

func (p *TiDBNodeServer) GetCurrentIP(req TiDBTestRequest, res *TiDBTestResponse) error {
	res.Respstr = util.GetLocalIp()
	return nil
}

func (p *TiDBNodeServer) KillServer(req TiDBTestRequest, res *TiDBTestResponse) error {
	_, err := exec.Command("sh", "-c", "pkill -SIGINT tidb-server").Output()
	if err != nil {
		return err
	} else {
		res.Respstr = "OK"
		return nil
	}
}

func (p *TiDBNodeServer) StressCPU(req TiDBTestRequest, res *TiDBTestResponse) error {
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
