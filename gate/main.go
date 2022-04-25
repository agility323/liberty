package main

import (
	"runtime"
	"runtime/debug"
	"os"
	"os/signal"
	"syscall"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtreg"
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

	// client server
	clientServer := lbtnet.NewTcpServer(Conf.ClientServerAddr, ClientConnectionCreator)
	logger.Info("create client server at %s", clientServer.GetAddr())
	clientServer.Start()
	clientManager.start()

	// service server (playing proxy role)
	serviceManager.start()

	// register
	lbtreg.InitWithEtcd(Conf.Etcd)
	go lbtreg.StartRegisterGate(11, make(chan bool), Conf.Host, Conf.ClientServerAddr)
	go lbtreg.StartDiscoverService(11, make(chan bool), OnDiscoverService, Conf.Host)

	// wait for stop
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh

	// on stop
	onStop()
}

func onStop() {
}
