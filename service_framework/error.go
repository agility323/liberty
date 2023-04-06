package service_framework

import "errors"

var (
	ErrRpcInvalidParams = errors.New("rpc fail invalid params")
	ErrRpcNoRoute = errors.New("rpc fail no route")
)
