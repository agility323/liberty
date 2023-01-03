package main

import (
	"plugin"

	"github.com/agility323/liberty/lbtreg"
	"github.com/agility323/liberty/hotfix"
	hitf "github.com/agility323/liberty/hotfix/itf"
)

func init() {
	lbtreg.RegisterCmdDataCreator("hotfix", func() lbtreg.CmdData { return &HotfixCmd{} })
	lbtreg.RegisterCmdDataCreator("broadcast", func() lbtreg.CmdData { return &BroadcastCmd{} })
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
	logger.Info("cmd begin %v", cmd)
	cmd.Data.Process()
	logger.Info("cmd end %v", cmd)
}

type HotfixCmd struct {
	lbtreg.HotfixCmdData
}

func (c *HotfixCmd) Process() {
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
	f.(func(hitf.HotfixInterface) error)(hotfix.Hotfix)
}

type BroadcastCmd struct {
	lbtreg.BroadcastCmdData
}

func (c *BroadcastCmd) Process() {
	data, err := makeBroadcastMsgData(c.Method, c.Param)
	if err != nil {
		logger.Error("broadcast fail 1 %v", err)
		return
	}
	clientManager.broadcastMsg(data)
}
