package service_framework

import (
	"sync"
)

type ClientCallback interface {
	OnClientDisconnect()
}

var clientCallbackMap sync.Map

func registerClientCallback(caddr string, cb ClientCallback) {
	clientCallbackMap.Store(caddr, cb)
}

func unregisterClientCallback(caddr string) {
	clientCallbackMap.Delete(caddr)
}

func getClientCallback(caddr string) ClientCallback {
	if v, ok := clientCallbackMap.Load(caddr); ok { return v.(ClientCallback) }
	return nil
}

func popClientCallback(caddr string) ClientCallback {
	if v, ok := clientCallbackMap.LoadAndDelete(caddr); ok { return v.(ClientCallback) }
	return nil
}
