package contents

import (
	"errors"
	"fmt"
	"sync"

	"strconv"

	"github.com/blueberryserver/tcpserver/network"
)

// channel
type Channel struct {
	no      uint32 // channel number
	chType  ChType
	sync    sync.Mutex       // sync obj
	members map[uint32]*User // channel user
	rooms   map[uint32]*Room // channel room
}

// channel type
type ChType uint32

// user status
const (
	_ChDefault ChType = 0
	_ChNormal  ChType = 10
	_ChLevel1  ChType = 1
	_ChLevel2  ChType = 2
)

// room status
var ChTypeName = map[ChType]string{
	0:  "DEFAULT",
	10: "NORMAL",
	1:  "LEVEL1",
	2:  "LEVEL2",
}

// room status
var ChTypeValue = map[string]ChType{
	"DEFAULT": 0,
	"NORMAL":  10,
	"LEVEL1":  1,
	"LEVEL2":  2,
}

var _channels map[uint32]*Channel

// generate channel
func NewChannel() {
	_channels = make(map[uint32]*Channel)

	for i := 0; i < 2; i++ {
		_channels[uint32(i)] = &Channel{
			no:      uint32(i),
			chType:  _ChDefault,
			members: make(map[uint32]*User),
			rooms:   make(map[uint32]*Room),
		}
	}
}

// load channel from redis
func LoadChannel() {
	// redis slelct db 2(room, ch)
	pipe := _redisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(2)
	_, err := pipe.Exec()
	if err != nil {
		fmt.Println(err)
		return
	}

	_channels = make(map[uint32]*Channel)

	var cursor uint64
	var outputs []string
	outputs, cursor, err = _redisClient.HScan("blue_server.ch.type", cursor, "", 10).Result()
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < len(outputs); i += 2 {
		// redis key
		no := outputs[i]
		iNo, _ := strconv.Atoi(no)
		// redis value
		chType := outputs[i+1]
		iChType := ChTypeValue[chType]

		_channels[uint32(iNo)] = &Channel{
			no:      uint32(iNo),
			chType:  iChType,
			members: make(map[uint32]*User),
			rooms:   make(map[uint32]*Room),
		}
		//fmt.Println(_channels[uint32(iNo)])
	}
	// room count
	cursor = 0
	outputs, cursor, err = _redisClient.HScan("blue_server.ch.room.count", cursor, "", 10).Result()
	//fmt.Println(outputs)
	// user count
	cursor = 0
	outputs, cursor, err = _redisClient.HScan("blue_server.ch.user.count", cursor, "", 10).Result()
	//fmt.Println(outputs)
}

// save channel to redis
func saveChannel() {
	pipe := _redisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(2)
	_, err := pipe.Exec()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, ch := range _channels {
		var mu = &ch.sync
		mu.Lock()
		defer mu.Unlock()
		_, err := _redisClient.HSet("blue_server.ch.type", strconv.Itoa(int(ch.no)), ChTypeName[ch.chType]).Result()
		if err != nil {
			fmt.Println(err)
			continue
		}
		roomCount := len(ch.rooms)
		_, err = _redisClient.HSet("blue_server.ch.room.count", strconv.Itoa(int(ch.no)), strconv.Itoa(roomCount)).Result()
		if err != nil {
			fmt.Println(err)
			continue
		}
		userCount := len(ch.members)
		_, err = _redisClient.HSet("blue_server.ch.user.count", strconv.Itoa(int(ch.no)), strconv.Itoa(userCount)).Result()
		if err != nil {
			fmt.Println(err)
			continue
		}

		// save room info
		for _, rm := range ch.rooms {
			rm.Save(_redisClient)
		}
	}
}

// enter channel
func EnterCh(chNo uint32, user *User) bool {
	if int(chNo) > len(_channels) || chNo < 0 {
		return false
	}

	fmt.Println("Enter channel no:", chNo, "user:", user.Name)
	_channels[chNo].members[user.ID] = user

	user.ChNo = chNo
	return true
}

// leave channel
func LeaveCh(user *User) {
	fmt.Println("Leave channel no:", user.ChNo, "user:", user.Name, "member count:", len(_channels[0].members))
	//leave defualt channel
	delete(_channels[0].members, user.ID)
	fmt.Println("Remind channel no: 0 member count:", len(_channels[0].members))

	var mu = &_channels[user.ChNo].sync
	mu.Lock()
	defer mu.Unlock()

	if user.RmNo != 0 && user.ChNo != 0 {
		_channels[user.ChNo].rooms[user.RmNo].LeaveMember(user)
	}
	// leave current channel
	if user.ChNo != 0 {
		delete(_channels[user.ChNo].members, user.ID)
	}
}

// move channel
func MoveCh(chNo uint32, user *User) {
	fmt.Println("Move channel no:", chNo, "user:", user.Name)

	// leave current channel
	delete(_channels[user.ChNo].members, user.ID)

	// enter channel
	EnterCh(chNo, user)
}

// find user
func FindUser(session *network.Session) (*User, error) {
	for _, v := range _channels[0].members {
		if v.Session == session {
			return v, nil
		}
	}
	return &User{}, errors.New("Not find user session")
}

//
func FindCh(chNo uint32) (*Channel, error) {
	if _channels[chNo] == nil {
		return nil, errors.New("Not find channel")
	}
	return _channels[chNo], nil
}

//
func (ch *Channel) EnterRm(rmNo uint32, user *User) error {
	var mu = &ch.sync
	mu.Lock()
	defer mu.Unlock()

	if ch.rooms[rmNo] == nil {
		rm, err := LoadRoom(rmNo, _redisClient)
		if err != nil {
			return err
		}
		ch.rooms[rmNo] = rm
		rm.EnterMember(user)
		return nil
	}

	ch.rooms[rmNo].EnterMember(user)
	return nil
}

//
func MonitorChannel() string {
	var str string
	for _, ch := range _channels {
		str += "ch: " + strconv.Itoa(int(ch.no)) + "\r\n"
		var mu = &ch.sync
		mu.Lock()
		defer mu.Unlock()

		for _, rm := range ch.rooms {
			str += "	rm: " + strconv.Itoa(int(rm.rID)) + "\r\n"
		}

		for _, ur := range ch.members {
			str += "	user: " + ur.Name + "\r\n"
		}
	}
	return str
}

// update channel info
func UpdateChannel() {

	saveChannel()
}
