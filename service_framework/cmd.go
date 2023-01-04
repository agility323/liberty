package service_framework

import (
	"plugin"

	"github.com/agility323/liberty/lbtreg"
	"github.com/agility323/liberty/hotfix"
)

var cmdMap = map[string]func(map[string]interface{}) {
	"hotfix": CMD_hotfix,
}

func OnWatchServiceCmd(typ int, key string, val []byte) {
	logger.Debug("OnWatchServiceCmd %d %s %q", typ, key, val)
	if typ != 0 { return }	// EventTypePut only
	cmd := lbtreg.CmdValue{}
	if err := cmd.Unmarshal(val); err != nil {
		logger.Warn("invalid cmd val %d %s %q", typ, key, val)
		return
	}
	if serviceConf.ServiceType != cmd.Node { return }
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
