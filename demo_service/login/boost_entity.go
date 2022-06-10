package main

import (
	"reflect"
	"context"

	"github.com/agility323/liberty/lbtnet"
	sf "github.com/agility323/liberty/service_framework"

)

const BoostEntityType = "BoostEntity"
var BoostEntityRpcList = []string {"CMD_login_startlogin_cs", }

var redisCtx = context.TODO()

func init() {
	sf.RegisterEntityType(BoostEntityType, reflect.TypeOf((*BoostEntity)(nil)), BoostEntityRpcList)
}

type BoostEntity struct {
	EC sf.EntityCore
	stub *sf.RemoteEntityStub
}

func (b *BoostEntity) Init(c *lbtnet.TcpConnection, srcAddr string) {
	b.stub = sf.NewRemoteEntityStub(&b.EC, c, srcAddr)
}

func (b *BoostEntity) Start() {
	logger.Debug("boost entity start %s", b.EC.GetId().Hex())
	b.stub.Bind(b)	// bind client proxy
	data := map[string]interface{} {
		"EC": b.EC.Dump(),
		"addr": b.stub.GetLocalAddr(),
	}
	b.stub.CreateEntity(data)
}

func (b *BoostEntity) OnClientDisconnect() {
	logger.Debug("boost entity OnClientDisconnect %s", b.EC.GetId().Hex())
}

func (b *BoostEntity) CMD_login_startlogin_cs(token string, sdkInfo map[string]interface{}) {
	logger.Debug("BoostEntity CMD_login_startlogin_cs %v %v", token, sdkInfo)
	// test
	args := []interface{}{"CMD_login_startlogin_cs received", }
	b.stub.CallClientMethod("CMD_show_msg_sc", args)
	// login
	b.stub.CallClientMethod("CMD_login_reply", []interface{}{0, })

	// load from db

	// create avatar

}
