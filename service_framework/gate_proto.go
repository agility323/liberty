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
	ServiceMethodHandler[lbtproto.Service.Method_service_reply] = Service_service_reply
	ServiceMethodHandler[lbtproto.Service.Method_client_service_request] = Service_client_service_request
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
	return nil
}

func Service_client_disconnect(c *lbtnet.TcpConnection, buf []byte) error {
	info := &lbtproto.BindClientInfo{}
	if err := lbtproto.DecodeMessage(buf, info); err != nil {
		return err
	}
	if cb := getClientCallback(info.Caddr); cb != nil { cb.OnClientDisconnect() }
	return nil
}

func Service_service_request(c *lbtnet.TcpConnection, buf []byte) error {
	request := &lbtproto.ServiceRequest{}
	if err := lbtproto.DecodeMessage(buf, request); err != nil {
		return err
	}
	/*
	replyData, err := processServiceMethod(c, request.Addr, request.Reqid, request.Method, request.Params)
	if err != nil {
		return err
	}
	if replyData == nil { return nil }
	sendServiceReply(c, request.Addr, request.Reqid, replyData)
	*/
	ma := newMethodActor(c, request, true)
	go ma.start()
	return nil
}

func Service_service_reply(c *lbtnet.TcpConnection, buf []byte) error {
	reply := &lbtproto.ServiceReply{}
	if err := lbtproto.DecodeMessage(buf, reply); err != nil {
		return err
	}
	processServiceReply(reply.Reqid, reply.Reply)
	return nil
}

func Service_client_service_request(c *lbtnet.TcpConnection, buf []byte) error {
	request := &lbtproto.ServiceRequest{}
	if err := lbtproto.DecodeMessage(buf, request); err != nil {
		return err
	}
	/*
	replyData, err := processServiceMethod(c, request.Addr, request.Reqid, request.Method, request.Params)
	if err != nil {
		return err
	}
	if replyData == nil { return nil }
	sendClientServiceReply(c, request.Addr, request.Reqid, replyData)
	*/
	ma := newMethodActor(c, request, false)
	go ma.start()
	return nil
}

func Service_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	if err := CallEntityMethodLocal(lbtutil.ObjectId(msg.Id), msg.Method, msg.Params); err != nil {
		return err
	}
	return nil
}

func Service_service_shutdown(c *lbtnet.TcpConnection, buf []byte) error {
	Stop()
	return nil
}
/********** ProtoHandler End **********/

/********** ProtoSender **********/
func sendRegisterService(c *lbtnet.TcpConnection) {
	msg := &lbtproto.ServiceInfo{
		Addr: c.LocalAddr(),
		Type: serviceConf.ServiceType,
		Entityid: "",
	}
	lbtproto.SendMessage(
		c,
		lbtproto.ServiceGate.Method_register_service,
		msg,
	)
}

func sendServiceReply(c *lbtnet.TcpConnection, addr, reqid string, data []byte) {
	reply := &lbtproto.ServiceReply{
		Addr: addr,
		Reqid: reqid,
		Reply: data,
	}
	if err := lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_service_reply, reply); err != nil {
		logger.Error("sendServiceReply failed %v", err)
	}
}

func sendClientServiceReply(c *lbtnet.TcpConnection, addr, reqid string, data []byte) {
	reply := &lbtproto.ServiceReply{
		Addr: addr,
		Reqid: reqid,
		Reply: data,
	}
	if err := lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_client_service_reply, reply); err != nil {
		logger.Error("sendClientServiceReply failed %v", err)
	}
}

func SendBindClient(c *lbtnet.TcpConnection, saddr, caddr string) {
	msg := &lbtproto.BindClientInfo{
		Caddr: caddr,
		Saddr: saddr,
	}
	lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_bind_client, msg)
}

func SendUnbindClient(c *lbtnet.TcpConnection, saddr, caddr string) {
	msg := &lbtproto.BindClientInfo{
		Caddr: caddr,
		Saddr: saddr,
	}
	lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_unbind_client, msg)
}

func SendCreateEntity(c *lbtnet.TcpConnection, addr, id, typ string, data interface{}) {
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
	err = lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_create_entity, msg)
	if err != nil {
		logger.Error("SendCreateEntity failed: SendMessage - %s", err.Error())
	}
}

func SendClientEntityMsg(c *lbtnet.TcpConnection, addr, id, method string, params interface{}) {
	b, err := msgpack.Marshal(&params)
	if err != nil {
		logger.Error("SendClientEntityMsg failed 1 %s", err.Error())
		return
	}
	logger.Debug("SendClientEntityMsg %s %s %s %v", addr, lbtutil.ObjectId(id).Hex(), method, params)
	msg := &lbtproto.EntityMsg{
		Addr: addr,
		Id: id,
		Method: method,
		Params: b,
	}
	err = lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_client_entity_msg, msg)
	if err != nil {
		logger.Error("SendClientEntityMsg failed 2 %s", err.Error())
	}
}

/********** ProtoSender End **********/
