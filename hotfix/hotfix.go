package hotfix

import (
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"

	"github.com/agiledragon/gomonkey/v2"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "sf")

type FuncEntry struct {
	target interface{}
	double interface{}
}

func NewFuncEntry(target interface{}, double interface{}) *FuncEntry {
	return &FuncEntry{target: target, double: double}
}

type MethodEntry struct {
	target interface{}
	method string
	double interface{}
}

func NewMethodEntry(target interface{}, method string, double interface{}) *MethodEntry {
	return &MethodEntry{target: target, method: method, double: double}
}

func ApplyHotfix(entries []interface{}) {
	logger.Info("apply hotfix begin %d", len(entries))
	for _, entry := range entries {
		ApplyHotfixEntry(entry)
	}
	logger.Info("apply hotfix end")
}

func ApplyHotfixEntry(i interface{}) {
	switch i.(type) {
	case *MethodEntry:
		e := i.(*MethodEntry)
		_ = gomonkey.ApplyMethod(e.target, e.method, e.double)
	case *FuncEntry:
		e := i.(*FuncEntry)
		_ = gomonkey.ApplyFunc(e.target, e.double)
	default:
		logger.Error("invalid hotfix entry: %v", i)
	}
}
