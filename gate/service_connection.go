package main

import (
	"net"

	"github.com/agility323/liberty/lbtnet"
)

type ServiceConnectionHandler struct {
}

func ServiceConnectionCreator(conn net.Conn) {
	handler := &ServiceConnectionHandler{
	}
	c := lbtnet.NewTcpConnection(conn, handler)
	c.Start()
	handler.OnConnectionReady(c)
}

func (handler *ServiceConnectionHandler) HandleProto(c *lbtnet.TcpConnection, data []byte) error {
	return processServiceProto(c, data)
}

func (handler *ServiceConnectionHandler) OnConnectionReady(c *lbtnet.TcpConnection) {
	postServiceManagerJob("connect", c)
}

func (handler *ServiceConnectionHandler) OnConnectionClose(c *lbtnet.TcpConnection) {
	postServiceManagerJob("disconnect", c)
}
