package main

import (
	"context"
	"reflect"
	"strings"
	"time"

	sf "github.com/agility323/liberty/service_framework"
	"github.com/agility323/liberty/demo_service/avatar/avatardata"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/agility323/liberty/demo_service/avatar/misc"
)

// cp from golang_conf_loader
type TrackingInfo struct {
	DataVersion string `msgpack:"data_version"`
	LaunchId string `msgpack:"launch_id"`
	Fpid string `msgpack:"fpid"`
	LogSource string `msgpack:"log_source"`
	SessionId string `msgpack:"session_id"`
	AppId string `msgpack:"app_id"`
	Properties *TrackingProperties `msgpack:"properties"`
}

type AccountData struct {
	// persistent data
	Aid string `bson:"_id" msgpack:"aid"`
	Avatars  []AvatarData `bson:"avatars" msgpack:"avatars"`
	// runtime data
	Token string `bson:"token" msgpack:"-"`
	Online int `bson:"online" msgpack:"-"`
	OnlineId int `bson:"online_id" msgpack:"-"`
	Saddr string `bson:"saddr" msgpack:"-"`
	Caddr string `bson:"caddr" msgpack:"-"`
}

func CreateAvatar(caddr, ctoken, avatarid string, sdkinfo *TrackingInfo) int {
	// retrieve aid, token
	arr := strings.SplitN(ctoken, "|", 2)
	if len(arr) != 2 {
		return 12345
	}
	aid := arr[0]
	token := arr[1]
	// getset online state in account data
	doc := &AccountData{}
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	err := misc.ColAccount.FindOneAndUpdate(
		ctx,
		bson.M{"_id": aid, "online": 0},
		bson.M{"$set": bson.M{
			"online": int(time.Now().Unix()),
			"online_id": avatarid,
			"saddr": misc.Conf.Service.GateServerAddr,
			"caddr": caddr,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(doc)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Error("create avatar fail 1 %s %s %v", aid, caddr, err)
		return 12345
	}
	// check online avatar
	if err == mongo.ErrNoDocuments {
		doc = GetAccountData(aid)
		if doc.Online > 0 && doc.OnlineId != "" && doc.Saddr != "" {
			// TODO relogin

		}

		return 12345
	}
	// load from db
	adoc := &avatardata.AvatarData{}
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	err := misc.ColEntity.FindOne(
		ctx,
		bson.M{"_id": avatarid},
		options.FindOne(),
	).Decode(adoc)
	if err != nil {
		return 12345
	}
	avatar := sf.CreateEntity("Avatar", avatarid).(*Avatar)
	avatar.Init(sf.NewRemoteEntityStub(&avatar.EC, c, srcAddr), adata, true)
	avatar.Start()

	avatarmanager.add(avatar)
	return 0
}

func GetAccountData(aid string) *AccountData {
	doc := &AccountData{}
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	err := misc.ColAccount.FindOne(
		ctx,
		bson.M{"_id": aid},
		options.FindOne(),
	).Decode(doc)
	if err != nil {
		logger.Error("get online fail 1 %s %v", aid, err)
		return nil
	}
	return doc
}
