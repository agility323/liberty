// +build !debug

package lbtutil

import (
	"runtime/debug"
)

func Recover(tag string, after func()) {
	if r := recover(); r != nil {
		log.Error("recover panic at %s [%v]", tag, r)
		debug.PrintStack()
		if after != nil { after() }
	}
}
