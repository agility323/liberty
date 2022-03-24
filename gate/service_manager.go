package main

import (
	"sync/atomic"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
)

var serviceManager ServiceManager

func init() {
	serviceManager = ServiceManager{
		started: 0,
		jobCh: make(chan serviceManagerJob, 20),
		serviceConnMap: make(map[string]*lbtnet.TcpConnection),
		serviceList: make([]serviceInfo, 0),
		serviceIndexMap: make(map[string]int),
	}
}

type serviceInfo struct {
	addr,
	typ,
	entityid string
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
	serviceConnMap map[string]*lbtnet.TcpConnection
	serviceList []serviceInfo
	serviceIndexMap map[string]int	// addr: index in serviceList
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
			sm.serviceRegister(job.jd.(serviceInfo))
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
	sm.serviceConnMap[c.RemoteAddr()] = c
}

func (sm *ServiceManager) serviceDisconnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	delete(sm.serviceConnMap, addr)
	if index, ok := sm.serviceIndexMap[addr]; ok {
		// swap index and last, abandon last
		l := sm.serviceList
		lastIndex := len(l) - 1
		lastEntry := l[lastIndex]
		entry := l[index]
		l[index] = lastEntry
		sm.serviceIndexMap[lastEntry.addr] = index
		l = l[:lastIndex]
		delete(sm.serviceIndexMap, entry.addr)
	}
}

func (sm *ServiceManager) serviceRegister(info serviceInfo) {
	// register
	addr := info.addr
	typ := info.typ
	if index, ok := sm.serviceIndexMap[addr]; !ok {
		sm.serviceList = append(sm.serviceList, info)
		index = len(sm.serviceList) - 1
		sm.serviceIndexMap[addr] = index
		// close old service
		for i, s := range sm.serviceList {
			if s.typ == typ && i != index {
				if c, ok := sm.serviceConnMap[s.addr]; ok {
					if err := lbtproto.SendMessage(
						c,
						lbtproto.Service.Method_service_shutdown,
						&lbtproto.Void{},
					); err != nil {
						logger.Warn("service shutdown send fail 1 - " + err.Error())
					}
				}
			}
		}
	}
	logger.Info("register service %s", info)
	// reply
	if c, _ := sm.serviceConnMap[addr]; c == nil {
		logger.Warn("service register reply send fail 1")
	} else {
		msg := &lbtproto.ServiceInfo{
			Addr: addr,
			Type: typ,
			Entityid: info.entityid,
		}
		if err := lbtproto.SendMessage(
			c,
			lbtproto.Service.Method_register_reply,
			msg,
		); err != nil {
			logger.Warn("service register reply send fail 2 - " + err.Error())
		}
	}
}

func (sm *ServiceManager) serviceRequest(buf []byte) {
	// TODO: divide with type; route strategy; gate header + sendfile
	msg := &lbtproto.ServiceRequest{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("service request fail 1")
		return
	}
	typ := msg.Type
	for _, info := range sm.serviceList {
		if info.typ == typ {
			if c, _ := sm.serviceConnMap[info.addr]; c == nil {
				logger.Warn("service request fail 2 at [%s], continue ...", info.addr)
			} else {
				if err := c.SendData(buf); err != nil {
					logger.Warn("service request fail 3 at [%s] [%s], continue ...", info.addr, err.Error())
				} else {
					logger.Debug("service request sent to %s", info.addr)
					return
				}
			}
		} else {
			logger.Debug("service request skip service %s", info)
		}
	}
	logger.Warn("service request fail 4")
}

func (sm *ServiceManager) entityMsg(buf []byte) {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("entityMsg fail 1")
		return
	}
	addr := msg.Addr
	if c, _ := sm.serviceConnMap[addr]; c == nil {
		logger.Warn("entityMsg fail 2 at [%s]", addr)
	} else {
		if err := c.SendData(buf); err != nil {
			logger.Warn("entityMsg fail 3 at [%s] [%s]", addr, err.Error())
		} else {
			logger.Debug("entityMsg sent to %s", addr)
			return
		}
	}
}
