package legacy

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/vmihailenco/msgpack"
	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtutil"
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

	/*
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
	*/

	if !dep.AtService() {
		LP_SendConnectServerResp(c, lbtproto.ConnectServerResp_Busy, []byte{})
		return nil
	}
	LP_SendConnectServerResp(c, lbtproto.ConnectServerResp_Connected, []byte{})
	id := lbtutil.NewObjectID()
	typ := dep.ConnectServerEntity
	data := lbtutil.MsgpackEmptyMapBytes
	LP_SendCreateChannelEntity(c, id[:], []byte(typ), data)

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
			Addr: addr,	// this field is not used by service
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
		if len(context) < 5 {
			return fmt.Errorf("ClientGate_entityMessage fail: invalid msg.Context %d<5", len(context))
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
		if len(context) > 5 {
			hval, err := strconv.Atoi(context[5])
			if err != nil {
				logger.Warn("legacy.proto.LP_ClientGate_entityMessage invalid context hval %s %v", context[5], err)
				hval = 0
			}
			newmsg.Hval = int32(hval)
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
	//logger.Debug("proto recv ClientGate_channelMessage %v", buf)
	msg := &lbtproto.ChannelMessage{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		return err
	}
	chanMsg := make(map[string]interface{})
	addr := c.RemoteAddr()
	if err := msgpack.Unmarshal(msg.ChanMsg, &chanMsg); err != nil {
		logger.Warn("channel message msgpack decode fail %s", addr)
		return err
	}
	itf, ok := chanMsg["msgType"]
	if !ok {
		logger.Warn("channel message missng type %s %v", addr, chanMsg)
		return nil
	}
	var mt int = 0
	v := reflect.ValueOf(itf)
	if !v.CanConvert(reflect.TypeOf(mt)) {
		logger.Warn("channel message invalid msg type %s %v %v", addr, itf, chanMsg)
		return nil
	}
	mt = v.Convert(reflect.TypeOf(mt)).Interface().(int)
	if !ok {
		logger.Warn("channel message invalid type %s %v", addr, chanMsg)
		return nil
	}
	itf, ok = chanMsg["channelId"]
	var cid string
	if !ok {
		cid = "NULL"
	} else {
		oid := lbtutil.NewObjectID()
		s, ok := itf.(string)
		if !ok {
			cid = "INVALID"
		}
		copy(oid[:], s)
		cid = oid.String()
	}
	if mt == 3 {	// heartbeat
/*
    hbMsg.channelId = channelId
    hbMsg.msgType = ChannelConst.MSGTYPE_HEARTBEAT
*/
		c.OnHeartbeat(0)
		lbtproto.SendMessage(c, lbtproto.Client.Method_channelMessage, msg)

	} else if mt == 1 {	// connect
/*
    connMsg.channelId = channelId
    connMsg.channelName = channelName
    connMsg.peer = peer
    connMsg.msgType = ChannelConst.MSGTYPE_CONNECT
*/
		chanMsg["channelId"] = cid
		logger.Info("gate channel message connect %v", chanMsg)
	} else if mt == 2 {	// close
/*
    closeMsg.channelId = channelId
    closeMsg.msgType = ChannelConst.MSGTYPE_CLOSE
*/
		chanMsg["channelId"] = cid
		logger.Info("gate channel message close %v", chanMsg)
	} else {
		chanMsg["channelId"] = cid
		logger.Debug("gate channel message unknown %v", chanMsg)
	}
	return nil
}
/********** ProtoHandler End **********/

/********** ProtoSender **********/
func LP_SendResponseEncryptToken(c *lbtnet.TcpConnection, token int64) {
	msg := &lbtproto.EncryptToken{Token: token}
	if err := lbtproto.SendMessage(c, lbtproto.Client.Method_responseEncryptToken, msg); err != nil {
		addr := "nil"
		if c != nil { addr = c.RemoteAddr() }
		logger.Warn("LP_SendResponseEncryptToken fail %s %d [%v]", addr, token, err)
	}
}

func LP_SendConfirmEncryptKeyAck(c *lbtnet.TcpConnection) {
	msg := &lbtproto.Void{}
	if err := lbtproto.SendMessage(c, lbtproto.Client.Method_confirmEncryptKeyAck, msg); err != nil {
		addr := "nil"
		if c != nil { addr = c.RemoteAddr() }
		logger.Warn("LP_SendConfirmEncryptKeyAck fail %s [%v]", addr, err)
	}
}

func LP_SendConnectServerResp(c *lbtnet.TcpConnection, typ lbtproto.ConnectServerResp_RespType, sessionId []byte) {
	msg := &lbtproto.ConnectServerResp{
		Type: typ,
		SessionId: sessionId,
	}
	err := lbtproto.SendMessage(c, lbtproto.Client.Method_connectResponse, msg)
	if err != nil {
		addr := "nil"
		if c != nil { addr = c.RemoteAddr() }
		logger.Warn("LP_SendConnectServerResp fail %s [%v]", addr, err)
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
		addr := "nil"
		if c != nil { addr = c.RemoteAddr() }
		logger.Warn("LP_SendCreateChannelEntity fail %s [%v]", addr, err)
	}
}

func LP_MakeEntityMessageData(entityid []byte, method []byte, parameters []byte) ([]byte, error) {
	msg := &lbtproto.EntityMessage{
		EntityId: entityid,
		MethodName: method,
		Index: 0,
		Parameters: parameters,
		SessionId: []byte {},
		Context: []byte {},
	}
	return lbtproto.EncodeMessage(lbtproto.Client.Method_entityMessage, msg)
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
	err := lbtproto.SendMessage(c, lbtproto.Client.Method_entityMessage, msg)
	if err != nil {
		addr := "nil"
		if c != nil { addr = c.RemoteAddr() }
		logger.Warn("LP_SendEntityMessage fail %s [%v]", addr, err)
	}
}
/********** ProtoSender End **********/
