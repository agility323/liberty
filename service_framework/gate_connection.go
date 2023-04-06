package service_framework

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/agility323/liberty/lbtnet"
)

var GateConnectionHeartbeatTorlerance int64 = 36000	// time in milliseconds

type GateConnectionHandler struct {
	hbtime int64
}

func GateConnectionCreator(conn net.Conn) {
	handler := &GateConnectionHandler{
		hbtime: time.Now().UnixMilli(),
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
	atomic.StoreInt64(&handler.hbtime, time.Now().UnixMilli())
}

func (handler *GateConnectionHandler) CheckHeartbeat() bool {
	return time.Now().UnixMilli() - atomic.LoadInt64(&handler.hbtime) < GateConnectionHeartbeatTorlerance
}
