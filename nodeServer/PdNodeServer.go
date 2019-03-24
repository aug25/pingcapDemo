package nodeServer

import (
	"os/exec"
	"pingcapDemo/util"
)

// Deprecated: move to commond NodeServer
// Test that specific to PD node can add here
type PdNodeServer struct{}

type PdTestRequest struct {
	Echostr string
}

type PdTestResponse struct {
	Respstr string
}

func (p *PdNodeServer) Echostr(req PdTestRequest, res *PdTestResponse) error {
	res.Respstr = " PdServer get: [ " + req.Echostr + " ]\n"
	return nil
}

func (p *PdNodeServer) GetCurrentIP(req PdTestRequest, res *PdTestResponse) error {
	res.Respstr = util.GetLocalIp()
	return nil
}

func (p *PdNodeServer) KillServer(req TiDBTestRequest, res *TiDBTestResponse) error {
	_, err := exec.Command("sh", "-c", "pkill -SIGINT pd-server").Output()
	if err != nil {
		return err
	} else {
		res.Respstr = "OK"
		return nil
	}
}
