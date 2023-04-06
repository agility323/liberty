package lbtnet

import (
	"net"
	"strconv"
	"sync/atomic"
	"os"
	"time"

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
	reconnectTime int
}

const (
	DEFAULT_RECONNECT_TIME int = 10
)

func NewTcpClient(addr string, handler ConnectionHandler) *TcpClient {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic("tcp client addr fail " + addr)
	}
	client := &TcpClient{
		state: ST_NOT_CONNECTED,
		raddr: tcpAddr,
		timer: nil,
		logger: lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "TcpClient"),
		fconn: nil,
		handler: handler,
		reconnectTime: DEFAULT_RECONNECT_TIME,
	}
	return client
}

func (client *TcpClient) LocalAddr() string {
	if client.fconn != nil { return client.fconn.LocalAddr() }
	return ""
}

func (client *TcpClient) RemoteAddr() string {
	return client.raddr.String()
}

func (client *TcpClient) SetReconnectTime(reconnectTime int) {
	client.reconnectTime = reconnectTime
}

func (client *TcpClient) StartConnect(reconnectCount int) {
	if !atomic.CompareAndSwapInt32(&client.state, ST_NOT_CONNECTED, ST_CONNECTING) {
		client.logger.Info("client connect abort code 1")
		return
	}
	client.logger.Info("client start connect %s %d", client.raddr.String(), reconnectCount)
	conn, err := net.DialTCP("tcp", nil, client.raddr)
	if err != nil {
		// connect fail
		atomic.StoreInt32(&client.state, ST_NOT_CONNECTED)
		reconnectCount -= 1
		if reconnectCount <= 0 {
			client.logger.Info("client connect fail %s %s", client.raddr.String(), err.Error())
			client.handler.OnConnectionFail(client)
			return
		}
		// delayed retry
		client.logger.Info("client connect fail %s %s retry in %d sec",
				client.raddr.String(), err.Error(), client.reconnectTime)
		client.timer = time.AfterFunc(
			time.Duration(client.reconnectTime) * time.Second,
			func() { client.StartConnect(reconnectCount) },
		)
		return
	}
	// connect success
	client.timer = nil
	client.logger.Info("client connect success %s", client.raddr.String())
	conf := ConnectionConfig{WriteChLen: DefaultWriteChLen, WriteChWaitTime: DefaultWriteChWaitTime, ErrLog: true}
	client.fconn = NewTcpConnection(conn, client.handler, conf)
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

// never called
func (client *TcpClient) OnConnectionClose() {
	if !atomic.CompareAndSwapInt32(&client.state, ST_CONNECTED, ST_STOPPING) { return }
	client.logger.Info("client connection close %s", client.raddr.String())
	client.fconn = nil
	atomic.StoreInt32(&client.state, ST_NOT_CONNECTED)
	client.timer = time.AfterFunc(
		time.Duration(client.reconnectTime) * time.Second,
		func() { client.StartConnect(1) },
	)
}

func (client *TcpClient) SendData(buf []byte) error {
	if atomic.LoadInt32(&client.state) != ST_CONNECTED { return ErrSendClientNotReady }
	return client.fconn.SendData(buf)
}
