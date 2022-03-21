package main

import (
	"os"

	"github.com/agility323/liberty/lbtutil"
	sf "github.com/agility323/liberty/service_framework"
)

var defaultConfData map[string]interface{} = map[string]interface{} {
	"service_type": "login_service",
}

func main() {
	lbtutil.LoadConfFromCmdLine(defaultConf, os.Args[1:], &Conf)
	sf.ServiceType = Conf.ServiceType

	sf.RegisterMethodHandlerCreator("login", func() sf.MethodHandler {return new(loginHandler)})

	sf.Start()
}
