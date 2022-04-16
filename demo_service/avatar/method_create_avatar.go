package main

import (
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/demo_service/avatar/avatardata"
	sf "github.com/agility323/liberty/service_framework"
)

type CreateAvatarRequest struct {
	Account string `msgpack:"account"`
	LoginToken string `msgpack:"login_token"`
}

type CreateAvatarReply struct {
	Code int `msgpack:"code"`
}

type createAvatarHandler struct {
	request CreateAvatarRequest
	reply lbtutil.Void
}

func (h *createAvatarHandler) GetRequest() interface{} {return &(h.request)}
func (h *createAvatarHandler) GetReply() interface{} {return &(h.reply)}

func (h *createAvatarHandler) Process(c *lbtnet.TcpConnection, srcAddr string) error {
	acc := h.request.Account
	token := h.request.LoginToken
	// TODO: verify token on redis
	logger.Debug("verify token %s %s", acc, token)
	// create avatar
	avatar := sf.CreateEntity("Avatar").(*Avatar)
	avatar.Init(c, srcAddr, &avatardata.AvatarData{}, true)
	avatar.Start()
	return nil
}
