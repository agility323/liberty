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
	handler.OnConnectionReady(c)	// must be called before c.Start, or service_register may fail and later ServiceManagerJob (like serviceDisconnect) may result in panic
	c.Start()
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
