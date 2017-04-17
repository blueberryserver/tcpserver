package network

import (
	"errors"
	"sync"
)

type _Server struct {
	// server net session
	_server *NetServer

	// sync obj
	_lockSession sync.Mutex
}

//
var _netServet *_Server

//
func SetGlobalNetServer(server *_Server) {
	_netServet = server
}

//
func NewServer(net string, addr string, closeHandler interface{}) *_Server {
	return &_Server{
		_server: NewNetServer(net, addr, nil, nil, closeHandler),
	}
}

//
func (server *_Server) Listen() error {
	c := make(chan bool)
	go server._server.Listen(c)

	_ = <-c
	return nil
}

//
func (server *_Server) Stop() {
	server.Stop()
}

//
func (server *_Server) AddMsgHandler(msgID int32, handler _MsgHandler) error {
	if server._server._handler[msgID] != nil {
		return errors.New("already handler binding")
	}

	server._server._handler[msgID] = handler
	return nil
}
