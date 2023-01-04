package lbtreg

import (
	"encoding/json"
)

type Value interface {
	Marshal() (string, error)
	Unmarshal([]byte) error
}

type CmdValue struct {
	Cmd string	`json:"cmd"`
	Node string	`json:"node"`
	Param map[string]interface{}	`json:"param"`
}

func (v *CmdValue) Marshal() (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}

func (v *CmdValue) Unmarshal(b []byte) error {
	return json.Unmarshal(b, v)
}
