package service_framework

import (
	"plugin"

	"github.com/agility323/liberty/lbtreg"
	"github.com/agility323/liberty/hotfix"
	"github.com/agility323/liberty/hotfix/itf"
)

var cmdMap = map[string]func(map[string]interface{}) {
	"hotfix": CMD_hotfix,
	"quit": CMD_quit,
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
	f.(func(itf.HotfixInterface) error)(hotfix.Hotfix)
}

func CMD_quit(param map[string]interface{}) {
	Stop()
}
