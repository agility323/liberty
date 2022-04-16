package service_framework

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtreg"
)

var stopCh = make(chan os.Signal, 1)

func Start(cb func()) {
	// check
	if !checkServiceConf() {
		panic("service start failed: invalid serviceConf")
	} else {
		logger.Info("service start with conf: %v", serviceConf)
	}

	// gate server
	gateServer := lbtnet.NewTcpServer(serviceConf.GateServerAddr, GateConnectionCreator)
	logger.Info("create service server at %s", gateServer.GetAddr())
	gateServer.Start()
	gateManager.start()

	// register
	lbtreg.InitWithEtcd(serviceConf.Etcd)
	go lbtreg.StartRegisterService(31, make(chan bool), 101, serviceConf.ServiceType, serviceConf.GateServerAddr)

	// wait for stop
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh

	// on stop
	onStop()
	cb()
}

func onStop() {
	logger.Info("service stopped %s", serviceConf.ServiceType)
}

func Stop() {
	gateManager.stop()
	stopCh <- syscall.SIGTERM
}
