package method

import (
	"github.com/agility323/liberty/lbtnet"

	"github.com/agility323/liberty/demo_service/login/core"
)

type CreateRoleRequest struct {
	Account string `msgpack:"account"`
	Token string `msgpack:"token"`
	AvatarInfo *core.AvatarData `msgpack:"avatar_info"`
}

type CreateRoleReply struct {
	Code int `msgpack:"code"`
	AvatarId string `msgpack:"avatar_id"`
}

type CreateRoleHandler struct {
	request CreateRoleRequest
	reply CreateRoleReply
}

func (h *CreateRoleHandler) GetRequest() interface{} {return &(h.request)}
func (h *CreateRoleHandler) GetReply() interface{} {return &(h.reply)}

func (h *CreateRoleHandler) Process(c *lbtnet.TcpConnection, srcAddr string) error {
	acc := h.request.Account
	info := h.request.AvatarInfo
	code, avatarId := core.CreateRole(srcAddr, acc, info)
	h.reply = CreateRoleReply{Code: code, AvatarId: avatarId}
	return nil
}
