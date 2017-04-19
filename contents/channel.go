package contents

import (
	"errors"
	"log"
	"sync"
	"time"

	redis "gopkg.in/redis.v4"

	"strconv"

	"github.com/blueberryserver/tcpserver/network"
)

// room,  channel db(2)
var rmchRedisClient *redis.Client

//
func SetRmChRedisClient(client *redis.Client) {
	rmchRedisClient = client

	pipe := rmchRedisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(2)
	_, _ = pipe.Exec()
}

var _chSync sync.Mutex // sync obj
// channel
type Channel struct {
	no      uint32 // channel number
	chType  ChType
	members map[uint32]*User // channel user
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
		}
	}
}

// load channel from redis
func LoadChannel() {
	log.Println("loading channel info")
	_channels = make(map[uint32]*Channel)

	var cursor uint64
	var outputs []string
	outputs, cursor, err := rmchRedisClient.HScan("blue_server.ch.type", cursor, "", 10).Result()
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
		}
	}
	// user count
	//cursor = 0
	//outputs, cursor, err = rmchRedisClient.HScan("blue_server.ch.user.count", cursor, "", 10).Result()
	//log.Println(outputs)
}

// save channel to redis
func saveChannel(id int) {
	{
		// save channel
		var chMutex = &_chSync
		chMutex.Lock()
		defer chMutex.Unlock()
		for _, ch := range _channels {

			_, err := rmchRedisClient.HSet("blue_server.ch.type", strconv.Itoa(int(ch.no)), ChTypeName[ch.chType]).Result()
			if err != nil {
				log.Println(err)
				continue
			}

			userCount := len(ch.members)
			_, err = rmchRedisClient.HSet("blue_server.ch.user.count", strconv.Itoa(int(ch.no)), strconv.Itoa(userCount)).Result()
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}

	{
		// save room
		var mutex = &_roomSync
		mutex.Lock()
		defer mutex.Unlock()
		for _, rm := range _rooms {
			rm.save()
		}
	}
}

// enter channel
func EnterCh(chNo uint32, user *User) bool {
	if int(chNo) > len(_channels) || chNo < 0 {
		return false
	}

	var chMutex = &_chSync
	chMutex.Lock()
	defer chMutex.Unlock()
	log.Println("Enter channel no:", chNo, "user:", user.Name)
	_channels[chNo].members[user.ID] = user

	if chNo != 0 {
		user.ChNo = chNo
	}
	return true
}

// leave channel
func LeaveCh(user *User) {
	log.Println("Leave channel no:", user.ChNo, "user:", user.Name, "member count:", len(_channels[0].members))

	var chMutex = &_chSync
	chMutex.Lock()
	defer chMutex.Unlock()
	//leave defualt channel
	delete(_channels[0].members, user.ID)
	log.Println("Remind channel no: 0 member count:", len(_channels[0].members))

	// leave current channel
	if user.ChNo != 0 {
		delete(_channels[user.ChNo].members, user.ID)
	}
}

// move channel
func MoveCh(chNo uint32, user *User) {
	log.Println("Move channel no:", chNo, "user:", user.Name)
	var chMutex = &_chSync
	chMutex.Lock()
	defer chMutex.Unlock()
	// leave current channel
	delete(_channels[user.ChNo].members, user.ID)

	// enter channel
	EnterCh(chNo, user)
}

// find user
func FindUser(session *network.Session) (*User, error) {
	var chMutex = &_chSync
	chMutex.Lock()
	defer chMutex.Unlock()
	for _, v := range _channels[0].members {
		if v.Session == session {
			return v, nil
		}
	}
	return &User{}, errors.New("Not find user session")
}

// find user
func FindUserByID(id uint32) (*User, error) {
	var chMutex = &_chSync
	chMutex.Lock()
	defer chMutex.Unlock()
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
func MonitorChannel(id int) {
	{
		var chMutex = &_chSync
		chMutex.Lock()
		defer chMutex.Unlock()
		for _, ch := range _channels {
			log.Println(id, "ch: "+strconv.Itoa(int(ch.no)))

			for _, ur := range ch.members {
				log.Println(id, "	user: "+ur.Name)
			}
		}
	}

	{
		var mutex = &_roomSync
		mutex.Lock()
		defer mutex.Unlock()
		for _, rm := range _rooms {
			log.Println(id, "rm: "+strconv.Itoa(int(rm.rID)))

			for _, ur := range rm.members {
				log.Println(id, "	user: "+ur.Name)
			}
		}
	}
}

// update channel info
func UpdateChannel(id int) {
	log.Println(id, "update channel")
	checkLogoutUser(id)
	saveChannel(id)
}

// generate id
func RoomGenID() uint32 {

	genID, _ := rmchRedisClient.Incr("blue_server.manager.room.genid").Result()
	return uint32(genID)
}

func checkLogoutUser(id int) {
	log.Println(id, "check logoff user")
	{
		var rmMutex = &_roomSync
		rmMutex.Lock()
		defer rmMutex.Unlock()
		for _, rm := range _rooms {
			for _, ur := range rm.members {
				if ur.Status == _LogOff && time.Now().After(ur.LogoutTime.Add(30*time.Second)) {
					rm.LeaveMember(ur)
				}
			}
		}
	}

	{
		var chMutex = &_chSync
		chMutex.Lock()
		defer chMutex.Unlock()
		for _, v := range _channels[0].members {
			if v.Status == _LogOff && time.Now().After(v.LogoutTime.Add(30*time.Second)) {
				// leave ch
				delete(_channels[0].members, v.ID)
				log.Println(id, "Remind channel no: 0 member count:", len(_channels[0].members))

				// leave current channel
				if v.ChNo != 0 {
					delete(_channels[v.ChNo].members, v.ID)
				}

				// save data
				v.Save()
			}
		}
	}
}
