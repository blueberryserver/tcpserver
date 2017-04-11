package network

import (
	"errors"
	"sync"
)

type _Server struct {
	// 서버 네트워크 포인터
	_server *NetServer

	// 세션 맵 동기화 객체
	_lockSession sync.Mutex
	// 세션 맵
	//_sessions _SessionMap
}

var __netServet *_Server

func SetGlobalNetServer(server *_Server) {
	__netServet = server
}

func NewServer(net string, addr string) *_Server {
	return &_Server{
		//_sessions: make(_SessionMap),
		_server: NewNetServer(net, addr, nil, nil),
	}
}

func (server *_Server) Listen() error {
	go server._server.Listen()
	return nil
}

func (server *_Server) AddMsgHandler(msgId int32, handler _MsgHandler) error {
	if server._server._handler[msgId] != nil {
		return errors.New("already handler binding")
	}

	server._server._handler[msgId] = handler
	return nil
}
