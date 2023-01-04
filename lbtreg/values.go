package lbtreg

import (
	"encoding/json"
)

type Value interface {
	Marshal() (string, error)
	Unmarshal([]byte) error
}

type CmdValue struct {
	Cmd string
	Node string
	Param map[string]interface{}
}

func (v *CmdValue) Marshal() (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}

func (v *CmdValue) Unmarshal(b []byte) error {
	return json.Unmarshal(b, v)
}
