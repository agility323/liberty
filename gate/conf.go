package main

import (
	"github.com/agility323/liberty/lbtutil"
)

type connectServerHandler struct {
	Service string `json:"service"`
	Method string `json:"method"`
}

type confType struct {
	LogLevel int `json:"log_level"`
	ClientServerAddr string `json:"client_server_addr"`
	ServiceServerAddr string `json:"service_server_addr"`
	ConnectServerHandler connectServerHandler `json:"connect_server_handler"`
}

var Conf confType

var defaultConf map[string]interface{} = map[string]interface{} {
	"log_level": lbtutil.Ldebug,
	"client_server_addr": "127.0.0.1:4001",
	"service_server_addr": "127.0.0.1:5001",
	"connect_server_handler": map[string]string {
		"service": "login_service",
		"method": "connect_server",
	},
}
