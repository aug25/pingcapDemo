package nodeServer

import (
	"os/exec"
	"pingcapDemo/util"
)

// Deprecated: move to commond NodeServer
// Test that specific to TiKV node can add here
type TiKVNodeServer struct{}

type TiKVTestRequest struct {
	Echostr string
}

type TiKVTestResponse struct {
	Respstr string
}

func (p *TiKVNodeServer) Echostr(req TiKVTestRequest, res *TiKVTestResponse) error {
	res.Respstr = " TiKV Server get: [ " + req.Echostr + " ]\n"
	return nil
}

func (p *TiKVNodeServer) GetCurrentIP(req TiKVTestRequest, res *TiKVTestResponse) error {
	res.Respstr = util.GetLocalIp()
	return nil
}

func (p *TiKVNodeServer) KillServer(req TiDBTestRequest, res *TiDBTestResponse) error {
	_, err := exec.Command("sh", "-c", "pkill -SIGINT tikv-server").Output()
	if err != nil {
		return err
	} else {
		res.Respstr = "OK"
		return nil
	}
}
