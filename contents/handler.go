package contents

import (
	"log"
	"strconv"
	"time"

	"github.com/blueberryserver/tcpserver/msg"
	"github.com/blueberryserver/tcpserver/network"
	"github.com/blueberryserver/tcpserver/util"
	"github.com/funny/pprof"
	"github.com/golang/protobuf/proto"
	redis "gopkg.in/redis.v4"
)

var _redisClient *redis.Client

var _recorder *pprof.TimeRecorder

// set global redis client
func SetRedisClient(client *redis.Client) {
	_redisClient = client
}

// set golbal time recorder
func SetTimeRecorder(recorder *pprof.TimeRecorder) {
	_recorder = recorder
}

// server handler
//-------------------------------------------------------------------------------------

// session disconnection call function
func CloseHandler(session *network.Session) {
	log.Printf("Server close handler call \r\n")
	user, err := FindUser(session)
	if err != nil {
		log.Println(err)
		return
	}

	// leave channel
	//LeaveCh(user)

	// leave room
	//rm, err := FindRm(user.RmNo)
	//if err == nil {
	//	rm.LeaveMember(user)
	//}

	// logout
	user.Status = UserStatusValue["LOGOFF"]
	user.LogoutTime = time.Now()
	user.Save()
}

// req ping
type ReqPing struct {
	msgID int32
}

// req login
func GetHandlerReqPing() ReqPing {
	return ReqPing{msgID: msg.Msg_Id_value["Ping_Req"]}
}

// req login
func (m ReqPing) Execute(session *network.Session, data []byte, length uint16) bool {
	// unmarshaling
	req := &msg.PingReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.PongAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Pong_Ans"], abuff, uint16(len(abuff)))
		return false
	}

	log.Printf("Server ReqPing msg: %d %s\r\n", m.msgID, req.String())

	// ans
	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.PongAns{
		Err: &errCode,
	}

	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Pong_Ans"], abuff, uint16(len(abuff)))
	return true
}

// req regist
type ReqRegist struct {
	msgID int32
}

// req regist
func GetHandlerReqRegist() ReqRegist {
	return ReqRegist{msgID: msg.Msg_Id_value["Regist_Req"]}
}

// req regist
func (m ReqRegist) Execute(session *network.Session, data []byte, length uint16) bool {
	//t1 := time.Now()
	// unmarshaling
	req := &msg.RegistReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.RegistAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Regist_Ans"], abuff, uint16(len(abuff)))
		return false
	}
	log.Printf("Server ReqRegist msg: %d %s\r\n", m.msgID, req.String())

	// redis query by user id

	pipe := _redisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(1)
	_, _ = pipe.Exec()

	uID, err := _redisClient.HGet("blue_server.user.id", *req.Name).Result()
	if uID != "" {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_EXIST_NAME_FAIL)
		ans := &msg.RegistAns{
			Err: &errCode,
		}

		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Regist_Ans"], abuff, uint16(len(abuff)))
		return false
	}

	// create user obj
	user := NewUser()
	user.ID = UserGenID()
	user.Name = *req.Name
	user.Platform = UserPlatform(*req.Platform)
	user.Status = UserStatusValue["LOGON"]
	user.VcGem = 100
	user.VcGold = 100
	user.CreateTime = time.Now()
	user.LoginTime = time.Now()
	user.Key = util.RandStr(16)
	user.ChNo = 0
	user.RmNo = 0
	err = user.Save()

	// ans
	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.RegistAns{
		Err: &errCode,
	}

	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Regist_Ans"], abuff, uint16(len(abuff)))
	//_recorder.Record("ReqRegist", time.Since(t1))
	return true
}

// req login
type ReqLogin struct {
	msgID int32
}

// req login
func GetHandlerReqLogin() ReqLogin {
	return ReqLogin{msgID: msg.Msg_Id_value["Login_Req"]}
}

// req login
func (m ReqLogin) Execute(session *network.Session, data []byte, length uint16) bool {
	//t1 := time.Now()
	// unmarshaling
	req := &msg.LoginReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.LoginAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Login_Ans"], abuff, uint16(len(abuff)))
		return false
	}
	log.Printf("Server ReqLogin msg: %d %s\r\n", m.msgID, req.String())

	// redis query by user id
	pipe := _redisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(1)
	_, _ = pipe.Exec()

	uID, err := _redisClient.HGet("blue_server.user.id", *req.Id).Result()
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.LoginAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Login_Ans"], abuff, uint16(len(abuff)))
		return false
	}
	id, _ := strconv.Atoi(uID)
	user, err := FindUserByID(uint32(id))
	if err != nil {
		// load User data from redis
		user, err = LoadUser(uint32(id))
		if err != nil {
			log.Println(err)
			errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
			ans := &msg.LoginAns{
				Err: &errCode,
			}
			abuff, _ := proto.Marshal(ans)
			session.SendPacket(msg.Msg_Id_value["Login_Ans"], abuff, uint16(len(abuff)))
			return false
		}
	}
	// print loaded user info
	//log.Println(user.ToString())

	// session binding
	user.Session = session

	// update login time
	user.Status = UserStatusValue["LOGON"]
	user.LoginTime = time.Now()

	// enter default channel(0)
	EnterCh(0, user)

	// save user info
	user.Save()

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
		ChNo:     &user.ChNo,
		RmNo:     &user.RmNo,
	}

	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Login_Ans"], abuff, uint16(len(abuff)))
	//_recorder.Record("ReqLogin", time.Since(t1))
	return true
}

// req relay
type ReqRelay struct {
	msgID int32
}

// req relay
func GetHandlerReqRelay() ReqRelay {
	return ReqRelay{msgID: msg.Msg_Id_value["Relay_Req"]}
}

// req relay
func (m ReqRelay) Execute(session *network.Session, data []byte, length uint16) bool {
	//t1 := time.Now()
	req := &msg.RelayReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.RelayAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Relay_Ans"], abuff, uint16(len(abuff)))
		return false
	}
	log.Printf("Server ReqRelay msg: %d %s \r\n", m.msgID, req.String())

	user, err := FindUser(session)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.RelayAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Relay_Ans"], abuff, uint16(len(abuff)))
		return false
	}

	rm, err := FindRm(user.RmNo)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.RelayAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Relay_Ans"], abuff, uint16(len(abuff)))
		return false
	}
	// room bradcasting
	not := &msg.RelayNot{
		RmNo: req.RmNo,
		Data: req.Data,
	}
	nbuff, _ := proto.Marshal(not)
	rm.Broadcast(msg.Msg_Id_value["Relay_Not"], nbuff, uint16(len(nbuff)))

	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.RelayAns{
		Err: &errCode,
	}
	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Relay_Ans"], abuff, uint16(len(abuff)))

	//session.SendPacket(msg.Msg_Id_value["Relay_Not"], nbuff, uint16(len(nbuff)))
	//_recorder.Record("ReqRelay", time.Since(t1))
	return true
}

// req enter channel
type ReqEnterCh struct {
	msgID int32
}

// req enter channel
func GetHandlerReqEnterCh() ReqEnterCh {
	return ReqEnterCh{msgID: msg.Msg_Id_value["Enter_Ch_Req"]}
}

// req enter channel
func (m ReqEnterCh) Execute(session *network.Session, data []byte, length uint16) bool {
	//t1 := time.Now()
	// unmarshaling
	req := &msg.EnterChReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.EnterChAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Enter_Ch_Ans"], abuff, uint16(len(abuff)))
		return false
	}
	log.Printf("Server ReqEnterCh msg: %d %s\r\n", m.msgID, req.String())

	user, err := FindUser(session)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.EnterChAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Enter_Ch_Ans"], abuff, uint16(len(abuff)))
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
	//_recorder.Record("ReqEnterCh", time.Since(t1))
	return true
}

// req enter room
type ReqEnterRm struct {
	msgID int32
}

// req enter room
func GetHandlerReqEnterRm() ReqEnterRm {
	return ReqEnterRm{msgID: msg.Msg_Id_value["Enter_Rm_Req"]}
}

// req enter channel
func (m ReqEnterRm) Execute(session *network.Session, data []byte, length uint16) bool {
	//t1 := time.Now()
	// unmarshaling
	req := &msg.EnterRmReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.EnterRmAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Enter_Rm_Ans"], abuff, uint16(len(abuff)))
		return false
	}

	log.Printf("Server ReqEnterRm msg: %d %s\r\n", m.msgID, req.String())

	// find user
	user, err := FindUser(session)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.EnterRmAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Enter_Rm_Ans"], abuff, uint16(len(abuff)))
		return false
	}

	err = EnterRm(*req.RmNo, user)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.EnterRmAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Enter_Rm_Ans"], abuff, uint16(len(abuff)))
		return false
	}
	// ans
	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.EnterRmAns{
		Err: &errCode,
	}
	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Enter_Rm_Ans"], abuff, uint16(len(abuff)))
	//_recorder.Record("ReqEtnerRm", time.Since(t1))
	return true
}

// req leave room
type ReqLeaveRm struct {
	msgID int32
}

// req leave room
func GetHandlerReqLeaveRm() ReqLeaveRm {
	return ReqLeaveRm{msgID: msg.Msg_Id_value["Leave_Rm_Req"]}
}

// req leave channel
func (m ReqLeaveRm) Execute(session *network.Session, data []byte, length uint16) bool {
	//t1 := time.Now()
	// unmarshaling
	req := &msg.LeaveRmReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.LeaveRmAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Leave_Rm_Ans"], abuff, uint16(len(abuff)))
		return false
	}

	log.Printf("Server ReqLeaveRm msg: %d %s\r\n", m.msgID, req.String())

	// find user
	user, err := FindUser(session)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.LeaveRmAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Leave_Rm_Ans"], abuff, uint16(len(abuff)))
		return false
	}

	err = LeaveRm(*req.RmNo, user)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.LeaveRmAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["Leave_Rm_Ans"], abuff, uint16(len(abuff)))
		return false
	}
	// ans
	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.LeaveRmAns{
		Err: &errCode,
	}
	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["Leave_Rm_Ans"], abuff, uint16(len(abuff)))
	//_recorder.Record("ReqEtnerRm", time.Since(t1))
	return true
}

//
type ReqListRm struct {
	msgID int32
}

//
func GetHandlerReqListRm() ReqListRm {
	return ReqListRm{msgID: msg.Msg_Id_value["List_Rm_Req"]}
}

//
func (m ReqListRm) Execute(session *network.Session, data []byte, length uint16) bool {
	//t1 := time.Now()
	// unmarshaling
	req := &msg.ListRmReq{}
	err := proto.Unmarshal(data[:length], req)
	if err != nil {
		log.Println(err)
		errCode := msg.ErrorCode(msg.ErrorCode_ERR_SYSTEM_FAIL)
		ans := &msg.ListRmAns{
			Err: &errCode,
		}
		abuff, _ := proto.Marshal(ans)
		session.SendPacket(msg.Msg_Id_value["List_Rm_Ans"], abuff, uint16(len(abuff)))
		return false
	}

	rmList := GetRoomList()
	log.Println(rmList)
	// ans
	errCode := msg.ErrorCode(msg.ErrorCode_ERR_SUCCESS)
	ans := &msg.ListRmAns{
		Err: &errCode,
	}
	ans.RmLists = rmList
	abuff, _ := proto.Marshal(ans)
	session.SendPacket(msg.Msg_Id_value["List_Rm_Ans"], abuff, uint16(len(abuff)))
	//_recorder.Record("ReqListRm", time.Since(t1))
	return true
}

// client handler
//-------------------------------------------------------------------------------------

// ans pong
type AnsPong struct {
	msgID int32
}

// ans pong
func GetHandlerAnsPong() AnsPong {
	return AnsPong{msgID: msg.Msg_Id_value["Pong_Ans"]}
}

// ans pong
func (m AnsPong) Execute(session *network.Session, data []byte, length uint16) bool {
	// and un marshaling
	ans := &msg.PongAns{}
	err := proto.Unmarshal(data[:length], ans)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Printf("Client AnsPong msg: %d %s\r\n", m.msgID, ans.String())
	return true
}

// ans login
type AnsLogin struct {
	msgID int32
}

// ans login
func GetHandlerAnsLogin() AnsLogin {
	return AnsLogin{msgID: msg.Msg_Id_value["Login_Ans"]}
}

// ans login
func (m AnsLogin) Execute(session *network.Session, data []byte, length uint16) bool {
	// and un marshaling
	ans := &msg.LoginAns{}
	err := proto.Unmarshal(data[:length], ans)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Printf("Client AnsLogin msg: %d %s\r\n", m.msgID, ans.String())

	return true
}

// ans relay
type AnsRelay struct {
	msgID int32
}

// ans relay
func GetHandlerAnsRelay() AnsRelay {
	return AnsRelay{msgID: msg.Msg_Id_value["Relay_Ans"]}
}

// ans login
func (m AnsRelay) Execute(session *network.Session, data []byte, length uint16) bool {
	// and un marshaling
	ans := &msg.RelayAns{}
	err := proto.Unmarshal(data[:length], ans)
	if err != nil {
		log.Println(err)
		return false
	}

	log.Printf("Client AnsRelay msg: %d %s\r\n", m.msgID, ans.String())
	return true
}

// not relay
type NotRelay struct {
	msgID int32
}

// not relay
func GetHandlerNotRelay() NotRelay {
	return NotRelay{msgID: msg.Msg_Id_value["Relay_Not"]}
}

// not login
func (m NotRelay) Execute(session *network.Session, data []byte, length uint16) bool {
	// not un marshaling
	not := &msg.RelayNot{}
	err := proto.Unmarshal(data[:length], not)
	if err != nil {
		log.Println(err)
		return false
	}

	log.Printf("Client NotRelay msg: %d %s\r\n", m.msgID, not.String())
	return true
}

// ans enter ch
type AnsEnterCh struct {
	msgID int32
}

// ans enter ch
func GetHandlerAnsEnterCh() AnsEnterCh {
	return AnsEnterCh{msgID: msg.Msg_Id_value["Enter_Ch_Ans"]}
}

// ans enter ch
func (m AnsEnterCh) Execute(session *network.Session, data []byte, length uint16) bool {
	log.Printf("Client AnsEnterCh msg: %d \r\n", m.msgID)
	return true
}

// ans enter rm
type AnsEnterRm struct {
	msgID int32
}

// ans enter rm
func GetHandlerAnsEnterRm() AnsEnterRm {
	return AnsEnterRm{msgID: msg.Msg_Id_value["Enter_Rm_Ans"]}
}

// ans enter rm
func (m AnsEnterRm) Execute(session *network.Session, data []byte, length uint16) bool {
	ans := &msg.EnterRmAns{}
	err := proto.Unmarshal(data[:length], ans)
	if err != nil {
		log.Println(err)
		return false
	}

	log.Printf("Client AnsEnterRm msg: %d %s\r\n", m.msgID, ans.String())
	return true
}

// not enter rm
type NotEnterRm struct {
	msgID int32
}

// not enter rm
func GetHandlerNotEnterRm() NotEnterRm {
	return NotEnterRm{msgID: msg.Msg_Id_value["Enter_Rm_Not"]}
}

// not enter rm
func (m NotEnterRm) Execute(session *network.Session, data []byte, length uint16) bool {
	// not un marshaling
	not := &msg.EnterRmNot{}
	err := proto.Unmarshal(data[:length], not)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Printf("Client NotEnterRm msg: %d %s\r\n", m.msgID, not.String())
	return true
}
