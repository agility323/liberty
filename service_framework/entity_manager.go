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

var entmgr *EntityManager

func init() {
	entmgr = &EntityManager{
	}
	for slot := range entmgr.entitySlots {
		entmgr.entitySlots[slot] = make(map[lbtutil.ObjectID]interface{})
	}
}

const EntitySlotNum = 64

type EntityManager struct {
	locks [EntitySlotNum]sync.RWMutex
	entitySlots [EntitySlotNum]map[lbtutil.ObjectID]interface{}
}

func addEntity(id lbtutil.ObjectID, e interface{}) {
	slot := lbtutil.StringHash(id.Hex()) % EntitySlotNum
	entmgr.locks[slot].Lock()
	defer entmgr.locks[slot].Unlock()

	entmgr.entitySlots[slot][id] = e
}

func removeEntity(id lbtutil.ObjectID) {
	slot := lbtutil.StringHash(id.Hex()) % EntitySlotNum
	entmgr.locks[slot].Lock()
	defer entmgr.locks[slot].Unlock()

	delete(entmgr.entitySlots[slot], id)
}

func GetEntity(id lbtutil.ObjectID) interface{} {
	slot := lbtutil.StringHash(id.Hex()) % EntitySlotNum
	entmgr.locks[slot].RLock()
	defer entmgr.locks[slot].RUnlock()

	return entmgr.entitySlots[slot][id]
}

func (m *EntityManager) onStart() {
	tickmgr.AddTickJob(m.OnTick)
}

func (m *EntityManager) OnTick() {
	n := 0
	for slot := range m.entitySlots {
		m.locks[slot].RLock()
		n += len(m.entitySlots[slot])
		m.locks[slot].RUnlock()
	}
	logger.Info("entity manager tick %d", n)
}

func CallEntityMethodLocal(id lbtutil.ObjectID, method string, paramBytes []byte, hval int) error {
	// entity
	entity := GetEntity(id)
	if entity == nil {
		return errors.New(fmt.Sprintf("CallEntityMethodLocal fail: entity not found %s %s", id.Hex(), method))
	}
	// rpc method
	v := reflect.ValueOf(entity)
	pec := v.Elem().FieldByName(EntityCoreFieldName).Addr().Interface().(*EntityCore)
	task := func() {
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
			logger.Error("CallEntityMethodLocal fail: params is not array %v %v", method, paramBytes)
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
	}
	if hval > 0 {
		pec.PushHashedTask(task, hval - 1)
	} else {
		pec.PushMainTask(task)
	}
	return nil
}

func CallEntityMethod(addr string, id lbtutil.ObjectID, method string, params interface{}) error {
	return CallHashedEntityMethod(addr, id, method, params, -1)
}

func CallHashedEntityMethod(addr string, id lbtutil.ObjectID, method string, params interface{}, hval int) error {
	b, err := msgpack.Marshal(&params)
	if err != nil {
		logger.Error("CallEntityMethod fail 1 %s", err.Error())
		return ErrRpcInvalidParams
	}
	//logger.Debug("CallEntityMethod %s %s %s %v", addr, id.Hex(), method, params)
	msg := &lbtproto.EntityMsg{
		Addr: addr,
		Id: id[:],
		Method: method,
		Params: b,
		Hval: int32(hval),
	}
	c := gateManager.getPrimaryGate()
	if c == nil {
		logger.Error("CallEntityMethod fail 2 no gate connection")
		return ErrRpcNoRoute
	}
	if err = lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_entity_msg, msg); err != nil {
		logger.Error("CallEntityMethod fail 3 %v", err)
		return err
	}
	return nil
}
