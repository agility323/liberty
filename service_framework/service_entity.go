package service_framework

import (
	"github.com/agility323/liberty/lbtutil"
	"reflect"
)

var ServiceInstance *ServiceEntity

func init() {
	rpclist := []string{"CMD_hashed_task"}
	RegisterEntityType("ServiceEntity", reflect.TypeOf((*ServiceEntity)(nil)), rpclist)
	ServiceInstance = CreateEntity("ServiceEntity", lbtutil.NilObjectID).(*ServiceEntity)
}

type ServiceEntity struct {
	EC EntityCore
	fmap map[string]func(...interface{})
}

func (e *ServiceEntity) Init(qlen int, hsize int, hqlen int, fmap map[string]func(...interface{})) {
	e.EC.InitMainWorker(qlen)
	e.EC.initHashedWorker(hsize, hqlen)
	e.fmap = fmap
}

func (e *ServiceEntity) Start() {
	e.EC.StartWorker()
}

func (e *ServiceEntity) Stop() {
	e.EC.StopWorker()
}

func (e *ServiceEntity) CMD_hashed_task(method string, params ...interface{}) {
	if f, ok := e.fmap[method]; ok { f(params...) }
}

