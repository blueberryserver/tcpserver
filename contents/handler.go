package contents

import (
	"fmt"
	"strconv"
	"time"

	"github.com/blueberryserver/tcpserver/msg"
	"github.com/blueberryserver/tcpserver/network"
	"github.com/golang/protobuf/proto"
	redis "gopkg.in/redis.v4"
)

var _redisClient *redis.Client

// set global redis client
func SetRedisClient(client *redis.Client) {
	_redisClient = client
}

// server handler
//-------------------------------------------------------------------------------------

//
func CloseHandler(session *network.Session) {
	fmt.Printf("Server close handler call \r\n")
	user, err := FindUser(session)
	if err != nil {
		fmt.Println(err)
		return
	}

	LeaveCh(user)
}

// req ping
type ReqPing struct {
	msgID int32
}

// req login
func GetHandlerReqPing(msgid int32) ReqPing {
	return ReqPing{msgID: msgid}
}

// req login
func (m ReqPing) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Server ReqPing msg: %d data: %s\r\n", m.msgID, data[:length])

	session.SendPacket(msg.Msg_Id_value["Pong_Ans"], data, length)
	return true
}

// req regist
type ReqRegist struct {
	msgID int32
}

// req regist
func GetHandlerReqRegist(msgid int32) ReqRegist {
	return ReqRegist{msgID: msgid}
}

// req regist
func (m ReqRegist) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Server ReqPing msg: %d\r\n", m.msgID)

	// unmarshaling
	req := &msg.RegistReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// create user obj
	user := NewUser()
	user.ID = GenID()
	user.Name = *req.Name
	user.Platform = UserPlatform(*req.Platform)
	user.Status = UserStatusValue["LOGON"]
	user.VcGem = 0
	user.VcGold = 0
	user.CreateTime = time.Now()
	user.LoginTime = time.Now()
	err = user.Save(_redisClient)

	// ans
	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.RegistAns{
		Err: &errCode,
	}

	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Regist_Ans"], abuff, uint16(len(abuff)))
	return true
}

// req login
type ReqLogin struct {
	msgID int32
}

// req login
func GetHandlerReqLogin(msgid int32) ReqLogin {
	return ReqLogin{msgID: msgid}
}

// req login
func (m ReqLogin) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Server ReqLogin msg: %d \r\n", m.msgID)

	// unmarshaling
	req := &msg.LoginReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// redis query by user id
	pipe := _redisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(1)
	_, _ = pipe.Exec()

	uID, err := _redisClient.HGet("blue_server.user.id", *req.Id).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}

	// load User data from redis
	id, _ := strconv.Atoi(uID)
	user, err := LoadUser(uint32(id), _redisClient)
	if err != nil {
		fmt.Println(err)
		return false
	}
	// print loaded user info
	fmt.Println(user.ToString())

	// session binding
	user.Session = session

	// enter default channel(0)
	EnterCh(0, user)

	// ans
	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	platform := uint32(user.Platform)
	ans := &msg.LoginAns{
		Err:      &errCode,
		Id:       &user.ID,
		Name:     &user.Name,
		Platform: &platform,
		Gem:      &user.VcGem,
		Gold:     &user.VcGold,
		SecKey:   &user.Key,
	}

	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Login_Ans"], abuff, uint16(len(abuff)))
	return true
}

// req relay
type ReqRelay struct {
	msgID int32
}

// req relay
func GetHandlerReqRelay(msgid int32) ReqRelay {
	return ReqRelay{msgID: msgid}
}

// req relay
func (m ReqRelay) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Server ReqRelay msg: %d \r\n", m.msgID)

	req := &msg.RelayReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.RelayAns{
		Err: &errCode,
	}

	fmt.Println(req)

	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Relay_Ans"], abuff, uint16(len(abuff)))

	// room bradcasting
	not := &msg.RelayNot{
		RmNo: req.RmNo,
		Data: req.Data,
	}

	nbuff, _ := proto.Marshal(not)
	session.SendPacket(msg.Msg_Id_value["Relay_Not"], nbuff, uint16(len(nbuff)))
	return true
}

// req enter channel
type ReqEnterCh struct {
	msgID int32
}

// req enter channel
func GetHandlerReqEnterCh(msgid int32) ReqEnterCh {
	return ReqEnterCh{msgID: msgid}
}

// req enter channel
func (m ReqEnterCh) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Server ReqEnterCh msg: %d \r\n", m.msgID)

	// unmarshaling
	req := &msg.EnterChReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	user, err := FindUser(session)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// channel enter
	EnterCh(*req.ChNo, user)

	// ans
	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.EnterChAns{
		Err: &errCode,
	}
	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Enter_Ch_Ans"], abuff, uint16(len(abuff)))
	return true
}

// req enter room
type ReqEnterRm struct {
	msgID int32
}

// req enter room
func GetHandlerReqEnterRm(msgid int32) ReqEnterRm {
	return ReqEnterRm{msgID: msgid}
}

// req enter channel
func (m ReqEnterRm) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Server ReqEnterRm msg: %d \r\n", m.msgID)

	// unmarshaling
	req := &msg.EnterRmReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// find user
	user, err := FindUser(session)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// find channel
	ch, err := FindCh(user.ChNo)
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = ch.EnterRm(*req.RmNo, user)
	if err != nil {
		fmt.Println(err)
		return false
	}
	// ans
	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.EnterRmAns{
		Err: &errCode,
	}
	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Enter_Rm_Ans"], abuff, uint16(len(abuff)))
	return true
}

// client handler
//-------------------------------------------------------------------------------------

// ans pong
type AnsPong struct {
	msgID int32
}

// ans pong
func GetHandlerAnsPong(msgid int32) AnsPong {
	return AnsPong{msgID: msgid}
}

// ans pong
func (m AnsPong) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Client Ans_Pong msg: %d data: %s\r\n", m.msgID, data[:length])
	return true
}

// ans login
type AnsLogin struct {
	msgID int32
}

// ans login
func GetHandlerAnsLogin(msgid int32) AnsLogin {
	return AnsLogin{msgID: msgid}
}

// ans login
func (m AnsLogin) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Client Ans_Login msg: %d \r\n", m.msgID)

	// and un marshaling
	ans := &msg.LoginAns{}
	err := proto.Unmarshal(data[:length], ans)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println(ans)
	return true
}

// ans relay
type AnsRelay struct {
	msgID int32
}

// ans relay
func GetHandlerAnsRelay(msgid int32) AnsRelay {
	return AnsRelay{msgID: msgid}
}

// ans login
func (m AnsRelay) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Client Ans_Relay msg: %d \r\n", m.msgID)

	// and un marshaling
	ans := &msg.RelayAns{}
	err := proto.Unmarshal(data[:length], ans)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println(ans)
	return true
}

// not relay
type NotRelay struct {
	msgID int32
}

// not relay
func GetHandlerNotRelay(msgid int32) NotRelay {
	return NotRelay{msgID: msgid}
}

// not login
func (m NotRelay) Execute(session *network.Session, data []byte, length uint16) bool {
	fmt.Printf("Client Not_Relay msg: %d \r\n", m.msgID)

	// not un marshaling
	not := &msg.RelayNot{}
	err := proto.Unmarshal(data[:length], not)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println(not)
	return true
}
