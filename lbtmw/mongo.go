package lbtmw

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DbNameGame = "gamedb"
)

var mongoClient *mongo.Client = nil

func InitMongoClient(uri string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	mongoClient = client
}

func CloseMongoClient() {
	if err := mongoClient.Disconnect(context.Background()); err != nil {
		panic(err)
	}
}
