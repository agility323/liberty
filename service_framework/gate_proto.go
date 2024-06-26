package service_framework

import (
	"errors"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtactor"

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
	ServiceMethodHandler[lbtproto.Service.Method_heartbeat] = Service_heartbeat
	ServiceMethodHandler[lbtproto.Service.Method_gate_stop] = Service_gate_stop
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
	if cb := popClientCallback(info.Caddr); cb != nil { cb.OnClientDisconnect() }
	return nil
}

func Service_service_request(c *lbtnet.TcpConnection, buf []byte) error {
	request := &lbtproto.ServiceRequest{}
	if err := lbtproto.DecodeMessage(buf, request); err != nil {
		return err
	}
	runServiceRequestTask(c, request, true)
	return nil
}

func runServiceRequestTask(c *lbtnet.TcpConnection, req *lbtproto.ServiceRequest, fromService bool) {
	task := func() {
		reqid := string(req.Reqid)
		replyData, err := processServiceMethod(c, req.Addr, reqid, req.Method, req.Params)
		if err != nil {
			logger.Warn("service request task fail [%v] [%v]", err, req)
			return
		}
		if replyData == nil { return }
		if fromService {
			sendServiceReply(c, req.Addr, reqid, replyData)
		} else {
			sendClientServiceReply(c, req.Addr, reqid, replyData)
		}
		return
	}
	hval := int(req.Hval - 1)
	if hval < 0 {
		lbtactor.RunTask("runServiceRequestTask." + req.Method, task)
		return
	}
	if HashedMethodWorker == nil {
		logger.Warn("HashedMethodWorker is nil")
		lbtactor.RunTask("runServiceRequestTask." + req.Method, task)
		return
	}
	HashedMethodWorker.PushTask(task, hval)
}

func Service_service_reply(c *lbtnet.TcpConnection, buf []byte) error {
	reply := &lbtproto.ServiceReply{}
	if err := lbtproto.DecodeMessage(buf, reply); err != nil {
		return err
	}
	reqid := *(*lbtutil.ObjectID)(reply.Reqid)
	processServiceReply(reqid, reply.Reply)
	return nil
}

func Service_client_service_request(c *lbtnet.TcpConnection, buf []byte) error {
	request := &lbtproto.ServiceRequest{}
	if err := lbtproto.DecodeMessage(buf, request); err != nil {
		return err
	}
	runServiceRequestTask(c, request, false)
	return nil
}

func Service_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	id := *(*lbtutil.ObjectID)(msg.Id)
	if err := CallEntityMethodLocal(id, msg.Method, msg.Params, int(msg.Hval)); err != nil {
		return err
	}
	return nil
}

func Service_heartbeat(c *lbtnet.TcpConnection, buf []byte) error {
	c.OnHeartbeat(0)
	return c.SendData(buf)
}

func Service_gate_stop(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.ServiceInfo{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("Service_gate_stop fail decode %v", err)
		return nil
	}
	gateManager.onGateStop(c.RemoteAddr())
	return nil
}
/********** ProtoHandler End **********/

/********** ProtoSender **********/
func sendRegisterService(c *lbtnet.TcpConnection) {
	msg := &lbtproto.ServiceInfo{
		Addr: serviceAddr,
		Type: serviceConf.ServiceType,
		Entityid: []byte {},
	}
	if err := lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_register_service, msg); err != nil {
		logger.Error("sendRegisterService failed %v", err)
	}
}

func sendServiceReply(c *lbtnet.TcpConnection, addr, reqid string, data []byte) {
	reply := &lbtproto.ServiceReply{
		Addr: addr,
		Reqid: []byte(reqid),
		Reply: data,
	}
	if err := lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_service_reply, reply); err != nil {
		logger.Error("sendServiceReply failed %v", err)
	}
}

func sendClientServiceReply(c *lbtnet.TcpConnection, addr, reqid string, data []byte) {
	reply := &lbtproto.ServiceReply{
		Addr: addr,
		Reqid: []byte(reqid),
		Reply: data,
	}
	if err := lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_client_service_reply, reply); err != nil {
		logger.Error("sendClientServiceReply failed %v", err)
	}
}

func SendBindClient(c *lbtnet.TcpConnection, saddr, caddr string) error {
	msg := &lbtproto.BindClientInfo{
		Caddr: caddr,
		Saddr: saddr,
	}
	return lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_bind_client, msg)
}

func SendUnbindClient(c *lbtnet.TcpConnection, saddr, caddr string) error {
	msg := &lbtproto.BindClientInfo{
		Caddr: caddr,
		Saddr: saddr,
	}
	return lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_unbind_client, msg)
}

func SendCreateEntity(c *lbtnet.TcpConnection, addr string, id lbtutil.ObjectID, typ string, data interface{}) error {
	b, err := msgpack.Marshal(&data)
	if err != nil {
		logger.Error("SendCreateEntity failed: marshal data - %s", err.Error())
		return err
	}
	logger.Debug("SendCreateEntity %s %s %s %v %v", addr, id.Hex(), typ, data, b)
	msg := &lbtproto.EntityData{
		Addr: addr,
		Id: id[:],
		Type: typ,
		Data: b,
	}
	err = lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_create_entity, msg)
	if err != nil {
		logger.Error("SendCreateEntity failed: SendMessage - %s", err.Error())
		return err
	}
	return nil
}

func SendClientEntityMsg(c *lbtnet.TcpConnection, addr string, id lbtutil.ObjectID, method string, params interface{}) error {
	b, err := msgpack.Marshal(&params)
	if err != nil {
		logger.Error("SendClientEntityMsg failed 1 %s", err.Error())
		return err
	}
	logger.Debug("SendClientEntityMsg %s %s %s %v", addr, id.Hex(), method, params)
	msg := &lbtproto.EntityMsg{
		Addr: addr,
		Id: id[:],
		Method: method,
		Params: b,
	}
	err = lbtproto.SendMessage(c, lbtproto.ServiceGate.Method_client_entity_msg, msg)
	if err != nil {
		logger.Error("SendClientEntityMsg failed 2 %s", err.Error())
		return err
	}
	return nil
}

func makeServiceStopData() ([]byte, error) {
	msg := &lbtproto.ServiceInfo{
		Addr: serviceAddr,
		Type: serviceConf.ServiceType,
		Entityid: []byte {},
	}
	return lbtproto.EncodeMessage(lbtproto.ServiceGate.Method_service_stop, msg)
}

/********** ProtoSender End **********/
