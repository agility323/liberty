package main

import (
	"os"

	"github.com/agility323/liberty/lbtutil"
	sf "github.com/agility323/liberty/service_framework"

	"github.com/agility323/liberty/demo_service/avatar/misc"
)

var logger = sf.Logger

func main() {
	// conf
	misc.Conf.Service = sf.GetServiceConf()
	lbtutil.LoadConfFromCmdLine(misc.DefaultConf, os.Args[1:], &misc.Conf)
	// db
	//misc.InitRedisClient(misc.Conf.Redis.Addrs, misc.Conf.Redis.MasterName)
	misc.InitMongoClient(misc.Conf.Mongo.Uri)
	// method
	sf.RegisterMethodHandlerCreator("create_avatar", func() sf.MethodHandler {return new(method.CreateAvatarHandler)})
	// start
	sf.Start(OnShutdown)
}

func OnShutdown() {
}
