package service_framework

import (
	"sync"
	"errors"
	"reflect"
	"fmt"
	"bytes"
	"io"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtproto"

	"github.com/vmihailenco/msgpack/v5"
)

var entities sync.Map

func registerEntity(id lbtutil.ObjectID, e interface{}) {
	entities.Store(id, e)
}

func removeEntity(id lbtutil.ObjectID) {
	entities.Delete(id)
}

func GetEntity(id lbtutil.ObjectID) interface{} {
	if v, ok := entities.Load(id); ok {
		return v
	}
	return nil
}

func CallEntityMethodLocal(id lbtutil.ObjectID, method string, paramBytes []byte) error {
	// entity
	entity := GetEntity(id)
	if entity == nil {
		return errors.New(fmt.Sprintf("CallEntityMethodLocal fail: entity not found %s", id.Hex()))
	}
	// rpc method
	v := reflect.ValueOf(entity)
	pec := v.Elem().FieldByName(EntityCoreFieldName).Addr().Interface().(*EntityCore)
	pec.PushActorTask(func() {
		typ := pec.GetType()
		rpc, ok := entityRpcMap[typ][method]
		if !ok {
			logger.Error("CallEntityMethodLocal fail: method not found %s %s %s", typ, id.Hex(), method)
			return
		}
		// parameters
		params := make([]reflect.Value, 1, len(rpc.pts) + 1)
		params[0] = v
		for _, pt := range rpc.pts {
			ptrVal := reflect.New(pt)
			params = append(params, ptrVal.Elem())
		}
		rawArray := lbtutil.MsgpackRawArray(paramBytes)
		if !rawArray.Valid() {
			logger.Error("CallEntityMethodLocal fail: params is not array %v", paramBytes)
			return
		}
		decoder := msgpack.NewDecoder(bytes.NewBuffer(rawArray.Body()))
		for i := 1; i < len(params); i++ {
			err := decoder.DecodeValue(params[i])
			if err == io.EOF {
				logger.Warn(fmt.Sprintf("CallEntityMethodLocal insufficient params: %s %s %s %d %d",
					typ, id.Hex(), method, len(params) - 1, i - 1))
				break
			}
			if err != nil {
				logger.Error("CallEntityMethodLocal fail: msgpack decode fail [%v] %s %v %v",
					err, method, paramBytes, rawArray.Body())
				return
			}
		}
		// call
		_ = rpc.m.Func.Call(params)
	})
	return nil

	/*
	typ := pec.GetType()
	rpc, ok := entityRpcMap[typ][method]
	if !ok {
		return fmt.Errorf("CallEntityMethodLocal fail: method not found %s %s %s", typ, id.Hex(), method)
	}
	// parameters
	params := make([]reflect.Value, 1, len(rpc.pts) + 1)
	params[0] = v
	for _, pt := range rpc.pts {
		ptrVal := reflect.New(pt)
		params = append(params, ptrVal.Elem())
	}
	rawArray := lbtutil.MsgpackRawArray(paramBytes)
	if !rawArray.Valid() {
		return fmt.Errorf("CallEntityMethodLocal fail: params is not array %v", paramBytes)
	}
	decoder := msgpack.NewDecoder(bytes.NewBuffer(rawArray.Body()))
	for i := 1; i < len(params); i++ {
		err := decoder.DecodeValue(params[i])
		if err == io.EOF {
			logger.Warn(fmt.Sprintf("CallEntityMethodLocal insufficient params: %s %s %s %d %d",
				typ, id.Hex(), method, len(params) - 1, i - 1))
			break
		}
		if err != nil {
			return errors.New(fmt.Sprintf("CallEntityMethodLocal fail: msgpack decode fail [%v] %s %v %v",
				err, method, paramBytes, rawArray.Body()))
		}
	}
	// call
	_ = rpc.m.Func.Call(params)
	return nil
	*/
}

func CallEntityMethod(addr string, id lbtutil.ObjectID, method string, params interface{}) {
	b, err := msgpack.Marshal(&params)
	if err != nil {
		logger.Error("CallEntityMethod failed 1 %s", err.Error())
		return
	}
	//logger.Debug("CallEntityMethod %s %s %s %v", addr, id.Hex(), method, params)
	msg := &lbtproto.EntityMsg{
		Addr: addr,
		Id: id[:],
		Method: method,
		Params: b,
	}
	postGateManagerJob("entity_msg", msg)
}
