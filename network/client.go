package network

import (
	"fmt"
)

type _Client struct {
	_client *NetClient
}

func NewClient() *_Client {
	return &_Client{
		_client: NewNetClient(nil, nil),
	}
}

func (client *_Client) connect(net string, addr string) error {
	fmt.Println("connect")
	err := client._client.Connect(net, addr)
	if err != nil {
		return err
	}
	return nil
}

// func (client *_Client) addMsgHandler(msgId uint16, handler _MsgHandler) error {
// 	if client._client._handler[msgId] != nil {
// 		return errors.New("already handler binding")
// 	}

// 	client._client._handler[msgId] = handler
// 	return nil
// }
