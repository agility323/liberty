package legacy

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtutil"

	"github.com/vmihailenco/msgpack"
	"github.com/golang/protobuf/proto"
)

/********** init **********/
func InitLegacyMethodHandler(handler map[uint16]lbtnet.ProtoHandlerType) {
	handler[lbtproto.ClientGate.Method_requestEncryptToken] = LP_ClientGate_requestEncryptToken
	handler[lbtproto.ClientGate.Method_confirmEncryptKey] = LP_ClientGate_confirmEncryptKey
	handler[lbtproto.ClientGate.Method_connectServer] = LP_ClientGate_connectServer
	handler[lbtproto.ClientGate.Method_entityMessage] = LP_ClientGate_entityMessage
	handler[lbtproto.ClientGate.Method_channelMessage] = LP_ClientGate_channelMessage

}
/********** init End **********/

/********** ProtoHandler **********/
func LP_ClientGate_requestEncryptToken(c *lbtnet.TcpConnection, buf []byte) error {
	token := rand.Int63()
	c.SetVar("encryptToken", token)
	LP_SendResponseEncryptToken(c, token)
	logger.Info("encryptToken set %s %d", c.RemoteAddr(), token)
	return nil
}

func LP_ClientGate_confirmEncryptKey(c *lbtnet.TcpConnection, buf []byte) error {
	encryptKeyString := &lbtproto.EncryptKeyString{}
	if err := lbtproto.DecodeMessage(buf, encryptKeyString); err != nil {
		return err
	}
	encryptKeyData, err := rsaDecoder.decode(encryptKeyString.GetKeyString())
	if err != nil {
		return err
	}
	encryptKey := &lbtproto.EncryptKey{}
	if err := proto.Unmarshal(encryptKeyData, encryptKey); err != nil {
		return err
	}
	if len(encryptKey.GetKey()) == 0 {
		return fmt.Errorf("LP_ClientGate_confirmEncryptKey fail 1 %s", c.RemoteAddr())
	}
	v := c.GetVar("encryptToken")
	if v == nil {
		return fmt.Errorf("LP_ClientGate_confirmEncryptKey fail 2 %s", c.RemoteAddr())
	}
	token := v.(int64)
	if encryptKey.GetToken() != token {
		return fmt.Errorf("LP_ClientGate_confirmEncryptKey fail 3 %s %d %d", c.RemoteAddr(), encryptKey.GetToken(), token)
	}
	key := encryptKey.GetKey()
	if err := c.EnableEncryptAndCompress(key); err != nil {
		return fmt.Errorf("LP_ClientGate_confirmEncryptKey fail 4 %s %v", c.RemoteAddr(), err)
	}
	LP_SendConfirmEncryptKeyAck(c)
	return nil
}

func LP_ClientGate_connectServer(c *lbtnet.TcpConnection, buf []byte) error {
	//logger.Debug("proto recv ClientGate_connectServer %v", buf)
	msg := &lbtproto.ConnectServerReq{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	pdata, err := msgpack.Marshal(&map[string]interface{}{})
	reqid := lbtutil.NewObjectID()
	newmsg := &lbtproto.ServiceRequest{
		Addr: c.RemoteAddr(),
		Reqid: reqid[:],
		Type: dep.ConnectServerService,
		Method: dep.ConnectServerMethod,
		Params: pdata,
	}
	data, err := lbtproto.EncodeMessage(
			lbtproto.Service.Method_client_service_request,
			newmsg,
		)
	if err != nil {
		return err
	}
	if dep.ServiceRequestHandler(data) {
		LP_SendConnectServerResp(c, lbtproto.ConnectServerResp_Connected, []byte{})
	} else {
		LP_SendConnectServerResp(c, lbtproto.ConnectServerResp_Busy, []byte{})
	}
	return nil
}

func LP_ClientGate_entityMessage(c *lbtnet.TcpConnection, buf []byte) error {
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
			return fmt.Errorf("ClientGate_entityMessage fail: invalid msg.Context %d!=3", len(context))
		}
		addr := context[1]
		id := context[2]
		newmsg := &lbtproto.EntityMsg{
			Addr: addr,
			Id: []byte(id),	// msg.EntityId from client is empty string (client use SessionId)
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
		saddr := dep.ServiceAddrGetter(caddr)
		if saddr == "" { return errors.New("no saddr for " + caddr) }
		dep.ServiceSender(saddr, newbuf)
	} else if msgType == "service" {
		if len(context) != 5 {
			return fmt.Errorf("ClientGate_entityMessage fail: invalid msg.Context %d!=5", len(context))
		}
		typ := context[1]
		id := context[2]
		routeType, _ := dep.LegacyRouteTypeMap[context[3]]
		routeParam := context[4]
		newmsg := &lbtproto.ServiceRequest{
			Addr: c.RemoteAddr(),
			Reqid: []byte(id),
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
		dep.ServiceRequestHandler(data)
	}
	return nil
}

func LP_ClientGate_channelMessage(c *lbtnet.TcpConnection, buf []byte) error {
	logger.Debug("proto recv ClientGate_channelMessage %v", buf)
	return nil
}
/********** ProtoHandler End **********/

/********** ProtoSender **********/
func LP_SendResponseEncryptToken(c *lbtnet.TcpConnection, token int64) {
	msg := &lbtproto.EncryptToken{Token: token}
	if err := lbtproto.SendMessage(c, lbtproto.Client.Method_responseEncryptToken, msg); err != nil {
		logger.Warn("LP_SendResponseEncryptToken fail %s %d %v", c.RemoteAddr(), token, err)
	}
}

func LP_SendConfirmEncryptKeyAck(c *lbtnet.TcpConnection) {
	msg := &lbtproto.Void{}
	if err := lbtproto.SendMessage(c, lbtproto.Client.Method_confirmEncryptKeyAck, msg); err != nil {
		logger.Warn("LP_SendConfirmEncryptKeyAck fail %s %v", c.RemoteAddr(), err)
	}
}

func LP_SendConnectServerResp(c *lbtnet.TcpConnection, typ lbtproto.ConnectServerResp_RespType, sessionId []byte) {
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
		logger.Error("LP_SendConnectServerResp failed: SendMessage - %s", err.Error())
	}
}

func LP_SendCreateChannelEntity(c *lbtnet.TcpConnection, id, typ, info []byte) {
	msg := &lbtproto.ChannelEntityInfo{
		Type: typ,
		Info: info,
		EntityId: id,
		SessionId: []byte{},
	}
	if err := lbtproto.SendMessage(c, lbtproto.Client.Method_createChannelEntity, msg); err != nil {
		logger.Error("LP_SendCreateChannelEntity fail 1 %s %v", c.RemoteAddr(), err)
	}
}

func LP_SendEntityMessage(c *lbtnet.TcpConnection, entityid []byte, method []byte, parameters []byte) {
	logger.Debug("LP_SendEntityMessage start: entityid=%v", string(entityid))
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
		logger.Error("LP_SendEntityMessage failed: SendMessage - %s", err.Error())
	}
}
/********** ProtoSender End **********/
