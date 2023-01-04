package service_framework

import (
	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
)

type methodActor struct {
	c *lbtnet.TcpConnection
	req *lbtproto.ServiceRequest
	fromService bool
}

func newMethodActor(c *lbtnet.TcpConnection, req *lbtproto.ServiceRequest, fromService bool) *methodActor {
	return &methodActor{c: c, req: req, fromService: fromService}
}

func (ma *methodActor) start() {
	replyData, err := processServiceMethod(ma.c, ma.req.Addr, ma.req.Reqid, ma.req.Method, ma.req.Params)
	if err != nil {
		logger.Warn("method actor fail [%v] [%v]", err, ma.req)
		return
	}
	if replyData == nil { return }
	if ma.fromService {
		sendServiceReply(ma.c, ma.req.Addr, ma.req.Reqid, replyData)
	} else {
		sendClientServiceReply(ma.c, ma.req.Addr, ma.req.Reqid, replyData)
	}
}


