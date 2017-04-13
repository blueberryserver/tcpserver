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
	cNo uint32 // channel number

	sync    sync.Mutex       // sync obj
	members map[uint32]*User // channel user
	rooms   map[uint32]*Room // channel room
}

var _channels map[uint32]*Channel

// generate channel
func NewChannel() {
	_channels = make(map[uint32]*Channel)

	for i := 0; i < 2; i++ {
		_channels[uint32(i)] = &Channel{
			cNo:     uint32(i),
			members: make(map[uint32]*User),
			rooms:   make(map[uint32]*Room),
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

func Monitor() string {
	var str string
	for _, ch := range _channels {
		str += "ch: " + strconv.Itoa(int(ch.cNo)) + "\r\n"
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
