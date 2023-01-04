package lbtreg

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Value interface {
	Marshal() (string, error)
	Unmarshal([]byte) error
}

type CmdValue struct {
	Cmd string	`json:"cmd"`
	Node string	`json:"node"`
	Data CmdData	`json:"data"`
}

func (v *CmdValue) Marshal() (string, error) {
	b, err := json.Marshal([]interface{}{v.Cmd, v.Node, v.Data})
	return string(b), err
}

func (v *CmdValue) Unmarshal(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	// decode token
	if _, err := dec.Token(); err != nil { return err }
	// decode cmd
	if !dec.More() { return nil }
	if err := dec.Decode(&(v.Cmd)); err != nil { return err }
	// decode node
	if !dec.More() { return nil }
	if err := dec.Decode(&(v.Node)); err != nil { return err }
	// decode data
	if v.Data = CreateCmdData(v.Cmd); v.Data == nil { return fmt.Errorf("CmdValue decode data fail 1") }
	if err := dec.Decode(v.Data); err != nil { return err }
	return nil
}
