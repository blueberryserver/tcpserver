package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
)

type Session struct {
	_id   uint16
	_conn net.Conn
}

func (session *Session) Read(data []byte) (int, error) {
	return session._conn.Read(data)
}

func (session *Session) Close() {
	session._conn.Close()
}

type _SessionMap map[uint16]*Session

type _MsgHandler interface {
	execute(*Session, []byte, uint16) bool
}

type _MsgHandlerMap map[uint16]_MsgHandler

//net server
type NetServer struct {
	_net            string
	_addr           string
	_genId          uint16
	_sessions       _SessionMap
	_handler        _MsgHandlerMap
	_connectHandler interface{}
	_recvHandler    interface{}
}

func NewNetServer(net string, addr string, connHandler interface{}, recvHandler interface{}) *NetServer {
	return &NetServer{
		_net:            net,
		_addr:           addr,
		_genId:          0,
		_sessions:       make(_SessionMap),
		_handler:        make(_MsgHandlerMap),
		_connectHandler: connHandler,
		_recvHandler:    recvHandler,
	}
}

func (server *NetServer) RemoveSession(session *Session) {
	server._sessions[session._id] = nil
}

func (server *NetServer) Listen() error {
	ln, err := net.Listen(server._net, server._addr)
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		server._genId++
		session := &Session{
			_id:   server._genId,
			_conn: conn,
		}

		server._sessions[session._id] = session

		if server._connectHandler != nil {
			go server._connectHandler.(func(*NetServer, *Session))(server, session)
		} else {
			go server.handlerConnect(session)
		}
	}
	return nil
}

func (server *NetServer) handlerConnect(session *Session) {
	// 연결 처리
	fmt.Printf("accept session sid:%d\r\n", session._id)
	if server._recvHandler != nil {
		go server._recvHandler.(func(*NetServer, *Session))(server, session)
	} else {
		go server.handlerRecv(session)
	}
}

func RecvHandler(server *NetServer, session *Session) {
	go server._recvHandler.(func(*NetServer, *Session))(server, session)
}

func (server *NetServer) handlerRecv(session *Session) {

	data := make([]byte, 4096)

	for {
		n, err := session._conn.Read(data)
		if err != nil {
			if err != syscall.EINVAL {
				// 연결 종료
				//fmt.Printf("close sid:%d\r\n", session._id)
				session.Close()
				server.RemoveSession(session)
				return
			}
			fmt.Println(err)
			return
		}

		// 프로토콜을 이용해서 메시지 파싱및 핸들러 호출
		server.packetParsing(session, data, n)
	}
}

func (server *NetServer) packetParsing(session *Session, data []byte, bytes int) {
	//fmt.Printf("server recv sid:%d bytes:%d\r\n", session._id, bytes)
	//length := binary.BigEndian.Uint16(data[:2])
	//msgId := binary.BigEndian.Uint16(data[2:4])

	length := binary.LittleEndian.Uint16(data[:2])
	msgId := binary.LittleEndian.Uint16(data[2:4])
	body := data[4:]

	server._handler[msgId].execute(session, body, length)
}

//net client
type NetClient struct {
	_session        *Session
	_handler        _MsgHandlerMap
	_connectHandler interface{}
	_recvHandler    interface{}
}

func NewNetClient(connHandler interface{}, recvHandler interface{}) *NetClient {
	return &NetClient{
		_handler:        make(_MsgHandlerMap),
		_connectHandler: connHandler,
		_recvHandler:    recvHandler,
	}
}

func (client *NetClient) Connect(n string, addr string) error {
	conn, err := net.Dial(n, addr)
	if err != nil {
		return err
	}

	client._session = &Session{
		_id:   0,
		_conn: conn,
	}

	if client._connectHandler != nil {

	} else {
		go client.handlerConnect(client._session)
	}
	return nil
}

func (client *NetClient) Connected() bool {
	if client._session == nil {
		return false
	} else {
		return true
	}
}

func (client *NetClient) handlerConnect(session *Session) {
	// 연결 처리
	fmt.Println("connection complate")
	if client._recvHandler != nil {
		go client._recvHandler.(func(*NetClient, *Session))(client, session)
	} else {
		go client.handlerRecv(session)
	}
}

func (client *NetClient) handlerRecv(session *Session) {

	data := make([]byte, 4096)

	for {
		n, err := session._conn.Read(data)
		if err != nil {

			fmt.Println(err)
			return
		}

		// 프로토콜을 이용해서 메시지 파싱및 핸들러 호출
		client.packetParsing(session, data, n)
	}
}

func (client *NetClient) packetParsing(session *Session, data []byte, bytes int) {
	//fmt.Println("NetClient Recv Data")
	//fmt.Println(string(data[:bytes]))
	length := binary.LittleEndian.Uint16(data[:2])
	msgId := binary.LittleEndian.Uint16(data[2:4])
	body := data[4:]

	client._handler[msgId].execute(session, body, length-4)
}

func (client *NetClient) SendPacket(data []byte) {
	client._session._conn.Write(data)
}

func (client *NetClient) Close() {
	client._session.Close()
	client._session = nil
}

func (session *Session) SendPacket(msgId uint16, data []byte, bytes uint16) error {
	buff := make([]byte, 4096)
	var msgLen uint16
	msgLen = bytes + 4
	binary.LittleEndian.PutUint16(buff[:], msgLen)
	binary.LittleEndian.PutUint16(buff[2:], msgId)
	copy(buff[4:], data)
	session._conn.Write(buff[:msgLen])

	//fmt.Printf("send packet bytes: %d\r\n", msgLen)
	return nil
}

func (session *Session) SendPacketStr(data []byte) error {
	session._conn.Write(data)
	return nil
}
