package main

import (
	"errors"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"

	"github.com/vmihailenco/msgpack"

	"github.com/agility323/liberty/gate/legacy"
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
	serviceManager.serviceRequest(buf)
	return nil
}

func ClientGate_entity_msg(c *lbtnet.TcpConnection, buf []byte) error {
	caddr := c.RemoteAddr()
	saddr := clientManager.getClientServiceAddr(caddr)
	if saddr == "" { return errors.New("no saddr for " + caddr) }
	serviceManager.sendToService(saddr, buf)
	return nil
}
/********** ProtoHandler End **********/

/********** ProtoSender **********/
func sendClientServiceReply(c *lbtnet.TcpConnection, reply []byte) {
	legacy.LP_SendEntityMessage(c, []byte {}, []byte("CMD_service_reply"), reply)
}

func sendCreateEntity(c *lbtnet.TcpConnection, id []byte, typ string, data []byte) {
	legacy.LP_SendCreateChannelEntity(c, id, []byte(typ), data)
}

func sendEntityMsg(c *lbtnet.TcpConnection, entityid []byte, method string, params []byte) {
	legacy.LP_SendEntityMessage(c, entityid, []byte(method), params)
}

func makeBroadcastMsgData(msg string) ([]byte, error) {
	parameters, err := msgpack.Marshal([]interface{} {msg, })
	if err != nil { return nil, err }
	return legacy.LP_MakeEntityMessageData([]byte {}, []byte("CMD_broadcast_msg"), parameters)
}
/********** ProtoSender End **********/
