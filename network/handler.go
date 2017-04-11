package network

import "fmt"

type Req_Ping struct {
	msgId int32
}

func (msg Req_Ping) execute(session *Session, data []byte, length uint16) bool {
	fmt.Printf("Server Req_Ping msg: %d data: %s\r\n", msg.msgId, data[:length])

	session.SendPacket(msg.msgId+1, data, length)
	return true
}
func GetHandler_Req_Ping(msgid int32) Req_Ping {
	return Req_Ping{msgId: msgid}
}

type Req_Login struct {
	msgId int32
}

func (msg Req_Login) execute(session *Session, data []byte, length uint16) bool {
	fmt.Printf("Server Req_Login msg: %d\r\n", msg.msgId)
	return true
}
func GetHandler_Req_Login(msgid int32) Req_Login {
	return Req_Login{msgId: msgid}
}

//-------------------------------------------------------------------------------------

type Ans_Pong struct {
	msgId int32
}

func (msg Ans_Pong) execute(session *Session, data []byte, length uint16) bool {
	fmt.Printf("Client Ans_Pong msg: %d data: %s\r\n", msg.msgId, data[:length])
	return true
}
func GetHandler_Ans_Pong(msgid int32) Ans_Pong {
	return Ans_Pong{msgId: msgid}
}
