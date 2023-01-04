package service_framework

import (
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtreg"

	"github.com/vmihailenco/msgpack/v5"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "sf")
var Logger = logger

var serviceAddr string = ""
var regData *lbtreg.BasicRegData

func init() {
	msgpack.SetEnableLuaMapDecode(true)
	regData = &lbtreg.BasicRegData{
		Version: lbtutil.ReadVersionFile(),
	}
}
