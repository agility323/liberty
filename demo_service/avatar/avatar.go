package main

import (
	"reflect"
	"context"
	"time"

	sf "github.com/agility323/liberty/service_framework"
	"github.com/agility323/liberty/demo_service/avatar/avatardata"

	"go.mongodb.org/mongo-driver/mongo/options"
)

const AvatarType = "Avatar"

func init() {
	sf.RegisterEntityType(AvatarType, reflect.TypeOf((*Avatar)(nil)), AvatarRpcList)
}

type Avatar struct {
	EC sf.EntityCore
	conAddr string	// connection addr
	srcAddr string	// client addr
	data *avatardata.AvatarData
}

func (a *Avatar) Init(conAddr, srcAddr string, data *avatardata.AvatarData, isNew bool) {
	a.conAddr = conAddr
	a.srcAddr = srcAddr

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
	logger.Debug("avatar start %s %s", a.EC.GetType(), a.EC.GetId().Hex())
	data := map[string]interface{} {
		"EC": a.EC.Dump(),
		"addr": a.conAddr,
		"data": a.data,
	}
	sf.SendCreateEntity(a.srcAddr, string(a.EC.GetId()), AvatarType, data)
	a.afterLogin()
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
	sf.SendEntityMsg(a.srcAddr, string(a.EC.GetId()), "CMD_update_prop", []interface{}{k, v, })
}

func (a *Avatar) showReward(rewards [][]int) {
	sf.SendEntityMsg(a.srcAddr, string(a.EC.GetId()), "CMD_show_reward", []interface{}{rewards, })
}
