package main

import (
	"github.com/agility323/liberty/lbtnet"
)

type ServiceConnectionHandler struct {
}

func (handler *ServiceConnectionHandler) HandleProto(c *lbtnet.TcpConnection, data []byte) error {
	return processServiceProto(c, data)
}

func (handler *ServiceConnectionHandler) OnConnectionReady(c *lbtnet.TcpConnection) {
	serviceManager.serviceConnect(c)
}

func (handler *ServiceConnectionHandler) OnConnectionClose(c *lbtnet.TcpConnection) {
	serviceManager.serviceDisconnect(c)
}

func (handler *ServiceConnectionHandler) OnConnectionFail(cli *lbtnet.TcpClient) {
	serviceManager.serviceConnectFail(cli)
}
