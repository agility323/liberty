package service_framework

import (
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
)

type EntityCore struct {
	typ string
	id lbtutil.ObjectId
}

func (ec *EntityCore) init(typ string) {
	ec.id = lbtutil.NewObjectId()
	ec.typ = typ
}

func (ec *EntityCore) GetType() string {
	return ec.typ
}

func (ec *EntityCore) GetId() lbtutil.ObjectId {
	return ec.id
}

func (ec *EntityCore) Dump() map[string]string {
	return map[string]string {
		"id": string(ec.id),
		"typ": ec.typ,
	}
}

type RemoteEntityStub struct {
	core *EntityCore
	c *lbtnet.TcpConnection
	remoteAddr string
}

func NewRemoteEntityStub(core *EntityCore, c *lbtnet.TcpConnection, remoteAddr string) *RemoteEntityStub {
	return &RemoteEntityStub{core: core, c: c, remoteAddr: remoteAddr}
}

func (stub *RemoteEntityStub) GetC() *lbtnet.TcpConnection {
	return stub.c
}

func (stub *RemoteEntityStub) GetRemoteAddr() string {
	return stub.remoteAddr
}

func (stub *RemoteEntityStub) BindProxy() {
	SendBindClient(stub.c, stub.c.LocalAddr(), stub.remoteAddr)
}

func (stub *RemoteEntityStub) CreateEntity(data interface{}) {
	SendCreateEntity(stub.c, stub.remoteAddr, string(stub.core.GetId()), stub.core.GetType(), data)
}

func (stub *RemoteEntityStub) EntityMsg(method string, params interface{}) {
	SendEntityMsg(stub.c, stub.remoteAddr, string(stub.core.GetId()), method, params)
}
