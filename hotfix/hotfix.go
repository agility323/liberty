package hotfix

import (
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"

	"github.com/agiledragon/gomonkey/v2"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "sf")

type HotfixInterface interface {
	NewFuncEntry(interface{}, interface{}) interface{}
	NewMethodEntry(interface{}, string, interface{}) interface{}
	ApplyHotfix([]interface{})
	ApplyHotfixEntry(interface{})
}

type FuncEntry struct {
	target interface{}
	double interface{}
}

type MethodEntry struct {
	target interface{}
	method string
	double interface{}
}

type HotfixImpl struct {}

var Hotfix *HotfixImpl

func (h *HotfixImpl) NewFuncEntry(target interface{}, double interface{}) interface{} {
	return &FuncEntry{target: target, double: double}
}

func (h *HotfixImpl) NewMethodEntry(target interface{}, method string, double interface{}) interface{} {
	return &MethodEntry{target: target, method: method, double: double}
}

func (h *HotfixImpl) ApplyHotfix(entries []interface{}) {
	logger.Info("apply hotfix begin %d", len(entries))
	for _, entry := range entries {
		h.ApplyHotfixEntry(entry)
	}
	logger.Info("apply hotfix end")
}

func (h *HotfixImpl) ApplyHotfixEntry(i interface{}) {
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
