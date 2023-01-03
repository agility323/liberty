package main

import (
	"errors"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"

	"gate/legacy"
)

var ClientGateMethodHandler map[uint16]lbtnet.ProtoHandlerType = make(map[uint16]lbtnet.ProtoHandlerType)

func init() {
	initClientGateMethodHandler()
}

func initClientGateMethodHandler() {
	ClientGateMethodHandler[lbtproto.ClientGate.Method_client_service_request] = ClientGate_client_service_request
	ClientGateMethodHandler[lbtproto.ClientGate.Method_entity_msg] = ClientGate_entity_msg
	legacy.InitLegacyMethodHandler(ClientGateMethodHandler)
}

func processClientProto(c *lbtnet.TcpConnection, buf []byte) error {
	methodIndex, err := lbtproto.DecodeMethodIndex(buf)
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
func ClientGate_client_service_request(c *lbtnet.TcpConnection, buf []byte) error {
	postServiceManagerJob("service_request", buf)
	return nil
}

func ClientGate_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	caddr := c.RemoteAddr()
	saddr := clientManager.getServiceAddr(caddr)	// TODO: concurrent issue
	if saddr == "" { return errors.New("no saddr for " + caddr) }
	postServiceManagerJob("entity_msg", []interface{} {saddr, buf})
	return nil
}
/********** ProtoHandler End **********/

/********** ProtoSender **********/
func sendClientServiceReply(c *lbtnet.TcpConnection, reqid string, reply []byte) {
	// reqid is reserved for future use
	legacy.LP_sendEntityMessage(c, []byte {}, []byte("CMD_service_reply"), reply)
}

func sendEntityMsg(c *lbtnet.TcpConnection, entityid, method string, params []byte) {
	legacy.LP_sendEntityMessage(c, []byte(entityid), []byte(method), params)
}
/********** ProtoSender End **********/
