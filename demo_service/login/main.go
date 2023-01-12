package main

import (
	"os"

	"github.com/agility323/liberty/lbtutil"
	sf "github.com/agility323/liberty/service_framework"

	"github.com/agility323/liberty/demo_service/login/misc"
	"github.com/agility323/liberty/demo_service/login/method"
)

var logger = sf.Logger

func main() {
	// conf
	misc.Conf.Service = sf.GetServiceConf()
	lbtutil.LoadConfFromCmdLine(misc.DefaultConf, os.Args[1:], &misc.Conf)
	// method
	sf.RegisterMethodHandlerCreator("connect_server", func() sf.MethodHandler {return new(method.ConnectServerHandler)})
	sf.RegisterMethodHandlerCreator("login", func() sf.MethodHandler {return new(method.LoginHandler)})
	sf.RegisterMethodHandlerCreator("create_role", func() sf.MethodHandler {return new(method.CreateRoleHandler)})
	// start
	sf.Start(OnShutdown)
}

func OnShutdown() {
}
