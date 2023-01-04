package lbtutil

import (
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

// MsgpackRawArray - Start
type MsgpackRawArray []byte

func (raw MsgpackRawArray) HeaderSize() int {
	if raw[0] >= 0x90 && raw[0] <= 0x9F { return 1 }
	if raw[0] == 0xDC { return 3 }
	return 5
}

func (raw MsgpackRawArray) Len() int {
	if (raw[0] >= 0x90 && raw[0] <= 0x9F) {
		return int(raw[0] & 0xF)
	}
	if raw[0] == 0xDC {
		return int(raw[1]) << 8 | int(raw[2])
	}
	return int(raw[1]) << 24 | int(raw[2]) << 16 | int(raw[3]) << 8 | int(raw[4])
}

func (raw MsgpackRawArray) Body() []byte {
	return []byte(raw[raw.HeaderSize():])
}

func (raw MsgpackRawArray) Valid() bool {
	rawSize := len(raw)
	if rawSize == 0 { return false }
	c := raw[0]
	if msgpcode.IsFixedArray(c) { return true }
	switch c {
	case msgpcode.Array16, msgpcode.Array32:
		return true
	}
	return false
}
// MsgpackRawArray - End
