package main

import (
	"runtime"
	"runtime/debug"
	"os"
	"os/signal"
	"syscall"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtutil"
)

func init() {
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(200)

	// conf
	lbtutil.LoadConfFromCmdLine(defaultConf, os.Args[1:], &Conf)

	// log
	lbtutil.SetLogLevel(Conf.LogLevel)

	// service server
	serviceServer := lbtnet.NewTcpServer(Conf.ServiceServerAddr, ServiceConnectionCreator)
	logger.Info("create service server at %s", serviceServer.GetAddr())
	serviceServer.Start()
	serviceManager.start()
	// client server
	clientServer := lbtnet.NewTcpServer(Conf.ClientServerAddr, ClientConnectionCreator)
	logger.Info("create client server at %s", clientServer.GetAddr())
	clientServer.Start()
	clientManager.start()

	// wait for stop
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh

	// on stop
	onStop()
}

func onStop() {
}
