package lbtutil

import (
	"fmt"
	"testing"
	"time"
)

func TestLogWithTag(t *testing.T) {
	s := time.Now()
	BICoreLog("login", map[string]interface{}{"event":"login", "detail":[]int{1,2,3,4}})
	BICustomLog("login", map[string]interface{}{"event":"login", "detail":[]int{1,2,3,4}})
	BICustomLog("login", map[string]interface{}{"event":"login", "detail":[]int{1,2,3,4}})
	BICustomLog("login", map[string]interface{}{"event":"login", "detail":[]int{1,2,3,4}})
	fmt.Println("cost: ", time.Since(s))
	s = time.Now()
	logger1 := NewLogger("avatar", "building")
	logger1.Info("test sdfsdfwer")
}
