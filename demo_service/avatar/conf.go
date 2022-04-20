package main

import (
	"github.com/agility323/liberty/lbtutil"
	sf "github.com/agility323/liberty/service_framework"
)

type redisConfType struct {
	Addrs []string `json:"addrs"`
	MasterName string `json:"master_name"`
}

type mongoConfType struct {
	Uri string `json:"uri"`
}

type confType struct {
	Service *sf.ServiceConfType `json:"service"`
	Redis redisConfType `json:"redis"`
	Mongo mongoConfType `json:"mongo"`
}

var Conf confType

var defaultConf map[string]interface{} = map[string]interface{} {
	"service": map[string]interface{} {
		"log_level": lbtutil.Ldebug,
		"host": 101,
		"service_type": "avatar_service",
		"gate_server_addr": "127.0.0.1:5001",
		"etcd": []string {
			"http://127.0.0.1:2379",
			"http://127.0.0.1:2479",
			"http://127.0.0.1:2579",
		},
	},
	"redis": map[string]interface{} {
		"addrs": []string{
			"127.0.0.1:6391",
			"127.0.0.1:6392",
			"127.0.0.1:6393",
			"127.0.0.1:6394",
			"127.0.0.1:6395",
			"127.0.0.1:6396",
		},
		"master_name": "",
	},
	"mongo": map[string]interface{} {
		"uri": "mongodb://gzl:gzl@127.0.0.1:25001,127.0.0.1:25002,127.0.0.1:25003/gamedb?w=1&minPoolSize=3&maxPoolSize=10&readPreference=primary",
	},
}
