package method

import (
	"github.com/agility323/liberty/lbtutil"
	sf "github.com/agility323/liberty/service_framework"
	"github.com/agility323/liberty/lbtnet"

	"github.com/agility323/liberty/demo_service/login/core"
)

type ConnectServerRequest struct {
}

type ConnectServerReply struct {
}

type ConnectServerHandler struct {
	request ConnectServerRequest
	reply ConnectServerReply
}

func (h *ConnectServerHandler) GetRequest() interface{} {return &(h.request)}
func (h *ConnectServerHandler) GetReply() interface{} {return nil}

func (h *ConnectServerHandler) Process(c *lbtnet.TcpConnection, srcAddr string) error {
	logger.Debug("connect server %s", srcAddr)
	// create avatar
	boost := sf.CreateEntity("BoostEntity", lbtutil.NilObjectID).(*core.BoostEntity)
	boost.Init(c, srcAddr)
	boost.Start()
	return nil
}
