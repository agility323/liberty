package main

import (
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
)

type LoginRequest struct {
	Account string `msgpack:"account"`
	Password string `msgpack:"password"`
}

type LoginReply struct {
	Code int `msgpack:"code"`
	Msg string `msgpack:"msg"`
}

type loginHandler struct {
	request LoginRequest
	reply LoginReply
}

func (h *loginHandler) GetRequest() interface{} {return &(h.request)}
func (h *loginHandler) GetReply() interface{} {return &(h.reply)}

func (h *loginHandler) Process(c *lbtnet.TcpConnection, srcAddr string) error {
	if !lbtutil.IsSimpleString(h.request.Account) {
		h.reply = LoginReply{Code: 101, Msg: "invalid account"}
	} else {
		h.reply = LoginReply{Code: 0, Msg: ""}
	}
	return nil
}
