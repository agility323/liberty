package service_framework

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/agility323/liberty/lbtnet"
)

var stopCh = make(chan os.Signal, 1)

func Start() {
	// check
	if !checkServiceConf() {
		panic("service start failed: invalid serviceConf")
	} else {
		logger.Info("service start with conf: %v", serviceConf)
	}

	// gate client
	gateClient = lbtnet.NewTcpClient(serviceConf.gateAddr, NewGateConnectionHandler())
	gateClient.StartConnect()

	// wait for stop
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh

	// on stop
	onStop()
}

func onStop() {
	logger.Info("service stopped %s", serviceConf.serviceType)
}

func Stop() {
	gateClient.Stop()
	stopCh <- syscall.SIGTERM
}
