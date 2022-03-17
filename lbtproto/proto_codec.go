package lbtproto

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/gogo/protobuf/proto"
)

const (
	IndexHeader uint32 = 4
	IndexBody uint32 = 6
	HeaderSize uint32 = IndexBody - IndexHeader
)

type MessageSender interface {
	SendData([]byte) error
}

func DecodeMethodIndex(buf []byte, order binary.ByteOrder) (uint16, error) {
	var methodIndex uint16 = 0
	if err := binary.Read(bytes.NewReader(buf[IndexHeader:IndexBody]), order, &methodIndex); err != nil {
		return 0, errors.New("lbtproto.DecodeMethodIndex fail - " + err.Error())
	}
	return methodIndex, nil
}

func DecodeMessage(buf []byte, msg proto.Message) error {
	return proto.Unmarshal(buf[IndexBody:], msg)
}

func EncodeMessage(methodIndex uint16, msg proto.Message, order binary.ByteOrder) ([]byte, error) {
	body, err := proto.Marshal(msg)
	if err != nil {
		return nil, errors.New("lbtproto.EncodeMessage fail 1 - " + err.Error())
	}
	size := HeaderSize + uint32(len(body))
	buffer := bytes.NewBuffer(make([]byte, 0, size + IndexHeader))
	binary.Write(buffer, order, size)
	binary.Write(buffer, order, methodIndex)
	binary.Write(buffer, order, body)
	data := buffer.Bytes()
	return data, nil
}

func SendMessage(sender MessageSender, methodIndex uint16, msg proto.Message, order binary.ByteOrder) error {
	// encode proto
	data, err := EncodeMessage(methodIndex, msg, order)
	if err != nil {
		return err
	}
	// send message
	//logger.Debug("lbtproto.SendMessage %v", data)
	sender.SendData(data)
	return nil
}
