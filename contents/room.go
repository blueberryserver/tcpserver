package contents

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	redis "gopkg.in/redis.v4"
)

// room obj
type Room struct {
	rID     uint32     // room id
	rType   RoomType   // room type ( normal, dual ... )
	rStatus RoomStatus // room status ( none, ready, ... )
	// etc
	members    map[uint32]User
	createTime time.Time
}

// room type
type RoomType uint32

// room type
const (
	_Normal RoomType = 1
	_Dual   RoomType = 2
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
	_None  RoomStatus = 1
	_Ready RoomStatus = 2
)

// room status
var RoomStatusName = map[RoomStatus]string{
	1: "NONE",
	2: "READY",
}

// room status
var RoomStatusValue = map[string]RoomStatus{
	"NONE":  1,
	"READY": 2,
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
		rType:   _Normal,
		rStatus: _None,
		members: make(map[uint32]User)}
}

// set room info
func (rm Room) SetRoom(rid uint32, rtype RoomType) {
	rm.rID = rid
	rm.rType = rtype
}

// add user to room
func (rm Room) AddMember(user User) {
	rm.members[user.ID] = user
}

// load redis db
func LoadRoom(id uint32, client *redis.Client) (Room, error) {

	// redis slelct db 2(room)
	pipe := client.Pipeline()
	defer pipe.Close()
	pipe.Select(2)
	_, err := pipe.Exec()
	if err != nil {
		return Room{}, err
	}

	// hget room
	rID := strconv.Itoa(int(id))
	rType, err := client.HGet("blue_server.room.type", rID).Result()
	if err != nil {
		return Room{}, err
	}
	rStatus, err := client.HGet("blue_server.room.status", rID).Result()
	if err != nil {
		return Room{}, err
	}
	createTime, err := client.HGet("blue_server.room.create.time", rID).Result()
	if err != nil {
		return Room{}, err
	}
	create, err := time.Parse("2006-01-02 15:04:05", createTime)
	if err != nil {
		return Room{}, err
	}
	iStatus := RoomStatusValue[rStatus]
	iType := RoomTypeValue[rType]

	// load member ???

	return Room{
		rID:        id,
		rType:      iType,
		rStatus:    iStatus,
		members:    make(map[uint32]User),
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

	if result == false {
		return errors.New("already set data")
	}

	return nil
}

// to string
func (rm Room) ToString() string {
	var users string
	for _, v := range rm.members {
		users += v.ToString()
	}

	return fmt.Sprintf("%d %s %s %s {%s}", rm.rID, RoomTypeName[rm.rType], RoomStatusName[rm.rStatus],
		rm.createTime.Format("2006-01-02 15:04:05"), users)
}
