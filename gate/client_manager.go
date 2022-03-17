package main

import (
	"sync/atomic"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
)

var clientManager ClientManager

func init() {
	clientManager = ClientManager{
		started: 0,
		jobCh: make(chan clientManagerJob, 20),
		clientMap: make(map[string]*lbtnet.TcpConnection),
	}
}

type clientManagerJob struct {
	op string
	jd interface{}
}

func postClientManagerJob(op string, jd interface{}) bool {
	if atomic.LoadInt32(&clientManager.started) == 0 { return false }
	select {
		case clientManager.jobCh <- clientManagerJob{op: op, jd: jd}:
			return true
		default:
			return false
	}
	return false
}

type ClientManager struct {
	started int32
	jobCh chan clientManagerJob
	clientMap map[string]*lbtnet.TcpConnection
}

func (cm *ClientManager) start() {
	if atomic.CompareAndSwapInt32(&cm.started, 0, 1) {
		logger.Info("client manager start ...")
		go cm.workLoop()
	}
}

func (cm *ClientManager) workLoop() {
	for job := range cm.jobCh {
		if job.op == "connect" {
			cm.clientConnect(job.jd.(*lbtnet.TcpConnection))
		} else if job.op == "disconnect" {
			cm.clientDisconnect(job.jd.(*lbtnet.TcpConnection))
		} else if job.op == "service_reply" {
			cm.serviceReply(job.jd.([]byte))
		} else if job.op == "create_entity" {
			cm.createEntity(job.jd.([]byte))
		} else if job.op == "entity_msg" {
			cm.entityMsg(job.jd.([]byte))
		} else {
			logger.Warn("ClientManager unrecogonized op %s", job.op)
		}
	}
}

func (cm *ClientManager) clientConnect(c *lbtnet.TcpConnection) {
	cm.clientMap[c.RemoteAddr()] = c
}

func (cm *ClientManager) clientDisconnect(c *lbtnet.TcpConnection) {
	delete(cm.clientMap, c.RemoteAddr())
}

func (cm *ClientManager) serviceReply(buf []byte) {
	msg := &lbtproto.ServiceReply{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("service reply fail 1")
		return
	}
	addr := msg.Addr
	c, ok := cm.clientMap[addr]
	if !ok {
		logger.Warn("service reply fail 2 %s", addr)
		return
	}
	if err := c.SendData(buf); err != nil {
		logger.Warn("service reply fail 3 [%s] [%s]", addr, err.Error())
		return
	}
}

func (cm *ClientManager) createEntity(buf []byte) {
	msg := &lbtproto.EntityData{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("createEntity fail 1 [%s]", err.Error())
		return
	}
	addr := msg.Addr
	c, ok := cm.clientMap[addr]
	if !ok {
		logger.Warn("createEntity fail 2 %s", addr)
		return
	}
	newmsg := &lbtproto.ChannelEntityInfo{
		Type: []byte(msg.Type),
		Info: msg.Data,
		EntityId: []byte(msg.Id),
		SessionId: []byte{},
	}
	newbuf, err := lbtproto.EncodeMessage(
			lbtproto.Client.Method_createChannelEntity,
			newmsg,
			lbtnet.ByteOrder,
		)
	if err != nil {
		logger.Warn("createEntity fail 3 [%s] [%s]", addr, err.Error())
		return
	}
	if err := c.SendData(newbuf); err != nil {
		logger.Warn("createEntity fail 4 [%s] [%s]", addr, err.Error())
		return
	}
}

func (cm *ClientManager) entityMsg(buf []byte) {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("entityMsg fail 1")
		return
	}
	addr := msg.Addr
	c, ok := cm.clientMap[addr]
	if !ok {
		logger.Warn("entityMsg fail 2 %s", addr)
		return
	}
	newmsg := &lbtproto.EntityMessage{
		EntityId: []byte(msg.Id),
		MethodName: []byte(msg.Method),
		Index: 0,
		Parameters: msg.Params,
		SessionId: []byte{},
		Context: []byte{},
	}
	newbuf, err := lbtproto.EncodeMessage(
			lbtproto.Client.Method_entityMessage,
			newmsg,
			lbtnet.ByteOrder,
		)
	if err != nil {
		logger.Warn("entityMsg fail 3 [%s] [%s]", addr, err.Error())
		return
	}
	if err := c.SendData(newbuf); err != nil {
		logger.Warn("entityMsg fail 4 [%s] [%s]", addr, err.Error())
		return
	}
}
