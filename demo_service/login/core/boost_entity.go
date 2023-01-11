package core

import (
	"reflect"
	"context"

	"github.com/agility323/liberty/lbtnet"
	sf "github.com/agility323/liberty/service_framework"

)

var BoostEntityRpcList = []string {"CMD_login_startlogin_cs", }

var redisCtx = context.TODO()

func init() {
	var rpclist = []string{"CMD_boost_handshake", }
	sf.RegisterEntityType("BoostEntity", reflect.TypeOf((*BoostEntity)(nil)), rpclist)
}

type BoostEntity struct {
	EC sf.EntityCore
	stub *sf.RemoteEntityStub
	destroyed bool
}

func (b *BoostEntity) Init(c *lbtnet.TcpConnection, srcAddr string) {
	b.stub = sf.NewRemoteEntityStub(&b.EC, c.RemoteAddr(), srcAddr, b.OnClientDisconnect)
	b.destroyed = false
}

func (b *BoostEntity) Start() {
	logger.Debug("boost entity start %s", b.EC.GetId().Hex())
	if !b.stub.Bind() {  // bind client proxy
		b.Destroy()
		return
	}
	data := map[string]interface{}{
		"EC":   b.EC.Dump(),
		"addr": b.stub.GetLocalAddr(),
	}
	if !b.stub.CreateEntity(data) {
		return
	}
	b.EC.StartActor(100)
}

func (b *BoostEntity) Destroy() {
	if b.destroyed { return }
	b.destroyed = true
	b.EC.StopActor()
	sf.DestroyEntity(b.EC.GetId())
}

func (b *BoostEntity) OnClientDisconnect() {
	logger.Info("BoostEntity.OnClientDisconnect %s", b.EC.GetId().Hex())
	b.EC.PushActorTask(func() {
		b.Destroy()
	})
}

func (b *BoostEntity) CMD_boost_handshake(msg string) {
	logger.Info("CMD_boost_handshake recv %s", msg)
	// handshake
	args := []interface{}{msg}
	if !b.stub.CallClientMethod("CMD_boost_handshake", args) {
		logger.Error("CMD_boost_handshake send fail")
		return
	}
}
