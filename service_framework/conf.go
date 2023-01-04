package service_framework

type ServiceConfType struct {
	LogLevel int `json:"log_level"`
	Host int `json:"host"`
	ServiceType string `json:"service_type"`
	GateServerAddr string `json:"gate_server_addr"`
	ProfilePort int `json:"profile_port"`
	Etcd []string `json:"etcd"`
}

var serviceConf ServiceConfType

func GetServiceConf() *ServiceConfType {
	return &serviceConf
}

func checkServiceConf() bool {
	if serviceConf.ServiceType == "" { return false }
	if serviceConf.GateServerAddr == "" { return false }
	return true
}
