package service_framework

type ClientCallback interface {
	OnClientDisconnect()
}

var clientCallbackMap = make(map[string]ClientCallback)	// TODO: thread safe

func registerClientCallback(caddr string, cb ClientCallback) {
	clientCallbackMap[caddr] = cb
}

func unregisterClientCallback(caddr string) {
	delete(clientCallbackMap, caddr)
}
