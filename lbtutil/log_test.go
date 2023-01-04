package lbtutil

import (
	"testing"
)

func TestLogWithTag(t *testing.T) {
	BICoreLog("login", map[string]interface{}{"event":"login", "detail":[]int{1,2,3,4}})
	BICustomLog("login", map[string]interface{}{"event":"login", "detail":[]int{1,2,3,4}})
	logger1 := NewLogger("avatar", "building")
	logger1.Info("test sdfsdfwer")
}
