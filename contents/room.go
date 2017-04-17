package contents

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/blueberryserver/tcpserver/msg"
	"github.com/golang/protobuf/proto"

	"strings"
)

// room obj
type Room struct {
	rID     uint32     // room id
	rType   RoomType   // room type ( normal, dual ... )
	rStatus RoomStatus // room status ( none, ready, ... )
	// etc
	members    map[uint32]*User
	createTime time.Time
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
	1: "NORMAL",
	2: "Dual",
}

// room type
var RoomTypeValue = map[string]RoomType{
	"NORMAL": 1,
	"Dual":   2,
}

// room status
type RoomStatus uint32

// room status
const (
	_RmNone  RoomStatus = 1
	_RmReady RoomStatus = 2
	_RmPlay  RoomStatus = 3
)

// room status
var RoomStatusName = map[RoomStatus]string{
	1: "NONE",
	2: "READY",
	3: "PLAY",
}

// room status
var RoomStatusValue = map[string]RoomStatus{
	"NONE":  1,
	"READY": 2,
	"PLAY":  3,
}

//
func InitRoom() {
	_rooms = make(map[uint32]*Room)
}

var _sync sync.Mutex // sync obj
var _rooms map[uint32]*Room

// create new room
func NewRoom() *Room {
	genID := RoomGenID()
	return &Room{
		rID:        genID,
		rType:      _RoomNormal,
		rStatus:    _RmNone,
		members:    make(map[uint32]*User),
		createTime: time.Now(),
	}
}

// set room info
func (rm Room) SetRoom(rid uint32, rtype RoomType) {
	rm.rID = rid
	rm.rType = rtype
}

// enter room
func (rm Room) EnterMember(user *User) {

	if len(rm.members) > 0 {
		not := &msg.EnterRmNot{}
		not.Names = make([]string, len(rm.members))
		i := 0
		for _, v := range rm.members {
			not.Names[i] = v.Name
			i++
		}

		nbuff, _ := proto.Marshal(not)
		user.Session.SendPacket(msg.Msg_Id_value["Enter_Rm_Not"], nbuff, uint16(len(nbuff)))
	}

	// enter not packet broad cast
	if len(rm.members) > 0 {
		not := &msg.EnterRmNot{}
		not.Names = make([]string, 1)
		not.Names[0] = user.Name
		nbuff, _ := proto.Marshal(not)

		for _, v := range rm.members {
			v.Session.SendPacket(msg.Msg_Id_value["Enter_Rm_Not"], nbuff, uint16(len(nbuff)))
		}
	}

	// add member
	rm.members[user.ID] = user
	user.RmNo = rm.rID

	log.Println("Enter Room no:", rm.rID, "member count:", len(rm.members))
}

//leave room
func (rm Room) LeaveMember(user *User) {
	delete(rm.members, user.ID)
	user.RmNo = 0

	// leave not packet broad cast
	if len(rm.members) > 0 {
		not := &msg.LeaveRmNot{}
		not.Names = make([]string, 1)
		not.Names[0] = user.Name
		nbuff, _ := proto.Marshal(not)

		for _, v := range rm.members {
			v.Session.SendPacket(msg.Msg_Id_value["Leave_Rm_Not"], nbuff, uint16(len(nbuff)))
		}
	}

	log.Println("Leave Room no:", rm.rID, "member count:", len(rm.members))
}

// set room setatus

// load redis db

func load(id uint32) (*Room, error) {

	log.Println("Load room id:", id)
	// redis slelct db 2(room)
	pipe := _redisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(2)
	_, err := pipe.Exec()
	if err != nil {
		return &Room{}, err
	}

	// hget room
	rID := strconv.Itoa(int(id))
	rType, err := _redisClient.HGet("blue_server.room.type", rID).Result()
	if err != nil {
		return &Room{}, err
	}
	rStatus, err := _redisClient.HGet("blue_server.room.status", rID).Result()
	if err != nil {
		return &Room{}, err
	}
	createTime, err := _redisClient.HGet("blue_server.room.create.time", rID).Result()
	if err != nil {
		return &Room{}, err
	}
	create, err := time.Parse("2006-01-02 15:04:05", createTime)
	if err != nil {
		return &Room{}, err
	}
	iStatus := RoomStatusValue[rStatus]
	iType := RoomTypeValue[rType]

	return &Room{
		rID:        id,
		rType:      iType,
		rStatus:    iStatus,
		members:    make(map[uint32]*User),
		createTime: create}, nil
}

// save room redis
func (rm Room) save() error {
	pipe := _redisClient.Pipeline()
	defer pipe.Close()

	pipe.Select(2)
	_, _ = pipe.Exec()

	id := strconv.Itoa(int(rm.rID))
	result, err := _redisClient.HSet("blue_server.room.type", id, strconv.Itoa(int(rm.rType))).Result()
	if err != nil {
		return err
	}

	result, err = _redisClient.HSet("blue_server.room.status", id, strconv.Itoa(int(rm.rStatus))).Result()
	if err != nil {
		return err
	}

	result, err = _redisClient.HSet("blue_server.room.create.time", id, rm.createTime.Format("2006-01-02 15:04:05")).Result()
	if err != nil {
		return err
	}
	var members string
	members = "["
	for _, v := range rm.members {
		members += v.Name + ", "
	}
	members = strings.Trim(members, ", ")
	members += "]"

	result, err = _redisClient.HSet("blue_server.room.member", id, members).Result()
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

// to string
func (rm Room) ToString() string {
	var users string
	for _, v := range rm.members {
		users += v.ToString() + "\r\n"
	}

	return fmt.Sprintf("%d %s %s %s \r\n{%s}", rm.rID, RoomTypeName[rm.rType], RoomStatusName[rm.rStatus],
		rm.createTime.Format("2006-01-02 15:04:05"), users)
}

//
func EnterRm(rmNo uint32, user *User) error {
	var mu = &_sync
	mu.Lock()
	defer mu.Unlock()

	// quick join
	if rmNo == 0 {
		for _, rm := range _rooms {
			if len(rm.members) < 2 {
				rmNo = rm.rID
				break
			}
		}
	}

	if rmNo == 0 {
		// create room
		rm := NewRoom()
		_rooms[rm.rID] = rm
		rmNo = rm.rID
	}

	if _rooms[rmNo] == nil {
		rm, err := load(rmNo)
		if err != nil {
			return err
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
	var mu = &_sync
	mu.Lock()
	defer mu.Unlock()

	if rmNo == 0 {
		rmNo = user.RmNo
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
	// redis slelct db 2(room, ch)
	pipe := _redisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(2)
	_, err := pipe.Exec()
	if err != nil {
		log.Println(err)
		return
	}

	_rooms = make(map[uint32]*Room)

	var cursor uint64
	var outputs []string
	outputs, cursor, err = _redisClient.HScan("blue_server.room.type", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		return
	}

	for i := 0; i < len(outputs); i += 2 {
		// redis key
		no := outputs[i]
		iNo, _ := strconv.Atoi(no)
		// redis value
		rType := outputs[i+1]
		irType := RoomTypeValue[rType]

		_rooms[uint32(iNo)] = &Room{
			rID:     uint32(iNo),
			rType:   irType,
			members: make(map[uint32]*User),
			//rooms:   make(map[uint32]*Room),
		}
		//log.Println(_channels[uint32(iNo)])
	}
	// room count
	//cursor = 0
	//outputs, cursor, err = _redisClient.HScan("blue_server.ch.room.count", cursor, "", 10).Result()
	//log.Println(outputs)
	// user count
	cursor = 0
	outputs, cursor, err = _redisClient.HScan("blue_server.room.status", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		return
	}
	//log.Println(outputs)
	for i := 0; i < len(outputs); i += 2 {
		// redis key
		no := outputs[i]
		iNo, _ := strconv.Atoi(no)
		// redis value
		rStatus := outputs[i+1]
		irStatus := RoomStatusValue[rStatus]

		_rooms[uint32(iNo)].rStatus = irStatus
	}

	cursor = 0
	outputs, cursor, err = _redisClient.HScan("blue_server.room.create.time", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		return
	}
	//log.Println(outputs)
	for i := 0; i < len(outputs); i += 2 {
		// redis key
		no := outputs[i]
		iNo, _ := strconv.Atoi(no)
		// redis value
		rcreateTime := outputs[i+1]
		createTime, _ := time.Parse("2006-01-02 15:04:05", rcreateTime)
		_rooms[uint32(iNo)].createTime = createTime
	}

	cursor = 0
	outputs, cursor, err = _redisClient.HScan("blue_server.room.member", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		return
	}
	//log.Println(outputs)
	for i := 0; i < len(outputs); i += 2 {
		// redis key
		no := outputs[i]
		_, _ = strconv.Atoi(no)
		// redis value
		_ = outputs[i+1]
		// find member add user
		//log.Println(iNo, rmember)
	}
}
