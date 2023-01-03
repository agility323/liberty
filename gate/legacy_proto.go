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
			lbtproto.Service.Method_client_service_request,
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

var legacyRouteTypeMap = map[string]int32 {
	"random": RouteTypeRandomOne,
	"hash": RouteTypeHash,
	"specific": RouteTypeSpecific,
	"all": RouteTypeAll,
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
	if len(context) == 0 {
		return errors.New("ClientGate_entityMessage fail: empty msg.Context")
	}
	msgType := context[0]
	if msgType == "entity" {
		if len(context) != 3 {
			return errors.New("ClientGate_entityMessage fail: invalid msg.Context")
		}
		addr := context[1]
		id := context[2]
		newmsg := &lbtproto.EntityMsg{
			Addr: addr,
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
		caddr := c.RemoteAddr()
		saddr := clientManager.getServiceAddr(caddr)   // TODO: concurrent issue
		if saddr == "" { return errors.New("no saddr for " + caddr) }
		postServiceManagerJob("entity_msg", []interface{} {saddr, newbuf})
	} else if msgType == "service" {
		if len(context) != 5 {
			return errors.New("ClientGate_entityMessage fail: invalid msg.Context")
		}
		typ := context[1]
		id := context[2]
		routeType, _ := legacyRouteTypeMap[context[3]]
		routeParam := context[4]
		newmsg := &lbtproto.ServiceRequest{
			Addr: c.RemoteAddr(),
			Reqid: id,
			Type: typ,
			Method: string(msg.MethodName),
			Params: msg.Parameters,
		}
		if routeType != 0 || len(routeParam) > 0 {
			newmsg.Routet = routeType
			newmsg.Routep = []byte(routeParam)
		}
		data, err := lbtproto.EncodeMessage(
				lbtproto.Service.Method_client_service_request,
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
	logger.Debug("lp_sendEntityMessage start: entityid=%v", string(entityid))
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
