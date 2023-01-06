package hotfix

import (
	"reflect"

	"github.com/agility323/liberty/hotfix/lookup"
)

type FuncEntry struct {
	target interface{}
	name string
	double interface{}
}

func (e *FuncEntry) Apply() {
	realt, err := lookup.MakeValueByFunctionName(e.target, e.name)
	if err != nil {
		logger.Error("hotfix.FuncEntry.Apply fail 1: %v", err)
		return
	}
	Hotfix.patches.ApplyCore(realt, reflect.ValueOf(e.double))	// patch main
	Hotfix.patches.ApplyCore(reflect.ValueOf(e.target), reflect.ValueOf(e.double))	// patch so
}
