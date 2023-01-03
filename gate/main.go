package main

import (
	"runtime"
	"runtime/debug"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtreg"

	"github.com/agility323/liberty/gate/legacy"
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

	// profile
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", Conf.ProfilePort), nil)

	// init legacy dependency
	dep := legacy.LegacyDependency{
		ConnectServerService: Conf.ConnectServerHandler.Service,
		ConnectServerMethod: Conf.ConnectServerHandler.Method,
		PostClientManagerJob: postClientManagerJob,
		PostServiceManagerJob: postServiceManagerJob,
		LegacyRouteTypeMap: map[string]int32 {
			"random": RouteTypeRandomOne,
			"hash": RouteTypeHash,
			"specific": RouteTypeSpecific,
			"all": RouteTypeAll,
		},
		ServiceAddrGetter: clientManager.getServiceAddr,
		PrivateRsaKey: Conf.PrivateRsaKey,
	}
	if err := legacy.InitLegacyDependency(dep); err != nil {
		panic(fmt.Sprintf("InitLegacyDependency fail %v\n\t%v", dep, err))
	}

	// client server
	listenAddr := Conf.ClientServerAddr
	if listenAddr[0] == ':' {
		if localip, err := lbtutil.GetLocalIP(); err != nil {
			panic(fmt.Sprintf("get local ip fail: %v", err))
		} else { listenAddr = localip + listenAddr }
	}
	clientServer := lbtnet.NewTcpServer(listenAddr, ClientConnectionCreator)
	logger.Info("create client server at %s entrance is %s", clientServer.GetAddr(), Conf.EntranceAddr)
	clientServer.Start()
	clientManager.start()

	// service server (playing proxy role)
	serviceManager.start()

	// register
	lbtreg.InitWithEtcd(Conf.Etcd)
	// usually, gate exposes its addr directly to clients, so the entrance addr should be unique
	// otherwise, such as there are proxies between gates and clients, the entrance varies
	// so here we use listen addr as unique addr
	go lbtreg.StartRegisterGate(11, make(chan bool), Conf.Host, listenAddr)
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
