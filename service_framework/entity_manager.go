package service_framework

import (
	"sync"
	"errors"
	"reflect"
	"fmt"

	"github.com/agility323/liberty/lbtutil"
)

var entities sync.Map

func registerEntity(id lbtutil.ObjectId, e interface{}) {
	entities.Store(id, e)
}

func removeEntity(id lbtutil.ObjectId) {
	entities.Delete(id)
}

func GetEntity(id lbtutil.ObjectId) interface{} {
	if v, ok := entities.Load(id); ok {
		return v
	}
	return nil
}

func CallEntityMethod(id lbtutil.ObjectId, method string, params []interface{}) error {
	// entity
	entity := GetEntity(id)
	if entity == nil {
		return errors.New(fmt.Sprintf("CallEntityMethod failed entity not found %s", id.Hex()))
	}
	// rpc method
	v := reflect.ValueOf(entity)
	pec := v.Elem().FieldByName(EntityCoreFieldName).Addr().Interface().(*EntityCore)
	typ := pec.GetType()
	rpc, ok := entityRpcMap[typ][method]
	if !ok {
		return errors.New(fmt.Sprintf("CallEntityMethod failed method not found %s %s %s", typ, id.Hex(), method))
	}
	// check params
	if len(params) != len(rpc.pts) {
		return errors.New(fmt.Sprintf("CallEntityMethod failed params mismatch 1 %s %s %s %d!=%d",
			typ, id.Hex(), method, len(params), len(rpc.pts)))
	}
	args := make([]reflect.Value, 1, len(params) + 1)
	args[0] = v
	for i, param := range params {
		vp := reflect.ValueOf(param)
		if vp.Type() != rpc.pts[i] {
			return errors.New(fmt.Sprintf("CallEntityMethod failed params mismatch 2 %s %s %s %v!=%v",
				typ, id.Hex(), method, vp.Type(), rpc.pts[i]))
		}
		args = append(args, vp)
	}
	// call
	_ = rpc.m.Func.Call(args)
	return nil
}
