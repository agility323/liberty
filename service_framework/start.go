package service_framework

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtutil"
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
	listenAddr := serviceConf.GateServerAddr
	if listenAddr[0] == ':' {
		if localip, err := lbtutil.GetLocalIP(); err != nil {
			panic(fmt.Sprintf("get local ip fail: %v", err))
		} else { listenAddr = localip + listenAddr }
	}
	gateServer := lbtnet.NewTcpServer(listenAddr, GateConnectionCreator)
	logger.Info("create service server at %s", gateServer.GetAddr())
	gateServer.Start()
	gateManager.start()

	// register
	lbtreg.InitWithEtcd(serviceConf.Etcd)
	go lbtreg.StartRegisterService(
		11,
		make(chan bool),
		serviceConf.Host,
		serviceConf.ServiceType,
		listenAddr,
	)

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
