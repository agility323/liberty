// +build !debug

package lbtutil

import (
	"runtime/debug"
)

func Recover(tag string, after func()) {
	if r := recover(); r != nil {
		log.Error("panic in %s [%v]", tag, r)
		debug.PrintStack()
		if after != nil { after() }
	}
}
