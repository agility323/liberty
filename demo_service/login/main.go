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

	// method
	sf.RegisterMethodHandlerCreator("connect_server", func() sf.MethodHandler {return new(connectServerHandler)})
	sf.RegisterMethodHandlerCreator("login", func() sf.MethodHandler {return new(loginHandler)})

	// start
	sf.Start(OnShutdown)
}

func OnShutdown() {
}
