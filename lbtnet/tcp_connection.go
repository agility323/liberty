package lbtnet

import (
	"fmt"
	"net"
	"sync/atomic"
	"encoding/binary"
	"io"
	"bytes"
	"time"

	"github.com/agility323/liberty/lbtutil"
)

const (
	SizeLen uint32 = 4
)

var (
	MaxMsgLenOnRead uint32 = 500000
	MaxMsgLenOnWrite uint32 = 450000
	DefaultWriteChLen int = 2000
	DefaultWriteChWaitTime time.Duration = 5	// time in second
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
	conf ConnectionConfig
	errlog func(format string, params ...interface{})
}

func NewTcpConnection(conn net.Conn, handler ConnectionHandler, conf ConnectionConfig) *TcpConnection {
	laddrStr := ""
	raddrStr := ""
	if laddr := conn.LocalAddr(); laddr != nil { laddrStr = laddr.String() }
	if raddr := conn.RemoteAddr(); raddr != nil { raddrStr = raddr.String() }
	bufsize := SizeLen + MaxMsgLenOnRead
	if conf.WriteChLen < 0 {
		conf.WriteChLen = 0
	}
	c := &TcpConnection{
		started: 0,
		laddr: laddrStr,
		raddr: raddrStr,
		buf: make([]byte, bufsize, bufsize),
		conn: conn,
		r: conn,
		w: conn,
		handler: handler,
		writeCh: make(chan []byte, conf.WriteChLen),
		stopCh: make(chan struct{}, 1),
		vars: make(map[string]interface{}),
		conf: conf,
	}
	c.readLoopFunc = c.readLoopOnce
	if conf.ErrLog {
		c.errlog = logger.Error
	} else {
		c.errlog = logger.Warn
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
	defer lbtutil.Recover(fmt.Sprintf("TcpConnection.readLoop %v", c.conn), c.Close)

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
		c.errlog("tcp conn %s read head fail %v", c.raddr, err)
		c.Close()
		return true
	}
	err = binary.Read(bytes.NewReader(bufHead), byteOrder, &bodyLen)
	if err != nil || bodyLen == 0 || bodyLen > uint32(len(c.buf)) - SizeLen {
		c.errlog("tcp conn %s invalid body len %d %v", c.raddr, bodyLen, err)
		c.Close()
		return true
	}
	// read body
	_, err = io.ReadFull(c.r, c.buf[SizeLen:SizeLen + bodyLen])
	if err != nil {
		c.errlog("tcp conn %s read body fail %d %v", c.raddr, bodyLen, err)
		c.Close()
		return true
	}
	// process proto
	data := make([]byte, SizeLen + bodyLen, SizeLen + bodyLen)
	copy(data, c.buf)
	//logger.Debug("tcp conn read %s %v", c.raddr, data)
	err = c.handler.HandleProto(c, data)
	if err != nil {
		c.errlog("tcp conn %s proto fail %v", c.raddr, err)
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
				c.errlog("tcp conn %s write fail %d %d %v", c.raddr, len(data), n, err)
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
		defer lbtutil.Recover(fmt.Sprintf("TcpConnection.CloseWithoutCallback %v", c.conn), nil)
		return true
	}
	return false
}

func (c *TcpConnection) SendData(data []byte) error {
	if c == nil {
		logger.Error("tcp conn send fail invalid connection")
		return ErrSendInvalidConnection
	}
	if data == nil {
		logger.Error("tcp conn send fail invalid data")
		return ErrSendInvalidData
	}
	if uint32(len(data)) > MaxMsgLenOnWrite {
		logger.Error("tcp conn send fail long data %d", len(data))
		return ErrSendLongData
	}
	if c.writeCh == nil {
		logger.Error("tcp conn send fail invalid chan")
		return ErrSendInvalidChan
	}
	// block for limited seconds
	select {
	case c.writeCh<- data:
		return nil
	default:
		if c.conf.WriteChWaitTime == 0 {
			logger.Error("tcp conn send fail chan full %s %s", c.laddr, c.raddr)
			c.Close()
			return ErrSendChanFull
		} else {
			ts := time.Now().UnixMilli()
			t := time.NewTimer(c.conf.WriteChWaitTime * time.Second)
			select {
			case c.writeCh<- data:
				if !t.Stop() {
					// do nothing
				}
				logger.Error("tcp conn send block for %d ms %s %s", time.Now().UnixMilli() - ts, c.laddr, c.raddr)
				return nil
			case <-t.C:
				logger.Error("tcp conn send fail chan full %s %s", c.laddr, c.raddr)
				c.Close()
				return ErrSendChanFull
			}
		}
	}
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
			logger.Warn("EnableEncryptAndCompress fail 4 %s %v", c.raddr, err)
			c.Close()
			return true
		}
		c.r = cr
		return c.readLoopOnce()
	}
	return nil
}

func (c *TcpConnection) OnHeartbeat(t int64) {
	c.handler.OnHeartbeat(c, t)
}
