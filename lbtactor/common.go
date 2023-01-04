package lbtactor

import (
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"
)

var logger lbtutil.Logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "lbtactor")
