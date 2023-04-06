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
	conf := lbtnet.ConnectionConfig{
		WriteChLen: lbtnet.DefaultWriteChLen,
		WriteChWaitTime: lbtnet.DefaultWriteChWaitTime,
		ErrLog: true,
	}
	c := lbtnet.NewTcpConnection(conn, handler, conf)
	handler.OnConnectionReady(c)
	c.Start()
}

func (handler *GateConnectionHandler) HandleProto(c *lbtnet.TcpConnection, buf []byte) error {
	return processGateProto(c, buf)
}

func (handler *GateConnectionHandler) OnConnectionReady(c *lbtnet.TcpConnection) {
	gateManager.gateConnect(c)
	sendRegisterService(c)
}

func (handler *GateConnectionHandler) OnConnectionClose(c *lbtnet.TcpConnection) {
	gateManager.gateDisconnect(c)
}

func (handler *GateConnectionHandler) OnConnectionFail(cli *lbtnet.TcpClient) {
}

func (handler *GateConnectionHandler) OnHeartbeat(c *lbtnet.TcpConnection, t int64) {
	// handled at proto layer, nothing here
}
