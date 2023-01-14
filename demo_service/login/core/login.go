package core

import (
	"context"
	"strconv"
	"time"

	"github.com/agility323/liberty/lbtutil"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/agility323/liberty/demo_service/login/misc"
)

const RoleNumLimit int = 1

const (
	LoginSuccess int = iota
	LoginPasswordError
	LoginAccountNotExist
	LoginAvatarNotExist
	LoginWrongState
	LoginRedisErr
	LoginParamErr
	LoginInternalErr
	LoginTokenError
	LoginSameRoleExist
	LoginClientVersionError
	LoginAvatarsFull
)

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

type AvatarData struct {
	Id string `bson:"id" msgpack:"id"`
	Name string `bson:"name" msgpack:"name"`
	Host int `bson:"host" msgpack:"host"`
}

type SdkAccountData struct {
	AppId string `json:"app_id"`
	ChannelId string `json:"channel_id"`
	PackageId string `json:"package_id"`
	AccountId string `json:"account_id"`
	Time int64 `json:"time"`
	Ext map[string]interface{}`json:"ext"`
}

func StartLogin(caddr, aid string) (int, map[string]*AvatarData, string) {
	// assure account data
	token := lbtutil.NewObjectID().Hex()
	ctoken := aid + "|" + token
	doc := AccountData{}
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	err := misc.ColAccount.FindOneAndUpdate(
		ctx,
		bson.M{"_id": aid},
		bson.M{"$setOnInsert": bson.M{
			"avatars": bson.M{},
			//"token": token,
			"online": 0,
			"saddr": "",
			"caddr": "",
		}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&doc)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Error("start login fail 1 %s %s %v", aid, caddr, err)
		return LoginInternalErr, nil, ""
	}
	if doc.Avatars == nil {
		logger.Error("start login fail 2 %s %s %v", aid, caddr, doc)
		return LoginInternalErr, nil, ""
	}
	// new account
	if len(doc.Avatars) == 0 {
		return LoginSuccess, nil, ctoken
	}
	// old account
	if len(doc.Avatars) > 0 {
		avatars := make(map[string]*AvatarData)
		for _, data := range doc.Avatars {
			avatars[data.Id] = &data
		}
		return LoginSuccess, avatars, ctoken
	}
	// unexpected
	logger.Error("start login fail 3 %s %s %v", aid, caddr, doc)
	return LoginInternalErr, nil, ""
}

func CreateRole(caddr, aid string, data *AvatarData) (int, string) {
	data.Id = lbtutil.NewObjectID().Hex()
	data.Host = misc.Conf.Service.Host
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	res, err := misc.ColAccount.UpdateOne(
		ctx,
		bson.M{
			"_id": aid,
			"avatars." + strconv.Itoa(RoleNumLimit - 1): bson.M{"$exists": false},
		},
		bson.M{"$push": data},
		options.Update(),
	)
	if err != nil {
		logger.Error("create role fail 1 %s %s %v %v", aid, caddr, *data, err)
		return LoginInternalErr, ""
	}
	if res.MatchedCount == 0 {
		return LoginAvatarsFull, ""
	}
	if res.ModifiedCount == 0{
		logger.Error("create role fail 2 %s %s %v", aid, caddr, *data)
		return LoginInternalErr, ""
	}
	return LoginSuccess, data.Id
}
