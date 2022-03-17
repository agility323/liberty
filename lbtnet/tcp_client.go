package lbtnet

import (
	"net"
	"strconv"
	"sync/atomic"
	"os"
	"time"
	"errors"

	"github.com/agility323/liberty/lbtutil"
)

const (
	ST_NOT_CONNECTED = iota
	ST_CONNECTING
	ST_CONNECTED
	ST_STOPPING
	ST_STOPPED
)

type TcpClient struct {
	state int32
	raddr *net.TCPAddr
	timer *time.Timer
	logger lbtutil.Logger
	fconn *TcpConnection
	handler ConnectionHandler
}

const (
	CLIENT_RECONNECT_TIME int = 5
)

func NewTcpClient(ip string, port int, handler ConnectionHandler) *TcpClient {
	addr, err := net.ResolveTCPAddr("tcp", ip + ":" + strconv.Itoa(port))
	if err != nil {
		panic("tcp client addr fail " + ip + ":" + strconv.Itoa(port))
	}
	client := &TcpClient{
		state: ST_NOT_CONNECTED,
		raddr: addr,
		timer: nil,
		logger: lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "TcpClient"),
		fconn: nil,
		handler: handler,
	}
	return client
}

func (client *TcpClient) LocalAddr() string {
	if client.fconn != nil { return client.fconn.LocalAddr() }
	return ""
}

func (client *TcpClient) StartConnect() {
	if !atomic.CompareAndSwapInt32(&client.state, ST_NOT_CONNECTED, ST_CONNECTING) {
		client.logger.Info("client connect abort code 1")
		return
	}
	conn, err := net.DialTCP("tcp", nil, client.raddr)
	if err != nil {
		// connect fail
		client.timer = time.AfterFunc(time.Duration(CLIENT_RECONNECT_TIME) * time.Second, client.StartConnect)
		client.logger.Info("client connect fail %s %s retry in %d sec",
				client.raddr.String(), err.Error(), CLIENT_RECONNECT_TIME)
		atomic.StoreInt32(&client.state, ST_NOT_CONNECTED)
		return
	}
	// connect success
	client.timer = nil
	client.logger.Info("client connect success %s", client.raddr.String())
	client.fconn = NewTcpConnection(conn, client.handler)
	atomic.StoreInt32(&client.state, ST_CONNECTED)
	// Start may cause OnConnectionClose, so run it after ST_CONNECTED
	client.fconn.Start()
	client.handler.OnConnectionReady(client.fconn)
}

func (client *TcpClient) Stop() {
	if !atomic.CompareAndSwapInt32(&client.state, ST_CONNECTED, ST_STOPPING) {
		client.logger.Info("client stop abort code 1")
		return
	}
	if client.timer != nil {
		client.timer.Stop()
		client.timer = nil
	}
	client.fconn.Close()
	client.fconn = nil
	atomic.StoreInt32(&client.state, ST_STOPPED)
}

func (client *TcpClient) OnConnectionClose() {
	if !atomic.CompareAndSwapInt32(&client.state, ST_CONNECTED, ST_STOPPING) { return }
	client.logger.Info("client connection close %s", client.raddr.String())
	client.fconn = nil
	atomic.StoreInt32(&client.state, ST_NOT_CONNECTED)
	client.timer = time.AfterFunc(time.Duration(CLIENT_RECONNECT_TIME) * time.Second, client.StartConnect)
}

func (client *TcpClient) SendData(buf []byte) error {
	if atomic.LoadInt32(&client.state) != ST_CONNECTED { return errors.New("TcpClient.SendData: not connected") }
	return client.fconn.SendData(buf)
}
