package lbtnet

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/agility323/liberty/lbtutil"
)


type TcpServer struct {
	started int32
	stop int32
	listener *net.TCPListener
	logger lbtutil.Logger
	cc connectionCreatorFunc
}

func NewTcpServer(addr string, cc connectionCreatorFunc) *TcpServer {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic("tcp server addr fail " + addr)
	}
	if tcpAddr.IP.String() == "" { logger.Warn("tcp server ip not specified") }
	if tcpAddr.Port == 0 { logger.Warn("tcp server port not specified") }
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic("tcp server listen fail " + addr)
	}
	server := &TcpServer{
		started: 0,
		stop: 0,
		listener: listener,
		cc: cc,
	}
	logger.Info("tcp server listen at %s", server.GetAddr())
	return server
}

func (server *TcpServer) Start() {
	if atomic.CompareAndSwapInt32(&server.started, 0, 1) {
		go server.acceptLoop()
	}
}

func (server *TcpServer) Stop() {
	if atomic.CompareAndSwapInt32(&server.stop, 0, 1) {
		server.listener.Close()
	}
}

func (server *TcpServer) GetAddr() string {
	if server.listener == nil { return "" }
	return server.listener.Addr().String()
}

func (server *TcpServer) acceptLoop() {
	defer lbtutil.Recover(fmt.Sprintf("TcpServer.acceptLoop %v", server.listener), func() { go server.acceptLoop() })

	for  {
		conn, err := server.listener.AcceptTCP()
		if err != nil {
			if atomic.LoadInt32(&server.stop) != 0 {
				logger.Info("tcp server close %s", server.GetAddr())
				return
			}
			continue
		}
		logger.Info("tcp server new conn %s", conn.RemoteAddr().String())
		server.cc(conn)
	}
}
