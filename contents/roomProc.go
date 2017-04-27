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
	Cmd    string `json:"cmd"`
	No     uint32 `json:"no"`
	Type   uint32 `json:"type"`
	User   *User  `json:"user"`
	Result chan *CmdResult
}

// go routine by channel commuity
func RoomProcFunc() {
	RoomCmd = make(chan *RoomCmdData)
	for {
		select {
		case cmd := <-RoomCmd:

			switch cmd.Cmd {
			case "LoadRoom":
				loadRm(cmd.Result)

			case "LeaveRoom":
				leaveRm(cmd.No, cmd.User, cmd.Result)

			case "EnterRoom":
				enterRm(cmd.No, cmd.Type, cmd.User, cmd.Result)

			case "ListRoomAns":
				listRmAns(cmd.Result)

			case "ListRoom":
				listRm(cmd.Result)

			case "FindRoom":
				findRm(cmd.No, cmd.Result)
			}
		}
	}
}

//
func loadRm(result chan *CmdResult) {
	sResult := <-result
	RoomList = make(map[uint32]*Room)

	var cursor uint64
	var outputs []string
	outputs, cursor, err := rmchRedisClient.HScan("blue_server.room.json", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		sResult.Err = err
		result <- sResult
		return
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
	sResult.Err = nil
	result <- sResult
}

//
func leaveRm(no uint32, user *User, result chan *CmdResult) {
	sResult := <-result
	if no == 0 {
		no = user.Data.RmNo
	}

	if no == 0 {
		sResult.Err = errors.New("not room member")
		result <- sResult
		return
	}

	if RoomList[no] == nil {
		rm, err := load(no)
		if err != nil {
			log.Println(err)
			sResult.Err = err
			result <- sResult
		}
		RoomList[no] = rm
		rm.LeaveMember(user)
		sResult.Err = nil
		result <- sResult
		return
	}

	RoomList[no].LeaveMember(user)
	sResult.Err = nil
	result <- sResult
}

//
func enterRm(no uint32, rtype uint32, user *User, result chan *CmdResult) {
	sResult := <-result

	if no == 0 {
		for _, rm := range RoomList {
			if RoomType(rtype) == _RoomNormal {
				if (len(rm.members) == 0) ||
					(len(rm.members) == 1 && _RoomSolo != rm.data.RmType) {
					no = rm.data.RmNo

					// change room type
					rm.data.RmType = RoomType(rtype)
					break
				}
			} else {
				if len(rm.members) == 0 {
					no = rm.data.RmNo
					// change room type
					rm.data.RmType = RoomType(rtype)
					break
				}
			}

		}
	}

	if no == 0 {
		// create room
		rm := NewRoom()
		rm.data.RmType = RoomType(rtype)
		RoomList[rm.data.RmNo] = rm
		no = rm.data.RmNo
	}

	if RoomList[no] == nil {
		rm, err := load(no)
		if err != nil {
			rm = NewRoom()
			rm.data.RmType = RoomType(rtype)
			no = rm.data.RmNo
		}
		RoomList[no] = rm
		rm.EnterMember(user)
		sResult.Err = nil
		result <- sResult
		return
	}

	RoomList[no].data.RmType = RoomType(rtype)
	RoomList[no].EnterMember(user)
	sResult.Err = nil
	result <- sResult
}

//
func listRmAns(result chan *CmdResult) {
	sResult := <-result

	list := make([]*msg.ListRmAns_RoomInfo, len(RoomList))
	index := 0
	for _, rm := range RoomList {
		list[index] = &msg.ListRmAns_RoomInfo{}
		list[index].RmNo = &rm.data.RmNo
		status := uint32(rm.data.RmStatus)
		list[index].RmStatus = &status
		rmtype := uint32(rm.data.RmType)
		list[index].RmType = &rmtype
		list[index].Names = make([]string, len(rm.members))

		userIndex := 0
		for _, ur := range rm.members {
			list[index].Names[userIndex] = ur.Data.Name
			userIndex++
		}
		index++
	}
	sResult.Data = list
	sResult.Err = nil
	result <- sResult
}

func listRm(result chan *CmdResult) {
	sResult := <-result
	var monitor string
	for _, rm := range RoomList {
		monitor += fmt.Sprintln("<p>Room No: " + strconv.Itoa(int(rm.data.RmNo)) + " Type: " +
			RoomTypeName[rm.data.RmType] + " Status: " +
			RoomStatusName[rm.data.RmStatus] + "</p>")
		for _, ur := range rm.members {
			monitor += "<p><blockquote>"
			monitor += fmt.Sprintf("User: %v", ur.Data)
			monitor += "</blockquote>"
		}
		rm.save()
	}

	sResult.Data = monitor
	sResult.Err = nil
	result <- sResult
}

//
func findRm(no uint32, result chan *CmdResult) {
	sResult := <-result
	sResult.Data = nil
	sResult.Err = errors.New("not find room")

	if RoomList[no] == nil {
		result <- sResult
		return
	}

	sResult.Data = RoomList[no]
	sResult.Err = nil
	result <- sResult
}
