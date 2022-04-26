package service_framework

import (
	"net"
	"github.com/agility323/liberty/lbtnet"
)

type GateConnectionHandler struct {
}

func GateConnectionCreator(conn net.Conn) {
	handler := &GateConnectionHandler{
	}
	c := lbtnet.NewTcpConnection(conn, handler)
	handler.OnConnectionReady(c)
	c.Start()
}

func (handler *GateConnectionHandler) HandleProto(c *lbtnet.TcpConnection, buf []byte) error {
	return processGateProto(c, buf)
}

func (handler *GateConnectionHandler) OnConnectionReady(c *lbtnet.TcpConnection) {
	postGateManagerJob("connect", c)
	sendRegisterService(c)
}

func (handler *GateConnectionHandler) OnConnectionClose(c *lbtnet.TcpConnection) {
	postGateManagerJob("disconnect", c)
}

func (handler *GateConnectionHandler) OnConnectionFail(cli *lbtnet.TcpClient) {
}
