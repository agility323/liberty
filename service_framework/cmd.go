package service_framework

import (
	"context"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/agility323/liberty/hotfix"
	hitf "github.com/agility323/liberty/hotfix/itf"
	"github.com/agility323/liberty/lbtreg"
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
	if serviceConf.ServerHotfixPath == "" {
		logger.Error("hotfix fail read path fail, ServerHotfixPath empty")
		return
	}
	files, err := os.ReadDir(serviceConf.ServerHotfixPath)
	if err != nil {
		logger.Error("hotfix fail read path fail %v", err)
		return
	}
	fileName := ""
	for _, file := range files {
		fn := file.Name()
		if !strings.HasPrefix(fn, "hotfix") { continue }
		if !strings.HasSuffix(fn, ".so") { continue }
		if fn > fileName { fileName = fn }
	}
	if fileName == "" {
		logger.Error("hotfix fail fail, has no hotfix*.so")
		return
	}
	logger.Info("begin load hotfix file %s", fileName)
	p, err1 := plugin.Open(filepath.Join(serviceConf.ServerHotfixPath, fileName))
	logger.Info("hotfix plugin open %p", p)
	if err1 != nil {
		logger.Error("hotfix fail 1 %v", err1)
		return
	}
	f, err2 := p.Lookup("Hotfix")
	if err2 != nil {
		logger.Error("hotfix fail 2 %v", err2)
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
		Stop()	// TODO soft stop
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
