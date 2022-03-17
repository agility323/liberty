package main

import (
	"errors"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
)

var ClientGateMethodHandler map[uint16]lbtnet.ProtoHandlerType = make(map[uint16]lbtnet.ProtoHandlerType)

func init() {
	initClientGateMethodHandler()
}

func initClientGateMethodHandler() {
	ClientGateMethodHandler[lbtproto.ClientGate.Method_service_request] = ClientGate_service_request
	ClientGateMethodHandler[lbtproto.ClientGate.Method_entity_msg] = ClientGate_entity_msg
}

func processClientProto(c *lbtnet.TcpConnection, buf []byte) error {
	methodIndex, err := lbtproto.DecodeMethodIndex(buf, lbtnet.ByteOrder)
	if err != nil {
		logger.Error("client proto fail read index %s", err.Error())
		return errors.New("read index")
	}
	f, ok := ClientGateMethodHandler[methodIndex]
	if !ok {
		logger.Error("client proto fail wrong index %d", methodIndex)
		return errors.New("wrong index")
	}
	return f(c, buf)
}

/********** ProtoHandler **********/
func ClientGate_service_request(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv ClientGate_service_request %v", buf)
	postServiceManagerJob("service_request", buf)
	return nil
}

func ClientGate_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv ClientGate_entity_msg %v", buf)
	postServiceManagerJob("entity_msg", buf)	// TODO test only: send to arbitrary single service
	return nil
}
/********** ProtoHandler End **********/

/********** ProtoSender **********/

/********** ProtoSender End **********/
