package contents

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/blueberryserver/tcpserver/msg"
	"github.com/golang/protobuf/proto"

	"strings"
)

type RMData struct {
	RmNo       uint32     `json:"rmno"`
	RmType     RoomType   `json:"rmtype"`
	RmStatus   RoomStatus `json:"rmstatus"`
	CreateTime time.Time  `json:"createtime"`
}

// room obj
type Room struct {
	data RMData
	// etc
	members map[uint32]*User
}

// room type
type RoomType uint32

// room type
const (
	_RoomNormal RoomType = 1
	_RoomDual   RoomType = 2
)

// room type
var RoomTypeName = map[RoomType]string{
	1: "Normal",
	2: "Dual",
}

// room type
var RoomTypeValue = map[string]RoomType{
	"Normal": 1,
	"Dual":   2,
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

//
func InitRoom() {
	_rooms = make(map[uint32]*Room)
}

var _roomSync sync.Mutex // sync obj
var _rooms map[uint32]*Room

// create new room
func NewRoom() *Room {
	genID := RoomGenID()
	return &Room{
		data: RMData{
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

	log.Println("Leave Room no:", rm.data.RmNo, "member count:", len(rm.members))
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
	rdata := RMData{}
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

//
func EnterRm(rmNo uint32, user *User) error {
	var mutex = &_roomSync
	mutex.Lock()
	defer mutex.Unlock()

	// quick join
	if rmNo == 0 {
		for _, rm := range _rooms {
			if len(rm.members) < 2 {
				rmNo = rm.data.RmNo
				break
			}
		}
	}

	if rmNo == 0 {
		// create room
		rm := NewRoom()
		_rooms[rm.data.RmNo] = rm
		rmNo = rm.data.RmNo
	}

	if _rooms[rmNo] == nil {
		rm, err := load(rmNo)
		if err != nil {
			//return err

			// create room
			rm = NewRoom()
			rmNo = rm.data.RmNo
		}
		_rooms[rmNo] = rm
		rm.EnterMember(user)
		return nil
	}

	_rooms[rmNo].EnterMember(user)
	return nil
}

//
func LeaveRm(rmNo uint32, user *User) error {
	var mu = &_roomSync
	mu.Lock()
	defer mu.Unlock()

	if rmNo == 0 {
		rmNo = user.Data.RmNo
	}

	if rmNo == 0 {
		return errors.New("Not room member")
	}

	if _rooms[rmNo] == nil {
		rm, err := load(rmNo)
		if err != nil {
			return err
		}
		_rooms[rmNo] = rm
		rm.LeaveMember(user)
		return nil
	}

	_rooms[rmNo].LeaveMember(user)

	// room destory
	if len(_rooms[rmNo].members) == 0 {
		//delete(ch.rooms, rmNo)
	}
	return nil
}

//
func FindRm(rmNo uint32) (*Room, error) {
	if _rooms[rmNo] == nil {
		return nil, errors.New("Not find room")
	}
	return _rooms[rmNo], nil
}

// load room from redis
func LoadRoom() {
	log.Println("loading room info")
	_rooms = make(map[uint32]*Room)

	var cursor uint64
	var outputs []string
	outputs, cursor, err := rmchRedisClient.HScan("blue_server.room.json", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		return
	}

	for i := 0; i < len(outputs); i += 2 {
		// redis key
		no := outputs[i]
		rmNo, _ := strconv.Atoi(no)
		// redis value
		rdata := RMData{}
		json.Unmarshal([]byte(outputs[i+1]), &rdata)

		_rooms[uint32(rmNo)] = &Room{
			data:    rdata,
			members: make(map[uint32]*User),
		}
	}
}

//
func GetRoomList() []*msg.ListRmAns_RoomInfo {
	rmList := make([]*msg.ListRmAns_RoomInfo, len(_rooms))
	var mu = &_roomSync
	mu.Lock()
	defer mu.Unlock()
	index := 0
	for _, rm := range _rooms {
		rmList[index] = &msg.ListRmAns_RoomInfo{}
		rmList[index].RmNo = &rm.data.RmNo
		status := uint32(rm.data.RmStatus)
		rmList[index].RmStatus = &status
		rmList[index].Names = make([]string, len(rm.members))

		userIndex := 0
		for _, ur := range rm.members {
			rmList[index].Names[userIndex] = ur.Data.Name
			userIndex++
		}
		index++
	}
	return rmList
}
