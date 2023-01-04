package service_framework

import (
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtactor"
)

type EntityCore struct {
	typ string
	id lbtutil.ObjectID
	actor *lbtactor.WorkerActor
}

func (ec *EntityCore) init(typ string) {
	ec.id = lbtutil.NewObjectID()
	ec.typ = typ
}

func (ec *EntityCore) GetType() string {
	return ec.typ
}

func (ec *EntityCore) GetId() lbtutil.ObjectID {
	return ec.id
}

func (ec *EntityCore) Dump() map[string]string {
	return map[string]string {
		"id": string(ec.id[:]),
		"typ": ec.typ,
	}
}

func (ec *EntityCore) InitActor() {
	ec.actor = lbtactor.NewWorkerActor(100)
}

func (ec *EntityCore) StartActor() bool {
	if ec.actor == nil { return false }
	ec.actor.Start()
	return true
}

func (ec *EntityCore) StopActor() {
	if ec.actor == nil { return }
	ec.actor.Stop()
}

func (ec *EntityCore) PushActorTask(task func()) bool {
	if ec.actor == nil { return false }
	return ec.actor.PushTask(task)
}

type RemoteEntityStub struct {
	core *EntityCore
	c *lbtnet.TcpConnection
	remoteAddr string
}

func NewRemoteEntityStub(core *EntityCore, c *lbtnet.TcpConnection, remoteAddr string) *RemoteEntityStub {
	return &RemoteEntityStub{core: core, c: c, remoteAddr: remoteAddr}
}

func (stub *RemoteEntityStub) GetLocalAddr() string {
	return stub.c.LocalAddr()
}

func (stub *RemoteEntityStub) GetRemoteAddr() string {
	return stub.remoteAddr
}

func (stub *RemoteEntityStub) Bind(cb ClientCallback) {
	SendBindClient(stub.c, stub.c.LocalAddr(), stub.remoteAddr)
	registerClientCallback(stub.remoteAddr, cb)
}

func (stub *RemoteEntityStub) Switch(c *lbtnet.TcpConnection, remoteAddr string) {
	SendUnbindClient(stub.c, stub.c.LocalAddr(), stub.remoteAddr)
	unregisterClientCallback(stub.remoteAddr)
	stub.c = c
	stub.remoteAddr = remoteAddr
}

func (stub *RemoteEntityStub) Yield(core *EntityCore) *RemoteEntityStub {
	return &RemoteEntityStub{core: core, c: stub.c, remoteAddr: stub.remoteAddr}
}

func (stub *RemoteEntityStub) CreateEntity(data interface{}) {
	SendCreateEntity(stub.c, stub.remoteAddr, stub.core.GetId(), stub.core.GetType(), data)
}

func (stub *RemoteEntityStub) CallClientMethod(method string, params interface{}) {
	SendClientEntityMsg(stub.c, stub.remoteAddr, stub.core.GetId(), method, params)
}
