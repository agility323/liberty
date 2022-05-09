package main

import (
	"sync/atomic"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtreg"
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

func OnDiscoverService(services map[string][]byte) {
	postServiceManagerJob("discover", services)
}

type serviceEntry struct {
	connected bool
	addr string
	typ string
	cli *lbtnet.TcpClient
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
		if job.op == "discover" {
			sm.serviceDiscover(job.jd.(map[string][]byte))
		} else if job.op == "connect" {
			sm.serviceConnect(job.jd.(*lbtnet.TcpConnection))
		} else if job.op == "connect_fail" {
			sm.serviceConnectFail(job.jd.(*lbtnet.TcpClient))
		} else if job.op == "disconnect" {
			sm.serviceDisconnect(job.jd.(*lbtnet.TcpConnection))
		} else if job.op == "register" {
			sm.serviceRegister(job.jd.(serviceEntry))
		} else if job.op == "client_disconnect" {
			sm.clientDisconnect(job.jd.(lbtproto.BindClientInfo))
		} else if job.op == "service_request" {
			sm.serviceRequest(job.jd.([]byte))
		} else if job.op == "entity_msg" {
			args := job.jd.([]interface{})
			sm.entityMsg(args[0].(string), args[1].([]byte))
		} else {
			logger.Warn("ServiceManager unrecogonized op %s", job.op)
		}
	}
}

func (sm *ServiceManager) serviceDiscover(services map[string][]byte) {
	for service, _ := range services {
		pair := lbtreg.SplitEtcdKey(service, 2)
		if len(pair) != 2 { continue }
		typ := pair[0]
		addr := pair[1]
		entry, ok := sm.serviceMap[addr]
		if !ok {
			sm.serviceMap[addr] = &serviceEntry{
				connected: false,
				addr: addr,
				typ: typ,
				cli: lbtnet.NewTcpClient(addr, &ServiceConnectionHandler{}),
			}
			sm.serviceMap[addr].cli.SetReconnectTime(10)
			sm.serviceMap[addr].cli.StartConnect(3)
			continue
		}
		if !entry.connected {
			logger.Warn("still connecting to service %s", addr)
			continue
		}
	}
}

func (sm *ServiceManager) serviceConnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	if entry, ok := sm.serviceMap[addr]; ok {
		entry.connected = true
		if _, ok = sm.serviceTypeToAddrSet[entry.typ]; !ok {
			sm.serviceTypeToAddrSet[entry.typ] = lbtutil.NewOrderedSet()
		}
		sm.serviceTypeToAddrSet[entry.typ].Add(addr)
	} else {
		logger.Warn("service connected with no client %s", addr)
		c.Close()
	}
}

func (sm *ServiceManager) serviceConnectFail(cli *lbtnet.TcpClient) {
	addr := cli.RemoteAddr()
	if entry, ok := sm.serviceMap[addr]; ok {
		if entry.connected{ return }
		delete(sm.serviceMap, addr)
	}
}

func (sm *ServiceManager) serviceDisconnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	if entry, ok := sm.serviceMap[addr]; ok {
		sm.serviceTypeToAddrSet[entry.typ].Remove(addr)
		delete(sm.serviceMap, addr)
		postClientManagerJob("service_disconnect", addr)
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
	/*
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
			if ent, ok := sm.serviceMap[ad]; ok && ent.connected {
				if err := lbtproto.SendMessage(
					ent.cli,
					lbtproto.Service.Method_service_shutdown,
					&lbtproto.Void{},
				); err != nil {
					logger.Warn("service shutdown send fail 1 - " + err.Error())
				}
			}
		}
	}
	*/
	logger.Info("register service %s %s", typ, addr)
	// reply
	msg := &lbtproto.ServiceInfo{
		Addr: addr,
		Type: typ,
		Entityid: "",
	}
	if err := lbtproto.SendMessage(
		entry.cli,
		lbtproto.Service.Method_register_reply,
		msg,
	); err != nil {
		logger.Warn("service register reply send fail - " + err.Error())
	}
}

func (sm *ServiceManager) clientDisconnect(info lbtproto.BindClientInfo) {
	if entry, ok := sm.serviceMap[info.Saddr]; ok && entry.connected {
		lbtproto.SendMessage(entry.cli, lbtproto.Service.Method_client_disconnect, &info)
	}
}

func (sm *ServiceManager) serviceRequest(buf []byte) {
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
	if entry, ok := sm.serviceMap[addr]; ok && entry.connected {
		if err := entry.cli.SendData(buf); err != nil {
			logger.Warn("service request fail 3 at %s %s", entry.addr, err.Error())
		} else {
			logger.Debug("service request sent to %s", entry.addr)
			return
		}
	} else {
		logger.Warn("service request fail 4")
	}
}

func (sm *ServiceManager) entityMsg(caddr string, buf []byte) {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("entity msg fail 1 %s", caddr)
		return
	}
	saddr := clientManager.getServiceAddr(caddr)
	if saddr == "" {
		logger.Warn("entity msg fail 2 %s", caddr)
		return
	}
	if entry, ok := sm.serviceMap[saddr]; ok && entry.connected {
		if err := entry.cli.SendData(buf); err != nil {
			logger.Warn("entity msg fail 3 %s-%s [%s]", caddr, saddr, err.Error())
		} else {
			logger.Debug("entity msg sent %s-%s", caddr, saddr)
			return
		}
	} else {
		logger.Warn("entity msg fail 4 %s-%s", caddr, saddr)
	}
}
