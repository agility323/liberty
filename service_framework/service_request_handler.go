package service_framework

import (
	"sync"
	"time"

	"github.com/vmihailenco/msgpack"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtproto"
)

type serviceCallbackFunc func(map[string]interface{})

type serviceCallback struct {
	f serviceCallbackFunc
	expire int64
}

var serviceCallbackMap sync.Map
var serviceCheckStopCh = make(chan bool)

func init() {
	go lbtutil.StartTickJob("checkServiceCallback", 17, serviceCheckStopCh, checkServiceCallback)
}

func CallServiceMethod(service, method string, params map[string]interface{}, cbf serviceCallbackFunc, expire int64) {
	b, err := msgpack.Marshal(&params)
	if err != nil {
		logger.Error("CallServiceMethod fail 1 %v", err)
		return
	}
	reqid := string(lbtutil.NewObjectId())
	msg := &lbtproto.ServiceRequest{
		Addr: serviceAddr,
		Reqid: reqid,
		Type: service,
		Method: method,
		Params: b,
	}
	postGateManagerJob("service_request", msg)
	if expire <= 0 { expire = 15 }
	serviceCallbackMap.Store(reqid, serviceCallback{f: cbf, expire: time.Now().Unix() + expire})
}

func processServiceReply(reqid string, reply []byte) {
	cb, _ := serviceCallbackMap.LoadAndDelete(reqid)
	if cb == nil { return }
	var args map[string]interface{}
	if err := msgpack.Unmarshal(reply, &args); err != nil {
		logger.Error("processServiceReply fail 1 %v %v", reply, err)
		return
	}
	cb.(serviceCallback).f(args)
}

func checkServiceCallback() {
	now := time.Now().Unix()
	serviceCallbackMap.Range(func(k, v interface{}) bool {
		if now > v.(serviceCallback).expire { serviceCallbackMap.Delete(k) }
		return true
	})
}
