package lbtnet

import "errors"

var (
	ErrSendInvalidData = errors.New("send fail invalid data")
	ErrSendLongData = errors.New("send fail long data")
	ErrSendInvalidConnection = errors.New("send fail invalid connection")
	ErrSendInvalidChan = errors.New("send fail invalid chan")
	ErrSendChanFull = errors.New("send fail chan full")
	ErrSendClientNotReady = errors.New("send fail client not ready")
	ErrSendHeartbeatExpired = errors.New("send fail heartbeat expired")
)
