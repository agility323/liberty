package lbtproto

import (
	"testing"
	"fmt"

	//"github.com/vmihailenco/msgpack"
)

func TestProto(t *testing.T) {
	buf := []byte{88, 0, 0, 0, 235, 3, 10, 15, 49, 50, 55, 46, 48, 46, 48, 46, 49, 58, 51, 57, 53, 53, 52, 18, 12, 97, 204, 55, 211, 225, 56, 35, 125, 196, 184, 130, 79, 26, 6, 65, 118, 97, 116, 97, 114, 34, 45, 130, 162, 69, 67, 130, 162, 105, 100, 172, 97, 204, 55, 211, 225, 56, 35, 125, 196, 184, 130, 79, 163, 116, 121, 112, 166, 65, 118, 97, 116, 97, 114, 164, 110, 97, 109, 101, 167, 116, 101, 115, 116, 49, 50, 51}
	msg := &EntityData{}
	if err := DecodeMessage(buf, msg); err != nil {
		fmt.Printf("EntityData decode fail: [%s]\n", err.Error())
	}
	fmt.Printf("EntityData: %v\n", *msg)
}
