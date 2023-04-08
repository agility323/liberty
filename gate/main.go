package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtreg"
	"github.com/agility323/liberty/lbtutil"

	"github.com/agility323/liberty/gate/legacy"
)

func init() {
}

var stopping = false
var (
	cancelRegister context.CancelFunc = nil
	cancelDiscover context.CancelFunc = nil
	cancelWatchCmd context.CancelFunc = nil
)

var (
	stopCh = make(chan os.Signal, 1)
	softStopCh = make(chan bool, 1)
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(200)

	// conf
	lbtutil.LoadConfFromCmdLine(defaultConf, os.Args[1:], &Conf)

	// log
	lbtutil.SetLogLevel(Conf.LogLevel)

	// profile
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", Conf.ProfilePort), nil)

	// init legacy dependency
	dep := legacy.LegacyDependency{
		ConnectServerEntity: Conf.ConnectServerHandler.Entity,
		LegacyRouteTypeMap: map[string]int32 {
			"random": lbtproto.RouteTypeRandomOne,
			"hash": lbtproto.RouteTypeHash,
			"specific": lbtproto.RouteTypeSpecific,
			"all": lbtproto.RouteTypeAll,
		},
		ServiceAddrGetter: clientManager.getClientServiceAddr,
		ServiceSender: serviceManager.sendToService,
		ServiceRequestHandler: serviceManager.serviceRequest,
		PrivateRsaKey: Conf.PrivateRsaKey,
		AtService: AtService,
	}
	if err := legacy.InitLegacyDependency(dep); err != nil {
		panic(fmt.Sprintf("InitLegacyDependency fail %v\n\t%v", dep, err))
	}

	// client server
	if Conf.ClientServerAddr[0] == ':' {
		if localip, err := lbtutil.GetLocalIP(); err != nil {
			panic(fmt.Sprintf("get local ip fail: %v", err))
		} else {
			Conf.ClientServerAddr = localip + Conf.ClientServerAddr
		}
	}
	gateAddr = Conf.ClientServerAddr
	clientServer := lbtnet.NewTcpServer(Conf.ClientServerAddr, ClientConnectionCreator)
	logger.Info("create client server at %s entrance is %s", clientServer.GetAddr(), Conf.EntranceAddr)
	clientServer.Start()

	// register, discrover, watch
	InitRegData()
	lbtreg.InitWithEtcd(Conf.Etcd)
	var ctxRegister, ctxDiscover, ctxWatchCmd context.Context
	ctxRegister, cancelRegister = context.WithCancel(context.Background())
	// usually, gate exposes its addr directly to clients, so the entrance addr should be unique
	// otherwise, such as there are proxies between gates and clients, the entrance varies
	// so here we use listen addr as unique addr
	go lbtreg.StartRegisterGate(ctxRegister, 11, Conf.Host, Conf.ClientServerAddr, regData)
	ctxDiscover, cancelDiscover = context.WithCancel(context.Background())
	go lbtreg.StartDiscoverService(ctxDiscover, 11, serviceManager.OnDiscoverService, Conf.Host)
	ctxWatchCmd, cancelWatchCmd = context.WithCancel(context.Background())
	go lbtreg.StartWatchGateCmd(ctxWatchCmd, OnWatchGateCmd, Conf.Host)

	// on start
	onStart()

	// wait for stop
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stopCh:
		stop()
	case <-softStopCh:
		done := beforeStop()
		select {
		case <-done:
			stop()
		}
	}
}

func onStart() {
	if Conf.TickTime > 0 {
		tickmgr.ResetTickTime(Conf.TickTime)
	}
	clientManager.onStart()
	serviceManager.onStart()
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

func AtService() bool {
	if stopping { return false }
	return true
}

func beforeStop() <-chan bool {
	stopping = true
	logger.Info("gate stop begin")
	cancelWatchCmd()
	cancelRegister()
	cancelDiscover()
	serviceManager.notifyGateStop()
	return clientManager.SoftStop(120, 20)
}

func stop() {
	tickmgr.Stop()
	logger.Info("gate stop finish")
}
