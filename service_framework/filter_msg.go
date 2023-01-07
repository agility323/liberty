package service_framework

import (

	"github.com/agility323/liberty/lbtproto"

	"github.com/vmihailenco/msgpack"
)

type Filter lbtproto.Filter

func RegisterFilterData(id string, fdata map[string]int32) {
	msg := &lbtproto.FilterData{
		Id: id,
		Data: fdata,
	}
	data, err := lbtproto.EncodeMessage(lbtproto.ServiceGate.Method_register_filter_data, msg)
	if err != nil {
		logger.Error("RegisterFilter fail %v", err)
		return
	}
	gates := gateManager.getAllGates()
	for _, c := range gates {
		c.SendData(data)
	}
}

func FilterMsg(method string, params interface{}, filters []*Filter) {
	b, err := msgpack.Marshal(&params)
	if err != nil {
		logger.Error("FilterMsg fail 1 %v", err)
		return
	}
	fs := []*lbtproto.Filter{}
	for _, f := range filters { fs = append(fs, (*lbtproto.Filter)(f)) }
	msg := &lbtproto.FilterMsg{
		Method: method,
		Params: b,
		Filters: fs,
	}
	data, err := lbtproto.EncodeMessage(lbtproto.ServiceGate.Method_filter_msg, msg)
	if err != nil {
		logger.Error("FilterMsg fail 2 %v", err)
		return
	}
	gates := gateManager.getAllGates()
	for _, c := range gates {
		c.SendData(data)
	}
}
