package main

import (
	"net"

	"github.com/agility323/liberty/lbtnet"
)

type ClientConnectionHandler struct {
}

func ClientConnectionCreator(conn net.Conn) {
	handler := &ClientConnectionHandler{}
	c := lbtnet.NewTcpConnection(conn, handler)
	c.Start()
	handler.OnConnectionReady(c)
	postClientManagerJob("connect", c)
}

func (handler *ClientConnectionHandler) HandleProto(c *lbtnet.TcpConnection, data []byte) error {
	return processClientProto(c, data)
}

func (handler *ClientConnectionHandler) OnConnectionReady(c *lbtnet.TcpConnection) {
}

func (handler *ClientConnectionHandler) OnConnectionClose(c *lbtnet.TcpConnection) {
	postClientManagerJob("disconnect", c)
}

func (handler *ClientConnectionHandler) OnConnectionFail(cli *lbtnet.TcpClient) {
}
