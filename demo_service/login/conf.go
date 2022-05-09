package main

import (
	sf "github.com/agility323/liberty/service_framework"
)

type confType struct {
	Service *sf.ServiceConfType `json:"service"`
}

var Conf confType

var defaultConf = map[string]interface{} {}
