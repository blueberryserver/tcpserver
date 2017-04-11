package network

import (
	"encoding/binary"
	"errors"
)

type _Client struct {
	_client *NetClient
}

func NewClient() *_Client {
	return &_Client{
		_client: NewNetClient(nil, nil),
	}
}

func (client *_Client) Connect(net string, addr string) error {
	err := client._client.Connect(net, addr)
	if err != nil {
		return err
	}
	return nil
}

func (client *_Client) AddMsgHandler(msgId int32, handler _MsgHandler) error {
	if client._client._handler[msgId] != nil {
		return errors.New("already handler binding")
	}

	client._client._handler[msgId] = handler
	return nil
}

func (client *_Client) SendPacket(msgId int32, data []byte, bytes uint16) error {
	buff := make([]byte, 4096)
	var msgLen uint16
	msgLen = bytes + 4
	binary.LittleEndian.PutUint16(buff[:], msgLen)
	binary.LittleEndian.PutUint16(buff[2:], uint16(msgId))
	copy(buff[4:], data)

	client._client.SendPacket(buff[:msgLen])
	return nil
}
