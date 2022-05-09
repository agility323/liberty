package main

import (
	"errors"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtutil"

	"github.com/vmihailenco/msgpack"
)

func init() {
	initLegacyMethodHandler()
}

func initLegacyMethodHandler() {
	ClientGateMethodHandler[lbtproto.ClientGate.Method_connectServer] = lp_ClientGate_connectServer
	ClientGateMethodHandler[lbtproto.ClientGate.Method_entityMessage] = lp_ClientGate_entityMessage
	ClientGateMethodHandler[lbtproto.ClientGate.Method_channelMessage] = lp_ClientGate_channelMessage

}

/********** ProtoHandler **********/
func lp_ClientGate_connectServer(c *lbtnet.TcpConnection, buf []byte) error {
	//logger.Debug("proto recv ClientGate_connectServer %v", buf)
	msg := &lbtproto.ConnectServerReq{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	pdata, err := msgpack.Marshal(&map[string]interface{}{})
	newmsg := &lbtproto.ServiceRequest{
		Addr: c.RemoteAddr(),
		Reqid: string(lbtutil.NewObjectId()),
		Type: Conf.ConnectServerHandler.Service,
		Method: Conf.ConnectServerHandler.Method,
		Params: pdata,
	}
	data, err := lbtproto.EncodeMessage(
			lbtproto.Service.Method_service_request,
			newmsg,
		)
	if err != nil {
		return err
	}
	if postServiceManagerJob("service_request", data) {
		lp_sendConnectServerResp(c, lbtproto.ConnectServerResp_Connected, []byte{})
	} else {
		lp_sendConnectServerResp(c, lbtproto.ConnectServerResp_Busy, []byte{})
	}
	return nil
}

func lp_ClientGate_entityMessage(c *lbtnet.TcpConnection, buf []byte) error {
	//logger.Debug("proto recv ClientGate_entityMessage %v", buf)
	msg := &lbtproto.EntityMessage{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	var context []string
	//logger.Debug("proto recv ClientGate_entityMessage context=%v", msg.Context)
	if err := msgpack.Unmarshal(msg.Context, &context); err != nil {
		return err
	}
	if len(context) != 3 {
		return errors.New("ClientGate_entityMessage fail: invalid msg.Context")
	}
	msgType := context[0]
	addrOrType := context[1]
	id := context[2]
	if msgType == "entity" {
		newmsg := &lbtproto.EntityMsg{
			Addr: addrOrType,
			Id: id,	// msg.EntityId from client is empty string (client use SessionId)
			Method: string(msg.MethodName),
			Params: msg.Parameters,
		}
		newbuf, err := lbtproto.EncodeMessage(
				lbtproto.Service.Method_entity_msg,
				newmsg,
			)
		if err != nil {
			return err
		}
		postServiceManagerJob("entity_msg", []interface{} {c.RemoteAddr(), newbuf})
	} else if msgType == "service" {
		newmsg := &lbtproto.ServiceRequest{
			Addr: c.RemoteAddr(),
			Reqid: id,
			Type: addrOrType,
			Method: string(msg.MethodName),
			Params: msg.Parameters,
			Context: msg.EntityId,
		}
		data, err := lbtproto.EncodeMessage(
				lbtproto.Service.Method_service_request,
				newmsg,
			)
		if err != nil {
			return err
		}
		postServiceManagerJob("service_request", data)
	}
	return nil
}

func lp_ClientGate_channelMessage(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv ClientGate_channelMessage %v", buf)
	return nil
}
/********** ProtoHandler End **********/

/********** ProtoSender **********/
func lp_sendConnectServerResp(c *lbtnet.TcpConnection, typ lbtproto.ConnectServerResp_RespType, sessionId []byte) {
	msg := &lbtproto.ConnectServerResp{
		Type: typ,
		SessionId: sessionId,
	}
	err := lbtproto.SendMessage(
		c,
		lbtproto.Client.Method_connectResponse,
		msg,
	)
	if err != nil {
		logger.Error("lp_sendConnectServerResp failed: SendMessage - %s", err.Error())
	}
}

func lp_sendEntityMessage(c *lbtnet.TcpConnection, entityid []byte, method []byte, parameters []byte) {
	msg := &lbtproto.EntityMessage{
		EntityId: entityid,
		MethodName: method,
		Index: 0,
		Parameters: parameters,
		SessionId: []byte {},
		Context: []byte {},
	}
	err := lbtproto.SendMessage(
		c,
		lbtproto.Client.Method_entityMessage,
		msg,
	)
	if err != nil {
		logger.Error("lp_sendEntityMessage failed: SendMessage - %s", err.Error())
	}
}
/********** ProtoSender End **********/
