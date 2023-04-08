package main

import (
	"net"
	"time"

	"github.com/agility323/liberty/lbtnet"
)

var (
	WriteChLen int = 200
	WriteChWaitTime time.Duration = 0
)

type ClientConnectionHandler struct {
}

func ClientConnectionCreator(conn net.Conn) {
	handler := &ClientConnectionHandler{}
	conf := lbtnet.ConnectionConfig{
		WriteChLen: WriteChLen,
		WriteChWaitTime: WriteChWaitTime,
		ErrLog: false,
	}
	c := lbtnet.NewTcpConnection(conn, handler, conf)
	c.Start()
	handler.OnConnectionReady(c)
	clientManager.clientConnect(c)
}

func (handler *ClientConnectionHandler) HandleProto(c *lbtnet.TcpConnection, data []byte) error {
	return processClientProto(c, data)
}

func (handler *ClientConnectionHandler) OnConnectionReady(c *lbtnet.TcpConnection) {
}

func (handler *ClientConnectionHandler) OnConnectionClose(c *lbtnet.TcpConnection) {
	clientManager.clientDisconnect(c)
}

func (handler *ClientConnectionHandler) OnConnectionFail(cli *lbtnet.TcpClient) {
}

func (handler *ClientConnectionHandler) OnHeartbeat(c *lbtnet.TcpConnection, t int64) {
	// do nothing
}

func (handler *ClientConnectionHandler) CheckHeartbeat() bool {
	// do nothing
	return true
}
