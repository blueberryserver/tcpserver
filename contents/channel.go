package contents

import (
	"errors"
	"log"
	"sync"
	"time"

	"strconv"

	"github.com/blueberryserver/tcpserver/network"
)

// channel
type Channel struct {
	no      uint32 // channel number
	chType  ChType
	sync    sync.Mutex       // sync obj
	members map[uint32]*User // channel user
	//rooms   map[uint32]*Room // channel room
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
			//rooms:   make(map[uint32]*Room),
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
		log.Println(err)
		return
	}

	_channels = make(map[uint32]*Channel)

	var cursor uint64
	var outputs []string
	outputs, cursor, err = _redisClient.HScan("blue_server.ch.type", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
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
			//rooms:   make(map[uint32]*Room),
		}
		//log.Println(_channels[uint32(iNo)])
	}
	// user count
	cursor = 0
	outputs, cursor, err = _redisClient.HScan("blue_server.ch.user.count", cursor, "", 10).Result()
	//log.Println(outputs)
}

// save channel to redis
func saveChannel() {
	pipe := _redisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(2)
	_, err := pipe.Exec()
	if err != nil {
		log.Println(err)
		return
	}

	for _, ch := range _channels {
		var mu = &ch.sync
		mu.Lock()
		defer mu.Unlock()
		_, err := _redisClient.HSet("blue_server.ch.type", strconv.Itoa(int(ch.no)), ChTypeName[ch.chType]).Result()
		if err != nil {
			log.Println(err)
			continue
		}
		// roomCount := len(ch.rooms)
		// _, err = _redisClient.HSet("blue_server.ch.room.count", strconv.Itoa(int(ch.no)), strconv.Itoa(roomCount)).Result()
		// if err != nil {
		// 	log.Println(err)
		// 	continue
		// }
		userCount := len(ch.members)
		_, err = _redisClient.HSet("blue_server.ch.user.count", strconv.Itoa(int(ch.no)), strconv.Itoa(userCount)).Result()
		if err != nil {
			log.Println(err)
			continue
		}

		// save user info
		// for _, ur := range ch.members {
		// 	ur.Save()
		// }

		// save room info
		// for _, rm := range ch.rooms {
		// 	rm.Save()
		// }
	}

	var mu = &_sync
	mu.Lock()
	defer mu.Unlock()
	for _, rm := range _rooms {
		rm.save()
	}
}

// enter channel
func EnterCh(chNo uint32, user *User) bool {
	if int(chNo) > len(_channels) || chNo < 0 {
		return false
	}

	log.Println("Enter channel no:", chNo, "user:", user.Name)
	_channels[chNo].members[user.ID] = user

	user.ChNo = chNo
	return true
}

// leave channel
func LeaveCh(user *User) {
	log.Println("Leave channel no:", user.ChNo, "user:", user.Name, "member count:", len(_channels[0].members))

	//leave defualt channel
	delete(_channels[0].members, user.ID)
	log.Println("Remind channel no: 0 member count:", len(_channels[0].members))

	var mu = &_channels[user.ChNo].sync
	mu.Lock()
	defer mu.Unlock()

	// leave room
	//if user.RmNo != 0 && user.ChNo != 0 {
	//	_channels[user.ChNo].rooms[user.RmNo].LeaveMember(user)
	//}
	// leave current channel
	if user.ChNo != 0 {
		delete(_channels[user.ChNo].members, user.ID)
	}
}

// move channel
func MoveCh(chNo uint32, user *User) {
	log.Println("Move channel no:", chNo, "user:", user.Name)

	// leave current channel
	delete(_channels[user.ChNo].members, user.ID)

	// enter channel
	EnterCh(chNo, user)
}

// find user
func FindUser(session *network.Session) (*User, error) {
	var mu = &_channels[0].sync
	mu.Lock()
	defer mu.Unlock()
	for _, v := range _channels[0].members {
		if v.Session == session {
			return v, nil
		}
	}
	return &User{}, errors.New("Not find user session")
}

// find user
func FindUserByID(id uint32) (*User, error) {
	var mu = &_channels[0].sync
	mu.Lock()
	defer mu.Unlock()
	for _, v := range _channels[0].members {
		if v.ID == id {
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
func MonitorChannel() string {
	var str string
	for _, ch := range _channels {
		str += "ch: " + strconv.Itoa(int(ch.no)) + "\r\n"
		var mu = &ch.sync
		mu.Lock()
		defer mu.Unlock()

		for _, ur := range ch.members {
			str += "	user: " + ur.Name + "\r\n"
		}
	}

	var mu = &_sync
	mu.Lock()
	defer mu.Unlock()
	for _, rm := range _rooms {
		str += "rm: " + strconv.Itoa(int(rm.rID)) + "\r\n"

		for _, ur := range rm.members {
			str += "	user: " + ur.Name + "\r\n"
		}
	}
	return str
}

// update channel info
func UpdateChannel() {
	checkLogoutUser()
	saveChannel()
}

// generate id
func RoomGenID() uint32 {
	pipe := _redisClient.Pipeline()
	defer pipe.Close()

	pipe.Select(2)
	_, _ = pipe.Exec()

	genID, _ := _redisClient.Incr("blue_server.manager.room.genid").Result()
	return uint32(genID)
}

func checkLogoutUser() {
	log.Println("Check Logout User")
	var mu = &_channels[0].sync
	mu.Lock()
	defer mu.Unlock()
	for _, v := range _channels[0].members {
		if v.Status == _LogOff && time.Now().After(v.LogoutTime.Add(30*time.Second)) {
			// leave ch
			LeaveCh(v)

			// leave rm
			LeaveRm(v.RmNo, v)

			// save data
			v.Save()
		}
	}
}
