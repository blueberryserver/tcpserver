package contents

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/blueberryserver/tcpserver/msg"
	"github.com/golang/protobuf/proto"

	"strings"

	redis "gopkg.in/redis.v4"
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

// generator number
var _genNo uint32

// init gen number
func InitRoom() {
	_genNo = 0
}

// create new room
func NewRoom() Room {
	_genNo++
	return Room{
		rID:     _genNo,
		rType:   _RoomNormal,
		rStatus: _RmNone,
		members: make(map[uint32]*User)}
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

	fmt.Println("Enter Room no:", rm.rID, "member count:", len(rm.members))
}

//leave room
func (rm Room) LeaveMember(user *User) {
	delete(rm.members, user.ID)

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

	fmt.Println("Leave Room no:", rm.rID, "member count:", len(rm.members))
}

// set room setatus

// load redis db
func LoadRoom(id uint32, client *redis.Client) (*Room, error) {

	fmt.Println("Load room id:", id)
	// redis slelct db 2(room)
	pipe := client.Pipeline()
	defer pipe.Close()
	pipe.Select(2)
	_, err := pipe.Exec()
	if err != nil {
		return &Room{}, err
	}

	// hget room
	rID := strconv.Itoa(int(id))
	rType, err := client.HGet("blue_server.room.type", rID).Result()
	if err != nil {
		return &Room{}, err
	}
	rStatus, err := client.HGet("blue_server.room.status", rID).Result()
	if err != nil {
		return &Room{}, err
	}
	createTime, err := client.HGet("blue_server.room.create.time", rID).Result()
	if err != nil {
		return &Room{}, err
	}
	create, err := time.Parse("2006-01-02 15:04:05", createTime)
	if err != nil {
		return &Room{}, err
	}
	iStatus := RoomStatusValue[rStatus]
	iType := RoomTypeValue[rType]

	// load member ???

	return &Room{
		rID:        id,
		rType:      iType,
		rStatus:    iStatus,
		members:    make(map[uint32]*User),
		createTime: create}, nil
}

// save room redis
func (rm Room) Save(client *redis.Client) error {
	pipe := client.Pipeline()
	defer pipe.Close()

	pipe.Select(2)
	_, _ = pipe.Exec()

	id := strconv.Itoa(int(rm.rID))
	result, err := client.HSet("blue_server.room.type", id, strconv.Itoa(int(rm.rType))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.room.status", id, strconv.Itoa(int(rm.rStatus))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.room.create.time", id, rm.createTime.Format("2006-01-02 15:04:05")).Result()
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

	result, err = client.HSet("blue_server.room.member", id, members).Result()
	if err != nil {
		return err
	}

	if result == false {
		return errors.New("already set data")
	}

	return nil
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
