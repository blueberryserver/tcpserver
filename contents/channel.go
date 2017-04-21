package contents

import (
	"encoding/json"
	"errors"
	"fmt"
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

// channel config data
type CHData struct {
	ChNo    uint32 `json:"chno"`
	ChType  ChType `json:"chtype"`
	ChLimit uint32 `json:"chlimit"`
}

// channel
type Channel struct {
	data    CHData
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

// load channel from redis
func LoadChannel() {
	log.Println("loading channel info")
	_channels = make(map[uint32]*Channel)

	var cursor uint64
	var outputs []string

	outputs, cursor, err := rmchRedisClient.HScan("blue_server.ch.json", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		return
	}

	for i := 0; i < len(outputs); i += 2 {
		// redis key
		no := outputs[i]
		chNo, _ := strconv.Atoi(no)
		// redis value
		chdata := CHData{}
		json.Unmarshal([]byte(outputs[i+1]), &chdata)
		_channels[uint32(chNo)] = &Channel{
			data:    chdata,
			members: make(map[uint32]*User),
		}
	}
}

// save channel to redis
func saveChannel(id int) {
	{
		// save channel
		var chMutex = &_chSync
		chMutex.Lock()
		defer chMutex.Unlock()
		for _, ch := range _channels {
			userCount := len(ch.members)
			_, err := rmchRedisClient.HSet("blue_server.ch.user.count", strconv.Itoa(int(ch.data.ChNo)), strconv.Itoa(userCount)).Result()
			if err != nil {
				log.Println(err)
				continue
			}
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
	log.Println("Enter channel no:", chNo, "user:", user.Data.Name)
	_channels[chNo].members[user.Data.ID] = user

	if chNo != 0 {
		user.Data.ChNo = chNo
	}
	return true
}

// leave channel
func LeaveCh(user *User) {
	log.Println("Leave channel no:", user.Data.ChNo, "user:", user.Data.Name, "member count:", len(_channels[0].members))

	var chMutex = &_chSync
	chMutex.Lock()
	defer chMutex.Unlock()
	//leave defualt channel
	delete(_channels[0].members, user.Data.ID)
	log.Println("Remind channel no: 0 member count:", len(_channels[0].members))

	// leave current channel
	if user.Data.ChNo != 0 {
		delete(_channels[user.Data.ChNo].members, user.Data.ID)
		user.Data.ChNo = 0
	}
}

// move channel
func MoveCh(chNo uint32, user *User) {
	log.Println("Move channel no:", chNo, "user:", user.Data.Name)
	var chMutex = &_chSync
	chMutex.Lock()
	defer chMutex.Unlock()
	// leave current channel
	delete(_channels[user.Data.ChNo].members, user.Data.ID)

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
		if v.Data.ID == id {
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
	{
		var chMutex = &_chSync
		chMutex.Lock()
		defer chMutex.Unlock()

		for i := 0; i < len(_channels); i++ {
			str += fmt.Sprintln("<p>Channel No: " + strconv.Itoa(int(_channels[uint32(i)].data.ChNo)) + " Type: " +
				ChTypeName[_channels[uint32(i)].data.ChType] + " Limit: " +
				strconv.Itoa(int(_channels[uint32(i)].data.ChLimit)) + "</p>")

			for _, ur := range _channels[uint32(i)].members {
				str += "<p><blockquote>"
				str += fmt.Sprintf("User: %v", ur.Data)
				str += "</blockquote>"
			}
		}
	}
	str += "<p>.....................................</p>"
	{
		var mutex = &_roomSync
		mutex.Lock()
		defer mutex.Unlock()
		for i := 1; i < len(_rooms)+1; i++ {
			str += fmt.Sprintln("<p>Room No: " + strconv.Itoa(int(_rooms[uint32(i)].data.RmNo)) + " Type: " +
				RoomTypeName[_rooms[uint32(i)].data.RmType] + " Status: " +
				RoomStatusName[_rooms[uint32(i)].data.RmStatus] + "</p>")

			for _, ur := range _rooms[uint32(i)].members {
				str += "<p><blockquote>"
				str += fmt.Sprintf("User: %v", ur.Data)
				str += "</blockquote>"
			}
		}
	}
	return str
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

				// logout time over
				if (ur.Data.Status == _LogOff && time.Now().After(ur.Data.LogoutTime.Add(30*time.Second))) ||
					(time.Now().After(ur.KeepaliveTime.Add(300 * time.Second))) {
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
			if (v.Data.Status == _LogOff && time.Now().After(v.Data.LogoutTime.Add(30*time.Second))) ||
				(time.Now().After(v.KeepaliveTime.Add(300 * time.Second))) {
				// leave ch
				delete(_channels[0].members, v.Data.ID)
				log.Println(id, "Remind channel no: 0 member count:", len(_channels[0].members))

				// leave current channel
				if v.Data.ChNo != 0 {
					delete(_channels[v.Data.ChNo].members, v.Data.ID)
				}

				// save data
				v.Save()

				// sesion init
				v.Session = nil
			}
		}
	}
}
