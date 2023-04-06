package service_framework

import (
	"gitlab-gz.funplus.io/liberty/liberty/lbtutil"
)

var ServiceInstance *ServiceEntity

func init() {
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

