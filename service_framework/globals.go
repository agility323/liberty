package service_framework

import (
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "sf")
var Logger = logger
