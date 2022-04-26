package service_framework

import (
	"sync"
	"errors"
	"reflect"
	"fmt"

	"github.com/agility323/liberty/lbtutil"

	"github.com/vmihailenco/msgpack"
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

func CallEntityMethod(id lbtutil.ObjectId, method string, paramBytes []byte) error {
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
	// parameters
	params := make([]interface{}, 0, len(rpc.pts))
	for _, pt := range rpc.pts {
		params = append(params, reflect.New(pt).Interface())
	}
	if err := msgpack.Unmarshal(paramBytes, &params); err != nil {
		return err
	}
	// call
	args := make([]reflect.Value, 1, len(params) + 1)
	args[0] = v
	for _, param := range params {
		args = append(args, reflect.ValueOf(param).Elem())
	}
	_ = rpc.m.Func.Call(args)
	return nil
}
