package service_framework

import (
	"fmt"
	"reflect"

	"github.com/agility323/liberty/lbtutil"
)

const EntityCoreFieldName = "EC"

type rpcInfo struct {
	m   reflect.Method
	pts []reflect.Type
}

var entityTypeMap map[string]reflect.Type = make(map[string]reflect.Type)
var entityRpcMap map[string]map[string]rpcInfo = make(map[string]map[string]rpcInfo)

func RegisterEntityType(name string, ptyp reflect.Type, rpcList []string) {
	typ := ptyp.Elem()
	if old, ok := entityTypeMap[name]; ok {
		panic(fmt.Sprintf("RegisterEntityType failed: duplicated name %s %s %s", name, old, typ))
	}
	if _, ok := typ.FieldByName(EntityCoreFieldName); !ok {
		panic(fmt.Sprintf("RegisterEntityType failed: missing EntityCore(%s) %s %s", EntityCoreFieldName, name, typ))
	}
	// record entity type
	entityTypeMap[name] = typ
	logger.Info("register entity type %s %s %v", name, typ.String(), rpcList)
	// generate rpc info
	entityRpcMap[name] = make(map[string]rpcInfo)
	for _, rpc := range rpcList {
		m, ok := ptyp.MethodByName(rpc)
		if !ok {
			logger.Warn("register entity rpc not found %s %s %s", name, typ.String(), rpc)
			continue
		}
		n := m.Type.NumIn()
		pts := make([]reflect.Type, 0, n - 1)
		for i := 1; i < n; i++ {
			pts = append(pts, m.Type.In(i))
		}
		entityRpcMap[name][rpc] = rpcInfo{m: m, pts: pts}
		logger.Info("register entity rpc %s %s %s %v", name, typ.String(), rpc, pts)
	}
}

func AppendEntityRpcMethod(name string, method string) bool {
	typ, ok := entityTypeMap[name]
	if !ok {
		logger.Error("append entity rpc fail no entity %s %s", name, method)
		return false
	}
	m, ok := reflect.PtrTo(typ).MethodByName(method)
	if !ok {
		logger.Error("append entity rpc fail no method %s %s", name, method)
		return false
	}
	n := m.Type.NumIn()
	pts := make([]reflect.Type, 0, n - 1)
	for i := 1; i < n; i++ {
		pts = append(pts, m.Type.In(i))
	}
	if _, ok = entityRpcMap[name]; !ok {
		entityRpcMap[name] = make(map[string]rpcInfo)
	}
	entityRpcMap[name][method] = rpcInfo{m: m, pts: pts}
	logger.Info("append entity rpc %s %s %s %v", name, typ.String(), method, pts)
	return true
}

func CreateEntity(name string, id lbtutil.ObjectID) interface{} {
	typ, ok := entityTypeMap[name]
	if !ok {
		panic(fmt.Sprintf("CreateEntity failed: %s not registered", name))
	}
	// instantiate
	ptr := reflect.New(typ)
	// init core
	ec := ptr.Elem().FieldByName(EntityCoreFieldName).Addr().Interface().(*EntityCore)
	ec.init(name, id)
	id = ec.GetId()
	// register
	e := ptr.Interface()
	addEntity(id, e)
	logger.Info("create entity %s %s", ec.GetType(), id.Hex())
	return e
}

func DestroyEntity(id lbtutil.ObjectID) {
	removeEntity(id)
	logger.Info("destroy entity %s", id.Hex())
}
