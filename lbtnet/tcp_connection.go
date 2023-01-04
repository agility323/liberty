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
	r io.Reader
	w io.Writer
	handler ConnectionHandler
	writeCh chan []byte
	stopCh chan struct{}
	vars map[string]interface{}	// customed variables
	readLoopFunc func() bool
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
		r: conn,
		w: conn,
		handler: handler,
		writeCh: make(chan []byte, WriteChLen),
		stopCh: make(chan struct{}, 1),
		vars: make(map[string]interface{}),
	}
	c.readLoopFunc = c.readLoopOnce
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
	for {
		if c.readLoopFunc() { return }
	}
}

func (c *TcpConnection) readLoopOnce() bool {
	var bodyLen uint32
	bufHead := c.buf[:SizeLen]
	// read head
	_, err := io.ReadFull(c.r, bufHead)
	if err != nil {
		logger.Warn("tcp conn %s read head fail %v", c.raddr, err)
		c.Close()
		return true
	}
	err = binary.Read(bytes.NewReader(bufHead), byteOrder, &bodyLen)
	if err != nil || bodyLen == 0 || bodyLen > MaxMsgLen {
		logger.Warn("tcp conn %s invalid body len %d %v", c.raddr, bodyLen, err)
		c.Close()
		return true
	}
	// read body
	_, err = io.ReadFull(c.r, c.buf[SizeLen:SizeLen + bodyLen])
	if err != nil {
		logger.Warn("tcp conn %s read body fail %d %v", c.raddr, bodyLen, err)
		c.Close()
		return true
	}
	// process proto
	data := make([]byte, SizeLen + bodyLen, SizeLen + bodyLen)
	copy(data, c.buf)
	//logger.Debug("tcp conn read %s %v", c.raddr, data)
	err = c.handler.HandleProto(c, data)
	if err != nil {
		logger.Warn("tcp conn %s proto fail %v", c.raddr, err)
	}
	return false
}

func (c *TcpConnection) writeLoop() {
	for {
		select {
		case <-c.stopCh:
			logger.Info("tcp conn %s write loop quit", c.raddr)
			return
		case data := <-c.writeCh:
			//logger.Debug("tcp conn write %s %v", c.raddr, data)
			n, err := c.w.Write(data)
			if err != nil {
				logger.Warn("tcp conn %s write fail %d %d %s", c.raddr, len(data), n, err.Error())
				c.Close()
				return
			}
		}

	}
}

func (c *TcpConnection) Close() {
	if c.CloseWithoutCallback() {
		c.handler.OnConnectionClose(c)
		select {
		case c.stopCh<- struct{}{}:
			return
		default:
			return
		}
	}
}

func (c *TcpConnection) CloseWithoutCallback() bool {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		logger.Info("tcp conn close %s", c.raddr)
		_ = c.conn.Close()
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				logger.Error("tcp conn panic in close %v %v", c.conn, r)
			}
		}()
		return true
	}
	return false
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
	select {
	case c.writeCh<- data:
	default:
		return errors.New("TcpConnection.SendData: fail 4")
	}
	return nil
}

func (c *TcpConnection) SetVar(k string, v interface{}) {
	c.vars[k] = v
}

func (c *TcpConnection) GetVar(k string) interface{} {
	return c.vars[k]
}

// This function is called from readLoop.
// After called, data read/write from connection is encrypted and compressed.
// The encrypt/compress order of read/write corresponds with the other side.
func (c *TcpConnection) EnableEncryptAndCompress(key []byte) error {
	// reader
	er, err := NewEncryptReader(c.r, key)
	if err != nil {
		logger.Error("EnableEncryptAndCompress fail 1 %s %v", c.raddr, err)
		c.Close()
		return err
	}
	c.r = er
	// writer
	ew, err := NewEncryptWriter(c.w, key)
	if err != nil {
		logger.Error("EnableEncryptAndCompress fail 2 %s %v", c.raddr, err)
		c.Close()
		return err
	}
	c.w = ew
	c.w = NewCompressWriter(c.w)
	// replace readLoopFunc to enable read compress
	c.readLoopFunc = func() bool {
		defer func() { c.readLoopFunc = c.readLoopOnce }()
		cr, err := NewCompressReader(c.r)
		if err != nil {
			logger.Error("EnableEncryptAndCompress fail 4 %s %v", c.raddr, err)
			c.Close()
			return true
		}
		c.r = cr
		return c.readLoopOnce()
	}
	return nil
}
