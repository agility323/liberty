package service_framework

import (
	"errors"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"

	"github.com/vmihailenco/msgpack"
)

/********** ProtoHandler **********/
var ServiceMethodHandler map[uint16]lbtnet.ProtoHandlerType = make(map[uint16]lbtnet.ProtoHandlerType)

func init() {
	initServiceMethodHandler()
}

func initServiceMethodHandler() {
	ServiceMethodHandler[lbtproto.Service.Method_register_reply] = Service_register_reply
	ServiceMethodHandler[lbtproto.Service.Method_client_disconnect] = Service_client_disconnect
	ServiceMethodHandler[lbtproto.Service.Method_service_request] = Service_service_request
	ServiceMethodHandler[lbtproto.Service.Method_entity_msg] = Service_entity_msg
	ServiceMethodHandler[lbtproto.Service.Method_service_shutdown] = Service_service_shutdown
}

func processGateProto(c *lbtnet.TcpConnection, buf []byte) error {
	methodIndex, err := lbtproto.DecodeMethodIndex(buf)
	if err != nil {
		logger.Error("gate proto fail read index %s", err.Error())
		return errors.New("read index")
	}
	f, ok := ServiceMethodHandler[methodIndex]
	if !ok {
		logger.Error("gate proto fail wrong index %d", methodIndex)
		return errors.New("wrong index")
	}
	return f(c, buf)
}

func Service_register_reply(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv Service_register_reply %v", buf)
	return nil
}

func Service_client_disconnect(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv Service_client_disconnect %v", buf)
	info := &lbtproto.BindClientInfo{}
	if err := lbtproto.DecodeMessage(buf, info); err != nil {
		return err
	}
	if cb, ok := clientCallbackMap[info.Caddr]; ok { cb.OnClientDisconnect() }
	return nil
}

func Service_service_request(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv Service_service_request %v", buf)
	request := &lbtproto.ServiceRequest{}
	if err := lbtproto.DecodeMessage(buf, request); err != nil {
		return err
	}
	replyData, err := processMethod(c.LocalAddr(), request.Addr, request.Method, request.Params)
	if err != nil {
		return err
	}
	if replyData == nil { return nil }
	SendServiceReply(request.Addr, request.Reqid, replyData)
	return nil
}

func Service_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv Service_entity_msg %v", buf)
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	var params []interface{}	// only support array
	if err := msgpack.Unmarshal(msg.Params, &params); err != nil {
		return err
	}
	if err := CallEntityMethod(lbtutil.ObjectId(msg.Id), msg.Method, params); err != nil {
		return err
	}
	return nil
}

func Service_service_shutdown(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv Service_service_shutdown %v", buf)
	Stop()
	return nil
}
/********** ProtoHandler End **********/

/********** ProtoSender **********/
func sendRegisterService(c *lbtnet.TcpConnection) {
	msg := &lbtproto.ServiceInfo{
		Addr: c.LocalAddr(),
		Type: serviceConf.serviceType,
		Entityid: "",
	}
	lbtproto.SendMessage(
		c,
		lbtproto.ServiceGate.Method_register_service,
		msg,
	)
}

func SendBindClient(saddr, caddr string) {
	msg := &lbtproto.BindClientInfo{
		Caddr: caddr,
		Saddr: saddr,
	}
	lbtproto.SendMessage(
		gateClient,
		lbtproto.ServiceGate.Method_bind_client,
		msg,
	)
}

func SendServiceReply(addr, reqid string, data []byte) {
	logger.Debug("SendServiceReply %s %s %v", addr, lbtutil.ObjectId(reqid).Hex(), data)
	reply := &lbtproto.ServiceReply{
		Addr: addr,
		Reqid: reqid,
		Reply: data,
	}
	err := lbtproto.SendMessage(
		gateClient,
		lbtproto.ServiceGate.Method_service_reply,
		reply,
	)
	if err != nil {
		logger.Error("SendServiceReply failed: SendMessage - %s", err.Error())
	}
}

func SendCreateEntity(addr, id, typ string, data interface{}) {
	b, err := msgpack.Marshal(&data)
	if err != nil {
		logger.Error("SendCreateEntity failed: marshal data - %s", err.Error())
		return
	}
	logger.Debug("SendCreateEntity %s %s %s %v %v", addr, lbtutil.ObjectId(id).Hex(), typ, data, b)
	msg := &lbtproto.EntityData{
		Addr: addr,
		Id: id,
		Type: typ,
		Data: b,
	}
	err = lbtproto.SendMessage(
		gateClient,
		lbtproto.ServiceGate.Method_create_entity,
		msg,
	)
	if err != nil {
		logger.Error("SendCreateEntity failed: SendMessage - %s", err.Error())
	}
}

func SendEntityMsg(addr, id, method string, params interface{}) {
	b, err := msgpack.Marshal(&params)
	if err != nil {
		logger.Error("SendEntityMsg failed 1 %s", err.Error())
		return
	}
	logger.Debug("SendEntityMsg %s %s %v %v", addr, lbtutil.ObjectId(id).Hex(), params, b)
	msg := &lbtproto.EntityMsg{
		Addr: addr,
		Id: id,
		Method: method,
		Params: b,
	}
	err = lbtproto.SendMessage(
		gateClient,
		lbtproto.ServiceGate.Method_entity_msg,
		msg,
	)
	if err != nil {
		logger.Error("SendEntityMsg failed 2 %s", err.Error())
	}
}

/********** ProtoSender End **********/
