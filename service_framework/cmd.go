package service_framework

import (
	"context"
	"plugin"

	"github.com/agility323/liberty/lbtreg"
	"github.com/agility323/liberty/hotfix"
	hitf "github.com/agility323/liberty/hotfix/itf"
)

func init() {
	lbtreg.RegisterCmdDataCreator("hotfix", func() lbtreg.CmdData { return &HotfixCmd{} })
	lbtreg.RegisterCmdDataCreator("quit", func() lbtreg.CmdData { return &QuitCmd{} })
}

func OnWatchServiceCmd(typ int, key string, val []byte) {
	logger.Debug("OnWatchServiceCmd %d %s %q", typ, key, val)
	if typ != 0 { return }	// EventTypePut only
	cmd := lbtreg.CmdValue{}
	if err := cmd.Unmarshal(val); err != nil {
		logger.Warn("invalid cmd val %d %s %q %s", typ, key, val, err.Error())
		return
	}
	if serviceConf.ServiceType != cmd.Node { return }
	logger.Info("cmd begin %v", cmd)
	cmd.Data.Process()
	logger.Info("cmd end %v", cmd)
}

type HotfixCmd struct {
	lbtreg.HotfixCmdData
}

func (c *HotfixCmd) Process() {
	p, err := plugin.Open(serviceConf.ServerHotfixPath)
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
	if c.Mode & lbtreg.QuitModeMaskSingleAll <= 0 && c.Addr != serviceConf.GateServerAddr { return }
	if c.Mode & lbtreg.QuitModeMaskSoftHard > 0 {
		Stop()
	} else {
		Stop()	// TODO
	}
}

func GateBroadcast(ctx context.Context, method string, param interface{}) {
	cmd := lbtreg.CmdValue{
		Cmd: "broadcast",
		Node: "",
		Data: &lbtreg.BroadcastCmdData{
			Method: method,
			Param: param,
		},
	}
	host := serviceConf.Host
	lbtreg.PutGateCmd(ctx, host, &cmd)
}
