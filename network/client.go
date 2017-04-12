package network

import (
	"encoding/binary"
	"errors"
)

// network client
type _Client struct {
	_client *NetClient
}

// create new client
func NewClient() *_Client {
	return &_Client{
		_client: NewNetClient(nil, nil),
	}
}

// network connect
func (client *_Client) Connect(net string, addr string) error {
	err := client._client.Connect(net, addr)
	if err != nil {
		return err
	}
	return nil
}

// network add message handler
func (client *_Client) AddMsgHandler(msgID int32, handler _MsgHandler) error {
	if client._client._handler[msgID] != nil {
		return errors.New("already handler binding")
	}

	client._client._handler[msgID] = handler
	return nil
}

// network send packet
func (client *_Client) SendPacket(msgID int32, data []byte, bytes uint16) error {
	buff := make([]byte, 4096)
	var msgLen uint16
	msgLen = bytes + 4
	binary.LittleEndian.PutUint16(buff[:], msgLen)
	binary.LittleEndian.PutUint16(buff[2:], uint16(msgID))
	copy(buff[4:], data)

	client._client.SendPacket(buff[:msgLen])
	return nil
}

// network close
func (client *_Client) Close() {
	client._client.Close()
}
