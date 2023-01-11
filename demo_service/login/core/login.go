package core

import (
	"context"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/agility323/liberty/demo_service/login/misc"
)

func StartLogin(acc string) (int, string) {
	// check and set account/role data
	filter := bson.M{"_id": acc}
	sessionid := lbtutil.NewObjectID().Hex()
	update := bson.M{"$setOnInsert": MakeOnlineData(sessionid)}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)//.SetProjection(bson.M{})
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second())
	err := misc.ColAccount.FindOneAndUpdate(ctx, bson.M{"_id": acc}, bson.D{{"$setOnInsert", newAcc}}, opts).Decode(&oldAcc) // 没有才更新
}

func MakeOnlineData(sessionid string) bson.M {
	return bson.M{
		"sessionid": sessionid,
		"online": int(time.time()),
		"saddr": "",
		"caddr": "",
	}
}
