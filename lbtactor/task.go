package lbtactor

import (

	"github.com/agility323/liberty/lbtutil"
)

func RunTask(name string, task func()) {
	go func() {
		defer lbtutil.Recover("RunTask." + name, nil)
		task()
	}()
}
