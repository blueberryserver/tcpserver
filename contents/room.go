package contents

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/blueberryserver/tcpserver/msg"
	"github.com/golang/protobuf/proto"

	"strings"
)

type RmData struct {
	RmNo       uint32     `json:"rmno"`
	RmType     RoomType   `json:"rmtype"`
	RmStatus   RoomStatus `json:"rmstatus"`
	CreateTime time.Time  `json:"createtime"`
}

// room obj
type Room struct {
	data RmData
	// etc
	members map[uint32]*User
}

// room type
type RoomType uint32

// room type
const (
	_RoomNormal RoomType = 1
	_RoomSolo   RoomType = 2
)

// room type
var RoomTypeName = map[RoomType]string{
	1: "Normal",
	2: "Solo",
}

// room type
var RoomTypeValue = map[string]RoomType{
	"Normal": 1,
	"Solo":   2,
}

// room status
type RoomStatus uint32

// room status
const (
	_RmNone  RoomStatus = 0
	_RmReady RoomStatus = 1
	_RmPlay  RoomStatus = 2
)

// room status
var RoomStatusName = map[RoomStatus]string{
	0: "NONE",
	1: "READY",
	2: "PLAY",
}

// room status
var RoomStatusValue = map[string]RoomStatus{
	"NONE":  0,
	"READY": 1,
	"PLAY":  2,
}

// create new room
func NewRoom() *Room {
	genID := RoomGenID()
	return &Room{
		data: RmData{
			RmNo:       genID,
			RmType:     _RoomNormal,
			RmStatus:   _RmNone,
			CreateTime: time.Now(),
		},
		members: make(map[uint32]*User),
	}
}

// enter room
func (rm Room) EnterMember(user *User) {

	if len(rm.members) > 0 {
		not := &msg.EnterRmNot{}
		not.Names = make([]string, len(rm.members))
		i := 0
		for _, v := range rm.members {
			not.Names[i] = v.Data.Name
			i++
		}

		nbuff, _ := proto.Marshal(not)
		user.Session.SendPacket(msg.Msg_Id_value["Enter_Rm_Not"], nbuff, uint16(len(nbuff)))
	}

	// enter not packet broad cast
	if len(rm.members) > 0 {
		not := &msg.EnterRmNot{}
		not.Names = make([]string, 1)
		not.Names[0] = user.Data.Name
		nbuff, _ := proto.Marshal(not)

		for _, v := range rm.members {
			v.Session.SendPacket(msg.Msg_Id_value["Enter_Rm_Not"], nbuff, uint16(len(nbuff)))
		}
	}

	// add member
	rm.members[user.Data.ID] = user
	user.Data.RmNo = rm.data.RmNo

	log.Println("Enter Room no:", rm.data.RmNo, "member count:", len(rm.members))
}

//leave room
func (rm Room) LeaveMember(user *User) {
	delete(rm.members, user.Data.ID)
	user.Data.RmNo = 0

	// leave not packet broad cast
	if len(rm.members) > 0 {
		not := &msg.LeaveRmNot{}
		not.Names = make([]string, 1)
		not.Names[0] = user.Data.Name
		nbuff, _ := proto.Marshal(not)

		for _, v := range rm.members {
			v.Session.SendPacket(msg.Msg_Id_value["Leave_Rm_Not"], nbuff, uint16(len(nbuff)))
		}
	}

	log.Println("Leave Room no:", rm.data.RmNo, user.Data.Name, "member count:", len(rm.members))
}

// load redis db
func load(id uint32) (*Room, error) {

	log.Println("Load room id:", id)

	// hget room
	rmNo := strconv.Itoa(int(id))
	jsonData, err := userRedisClient.HGet("blue_server.room.json", rmNo).Result()
	if err != nil {
		return &Room{}, err
	}
	rdata := RmData{}
	json.Unmarshal([]byte(jsonData), &rdata)
	return &Room{
		data:    rdata,
		members: make(map[uint32]*User)}, nil
}

// save room redis
func (rm Room) save() error {
	id := strconv.Itoa(int(rm.data.RmNo))
	data, _ := json.Marshal(rm.data)
	result, err := rmchRedisClient.HSet("blue_server.room.json", id, string(data)).Result()
	if err != nil {
		return err
	}

	// member info save
	var members string
	members = "["
	for _, v := range rm.members {
		members += v.Data.Name + ", "
	}
	members = strings.Trim(members, ", ")
	members += "]"

	result, err = rmchRedisClient.HSet("blue_server.room.member", id, members).Result()
	if err != nil {
		return err
	}

	if result == false {
		return errors.New("already set data")
	}

	return nil
}

// broadcasting
func (rm Room) Broadcast(msgID int32, data []byte, bytes uint16) bool {
	for _, ur := range rm.members {
		ur.Session.SendPacket(msgID, data, bytes)
	}
	return true
}

// generate id
func RoomGenID() uint32 {

	genID, _ := rmchRedisClient.Incr("blue_server.manager.room.genid").Result()
	return uint32(genID)
}

//
func FindRm(no uint32) (*Room, error) {
	rmcmd := &RoomCmdData{
		Cmd: "FindRoom",
		No:  no,
	}
	RoomCmd <- rmcmd
	rmcmd = <-RoomCmd
	if rmcmd.Result != nil {
		return nil, rmcmd.Result
	}

	return rmcmd.Room, nil
}

//
func EnterRm(no uint32, rtype uint32, user *User) error {
	// default normal
	if rtype == 0 {
		rtype = 1
	}
	rmcmd := &RoomCmdData{
		Cmd:  "EnterRoom",
		No:   no,
		Type: rtype,
		User: user,
	}
	RoomCmd <- rmcmd
	rmcmd = <-RoomCmd
	if rmcmd.Result != nil {
		return rmcmd.Result
	}

	return nil
}

//
func LeaveRm(no uint32, user *User) error {
	rmcmd := &RoomCmdData{
		Cmd:  "LeaveRoom",
		No:   no,
		User: user,
	}
	RoomCmd <- rmcmd
	rmcmd = <-RoomCmd
	if rmcmd.Result != nil {
		return rmcmd.Result
	}

	return nil
}

//
func GetRoomList() []*msg.ListRmAns_RoomInfo {
	rmcmd := &RoomCmdData{
		Cmd: "ListRoomAns",
	}
	RoomCmd <- rmcmd
	rmcmd = <-RoomCmd
	if rmcmd.Result != nil {
		return nil
	}
	return rmcmd.List
}

//
func LoadRoom() error {
	log.Println("loading room info")
	rmcmd := &RoomCmdData{
		Cmd: "LoadRoom",
	}
	RoomCmd <- rmcmd
	rmcmd = <-RoomCmd
	if rmcmd.Result != nil {
		return rmcmd.Result
	}

	return nil
}

//
func ListRoom() string {
	rmcmd := &RoomCmdData{
		Cmd:     "ListRoom",
		Monitor: "",
	}
	RoomCmd <- rmcmd
	rmcmd = <-RoomCmd
	if rmcmd.Result != nil {
		return ""
	}

	return rmcmd.Monitor
}
