package main

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtreg"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "gate")

type GateRegData struct {
	lbtreg.BasicRegData
	EntranceAddr string
}

var regData *GateRegData

func (d *GateRegData) Marshal() (string, error) {
	b, err := json.Marshal(d)
	return string(b), err
}

func InitRegData() {
	regData = &GateRegData{
		EntranceAddr: Conf.EntranceAddr,
	}
	regData.Version = lbtutil.ReadVersionFile()
}
