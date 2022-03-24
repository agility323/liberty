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
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_service_reply] = ServiceGate_service_reply
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_create_entity] = ServiceGate_create_entity
	ServiceGateMethodHandler[lbtproto.ServiceGate.Method_entity_msg] = ServiceGate_entity_msg
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
	return f(c, buf)
}

func ServiceGate_register_service(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv ServiceGate_register_service %v", buf)
	msg := &lbtproto.ServiceInfo{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	info := serviceInfo{
		addr: msg.Addr,
		typ: msg.Type,
		entityid: msg.Entityid,
	}
	postServiceManagerJob("register", info)
	return nil
}

func ServiceGate_service_reply(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv ServiceGate_service_reply %v", buf)
	postClientManagerJob("service_reply", buf)
	return nil
}

func ServiceGate_create_entity(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv ServiceGate_create_entity %v", buf)
	postClientManagerJob("create_entity", buf)
	return nil
}

func ServiceGate_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv ServiceGate_entity_msg %v", buf)
	postClientManagerJob("entity_msg", buf)
	return nil
}
