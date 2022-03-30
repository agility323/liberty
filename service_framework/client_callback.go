package service_framework

type ClientCallback interface {
	OnClientDisconnect()
}

var clientCallbackMap = make(map[string]ClientCallback)	// TODO: thread safe

func RegisterClientCallback(caddr string, cb ClientCallback) {
	clientCallbackMap[caddr] = cb	// TODO: handle overwrite
}
