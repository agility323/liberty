package main

import (
	"github.com/agility323/liberty/lbtutil"
)

type confType struct {
	LogLevel int `json:"log_level"`
	ClientServerAddr string `json:"client_server_addr"`
	ServiceServerAddr string `json:"service_server_addr"`
}

var Conf confType

var defaultConf map[string]interface{} = map[string]interface{} {
	"log_level": lbtutil.Ldebug,
	"client_server_addr": "127.0.0.1:4001",
	"service_server_addr": "127.0.0.1:5001",
}
