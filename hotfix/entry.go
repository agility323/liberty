package hotfix

import (
	"github.com/agility323/liberty/hotfix/itf"

	"github.com/agiledragon/gomonkey/v2"
)

type FuncEntry struct {
	target interface{}
	double interface{}
}

func (h *HotfixImpl) NewFuncEntry(target interface{}, double interface{}) itf.HotfixEntry {
	return &FuncEntry{target: target, double: double}
}

func (e *FuncEntry) Apply() {
	gomonkey.ApplyFunc(e.target, e.double)
}

type MethodEntry struct {
	target interface{}
	method string
	double interface{}
}

func (h *HotfixImpl) NewMethodEntry(target interface{}, method string, double interface{}) itf.HotfixEntry {
	return &MethodEntry{target: target, method: method, double: double}
}

func (e *MethodEntry) Apply() {
	gomonkey.ApplyMethod(e.target, e.method, e.double)
}

type MethodFuncEntry struct {
	target interface{}
	method string
	double interface{}
}

func (h *HotfixImpl) NewMethodFuncEntry(target interface{}, method string, double interface{}) itf.HotfixEntry {
	return &MethodFuncEntry{target: target, method: method, double: double}
}

func (e *MethodFuncEntry) Apply() {
	gomonkey.ApplyMethodFunc(e.target, e.method, e.double)
}
