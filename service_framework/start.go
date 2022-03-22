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
	if ServiceType == "" {
		panic("service start failed: ServiceType not set")
	} else {
		logger.Info("service start %s", ServiceType)
	}

	// forged addr
	ip := "127.0.0.1"
	port := 5009

	// gate client
	gateClient = lbtnet.NewTcpClient(ip, port, NewGateConnectionHandler())
	gateClient.StartConnect()

	// wait for stop
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh

	// on stop
	onStop()
}

func onStop() {
	logger.Info("service stopped %s", ServiceType)
}

func Stop() {
	gateClient.Stop()
	stopCh <- syscall.SIGTERM
}
