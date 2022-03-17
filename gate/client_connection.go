package main

import (
	"net"
	"math/rand"

	"github.com/agility323/liberty/lbtnet"
)

type ClientConnectionHandler struct {
	seed int64
}

func ClientConnectionCreator(conn net.Conn) {
	handler := &ClientConnectionHandler{
		seed: 0,
	}
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

func (handler *ClientConnectionHandler) generateSeed() int64 {
	seed := rand.Int63()
	handler.seed = seed
	return seed
}
