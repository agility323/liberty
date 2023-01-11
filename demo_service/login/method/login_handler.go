package method

import (
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"

	"github.com/agility323/liberty/demo_service/login/core"
)

const (
	LoginSuccess int = iota
	LoginPasswordError
	LoginAccountNotExist
	LoginAvatarNotExist
	LoginWrongState
	LoginRedisErr
	LoginParamErr
	LoginInternalErr
	LoginTokenError
	LoginSameRoleExist
	LoginClientVersionError
)

type LoginRequest struct {
	Account string `msgpack:"account"`
	Password string `msgpack:"password"`
}

type LoginReply struct {
	Code int `msgpack:"code"`
	Token string `msgpack:"token"`
}

type LoginHandler struct {
	request LoginRequest
	reply LoginReply
}

func (h *LoginHandler) GetRequest() interface{} {return &(h.request)}
func (h *LoginHandler) GetReply() interface{} {return &(h.reply)}

func (h *LoginHandler) Process(c *lbtnet.TcpConnection, srcAddr string) error {
	acc := h.request.Account
	if !lbtutil.IsSimpleString(acc) {
		h.reply = LoginReply{Code: LoginAccountNotExist, Token: ""}
		return nil
	}
	code, token := core.StartLogin(acc)
	h.reply = LoginReply{Code: code, Token: token}
	return nil
}
