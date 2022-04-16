package main

import (
	"reflect"
	"context"
	"time"

	sf "github.com/agility323/liberty/service_framework"
	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/demo_service/avatar/avatardata"

	"go.mongodb.org/mongo-driver/mongo/options"
)

const AvatarType = "Avatar"

const (
	ColNameEntity = "entities"
)

func init() {
	sf.RegisterEntityType(AvatarType, reflect.TypeOf((*Avatar)(nil)), AvatarRpcList)
}

type Avatar struct {
	EC sf.EntityCore
	stub *sf.RemoteEntityStub
	data *avatardata.AvatarData
}

func (a *Avatar) Init(c *lbtnet.TcpConnection, srcAddr string, data *avatardata.AvatarData, isNew bool) {
	a.stub = sf.NewRemoteEntityStub(&a.EC, c, srcAddr)
	sf.RegisterClientCallback(srcAddr, a)

	if data.Buildings == nil { data.Buildings = make(map[string]*avatardata.BuildingProp) }
	if data.Interacts == nil { data.Interacts = avatardata.MakeInitInteractData() }
	a.data = data

	if isNew {
		logger.Info("new avatar init %s", a.EC.GetId().Hex())
		a.onNewLogin()
	} else {
		logger.Info("avatar init %s", a.EC.GetId().Hex())
	}
	a.onLogin()
}

func (a *Avatar) Start() {
	logger.Debug("avatar start %s", a.EC.GetId().Hex())
	data := map[string]interface{} {
		"EC": a.EC.Dump(),
		"addr": a.stub.GetC().LocalAddr(),
		"data": a.data,
	}
	a.stub.CreateEntity(data)
	a.afterLogin()
}

func (a *Avatar) OnClientDisconnect() {
	logger.Debug("avatar OnClientDisconnect %s", a.EC.GetId().Hex())
}

func (a *Avatar) Save() {
	col := mongoClient.Database(DbNameGame).Collection(ColNameEntity)
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	opts := &options.UpdateOptions{}
	opts.SetUpsert(true)
	col.UpdateByID(ctx, a.EC.GetId(), a.data, opts)
}

func (a *Avatar) onNewLogin() {
}

func (a *Avatar) onLogin() {
}

func (a *Avatar) afterLogin() {
}

func (a *Avatar) updateProp(k string, v interface{}) {
	a.stub.EntityMsg("CMD_update_prop", []interface{}{k, v, })
}

func (a *Avatar) showReward(rewards [][]int) {
	a.stub.EntityMsg("CMD_show_reward", []interface{}{rewards, })
}
