package service_framework

type serviceConfType struct {
	serviceType string
	gateAddr string
}

var serviceConf serviceConfType

func InitServiceConf(serviceType string, gateAddr string) {
	serviceConf = serviceConfType{
		serviceType: serviceType,
		gateAddr: gateAddr,
	}
}

func checkServiceConf() bool {
	if serviceConf.serviceType == "" { return false }
	if serviceConf.gateAddr == "" { return false }
	return true
}
