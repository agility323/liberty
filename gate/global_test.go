package main

import (
	"fmt"
	"testing"

	"github.com/agility323/liberty/lbtutil"
	//"github.com/agility323/liberty/lbtreg"
)

func init() {
}

func TestRegData(t *testing.T) {
	data := &GateRegData{
		EntranceAddr: Conf.EntranceAddr,
	}
	data.Version = lbtutil.ReadVersionFile()
	b, err := data.Marshal()
	fmt.Printf("TestRegData: %v %v %v\n", data, b, err)
}
