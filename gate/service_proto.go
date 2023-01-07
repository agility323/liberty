package main

import (
	"errors"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
)

var ServiceGateMethodHandler map[uint16]lbtnet.ProtoHandlerType = make(map[uint16]lbtnet.ProtoHandlerType)

func init() {
	initServiceGateMethodHandler()
}

func initServiceGateMethodHandler() {
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_register_service] = ServiceGate_register_service
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_bind_client] = ServiceGate_bind_client
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_unbind_client] = ServiceGate_unbind_client
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_service_request] = ServiceGate_service_request
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_service_reply] = ServiceGate_service_reply
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_client_service_reply] = ServiceGate_client_service_reply
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_create_entity] = ServiceGate_create_entity
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_entity_msg] = ServiceGate_entity_msg
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_client_entity_msg] = ServiceGate_client_entity_msg
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_register_filter_data] = ServiceGate_register_filter_data
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_filter_msg] = ServiceGate_filter_msg
}

func processServiceProto(c *lbtnet.TcpConnection, buf []byte) error {
	methodIndex, err := lbtproto.DecodeMethodIndex(buf)
	if err != nil {
		logger.Error("service proto fail read index %s", err.Error())
		return errors.New("read index")
	}
	f, ok := ServiceGateMethodHandler[methodIndex]
	if !ok {
		logger.Error("service proto fail wrong index %d", methodIndex)
		return errors.New("wrong index")
	}
	//logger.Debug("proto recv %d %v", methodIndex, buf)
	return f(c, buf)
}

func ServiceGate_register_service(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.ServiceInfo{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	// register: nothing except log
	entry := serviceManager.getServiceEntry(msg.Addr)
	if entry == nil {
		logger.Warn("ServiceGate_register_service fail 1 - service not connected %v %v", entry, msg)
		return nil
	}
	logger.Info("register service %s %s", msg.Type, msg.Addr)
	// reply
	if err := lbtproto.SendMessage(entry.cli, lbtproto.Service.Method_register_reply, msg); err != nil {
		logger.Warn("service register reply send fail - %v %v", err, msg)
	}
	return nil
}

func ServiceGate_bind_client(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.BindClientInfo{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	clientManager.bindClient(*msg)
	return nil
}

func ServiceGate_unbind_client(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.BindClientInfo{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	clientManager.unbindClient(*msg)
	return nil
}

func ServiceGate_service_request(c *lbtnet.TcpConnection, buf []byte) error {
	serviceManager.serviceRequest(buf)
	return nil
}

func ServiceGate_service_reply(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.ServiceReply{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("service reply fail 1 %v", err)
		return nil
	}
	serviceManager.sendToService(msg.Addr, buf)
	return nil
}

func ServiceGate_client_service_reply(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.ServiceReply{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("ServiceGate_client_service_reply fail 1 %v", err)
		return nil
	}
	cc := clientManager.getClientConnection(msg.Addr)
	if cc == nil {
		logger.Warn("ServiceGate_client_service_reply fail 2")
		return nil
	}
	sendClientServiceReply(cc, msg.Reply)
	return nil
}

func ServiceGate_create_entity(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.EntityData{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("ServiceGate_create_entity fail 1 %v", err)
		return nil
	}
	cc := clientManager.getClientConnection(msg.Addr)
	if cc == nil {
		logger.Warn("ServiceGate_create_entity fail 2")
		return nil
	}
	sendCreateEntity(cc, msg.Id, msg.Type, msg.Data)
	return nil
}

func ServiceGate_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("ServiceGate_entity_msg fail 1 %s", c.RemoteAddr())
		return nil
	}
	serviceManager.sendToService(msg.Addr, buf)
	return nil
}

func ServiceGate_client_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("ServiceGate_client_entity_msg fail 1 %v", err)
		return nil
	}
	cc := clientManager.getClientConnection(msg.Addr)
	if cc == nil {
		logger.Warn("ServiceGate_client_entity_msg fail 2")
		return nil
	}
	sendEntityMsg(cc, msg.Id, msg.Method, msg.Params)
	return nil
}

func ServiceGate_register_filter_data(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.FilterData{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("ServiceGate_register_filter_data fail 1 %v", err)
		return nil
	}
	clientManager.registerFilterData(msg.Id, msg.Data)
	return nil
}

func ServiceGate_filter_msg(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.FilterMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("ServiceGate_filter_msg fail 1 %v", err)
		return nil
	}
	data, err := makeBroadcastMsgData(msg.Method, msg.Params)
	if err != nil {
		logger.Error("ServiceGate_filter_msg fail 2 %v", err)
		return nil
	}
	arr := clientManager.filterClients(msg.Filters)
	for _, clients := range arr {
		for _, c := range clients {
			c.SendData(data)
		}
	}
	return nil
}
