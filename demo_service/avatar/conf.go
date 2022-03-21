package main

import (
	"github.com/agility323/liberty/lbtutil"
)

type redisConfType struct {
	Addrs []string `json:"addrs"`
	MasterName string `json:"master_name"`
}

type mongoConfType struct {
	Uri string `json:"uri"`
}

type confType struct {
	ServiceType string `json:"service_type"`
	LogLevel int `json:"log_level"`
	Redis redisConfType `json:"redis"`
	Mongo mongoConfType `json:"mongo"`
}

var Conf confType

var defaultConf map[string]interface{} = map[string]interface{} {
	"service_type": "avatar_service",
	"log_level": lbtutil.Ldebug,
	"redis": map[string]interface{} {
		"addrs": []string{
			"10.1.71.45:6391",
			"10.1.71.45:6392",
			"10.1.71.45:6393",
			"10.1.71.45:6394",
			"10.1.71.45:6395",
			"10.1.71.45:6396",
		},
		"master_name": "",
	},
	"mongo": map[string]interface{} {
		"uri": "mongodb://gzl:gzl@127.0.0.1:25001,127.0.0.1:25002,127.0.0.1:25003/gamedb?w=1&minPoolSize=3&maxPoolSize=10&readPreference=primary",
	},
}
