package main

import (
	"github.com/agility323/liberty/lbtutil"
)

type confType struct {
	LogLevel int `json:"log_level"`
	ClientServerAddr string `json:"client_server_addr"`
	EntranceAddr string `json:"entrance_addr"`
	Host int `json:"host"`
	ConnectServerHandler struct {
		Service string `json:"service"`
		Method string `json:"method"`
	} `json:"connect_server_handler"`
	Etcd []string `json:"etcd"`
	PrivateRsaKey string `json:"private_rsa_key"`
}

var Conf confType

var defaultConf map[string]interface{} = map[string]interface{} {
	"log_level": lbtutil.Ldebug,
	"client_server_addr": "127.0.0.1:4001",
	"entrance_addr": "127.0.0.1:4001",
	"host": 101,
	"connect_server_handler": map[string]string {
		"service": "login_service",
		"method": "connect_server",
	},
	"etcd": []string {
		"http://127.0.0.1:2379",
		"http://127.0.0.1:2479",
		"http://127.0.0.1:2579",
	},
	"private_rsa_key": "./rsa_key",
}
