package main

import (
	"plugin"

	"github.com/agility323/liberty/lbtreg"
	"github.com/agility323/liberty/hotfix"
)

var cmdMap = map[string]func(map[string]interface{}) {
	"hotfix": CMD_hotfix,
	"broadcast": CMD_broadcast,
}

func OnWatchGateCmd(typ int, key string, val []byte) {
	logger.Debug("OnWatchGateCmd %d %s %q", typ, key, val)
	if typ != 0 { return }	// EventTypePut only
	cmd := lbtreg.CmdValue{}
	if err := cmd.Unmarshal(val); err != nil {
		logger.Warn("invalid cmd val %d %s %q", typ, key, val)
		return
	}
	//if "gate" != cmd.Node { return }
	f, ok := cmdMap[cmd.Cmd]
	if !ok {
		logger.Warn("invalid cmd type %v", cmd)
		return
	}
	logger.Info("cmd begin %v", cmd)
	f(cmd.Param)
	logger.Info("cmd end %v", cmd)
}

func CMD_hotfix(param map[string]interface{}) {
	p, err := plugin.Open("hotfix/hotfix.so")
	if err != nil {
		logger.Error("hotfix fail 1 %v", err)
		return
	}
	f, err := p.Lookup("Hotfix")
	if err != nil {
		logger.Error("hotfix fail 2 %v", err)
		return
	}
	f.(func(hotfix.HotfixInterface) error)(hotfix.Hotfix)
}

func CMD_broadcast(param map[string]interface{}) {
	itf, ok := param["method"]
	if !ok {
		logger.Error("broadcast fail 1 %v", param)
		return
	}
	method, ok := itf.(string)
	if !ok {
		logger.Error("broadcast fail 2 %v", param)
		return
	}
	methodParam, ok := param["param"].([]interface{})
	if !ok {
		logger.Error("broadcast fail 3 %v", param)
		return
	}
	data, err := makeBroadcastMsgData(method, methodParam)
	if err != nil {
		logger.Error("broadcast fail 4 %v", err)
		return
	}
	clientManager.broadcastMsg(data)
}
