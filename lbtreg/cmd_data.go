package lbtreg

type CmdData interface{
	Process()
}

var cmdDataCreatorMap = map[string]func() CmdData{}

func RegisterCmdDataCreator(cmd string, hc func() CmdData) {
	cmdDataCreatorMap[cmd] = hc
}

func CreateCmdData(cmd string) CmdData {
	hc, ok := cmdDataCreatorMap[cmd]
	if ok && hc != nil { return hc() }
	return nil
}

type BaseCmdData struct {
}

func (*BaseCmdData) Process() {
	panic("BaseCmdData.Process is a pure virtual function")
}

type HotfixCmdData struct {
	BaseCmdData
}

const (
	QuitModeMaskSoftHard = 1 << iota
	QuitModeMaskSingleAll
)

type QuitCmdData struct {
	BaseCmdData
	Mode int `json:"mode"`
	Addr string	`json:"addr"`
}

type BroadcastCmdData struct {
	BaseCmdData
	Method string `json:"method"`
	Param interface{} `json:"param"`
}

type BanServiceCmdData struct {
	BaseCmdData
	Type string `json:"type"`
	Addr string `json:"addr"`
}
