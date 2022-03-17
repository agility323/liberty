package service_framework

import (
	"github.com/vmihailenco/msgpack"
)

type MethodHandler interface {
	GetRequest() interface{}
	GetReply() interface{}
	Process(string, string) error
}

type methodHandlerCreator func() MethodHandler

type defaultMethodHandler struct {
	request []byte
	reply []byte
}
func (h *defaultMethodHandler) GetRequest() interface{} {return &(h.request)}
func (h *defaultMethodHandler) GetReply() interface{} {return &(h.reply)}
func (h *defaultMethodHandler) Process(conAddr, srcAddr string) error {
	h.reply = h.request
	return nil
}
func defaultMethodHandlerCreator() MethodHandler {
	return new(defaultMethodHandler)
}

var methodHandlerCreatorMap map[string]methodHandlerCreator = make(map[string]methodHandlerCreator)

func RegisterMethodHandlerCreator(method string, mhc methodHandlerCreator) {
	if _, ok := methodHandlerCreatorMap[method]; ok {
		logger.Warn("overwrite method handler creator of %s", method)
	}
	methodHandlerCreatorMap[method] = mhc
}

func getMethodHandlerCreator(method string) methodHandlerCreator {
	if mhc, ok := methodHandlerCreatorMap[method]; ok { return mhc }
	logger.Warn("method handler not found, use default %s", method)
	return defaultMethodHandlerCreator
}

func processMethod(conAddr, srcAddr string, method string, params []byte) ([]byte, error) {
	mhc := getMethodHandlerCreator(method)
	handler := mhc()
	if err := msgpack.Unmarshal(params, handler.GetRequest()); err != nil {
		return nil, err
	}
	logger.Debug("process method %s request %s", method, handler.GetRequest())
	if err := handler.Process(conAddr, srcAddr); err != nil {
		return nil, err
	}
	reply := handler.GetReply()
	logger.Debug("process method %s reply %v", method, reply)
	if reply == nil {
		return nil, nil
	}
	replyData, err := msgpack.Marshal(reply)
	if err != nil {
		return nil, err
	}
	return replyData, nil
}
