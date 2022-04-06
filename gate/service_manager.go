package main

import (
	"sync/atomic"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtutil"
)

var serviceManager ServiceManager

func init() {
	serviceManager = ServiceManager{
		started: 0,
		jobCh: make(chan serviceManagerJob, 20),
		serviceMap: make(map[string]*serviceEntry),
		serviceTypeToAddrSet: make(map[string]*lbtutil.OrderedSet),
	}
}

type serviceEntry struct {
	addr string
	typ string
	c *lbtnet.TcpConnection
}

type serviceManagerJob struct {
	op string
	jd interface{}
}

func postServiceManagerJob(op string, jd interface{}) bool {
	if atomic.LoadInt32(&serviceManager.started) == 0 { return false }
	select {
		case serviceManager.jobCh <- serviceManagerJob{op: op, jd: jd}:
			return true
		default:
			return false
	}
	return false
}

type ServiceManager struct {
	started int32
	jobCh chan serviceManagerJob
	serviceMap map[string]*serviceEntry
	serviceTypeToAddrSet map[string]*lbtutil.OrderedSet
}

func (sm *ServiceManager) start() {
	if atomic.CompareAndSwapInt32(&sm.started, 0, 1) {
		logger.Info("service manager start ...")
		go sm.workLoop()
	}
}

func (sm *ServiceManager) workLoop() {
	for job := range sm.jobCh {
		if job.op == "connect" {
			sm.serviceConnect(job.jd.(*lbtnet.TcpConnection))
		} else if job.op == "disconnect" {
			sm.serviceDisconnect(job.jd.(*lbtnet.TcpConnection))
		} else if job.op == "register" {
			sm.serviceRegister(job.jd.(serviceEntry))
		} else if job.op == "client_disconnect" {
			sm.clientDisconnect(job.jd.(lbtproto.BindClientInfo))
		} else if job.op == "service_request" {
			sm.serviceRequest(job.jd.([]byte))
		} else if job.op == "entity_msg" {
			sm.entityMsg(job.jd.([]byte))
		} else {
			logger.Warn("ServiceManager unrecogonized op %s", job.op)
		}
	}
}

func (sm *ServiceManager) serviceConnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	sm.serviceMap[addr] = &serviceEntry{addr: addr, typ: "", c: c}
}

func (sm *ServiceManager) serviceDisconnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	if entry, ok := sm.serviceMap[addr]; ok {
		sm.serviceTypeToAddrSet[entry.typ].Remove(addr)
		delete(sm.serviceMap, addr)
	}
}

func (sm *ServiceManager) serviceRegister(info serviceEntry) {
	addr := info.addr
	typ := info.typ
	// register
	entry, ok := sm.serviceMap[addr]
	if !ok {
		logger.Warn("serviceRegister fail 1 - service not connected %v %v", entry, info)
		return
	}
	if entry.typ != "" {
		logger.Warn("serviceRegister fail 2 - service existed %v %v", entry, info)
		return
	}
	entry.typ = typ
	if _, ok = sm.serviceTypeToAddrSet[typ]; !ok {
		sm.serviceTypeToAddrSet[typ] = lbtutil.NewOrderedSet()
	}
	sm.serviceTypeToAddrSet[typ].Add(addr)
	// close old service
	if sm.serviceTypeToAddrSet[typ].Size() > 1 {
		vs := sm.serviceTypeToAddrSet[typ].GetAll()
		for _, v := range vs {
			ad := v.(string)
			if ad == addr { continue }
			if ent, ok := sm.serviceMap[ad]; ok {
				if err := lbtproto.SendMessage(
					ent.c,
					lbtproto.Service.Method_service_shutdown,
					&lbtproto.Void{},
				); err != nil {
					logger.Warn("service shutdown send fail 1 - " + err.Error())
				}
			}
		}
	}
	logger.Info("register service %s %s", typ, addr)
	// reply
	msg := &lbtproto.ServiceInfo{
		Addr: addr,
		Type: typ,
		Entityid: "",
	}
	if err := lbtproto.SendMessage(
		entry.c,
		lbtproto.Service.Method_register_reply,
		msg,
	); err != nil {
		logger.Warn("service register reply send fail - " + err.Error())
	}
}

func (sm *ServiceManager) clientDisconnect(info lbtproto.BindClientInfo) {
	entry, ok := sm.serviceMap[info.Saddr]
	if !ok { return }
	lbtproto.SendMessage(
		entry.c,
		lbtproto.Service.Method_client_disconnect,
		&info,
	)
}

func (sm *ServiceManager) serviceRequest(buf []byte) {
	// TODO: divide with type; route strategy; gate header + sendfile
	msg := &lbtproto.ServiceRequest{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("service request fail 1")
		return
	}
	addrSet, ok := sm.serviceTypeToAddrSet[msg.Type]
	if !ok {
		logger.Warn("service request fail 1 - %v", msg)
		return
	}
	v := addrSet.RandomGetOne()
	if v == nil {
		logger.Warn("service request fail 2 - service list empty %v", msg)
		return
	}
	addr := v.(string)
	if entry, ok := sm.serviceMap[addr]; ok {
		if err := entry.c.SendData(buf); err != nil {
			logger.Warn("service request fail 3 at %s %s", entry.addr, err.Error())
		} else {
			logger.Debug("service request sent to %s", entry.addr)
			return
		}
	} else {
		logger.Warn("service request fail 4")
	}
}

func (sm *ServiceManager) entityMsg(buf []byte) {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("entityMsg fail 1")
		return
	}
	addr := msg.Addr
	if entry, ok := sm.serviceMap[addr]; ok {
		if err := entry.c.SendData(buf); err != nil {
			logger.Warn("entityMsg fail 2 at [%s] [%s]", addr, err.Error())
		} else {
			logger.Debug("entity msg sent to %s", addr)
			return
		}
	} else {
		logger.Warn("entityMsg fail 3 at [%s]", addr)
	}
}
