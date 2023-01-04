package service_framework

import (
	"github.com/agility323/liberty/lbtactor"
	"github.com/agility323/liberty/lbtutil"
)

type EntityCore struct {
	typ   string
	id    lbtutil.ObjectID
	actor *lbtactor.WorkerActor
}

func (ec *EntityCore) init(typ string, id lbtutil.ObjectID) {
	if id.IsZero() {
		id = lbtutil.NewObjectID()
	}
	ec.id = id
	ec.typ = typ
	ec.actor = lbtactor.NewWorkerActor()
}

func (ec *EntityCore) GetType() string {
	return ec.typ
}

func (ec *EntityCore) GetId() lbtutil.ObjectID {
	return ec.id
}

func (ec *EntityCore) Dump() map[string]string {
	return map[string]string{
		"id":  string(ec.id[:]),
		"typ": ec.typ,
	}
}

func (ec *EntityCore) StartActor(qlen int) bool {
	return ec.actor.Start(qlen)
}

func (ec *EntityCore) StopActor() bool {
	return ec.actor.Stop()
}

func (ec *EntityCore) PushActorTask(task func()) bool {
	return ec.actor.PushTask(task)
}

type RemoteEntityStub struct {
	core *EntityCore
	addr string	// communication address, remote addr of gate connection, key of gateManager.gateMap, c.RemoteAddr()
	localAddr string	// c.LocalAddr()
	remoteAddr string	// remote entity address
}

func NewRemoteEntityStub(core *EntityCore, addr, remoteAddr string) *RemoteEntityStub {
	c := gateManager.getGateByAddr(addr)
	if c == nil { return nil }
	localAddr := c.LocalAddr()
	return &RemoteEntityStub{
		core: core,
		addr: addr,
		localAddr: localAddr,
		remoteAddr: remoteAddr,
	}
}

func (stub *RemoteEntityStub) GetLocalAddr() string {
	return stub.localAddr
}

func (stub *RemoteEntityStub) GetRemoteAddr() string {
	return stub.remoteAddr
}

func (stub *RemoteEntityStub) Bind(cb ClientCallback) bool {
	c := gateManager.getGateByAddr(stub.addr)
	if c == nil { return false }
	err := SendBindClient(c, stub.localAddr, stub.remoteAddr)
	if err != nil {
		return false
	}
	registerClientCallback(stub.remoteAddr, cb)
	return true
}

func (stub *RemoteEntityStub) Switch(addr, remoteAddr string) bool {
	c := gateManager.getGateByAddr(stub.addr)
	if c == nil { return false }
	err := SendUnbindClient(c, stub.localAddr, stub.remoteAddr)
	if err != nil {
		return false
	}
	unregisterClientCallback(stub.remoteAddr)
	stub.addr = addr
	stub.localAddr = c.LocalAddr()
	stub.remoteAddr = remoteAddr
	return true
}

func (stub *RemoteEntityStub) Yield(core *EntityCore) *RemoteEntityStub {
	return &RemoteEntityStub{
		core: core,
		addr: stub.addr,
		localAddr: stub.localAddr,
		remoteAddr: stub.remoteAddr,
	}
}

func (stub *RemoteEntityStub) CreateEntity(data interface{}) bool {
	c := gateManager.getGateByAddr(stub.addr)
	if c == nil { return false }
	err := SendCreateEntity(c, stub.remoteAddr, stub.core.GetId(), stub.core.GetType(), data)
	if err != nil {
		return false
	}
	return true
}

func (stub *RemoteEntityStub) CallClientMethod(method string, params interface{}) bool {
	c := gateManager.getGateByAddr(stub.addr)
	if c == nil { return false }
	err := SendClientEntityMsg(c, stub.remoteAddr, stub.core.GetId(), method, params)
	if err != nil {
		return false
	}
	return true
}
