package network

import (
	"errors"
)

//
type BlueServer struct {
	// server net session
	server *NetServer
	// sync obj
	//_lockSession sync.Mutex
}

//
var _netServet *BlueServer

//
func SetGlobalNetServer(server *BlueServer) {
	_netServet = server
}

//
func NewServer(net string, addr string, closeHandler interface{}) *BlueServer {
	server := NewNetServer(net, addr, nil, nil, closeHandler)
	return &BlueServer{
		server: server,
	}
}

//
func (s *BlueServer) Listen(c *chan bool) {
	go s.server.Listen(c)
}

//
func (s *BlueServer) AddMsgHandler(msgID int32, handler _MsgHandler) error {
	if s.server._handler[msgID] != nil {
		return errors.New("already handler binding")
	}

	s.server._handler[msgID] = handler
	return nil
}

//
func StopServer() {
	_netServet.server.Stop()
}
