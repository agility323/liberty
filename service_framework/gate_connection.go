package service_framework

import (
	"github.com/agility323/liberty/lbtnet"
)

type GateConnectionHandler struct {
}

func NewGateConnectionHandler() lbtnet.ConnectionHandler {
	return &GateConnectionHandler{}
}

func (handler *GateConnectionHandler) HandleProto(c *lbtnet.TcpConnection, buf []byte) error {
	return processGateProto(c, buf)
}

func (handler *GateConnectionHandler) OnConnectionReady(c *lbtnet.TcpConnection) {
	sendServiceRegister(c)
}

func (handler *GateConnectionHandler) OnConnectionClose(c *lbtnet.TcpConnection) {
	gateClient.OnConnectionClose()
}
