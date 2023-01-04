package service_framework

import (
	"sync"
	"time"

	"github.com/vmihailenco/msgpack"

	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtutil"
)

type callbackHandler interface{
	Callback()
}

type serviceCallback struct {
	handler callbackHandler
	expire int64
}

var serviceCallbackMap sync.Map
var serviceCheckStopCh = make(chan bool)

func init() {
	go lbtutil.StartTickJob("checkServiceCallback", 17, serviceCheckStopCh, checkServiceCallback)
}

func CallServiceMethod(service, method string, params map[string]interface{}, handler callbackHandler, expire int64) {
	CallServiceMethodWithRoute(service, method, params, handler, expire, lbtproto.DefaultRouteType, []byte{})
}

func CallServiceMethodWithRoute(service, method string, params map[string]interface{}, handler callbackHandler, expire int64, routeType int32, routParam []byte) {
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
		Routet: routeType,
		Routep: routParam,
	}
	c := gateManager.getRandomGate()
	if c == nil {
		logger.Error("CallServiceMethod fail 2 no gate connection")
		return
	}
	if expire <= 0 { expire = 15 }
	// calback
	serviceCallbackMap.Store(reqid, serviceCallback{handler: handler, expire: time.Now().Unix() + expire})
	if err := lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_service_request, msg); err != nil {
		logger.Error("CallServiceMethod fail 3 %s", err.Error())
		serviceCallbackMap.Delete(reqid)
		return
	}
}

func processServiceReply(reqid lbtutil.ObjectID, replyByte []byte) {
	cbI, _ := serviceCallbackMap.LoadAndDelete(reqid)
	if cbI == nil { return }
	cb, _ := cbI.(serviceCallback)
	replyidStr := ""
	if err := msgpack.Unmarshal(replyByte, &[]interface{}{&replyidStr, cb.handler}); err != nil {
		logger.Error("processServiceReply fail 1 %v %v", replyByte, err)
		return
	}
	cb.handler.Callback()
}

func checkServiceCallback() {
	now := time.Now().Unix()
	serviceCallbackMap.Range(func(k, v interface{}) bool {
		if now > v.(serviceCallback).expire { serviceCallbackMap.Delete(k) }
		return true
	})
}
