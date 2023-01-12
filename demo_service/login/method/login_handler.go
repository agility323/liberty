package method

import (
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"

	"github.com/agility323/liberty/demo_service/login/core"
)

type LoginRequest struct {
	Account string `msgpack:"account"`
	Password string `msgpack:"password"`
}

type LoginReply struct {
	Code int `msgpack:"code"`
	Avatars map[string]*core.AvatarData `msgpack:"avatars"`
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
		h.reply = LoginReply{Code: core.LoginAccountNotExist, Token: ""}
		return nil
	}
	code, avatars, token := core.StartLogin(srcAddr, acc)
	h.reply = LoginReply{Code: code, Avatars: avatars, Token: token}
	return nil
}
