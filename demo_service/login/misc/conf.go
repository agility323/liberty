package misc

import (
	sfitf "github.com/agility323/liberty/service_framework/itf"
)

type confType struct {
	Service *sfitf.ServiceConfType `json:"service"`
	Redis struct {
		Addrs []string `json:"addrs"`
		MasterName string `json:"master_name"`
		PoolSize int `json:"pool_size"`
		MinIdleConns int `json:"min_idle_conns"`
	} `json:"redis"`
	Mongo struct {
		Uri string `json:"uri"`
		Db string `json:"db"`
	} `json:"mongo"`
	ServerType string `json:"server_type"`
	ClientVersion string `json:"client_version"`
}

var Conf confType

var DefaultConf = map[string]interface{} {
	"service": map[string]interface{} {
		//"log_level": lbtutil.Ldebug,
		"host": 101,
		"service_type": "login_service",
		"gate_server_addr": "127.0.0.1:5002",
		"etcd": []string {
			"http://127.0.0.1:2379",
			"http://127.0.0.1:2479",
			"http://127.0.0.1:2579",
		},
	},
	"redis": map[string]interface{}{
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
	"mongo": map[string]interface{}{
		"uri": "mongodb://gzl:gzl@127.0.0.1:25001,127.0.0.1:25002,127.0.0.1:25003/gamedb?w=1&minPoolSize=3&maxPoolSize=10&readPreference=primary",
	},
}
