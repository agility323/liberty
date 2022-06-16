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
	entry := serviceEntry{
		addr: msg.Addr,
		typ: msg.Type,
	}
	postServiceManagerJob("register", entry)
	return nil
}

func ServiceGate_bind_client(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.BindClientInfo{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	postClientManagerJob("bind_client", *msg)
	return nil
}

func ServiceGate_unbind_client(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.BindClientInfo{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	postClientManagerJob("unbind_client", *msg)
	return nil
}

func ServiceGate_service_request(c *lbtnet.TcpConnection, buf []byte) error {
	postServiceManagerJob("service_request", buf)
	return nil
}

func ServiceGate_service_reply(c *lbtnet.TcpConnection, buf []byte) error {
	postServiceManagerJob("service_reply", buf)
	return nil
}

func ServiceGate_client_service_reply(c *lbtnet.TcpConnection, buf []byte) error {
	postClientManagerJob("client_service_reply", buf)
	return nil
}

func ServiceGate_create_entity(c *lbtnet.TcpConnection, buf []byte) error {
	postClientManagerJob("create_entity", buf)
	return nil
}

func ServiceGate_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	msg := &lbtproto.EntityMsg{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("ServiceGate_entity_msg fail 1 %s", c.RemoteAddr())
		return nil
	}
	postServiceManagerJob("entity_msg", []interface{} {msg.Addr, buf})
	return nil
}

func ServiceGate_client_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	postClientManagerJob("client_entity_msg", buf)
	return nil
}
