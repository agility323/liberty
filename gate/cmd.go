package main

import (
	"github.com/vmihailenco/msgpack"
	"plugin"

	"github.com/agility323/liberty/hotfix"
	hitf "github.com/agility323/liberty/hotfix/itf"
	"github.com/agility323/liberty/lbtreg"
)

func init() {
	lbtreg.RegisterCmdDataCreator("hotfix", func() lbtreg.CmdData { return &HotfixCmd{} })
	lbtreg.RegisterCmdDataCreator("quit", func() lbtreg.CmdData { return &QuitCmd{} })
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

type QuitCmd struct {
	lbtreg.QuitCmdData
}

func(c *QuitCmd) Process() {
	if c.Addr != Conf.ClientServerAddr { return }
	InitiateSoftStop()
}

type BroadcastCmd struct {
	lbtreg.BroadcastCmdData
}

func (c *BroadcastCmd) Process() {
	paramBytes, err := msgpack.Marshal(c.Param)
	if err != nil {
		logger.Error("broadcast msgpack marshal err 1 %v", c.Param)
		return
	}
	data, err := makeBroadcastMsgData(c.Method, paramBytes)
	if err != nil {
		logger.Error("broadcast fail 1 %v", err)
		return
	}
	clientManager.broadcastMsg(data)
}

type BanServiceCmd struct {
	lbtreg.BanServiceCmdData
}

func (c *BanServiceCmd) Process() {
	serviceManager.banService(c.Type, c.Addr)
}
