package lbtreg

type CmdData interface{
	Process()
}

type BaseCmdData struct {
}

func (*BaseCmdData) Process() {
	panic("BaseCmdData.Process is a pure virtual function")
}

type HotfixCmdData struct {
	BaseCmdData
}

type QuitCmdData struct {
	BaseCmdData
	Addr string	`json:"addr"`
}

type BroadcastCmdData struct {
	BaseCmdData
	Method string `json:"method"`
	Param interface{} `json:"param"`

	/*
	Param *struct {
		Type int `json:"type"`
		Msg string `json:"msg"`
		T int `json:"t"`
		SenderData map[string]interface{} `json:"sender_data"`
	} `json:"param"`
	*/
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
