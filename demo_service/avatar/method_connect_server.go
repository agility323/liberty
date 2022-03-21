package main

import (
	sf "github.com/agility323/liberty/service_framework"
)

type ConnectServerRequest struct {
}

type ConnectServerReply struct {
}

type connectServerHandler struct {
	request ConnectServerRequest
	reply ConnectServerReply
}

func (h *connectServerHandler) GetRequest() interface{} {return &(h.request)}
func (h *connectServerHandler) GetReply() interface{} {return nil}

func (h *connectServerHandler) Process(conAddr, srcAddr string) error {
	logger.Debug("connect server %s", srcAddr)
	// create avatar
	boost := sf.CreateEntity("BoostEntity").(*BoostEntity)
	boost.Init(conAddr, srcAddr)
	boost.Start()
	return nil
}
