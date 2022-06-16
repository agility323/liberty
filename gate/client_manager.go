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
		clientMap: make(map[string]*clientEntry),
		boundClients: make(map[string]map[string]int),
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

type clientEntry struct {
	c *lbtnet.TcpConnection
	serviceAddr string
}

type ClientManager struct {
	started int32
	jobCh chan clientManagerJob
	clientMap map[string]*clientEntry
	boundClients map[string]map[string]int
}

// TODO: concurrent issue
func (cm *ClientManager) getServiceAddr(caddr string) string {
	if entry, ok := cm.clientMap[caddr]; ok {
		return entry.serviceAddr
	}
	return ""
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
		} else if job.op == "bind_client" {
			cm.bindClient(job.jd.(lbtproto.BindClientInfo))
		} else if job.op == "unbind_client" {
			cm.unbindClient(job.jd.(lbtproto.BindClientInfo))
		} else if job.op == "client_service_reply" {
			cm.clientServiceReply(job.jd.([]byte))
		} else if job.op == "create_entity" {
			cm.createEntity(job.jd.([]byte))
		} else if job.op == "client_entity_msg" {
			cm.clientEntityMsg(job.jd.([]byte))
		} else if job.op == "service_disconnect" {
			cm.serviceDisconnect(job.jd.(string))
		} else {
			logger.Warn("ClientManager unrecogonized op %s", job.op)
		}
	}
}

func (cm *ClientManager) clientConnect(c *lbtnet.TcpConnection) {
	cm.clientMap[c.RemoteAddr()] = &clientEntry{c: c, serviceAddr: ""}
}

func (cm *ClientManager) clientDisconnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	entry, ok := cm.clientMap[addr]
	if !ok { return }
	delete(cm.clientMap, addr)
	if entry.serviceAddr != "" {
		info := lbtproto.BindClientInfo{Caddr: addr, Saddr: entry.serviceAddr}
		postServiceManagerJob("client_disconnect", info)
	}
}

func (cm *ClientManager) bindClient(info lbtproto.BindClientInfo) {
	entry, ok := cm.clientMap[info.Caddr]
	if !ok { return }
	entry.serviceAddr = info.Saddr
	if _, ok := cm.boundClients[info.Saddr]; !ok {
		cm.boundClients[info.Saddr] = make(map[string]int)
	}
	cm.boundClients[info.Saddr][info.Caddr] = 1
}

func (cm *ClientManager) unbindClient(info lbtproto.BindClientInfo) {
	saddr := info.Saddr
	if entry, ok := cm.clientMap[info.Caddr]; ok {
		delete(cm.clientMap, info.Caddr)
		entry.c.CloseWithoutCallback()
		saddr = entry.serviceAddr
	}
	if m, ok := cm.boundClients[saddr]; ok {
		delete(m, info.Caddr)
		if len(m) == 0 {
			delete(cm.boundClients, saddr)
		}
	}
}

func (cm *ClientManager) clientServiceReply(buf []byte) {
	msg := &lbtproto.ServiceReply{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("clientServiceReply fail 1")
		return
	}
	addr := msg.Addr
	entry, ok := cm.clientMap[addr]
	if !ok {
		logger.Warn("clientServiceReply fail 2 %s", addr)
		return
	}
	sendClientServiceReply(entry.c, msg.Reqid, msg.Reply)
}

func (cm *ClientManager) createEntity(buf []byte) {
	msg := &lbtproto.EntityData{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("createEntity fail 1 [%s]", err.Error())
		return
	}
	addr := msg.Addr
	entry, ok := cm.clientMap[addr]
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
		)
	if err != nil {
		logger.Warn("createEntity fail 3 [%s] [%s]", addr, err.Error())
		return
	}
	if err := entry.c.SendData(newbuf); err != nil {
		logger.Warn("createEntity fail 4 [%s] [%s]", addr, err.Error())
		return
	}
}

func (cm *ClientManager) clientEntityMsg(buf []byte) {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("clientEntityMsg fail 1")
		return
	}
	addr := msg.Addr
	entry, ok := cm.clientMap[addr]
	if !ok {
		logger.Warn("clientEntityMsg fail 2 %s", addr)
		return
	}
	sendEntityMsg(entry.c, msg.Id, msg.Method, msg.Params)
}

func (cm *ClientManager) serviceDisconnect(saddr string) {
	m, ok := cm.boundClients[saddr]
	if !ok { return }
	delete(cm.boundClients, saddr)
	for caddr, _ := range m {
		if entry, ok := cm.clientMap[caddr]; ok {
			delete(cm.clientMap, caddr)
			entry.c.CloseWithoutCallback()
		}
	}
}
