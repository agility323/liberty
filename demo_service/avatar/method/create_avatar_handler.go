package main

import (
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
	sf "github.com/agility323/liberty/service_framework"

	"github.com/agility323/liberty/demo_service/avatar/avatardata"
	"github.com/agility323/liberty/demo_service/avatar/core"
)

type CreateAvatarRequest struct {
	Account string `msgpack:"account"`
	LoginToken string `msgpack:"login_token"`
}

type CreateAvatarReply struct {
	Code int `msgpack:"code"`
}

type CreateAvatarHandler struct {
	request CreateAvatarRequest
	reply lbtutil.Void
}

func (h *CreateAvatarHandler) GetRequest() interface{} {return &(h.request)}
func (h *CreateAvatarHandler) GetReply() interface{} {return &(h.reply)}

func (h *CreateAvatarHandler) Process(c *lbtnet.TcpConnection, srcAddr string) error {
	acc := h.request.Account
	token := h.request.LoginToken
	// TODO: verify token on redis
	logger.Debug("verify token %s %s", acc, token)
	// create avatar
	core.xxx.CreateAvatar()
	avatar := sf.CreateEntity("Avatar").(*Avatar)
	avatar.Init(sf.NewRemoteEntityStub(&avatar.EC, c, srcAddr), &avatardata.AvatarData{}, true)
	avatar.Start()
	return nil
}
