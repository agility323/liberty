package lbtactor

import (

	"github.com/agility323/liberty/lbtutil"
)

func RunTaskActor(name string, task func()) {
	go func() {
		defer lbtutil.Recover("RunTaskActor." + name, nil)
		task()
	}()
}
