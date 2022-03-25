package service_framework

import (
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "sf")
var Logger = logger

var gateClient *lbtnet.TcpClient
