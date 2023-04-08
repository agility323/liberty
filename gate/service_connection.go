package main

import (
	"sync/atomic"
	"time"

	"github.com/agility323/liberty/lbtnet"
)

var (
	ServiceConnectionTickTime time.Duration = 10
	ServiceConnectionHeartbeatCap int32 = 3
	ServiceConnectionRttWarnTime int64 = 100
)

type ServiceConnectionHandler struct {
	addr string
	hbpending int32
	hbstop chan struct{}
}

func (handler *ServiceConnectionHandler) HandleProto(c *lbtnet.TcpConnection, data []byte) error {
	return processServiceProto(c, data)
}

func (handler *ServiceConnectionHandler) OnConnectionReady(c *lbtnet.TcpConnection) {
	serviceManager.serviceConnect(c)
	handler.addr = c.RemoteAddr()
	handler.hbstop = make(chan struct{}, 1)
	go handler.StartHeartbeat(c)
}

func (handler *ServiceConnectionHandler) OnConnectionClose(c *lbtnet.TcpConnection) {
	serviceManager.serviceDisconnect(c)
	select {
	case handler.hbstop<- struct{}{}:
	default:
	}
}

func (handler *ServiceConnectionHandler) OnConnectionFail(cli *lbtnet.TcpClient) {
	serviceManager.serviceConnectFail(cli)
}

func (handler *ServiceConnectionHandler) StartHeartbeat(c *lbtnet.TcpConnection) {
	ticker := time.NewTicker(ServiceConnectionTickTime * time.Second)
	defer ticker.Stop()
	atomic.StoreInt32(&handler.hbpending, 0)
	for {
		select {
		case <-handler.hbstop:
			logger.Info("service heartbeat quit %s", handler.addr)
			return
		case <-ticker.C:
			// check fail
			hbpending := atomic.AddInt32(&handler.hbpending, 1)
			if hbpending > ServiceConnectionHeartbeatCap {
				// heartbeat fail
				logger.Error("service heartbeat fail %s", handler.addr)
				c.Close()
				return
			}
			if hbpending > 1 {
				logger.Warn("service heartbeat loss %s %d", handler.addr, hbpending - 1)
			}
			// send hb
			t := time.Now().UnixMilli()
			if !SendHeartbeat(c, t) {
				atomic.AddInt32(&handler.hbpending, -1)
				logger.Error("service heartbeat send fail %s", handler.addr)
			}
		}
	}
}

func (handler *ServiceConnectionHandler) OnHeartbeat(c *lbtnet.TcpConnection, t int64) {
	hbpending := atomic.AddInt32(&handler.hbpending, -1)
	if hbpending != 0 {
		logger.Warn("service heartbeat pending %s %d", handler.addr, hbpending)
	}
	t = time.Now().UnixMilli() - t
	if t >= ServiceConnectionRttWarnTime {
		logger.Warn("service heartbeat rtt %s %d", handler.addr, t)
	}
}

func (handler *ServiceConnectionHandler) CheckHeartbeat() bool {
	// do nothing
	return true
}
