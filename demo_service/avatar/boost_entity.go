package main

import (
	"reflect"
	"context"
	"time"

	"github.com/agility323/liberty/lbtnet"
	sf "github.com/agility323/liberty/service_framework"
	"github.com/agility323/liberty/demo_service/avatar/avatardata"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
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
	sf.RegisterClientCallback(srcAddr, b)
}

func (b *BoostEntity) Start() {
	logger.Debug("boost entity start %s", b.EC.GetId().Hex())
	b.stub.Bind()	// bind client proxy
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
	b.stub.EntityMsg("CMD_show_msg_sc", args)
	// login
	key := RedisKey([]string{RedisKeyLogin, token})
	val := b.stub.GetLocalAddr()
	res, err := redisClient.SetNX(key, val, 0).Result()
	if err != nil || !res {
	}
	b.stub.EntityMsg("CMD_login_reply", []interface{}{0, })

	// load from db
	col := mongoClient.Database(DbNameGame).Collection(ColNameEntity)
	data := &avatardata.AvatarData{}
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	err = col.FindOne(ctx, bson.D{{"_id", token}}).Decode(data)
	isNew := (err == mongo.ErrNoDocuments)
	if !isNew && err != nil {
		logger.Error("find avatar err %s %v", token, err)
		return
	}

	// create avatar
	avatar := sf.CreateEntity("Avatar").(*Avatar)
	avatar.Init(b.stub.Yield(&avatar.EC), data, isNew)
	avatar.Start()
}
