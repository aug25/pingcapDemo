package nodeServer

import(
	"os/exec"
	"pingcapDemo/util"
)
type TiDBNodeServer struct{}

type TiDBTestRequest struct{
	Echostr string
}

type TiDBTestResponse struct{
	Respstr string
}
func (p *TiDBNodeServer) Echostr(req TiDBTestRequest, res *TiDBTestResponse) error{
	res.Respstr = " TiDB Server get: [ "+req.Echostr + " ]\n"
	return nil
}

func (p *TiDBNodeServer) GetCurrentIP(req TiDBTestRequest, res *TiDBTestResponse) error{
	res.Respstr=util.GetLocalIp()
	return nil
}

func (p *TiDBNodeServer) KillTiDBServer(req TiDBTestRequest, res *TiDBTestResponse) error{
	_, err := exec.Command("sh", "-c", "pkill -SIGINT tidb-server").Output()
	if(err !=nil){
		 return err
	}else{
		 res.Respstr="OK"
		 return nil
	}
}

