package service_framework

import (
	"sync"
	"time"

	"github.com/vmihailenco/msgpack"

	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtutil"
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
	// marshal
	b, err := msgpack.Marshal(&params)
	if err != nil {
		logger.Error("CallServiceMethod fail 1 %v", err)
		return
	}
	// proto msg
	reqid := lbtutil.NewObjectID()
	msg := &lbtproto.ServiceRequest{
		Addr: serviceAddr,
		Reqid: reqid[:],
		Type: service,
		Method: method,
		Params: b,
	}
	c := gateManager.getRandomGate()
	if c == nil {
		logger.Error("CallServiceMethod fail 2 no gate connection")
		return
	}
	if err := lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_service_request, msg); err != nil {
		logger.Error("CallServiceMethod fail 3 %s", err.Error())
		return
	}
	// calback
	if expire <= 0 { expire = 15 }
	serviceCallbackMap.Store(reqid, serviceCallback{f: cbf, expire: time.Now().Unix() + expire})
}

func processServiceReply(reqid lbtutil.ObjectID, reply []byte) {
	cb, _ := serviceCallbackMap.LoadAndDelete(reqid)
	if cb == nil { return }
	var args []interface{}
	if err := msgpack.Unmarshal(reply, &args); err != nil {
		logger.Error("processServiceReply fail 1 %v %v", reply, err)
		return
	}
	params, _ := args[1].(map[string]interface{})
	cb.(serviceCallback).f(params)
}

func checkServiceCallback() {
	now := time.Now().Unix()
	serviceCallbackMap.Range(func(k, v interface{}) bool {
		if now > v.(serviceCallback).expire { serviceCallbackMap.Delete(k) }
		return true
	})
}
