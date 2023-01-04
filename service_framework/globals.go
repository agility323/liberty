package service_framework

import (
	"os"
	"strconv"
	"time"

	"github.com/agility323/liberty/lbtutil"

	"github.com/vmihailenco/msgpack/v5"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "sf")
var Logger = logger

var serviceAddr string = ""

var serviceRequestTimeout = 20 * time.Second

func init() {
	msgpack.SetEnableLuaMapDecode(true)
}
