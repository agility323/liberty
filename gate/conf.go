package main

import (
	"github.com/agility323/liberty/lbtutil"
)

type confServerType struct {
	Ip string `json:"ip"`
	Port int `json:"port"`
}

type confType struct {
	LogLevel int `json:"log_level"`
	ClientServer confServerType `json:"client_server"`
	ServiceServer confServerType `json:"service_server"`
}

var Conf confType

var defaultConf map[string]interface{} = map[string]interface{} {
	"test_ip": "test_localhost",
	"test_port": 2799,
	"test_etcd": []int{1,2,3},

	"log_level": lbtutil.Ldebug,
	"client_server": map[string]interface{} {
		"ip": "127.0.0.1",
		"port": 4001,
	},
	"service_server": map[string]interface{} {
		"ip": "127.0.0.1",
		"port": 5001,
	},
}
