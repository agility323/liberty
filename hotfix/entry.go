package hotfix

import (
	"github.com/agiledragon/gomonkey/v2"
)

type FuncEntry struct {
	target interface{}
	double interface{}
}

func (e *FuncEntry) Apply() {
	gomonkey.ApplyFunc(e.target, e.double)
}

type MethodEntry struct {
	target interface{}
	method string
	double interface{}
}

func (e *MethodEntry) Apply() {
	gomonkey.ApplyMethod(e.target, e.method, e.double)
}
