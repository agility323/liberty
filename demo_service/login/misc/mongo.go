package misc

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

const (
	ColNameAccount = "account"
)

var (
	ColAccount *mongo.Collection
)

func InitMongoClient(uri string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	MongoClient = client
	// collections
	ColAccount = GetColletion(ColNameAccount)
}

func CloseMongoClient() {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := MongoClient.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func GetColletion(colname string) {
	return MongoClient.Database(Conf.Mongo.Db).Collection(colname)
}
