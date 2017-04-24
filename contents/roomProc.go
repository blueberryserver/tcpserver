package contents

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/blueberryserver/tcpserver/msg"
)

// user command channel
var RoomCmd chan *RoomCmdData

// room data list map
var RoomList map[uint32]*Room

// cmd
type RoomCmdData struct {
	Cmd     string                    `json:"cmd"`
	No      uint32                    `json:"no"`
	Result  error                     `json:"result"`
	User    *User                     `json:"user"`
	Room    *Room                     `json:"room"`
	List    []*msg.ListRmAns_RoomInfo `json:"List"`
	Monitor string                    `json:"monitor"`
}

// go routine by channel commuity
func RoomProcFunc() {
	RoomCmd = make(chan *RoomCmdData)
	for {
		select {
		case cmd := <-RoomCmd:

			switch cmd.Cmd {
			case "LoadRoom":
				cmd.Result = loadRm()
			case "LeaveRoom":
				cmd.Result = leaveRm(cmd.No, cmd.User)
			case "EnterRoom":
				cmd.Result = enterRm(cmd.No, cmd.User)
			case "ListRoomAns":
				cmd.List, cmd.Result = listRmAns()
			case "ListRoom":
				cmd.Result = listRm(&cmd.Monitor)
			case "FindRoom":
				cmd.Room, cmd.Result = findRm(cmd.No)
			}
			RoomCmd <- cmd
		}
	}
}

//
func loadRm() error {
	RoomList = make(map[uint32]*Room)

	var cursor uint64
	var outputs []string
	outputs, cursor, err := rmchRedisClient.HScan("blue_server.room.json", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		return err
	}

	for i := 0; i < len(outputs); i += 2 {
		// redis key
		no := outputs[i]
		rmNo, _ := strconv.Atoi(no)
		// redis value
		rdata := RmData{}
		json.Unmarshal([]byte(outputs[i+1]), &rdata)

		RoomList[uint32(rmNo)] = &Room{
			data:    rdata,
			members: make(map[uint32]*User),
		}
	}
	return nil
}

//
func leaveRm(no uint32, user *User) error {
	if no == 0 {
		no = user.Data.RmNo
	}

	if no == 0 {
		return errors.New("not room member")
	}

	if RoomList[no] == nil {
		rm, err := load(no)
		if err != nil {
			log.Println(err)
			return err
		}
		RoomList[no] = rm
		rm.LeaveMember(user)
		return nil
	}

	RoomList[no].LeaveMember(user)
	return nil
}

//
func enterRm(no uint32, user *User) error {
	if no == 0 {
		for _, rm := range RoomList {
			if len(rm.members) < 2 {
				no = rm.data.RmNo
				break
			}
		}
	}

	if no == 0 {
		// create room
		rm := NewRoom()
		RoomList[rm.data.RmNo] = rm
		no = rm.data.RmNo
	}

	if RoomList[no] == nil {
		rm, err := load(no)
		if err != nil {
			rm = NewRoom()
			no = rm.data.RmNo
		}
		RoomList[no] = rm
		rm.EnterMember(user)
		return nil
	}

	RoomList[no].EnterMember(user)
	return nil
}

//
func listRmAns() ([]*msg.ListRmAns_RoomInfo, error) {
	list := make([]*msg.ListRmAns_RoomInfo, len(RoomList))

	index := 0
	for _, rm := range RoomList {
		list[index] = &msg.ListRmAns_RoomInfo{}
		list[index].RmNo = &rm.data.RmNo
		status := uint32(rm.data.RmStatus)
		list[index].RmStatus = &status
		list[index].Names = make([]string, len(rm.members))

		userIndex := 0
		for _, ur := range rm.members {
			list[index].Names[userIndex] = ur.Data.Name
			userIndex++
		}
		index++
	}
	return list, nil
}

func listRm(monitor *string) error {
	for i := 0; i < len(RoomList)+1; i++ {
		if RoomList[uint32(i)] == nil {
			continue
		}

		*monitor += fmt.Sprintln("<p>Room No: " + strconv.Itoa(int(RoomList[uint32(i)].data.RmNo)) + " Type: " +
			RoomTypeName[RoomList[uint32(i)].data.RmType] + " Status: " +
			RoomStatusName[RoomList[uint32(i)].data.RmStatus] + "</p>")
		for _, ur := range RoomList[uint32(i)].members {
			*monitor += "<p><blockquote>"
			*monitor += fmt.Sprintf("User: %v", ur.Data)
			*monitor += "</blockquote>"
		}
	}
	return nil
}

//
func findRm(no uint32) (*Room, error) {
	if RoomList[no] == nil {
		return nil, errors.New("not find room")
	}
	return RoomList[no], nil
}
