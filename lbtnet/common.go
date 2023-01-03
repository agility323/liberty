package lbtnet

import (
	"os"
	"strconv"
	"encoding/binary"
	"net"
	"time"

	"github.com/agility323/liberty/lbtutil"
)

var logger lbtutil.Logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "lbtnet")

var byteOrder binary.ByteOrder = binary.LittleEndian

type ConnectionHandler interface {
	// connection only
	HandleProto(*TcpConnection, []byte) error
	OnConnectionClose(*TcpConnection)
	// client only
	OnConnectionReady(*TcpConnection)
	OnConnectionFail(*TcpClient)
}

type ProtoHandlerType func(*TcpConnection, []byte) error

type connectionCreatorFunc func(net.Conn)

const WriteWaitTime = 200 * time.Millisecond
