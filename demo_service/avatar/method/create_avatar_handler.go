package main

import (
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
	sf "github.com/agility323/liberty/service_framework"

	"github.com/agility323/liberty/demo_service/avatar/core"
)

type CreateAvatarRequest struct {
	Token string `msgpack:"token"`
	AvatarId string `msgpack:"avatar_id"`
	AvatarInfo gameplay.NewAvatarInfo `msgpack:"avatar_info"`
	SDKInfo *core.TrackingInfo `msgpack:"tracking_info"`
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
	// TODO: verify token on redis
	logger.Debug("verify token %s %s", acc, token)
	// create avatar
	code := core.CreateAvatar(srcAddr, h.request.Token, h.request.AvatarId, h.request.SDKInfo)
	h.reply = CreateAvatarReply{Code: code}
	return nil
}
