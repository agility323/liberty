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

type ServiceMethodCaller struct {
	service string
	method string
	params map[string]interface{}
	handler callbackHandler
	expire int64

	routet int32
	routep []byte
	hval int32
}

func NewServiceMethodCaller(service, method string, params map[string]interface{}, handler callbackHandler, expire int64) *ServiceMethodCaller {
	if expire <= 0 { expire = 15 }
	return &ServiceMethodCaller{
		service: service,
		method: method,
		params: params,
		handler: handler,
		expire: expire,
	}
}

func (caller *ServiceMethodCaller) SetRoute(routet int32, routep []byte) {
	caller.routet = routet
	caller.routep = routep
}

func (caller *ServiceMethodCaller) SetHval(hval int) {
	caller.hval = int32(hval)
}

func (caller *ServiceMethodCaller) Call() error {
	// marshal
	b, err := msgpack.Marshal(&caller.params)
	if err != nil {
		logger.Error("service method call fail msgpack marshal %v", err)
		return ErrRpcInvalidParams
	}
	// proto msg
	reqid := lbtutil.NewObjectID()
	msg := &lbtproto.ServiceRequest{
		Addr: serviceAddr,
		Reqid: reqid[:],
		Type: caller.service,
		Method: caller.method,
		Params: b,
	}
	if caller.routet > 0 {
		msg.Routet = caller.routet
		msg.Routep = caller.routep
	}
	if caller.hval > 0 {
		msg.Hval = caller.hval
	}
	gate := gateManager.getRandomGate()
	if gate == nil {
		logger.Error("service method call fail no gate connection")
		return ErrRpcNoRoute
	}
	// calback
	serviceCallbackMap.Store(reqid, serviceCallback{handler: caller.handler, expire: time.Now().Unix() + caller.expire})
	if err := lbtproto.SendMessage(gate, lbtproto.ServiceGate.Method_service_request, msg); err != nil {
		logger.Error("service method fail send %v", err)
		serviceCallbackMap.Delete(reqid)
		return err
	}
	return nil
}

func CallServiceMethod(service, method string, params map[string]interface{}, handler callbackHandler, expire int64) error {
	return NewServiceMethodCaller(service, method, params, handler, expire).Call()
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
