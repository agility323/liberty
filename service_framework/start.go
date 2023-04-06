package service_framework

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"context"

	"net/http"
	_ "net/http/pprof"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtreg"
)

var (
	stopCh = make(chan os.Signal, 1)
	softStopCh = make(chan bool, 1)
)

var (
	cancelRegister context.CancelFunc = nil
)

func Start(cb func()) {
	// check
	if !checkServiceConf() {
		panic("service start failed: invalid serviceConf")
	} else {
		logger.Info("service start with conf: %v", serviceConf)
	}

	// profile
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", serviceConf.ProfilePort), nil)

	// gate server
	serviceAddr = serviceConf.GateServerAddr
	if serviceAddr[0] == ':' {
		if localip, err := lbtutil.GetLocalIP(); err != nil {
			panic(fmt.Sprintf("get local ip fail: %v", err))
		} else { serviceAddr = localip + serviceAddr }
	}
	gateServer := lbtnet.NewTcpServer(serviceAddr, GateConnectionCreator)
	logger.Info("create service server at %s", gateServer.GetAddr())
	gateServer.Start()

	// register
	lbtreg.InitWithEtcd(serviceConf.Etcd)
	var ctx context.Context
	ctx, cancelRegister = context.WithCancel(context.Background())
	go lbtreg.StartRegisterService(
		ctx,
		31,
		serviceConf.Host,
		serviceConf.ServiceType,
		serviceAddr,
		regData,
	)
	// watch
	go lbtreg.StartWatchServiceCmd(ctx, OnWatchServiceCmd, serviceConf.Host)

	// on start
	onStart()

	// wait for stop
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stopCh:
		stop()
	case <-softStopCh:
		beforeStop(cb)
		stop()
	}
}

func onStart() {
	if serviceConf.TickTime > 0 {
		tickmgr.ResetTickTime(serviceConf.TickTime)
	}
	entmgr.onStart()
	ccbmgr.onStart()
	tickmgr.Start()
}

func InitiateStop() {
	select {
	case stopCh<- syscall.SIGTERM:
	default:
	}
}

func InitiateSoftStop() {
	select {
	case softStopCh<- true:
	default:
	}
}

func beforeStop(cb func()) {
	logger.Info("service stop begin %s", serviceConf.ServiceType)
	cancelRegister()
	gateManager.notifyServiceStop()
	cb()
}

func stop() {
	select {
	case serviceCheckStopCh<- true:
	default:
	}
	tickmgr.Stop()
	logger.Info("service stop finish %s", serviceConf.ServiceType)
}
