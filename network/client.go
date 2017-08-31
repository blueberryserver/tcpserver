package network

import (
	"encoding/binary"
	"errors"
)

type BlueClient struct {
	// server net session
	client *NetClient
}

// network client
type _Client struct {
	_client *NetClient
}

// create new client
func NewClient() *BlueClient {
	return &BlueClient{
		client: NewNetClient(nil, nil),
	}
}

// network connect
func (client *BlueClient) Connect(net string, addr string) error {
	err := client.client.Connect(net, addr)
	if err != nil {
		return err
	}
	return nil
}

// network add message handler
func (client *BlueClient) AddMsgHandler(msgID int32, handler _MsgHandler) error {
	if client.client._handler[msgID] != nil {
		return errors.New("already handler binding")
	}

	client.client._handler[msgID] = handler
	return nil
}

// network send packet
func (client *BlueClient) SendPacket(msgID int32, data []byte, bytes uint16) error {
	buff := make([]byte, 4096)
	var msgLen uint16
	msgLen = bytes + 4
	binary.LittleEndian.PutUint16(buff[:], msgLen)
	binary.LittleEndian.PutUint16(buff[2:], uint16(msgID))
	copy(buff[4:], data)

	client.client.SendPacket(buff[:msgLen])
	return nil
}

// network close
func (client *BlueClient) Close() {
	client.client.Close()
}

//
func (client *BlueClient) GetSession() *Session {
	return client.client._session
}
