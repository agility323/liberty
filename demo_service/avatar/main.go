package main

import (
	"os"

	"github.com/agility323/liberty/lbtutil"
	sf "github.com/agility323/liberty/service_framework"
)

var logger = sf.Logger

func main() {
	// conf
	Conf.Service = sf.GetServiceConf()
	lbtutil.LoadConfFromCmdLine(defaultConf, os.Args[1:], &Conf)

	// db
	InitRedisClient(Conf.Redis.Addrs, Conf.Redis.MasterName)
	InitMongoClient(Conf.Mongo.Uri)

	// method
	sf.RegisterMethodHandlerCreator("connect_server", func() sf.MethodHandler {return new(connectServerHandler)})
	sf.RegisterMethodHandlerCreator("create_avatar", func() sf.MethodHandler {return new(createAvatarHandler)})

	// start
	sf.Start(OnShutdown)
}

func OnShutdown() {
}
