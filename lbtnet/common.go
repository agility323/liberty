package lbtnet

import (
	"os"
	"strconv"
	"encoding/binary"
	"net"

	"github.com/agility323/liberty/lbtutil"
)

var logger lbtutil.Logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "lbtnet")

var byteOrder binary.ByteOrder = binary.LittleEndian

type Connection interface {
	Addr() string
	Start()
	Close()
	SendData([]byte) error
}

type ConnectionHandler interface {
	HandleProto(*TcpConnection, []byte) error
	OnConnectionReady(*TcpConnection)
	OnConnectionClose(*TcpConnection)
}

type ProtoHandlerType func(*TcpConnection, []byte) error

type connectionCreatorFunc func(net.Conn)
