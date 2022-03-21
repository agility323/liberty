package main

import (
	"github.com/agility323/liberty/lbtutil"
)

type confType struct {
	ServiceType string `json:"service_type"`
	LogLevel int `json:"log_level"`
}

var Conf confType

var defaultConf map[string]interface{} = map[string]interface{} {
	"service_type": "login_service",
	"log_level": lbtutil.Ldebug,
}
