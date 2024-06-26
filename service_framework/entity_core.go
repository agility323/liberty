package service_framework

import (
	//"sync/atomic"

	"github.com/agility323/liberty/lbtactor"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtutil"
)

type EntityCore struct {
	typ string
	id lbtutil.ObjectID
	mw *lbtactor.Worker	// main worker
	hw *lbtactor.HashedWorker	// hashed worker
}

func (ec *EntityCore) init(typ string, id lbtutil.ObjectID) {
	if id.IsZero() {
		id = lbtutil.NewObjectID()
	}
	ec.id = id
	ec.typ = typ
	ec.mw = nil
	ec.hw = nil
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

func (ec *EntityCore) InitMainWorker(qlen int) {
	ec.mw = lbtactor.NewWorker("entity." + ec.id.Hex(), qlen)
}

func (ec *EntityCore) initHashedWorker(qlen, size int) {
	ec.hw = lbtactor.NewHashedWorker("entity." + ec.id.Hex(), qlen, size)
}

func (ec *EntityCore) StartWorker() {
	ec.mw.Start()
	if ec.hw != nil {
		ec.hw.Start()
	}
}

func (ec *EntityCore) StopWorker() {
	ec.mw.Stop()
	if ec.hw != nil {
		ec.hw.Stop()
	}
}

func (ec *EntityCore) PushMainTask(task func()) bool {
	return ec.mw.PushTask(task)
}

//if not init ec.hw, will push to ec.mw
func (ec *EntityCore) PushHashedTask(task func(), hval int) bool {
	if ec.hw == nil {
		return ec.mw.PushTask(task)
	}
	return ec.hw.PushTask(task, hval)
}

type RemoteEntityStub struct {
	core *EntityCore
	addr string	// communication address, remote addr of gate connection, key of gateManager.gateMap, c.RemoteAddr()
	localAddr string	// c.LocalAddr()
	remoteAddr string	// remote entity address
	disconnected int32
	disconnectCallback func()
}

func NewRemoteEntityStub(core *EntityCore, addr, remoteAddr string, cb func()) *RemoteEntityStub {
	c := gateManager.getGateByAddr(addr)
	if c == nil { return nil }
	localAddr := c.LocalAddr()
	return &RemoteEntityStub{
		core: core,
		addr: addr,
		localAddr: localAddr,
		remoteAddr: remoteAddr,
		disconnected: 0,
		disconnectCallback: cb,
	}
}

func (stub *RemoteEntityStub) GetLocalAddr() string {
	if stub == nil { return "" }
	return stub.localAddr
}

func (stub *RemoteEntityStub) GetRemoteAddr() string {
	if stub == nil { return "" }
	return stub.remoteAddr
}

func (stub *RemoteEntityStub) Bind() bool {
	if stub == nil { return false }
	c := gateManager.getGateByAddr(stub.addr)
	if c == nil {
		stub.disconnected = 1
		return false
	}
	err := SendBindClient(c, stub.localAddr, stub.remoteAddr)
	if err != nil {
		stub.disconnected = 1
		return false
	}
	registerClientCallback(stub.remoteAddr, stub)
	stub.disconnected = 0
	return true
}

func (stub *RemoteEntityStub) Switch(addr, remoteAddr string) bool {
	if stub == nil { return false }
	c := gateManager.getGateByAddr(addr)
	if c == nil { return false }
	if oldc := gateManager.getGateByAddr(stub.addr); oldc != nil {
		SendUnbindClient(oldc, stub.localAddr, stub.remoteAddr)
	}
	unregisterClientCallback(stub.remoteAddr)
	stub.addr = addr
	stub.localAddr = c.LocalAddr()
	stub.remoteAddr = remoteAddr
	return true
}

func (stub *RemoteEntityStub) Disconnect() bool {
	if stub == nil { return false }
	if stub.disconnected == 1 { return false }
	stub.disconnected = 1
	if c := gateManager.getGateByAddr(stub.addr); c != nil {
		SendUnbindClient(c, stub.localAddr, stub.remoteAddr)
	}
	unregisterClientCallback(stub.remoteAddr)
	stub.core.PushMainTask(stub.disconnectCallback)
	return true
}

func (stub *RemoteEntityStub) OnClientDisconnect() {
	if stub == nil { return }
	if stub.disconnected == 1 { return }
	stub.disconnected = 1
	stub.core.PushMainTask(stub.disconnectCallback)
}

func (stub *RemoteEntityStub) Yield(core *EntityCore) *RemoteEntityStub {
	if stub == nil { return nil }
	return &RemoteEntityStub{
		core: core,
		addr: stub.addr,
		localAddr: stub.localAddr,
		remoteAddr: stub.remoteAddr,
	}
}

func (stub *RemoteEntityStub) CreateEntity(data interface{}) bool {
	if stub == nil { return false }
	c := gateManager.getGateByAddr(stub.addr)
	if c == nil {
		stub.Disconnect()
		return false
	}
	err := SendCreateEntity(c, stub.remoteAddr, stub.core.GetId(), stub.core.GetType(), data)
	if err != nil {
		stub.Disconnect()
		return false
	}
	return true
}

func (stub *RemoteEntityStub) CallClientMethod(method string, params interface{}) bool {
	if stub == nil { return false }
	c := gateManager.getGateByAddr(stub.addr)
	if c == nil {
		stub.Disconnect()
		return false
	}
	err := SendClientEntityMsg(c, stub.remoteAddr, stub.core.GetId(), method, params)
	if err != nil {
		stub.Disconnect()
		return false
	}
	return true
}

func (stub *RemoteEntityStub) SetFilterData(fdata map[string]int32) {
	if stub == nil { return }
	SendSetFilterData(lbtproto.FilterData_SET, stub.addr, stub.remoteAddr, fdata)
}

func (stub *RemoteEntityStub) UpdateFilterData(fdata map[string]int32) {
	if stub == nil { return }
	SendSetFilterData(lbtproto.FilterData_UPDATE, stub.addr, stub.remoteAddr, fdata)
}

func (stub *RemoteEntityStub) DeleteFilterData(fdata map[string]int32) {
	if stub == nil { return }
	SendSetFilterData(lbtproto.FilterData_DELETE, stub.addr, stub.remoteAddr, fdata)
}
