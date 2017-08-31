package network

import (
	"encoding/binary"
	_ "fmt"
	"log"
	"net"
	"syscall"
)

//
type Session struct {
	_id   uint16
	_conn net.Conn
}

func (session *Session) Read(data []byte) (int, error) {
	return session._conn.Read(data)
}

//
func (session *Session) Close() {
	session._conn.Close()
}

type _SessionMap map[uint16]*Session

type _MsgHandler interface {
	Execute(*Session, []byte, uint16) bool
}

type _MsgHandlerMap map[int32]_MsgHandler

//net server
type NetServer struct {
	_net            string
	_addr           string
	_genID          uint16
	_sessions       _SessionMap
	_handler        _MsgHandlerMap
	_connectHandler interface{}
	_recvHandler    interface{}
	_closeHandler   interface{}
	_running        bool
	_listener       net.Listener
	_quit           chan bool
}

//
func NewNetServer(net string, addr string, connHandler interface{}, recvHandler interface{}, closeHandler interface{}) *NetServer {

	return &NetServer{
		_net:            net,
		_addr:           addr,
		_genID:          0,
		_sessions:       make(_SessionMap),
		_handler:        make(_MsgHandlerMap),
		_connectHandler: connHandler,
		_recvHandler:    recvHandler,
		_closeHandler:   closeHandler,
		_running:        true,
		_listener:       nil,
		_quit:           make(chan bool),
	}
}

//
func (server *NetServer) RemoveSession(session *Session) {
	server._sessions[session._id] = nil
}

//
func (server *NetServer) Listen(c *chan bool) error {
	ln, err := net.Listen(server._net, server._addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	server._listener = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)

			select {
			case <-server._quit:
				*c <- true
				return nil
			default:
			}
			continue
		}

		server._genID++
		session := &Session{
			_id:   server._genID,
			_conn: conn,
		}

		server._sessions[session._id] = session

		if server._connectHandler != nil {
			go server._connectHandler.(func(*NetServer, *Session))(server, session)
		} else {
			go server.handlerConnect(session)
		}
	}
}

//
func (server *NetServer) Stop() {
	close(server._quit)
	server._listener.Close()
}

func (server *NetServer) handlerConnect(session *Session) {

	log.Printf("accept session sid:%d\r\n", session._id)
	if server._recvHandler != nil {
		go server._recvHandler.(func(*NetServer, *Session))(server, session)
	} else {
		go server.handlerRecv(session)
	}
}

// default recv packet handler
func RecvHandler(server *NetServer, session *Session) {
	go server._recvHandler.(func(*NetServer, *Session))(server, session)
}

func (server *NetServer) handlerRecv(session *Session) {
	data := make([]byte, 4096)
	var pos int

	for {
		n, err := session._conn.Read(data[pos:])
		if err != nil {
			if err != syscall.EINVAL {

				//fmt.Printf("close sid:%d\r\n", session._id)
				session.Close()
				server._closeHandler.(func(*Session))(session)
				server.RemoveSession(session)
				return
			}
			log.Println(err)
			return
		}

		// protocol parsing
		pos = int(server.packetParsing(session, data, n))

		if int(pos) < n {
			remind := data[pos:]
			copy(data, remind)
			pos = n - pos
		} else {
			pos = 0
		}
	}
}

func (server *NetServer) packetParsing(session *Session, data []byte, bytes int) uint16 {
	//fmt.Printf("server recv sid:%d bytes:%d\r\n", session._id, bytes)
	var pos uint16
	for {
		if uint16(bytes)-pos < 4 {
			break
		}
		length := binary.LittleEndian.Uint16(data[pos : pos+2])
		msgID := binary.LittleEndian.Uint16(data[pos+2 : pos+4])

		if uint16(bytes)-pos < length {
			break
		}
		body := data[pos+4 : pos+length]
		if server._handler[int32(msgID)] == nil {
			log.Println("server not find handler msgid:", msgID)
			return pos
		}

		server._handler[int32(msgID)].Execute(session, body, length-4)
		pos = pos + length
		if pos < uint16(bytes) {
			log.Println("multi packet parsing")
		}
	}
	return pos
}

//net client
type NetClient struct {
	_session        *Session
	_handler        _MsgHandlerMap
	_connectHandler interface{}
	_recvHandler    interface{}
}

//
func NewNetClient(connHandler interface{}, recvHandler interface{}) *NetClient {
	return &NetClient{
		_handler:        make(_MsgHandlerMap),
		_connectHandler: connHandler,
		_recvHandler:    recvHandler,
	}
}

//
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

//
func (client *NetClient) Connected() bool {
	if client._session == nil {
		return false
	} else {
		return true
	}
}

func (client *NetClient) handlerConnect(session *Session) {
	// 연결 처리
	log.Println("connection complate")
	if client._recvHandler != nil {
		go client._recvHandler.(func(*NetClient, *Session))(client, session)
	} else {
		go client.handlerRecv(session)
	}
}

func (client *NetClient) handlerRecv(session *Session) {

	log.Println("recving")
	data := make([]byte, 4096)
	var pos int

	for {
		n, err := session._conn.Read(data[pos:])
		if err != nil {
			log.Println(err)
			return
		}

		// protocol parsing
		pos = int(client.packetParsing(session, data, n))

		if int(pos) < n {
			log.Println("parsing remind")
			remind := data[pos:]
			copy(data, remind)
			pos = n - pos
		} else {
			pos = 0
		}
	}
}

func (client *NetClient) packetParsing(session *Session, data []byte, bytes int) uint16 {
	//fmt.Println("NetClient Recv Data len:", bytes )

	var pos uint16
	for {
		if uint16(bytes)-pos < 4 {
			break
		}
		length := binary.LittleEndian.Uint16(data[pos : pos+2])
		msgID := binary.LittleEndian.Uint16(data[pos+2 : pos+4])

		if uint16(bytes)-pos < length {
			break
		}
		body := data[pos+4 : pos+length]

		//fmt.Println("length:", length, " id:", msgID )

		if client._handler[int32(msgID)] == nil {
			log.Println("client not find handler msgid:", msgID)
			return pos
		}

		client._handler[int32(msgID)].Execute(session, body, length-4)
		pos = pos + length
		if pos < uint16(bytes) {
			log.Println("multi packet parsing")
		}
	}
	return pos
}

//
func (client *NetClient) SendPacket(data []byte) {
	client._session._conn.Write(data)
}

//
func (client *NetClient) Close() {
	client._session.Close()
	client._session = nil
}

//
func (session *Session) SendPacket(msgId int32, data []byte, bytes uint16) error {
	buff := make([]byte, 4096)
	var msgLen uint16
	msgLen = bytes + 4
	binary.LittleEndian.PutUint16(buff[:], msgLen)
	binary.LittleEndian.PutUint16(buff[2:], uint16(msgId))
	copy(buff[4:], data)
	session._conn.Write(buff[:msgLen])

	//fmt.Printf("send packet bytes: %d\r\n", msgLen)
	return nil
}

//
func (session *Session) SendPacketStr(data []byte) error {
	session._conn.Write(data)
	return nil
}
