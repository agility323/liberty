package service_framework

import (
	itf "github.com/agility323/liberty/service_framework/itf"
)

var serviceConf = itf.ServiceConfType{ServiceRequestTimeout: 20}

func GetServiceConf() *itf.ServiceConfType {
	return &serviceConf
}

func checkServiceConf() bool {
	if serviceConf.ServiceType == "" { return false }
	if serviceConf.GateServerAddr == "" { return false }
	if serviceConf.ServiceRequestTimeout < 1 { return false }
	return true
}
