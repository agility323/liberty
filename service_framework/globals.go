package service_framework

import (
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"

	"github.com/vmihailenco/msgpack/v5"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "sf")
var Logger = logger

var serviceAddr string = ""

func init() {
	msgpack.SetEnableLuaMapDecode(true)
}
