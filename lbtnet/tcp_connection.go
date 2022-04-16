package lbtnet

import (
	"net"
	"sync/atomic"
	"encoding/binary"
	"io"
	"bytes"
	"runtime/debug"
	"errors"
)

const (
	SizeLen uint32 = 4
	MaxMsgLen uint32 = 100000
	WriteChLen int = 20
)

type TcpConnection struct {
	started int32
	closed int32
	laddr string
	raddr string
	buf	[]byte
	conn net.Conn
	handler ConnectionHandler
	writeCh chan []byte
}

func NewTcpConnection(conn net.Conn, handler ConnectionHandler) *TcpConnection {
	laddrStr := ""
	raddrStr := ""
	if laddr := conn.LocalAddr(); laddr != nil { laddrStr = laddr.String() }
	if raddr := conn.RemoteAddr(); raddr != nil { raddrStr = raddr.String() }
	c := &TcpConnection{
		started: 0,
		laddr: laddrStr,
		raddr: raddrStr,
		buf: make([]byte, SizeLen + MaxMsgLen),
		conn: conn,
		handler: handler,
		writeCh: make(chan []byte, WriteChLen),
	}
	return c
}

func (c *TcpConnection) LocalAddr() string {
	return c.laddr
}

func (c *TcpConnection) RemoteAddr() string {
	return c.raddr
}

func (c *TcpConnection) Start() {
	if atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		go c.readLoop()
		go c.writeLoop()
	}
}

func (c *TcpConnection) readLoop() {
		/*
	defer func(){
		if r := recover(); r != nil {
			logger.Error("tcp conn panic in read %v %v", c.conn, r)
			c.Close()
		}
	}()
		*/
	var bodyLen uint32
	bufHead := c.buf[:SizeLen]
	for {
		// read head
		_, err := io.ReadFull(c.conn, bufHead)
		if err != nil {
			logger.Debug("tcp conn %s read head fail %s", c.raddr, err.Error())
			c.Close()
			return
		}
		err = binary.Read(bytes.NewReader(bufHead), byteOrder, &bodyLen)
		if err != nil || bodyLen == 0 || bodyLen > MaxMsgLen {
			errmsg := ""
			if err != nil { errmsg = err.Error() }
			logger.Warn("tcp conn %s invalid body len %d %s", c.raddr, bodyLen, errmsg)
			c.Close()
			return
		}
		// read body
		_, err = io.ReadFull(c.conn, c.buf[SizeLen:SizeLen + bodyLen])
		if err != nil {
			logger.Debug("tcp conn %s read body fail %d %s", c.raddr, bodyLen, err.Error())
			c.Close()
			return
		}
		// process proto
		data := make([]byte, SizeLen + bodyLen, SizeLen + bodyLen)
		copy(data, c.buf)
		//logger.Debug("tcp conn read %s %v", c.raddr, data)
		err = c.handler.HandleProto(c, data)
		if err != nil {
			logger.Warn("tcp conn %s proto fail %s", c.raddr, err.Error())
			c.Close()
			return
		}
	}
}

func (c *TcpConnection) writeLoop() {
	for data := range c.writeCh {
		//logger.Debug("tcp conn write %s %v", c.raddr, data)
		n, err := c.conn.Write(data)
		if err != nil {
			logger.Warn("tcp conn %s write fail %d %d %s", c.raddr, len(data), n, err.Error())
			c.Close()
			return
		}
	}
}

func (c *TcpConnection) Close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		logger.Info("tcp conn close %s", c.raddr)
		_ = c.conn.Close()
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				logger.Error("tcp conn panic in close %v %v", c.conn, r)
			}
		}()
		c.onClose()
	}
}

func (c *TcpConnection) onClose() {
	c.handler.OnConnectionClose(c)
}

func (c *TcpConnection) SendData(data []byte) error {
	if data == nil {
		logger.Warn("tcp conn send fail 1")
		return errors.New("TcpConnection.SendData: fail 1")
	}
	if c == nil {
		logger.Warn("tcp conn send fail 2")
		return errors.New("TcpConnection.SendData: fail 2")
	}
	if c.writeCh == nil {
		logger.Warn("tcp conn send fail 3")
		return errors.New("TcpConnection.SendData: fail 3")
	}
	c.writeCh <- data
	return nil
}
