package main

import (
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtreg"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "gate")
var regData *lbtreg.BasicRegData

func init() {
	regData = &lbtreg.BasicRegData{
		Version: lbtutil.ReadVersionFile(),
	}
}
