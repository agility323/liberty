package service_framework

import (
	"sync"
	"errors"
	"reflect"
	"fmt"
	"bytes"
	"io"

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
		return errors.New(fmt.Sprintf("CallEntityMethod fail: entity not found %s", id.Hex()))
	}
	// rpc method
	v := reflect.ValueOf(entity)
	pec := v.Elem().FieldByName(EntityCoreFieldName).Addr().Interface().(*EntityCore)
	typ := pec.GetType()
	rpc, ok := entityRpcMap[typ][method]
	if !ok {
		return errors.New(fmt.Sprintf("CallEntityMethod fail: method not found %s %s %s", typ, id.Hex(), method))
	}
	// parameters
	params := make([]reflect.Value, 1, len(rpc.pts) + 1)
	params[0] = v
	ptrs := make([]interface{}, 0, len(rpc.pts))
	for _, pt := range rpc.pts {
		ptrVal := reflect.New(pt)
		params = append(params, ptrVal.Elem())
		ptrs = append(ptrs, ptrVal.Interface())
	}
	rawArray := lbtutil.MsgpackRawArray(paramBytes)
	decoder := msgpack.NewDecoder(bytes.NewBuffer(rawArray.Body()))
	for i, ptr := range ptrs {
		err := decoder.Decode(ptr)
		if err == io.EOF {
			logger.Warn(fmt.Sprintf("CallEntityMethod insufficient params: %s %s %s %d %d",
				typ, id.Hex(), method, len(ptrs), i))
			break
		}
		if err != nil {
			return errors.New(fmt.Sprintf("CallEntityMethod fail: msgpack decode fail [%v] %s %v %v",
				err, method, paramBytes, rawArray.Body()))
		}
	}
	// call
	_ = rpc.m.Func.Call(params)
	return nil
}
