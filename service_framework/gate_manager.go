package service_framework

import (
	"sync/atomic"
	"math/rand"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
)

var gateManager GateManager

func init() {
	gateManager = GateManager{
		started: 0,
		jobCh: make(chan gateManagerJob, 20),
		gateMap: make(map[string]*lbtnet.TcpConnection),
	}
}

type gateManagerJob struct {
	op string
	jd interface{}
}

func postGateManagerJob(op string, jd interface{}) bool {
	if atomic.LoadInt32(&gateManager.started) == 0 { return false }
	select {
		case gateManager.jobCh <- gateManagerJob{op: op, jd: jd}:
			return true
		default:
			return false
	}
	return false
}

type GateManager struct {
	started int32
	jobCh chan gateManagerJob
	gateMap map[string]*lbtnet.TcpConnection
}

func (gm *GateManager) start() {
	if atomic.CompareAndSwapInt32(&gm.started, 0, 1) {
		logger.Info("gate manager start ...")
		go gm.workLoop()
	}
}

func (gm *GateManager) stop() {
	//TODO.Stop()
}

func (gm *GateManager) workLoop() {
	for job := range gm.jobCh {
		if job.op == "connect" {
			gm.gateConnect(job.jd.(*lbtnet.TcpConnection))
		} else if job.op == "disconnect" {
			gm.gateDisconnect(job.jd.(*lbtnet.TcpConnection))
		} else if job.op == "entity_msg" {
			gm.entityMsg(job.jd.(*lbtproto.EntityMsg))
		} else if job.op == "service_request" {
			gm.serviceRequest(job.jd.(*lbtproto.ServiceRequest))
		} else {
			logger.Warn("GateManager unrecogonized op %s", job.op)
		}
	}
}

func (gm *GateManager) gateConnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	gm.gateMap[addr] = c
}

func (gm *GateManager) gateDisconnect(c *lbtnet.TcpConnection) {
	//TODO.OnConnectionClose()
	addr := c.RemoteAddr()
	delete(gm.gateMap, addr)
}

func (gm *GateManager) entityMsg(msg *lbtproto.EntityMsg) {
	n := rand.Intn(len(gm.gateMap))
	for _, c := range gm.gateMap {
		if n--; n >= 0 { continue }
		if err := lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_entity_msg, msg); err != nil {
			logger.Error("entityMsg failed 1 %s", err.Error())
		}
		return
	}
	logger.Error("entityMsg failed 2 no gate connection")
}

func (gm *GateManager) serviceRequest(msg *lbtproto.ServiceRequest) {
	n := rand.Intn(len(gm.gateMap))
	for _, c := range gm.gateMap {
		if n--; n >= 0 { continue }
		if err := lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_service_request, msg); err != nil {
			logger.Error("serviceRequest failed 1 %s", err.Error())
		}
		return
	}
	logger.Error("serviceRequest failed 2 no gate connection")
}
