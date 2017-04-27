package contents

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
)

// channel command channel
var ChCmd chan *ChCmdData

// channel data list map
var ChannelList map[uint32]*Channel

// cmd
type ChCmdData struct {
	Cmd    string `json:"cmd"`
	No     uint32 `json:"no"`
	User   *User  `json:"user"`
	UserID uint32 `json:"userid"`
	Result chan *CmdResult
}

// go routine by channel commuity
func ChannelProcFunc() {
	ChCmd = make(chan *ChCmdData)
	for {
		select {
		case cmd := <-ChCmd:

			switch cmd.Cmd {
			case "LoadCh":
				loadCh(cmd.Result)
			case "EnterCh":
				enterCh(cmd.No, cmd.User, cmd.Result)
			case "LeaveCh":
				leaveCh(cmd.User, cmd.Result)
			case "ListCh":
				listCh(cmd.Result)
			}
		}
	}
}

//
func loadCh(result chan *CmdResult) {
	sResult := <-result
	ChannelList = make(map[uint32]*Channel)

	var cursor uint64
	var outputs []string

	outputs, cursor, err := rmchRedisClient.HScan("blue_server.ch.json", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		sResult.Err = err
		result <- sResult
		return
	}

	for i := 0; i < len(outputs); i += 2 {
		// redis key
		no := outputs[i]
		chNo, _ := strconv.Atoi(no)
		// redis value
		chdata := ChData{}
		json.Unmarshal([]byte(outputs[i+1]), &chdata)
		ChannelList[uint32(chNo)] = &Channel{
			data:    chdata,
			members: make(map[uint32]*User),
		}
	}
	sResult.Err = nil
	result <- sResult
}

func enterCh(no uint32, user *User, result chan *CmdResult) {
	sResult := <-result
	if int(no) > len(ChannelList) || no < 0 {
		sResult.Err = errors.New("invalid channel number")
		result <- sResult
		return
	}
	log.Println("Enter channel no:", no, "user:", user.Data.Name)
	ChannelList[no].members[user.Data.ID] = user

	if no != 0 {
		user.Data.ChNo = no
	}
	sResult.Err = nil
	result <- sResult
}

func leaveCh(user *User, result chan *CmdResult) {
	sResult := <-result
	log.Println("Leave channel no:", user.Data.ChNo, "user:", user.Data.Name, "member count:", len(ChannelList[0].members))

	//leave defualt channel
	delete(ChannelList[0].members, user.Data.ID)
	log.Println("Remind channel no: 0 member count:", len(ChannelList[0].members))

	// leave current channel
	if user.Data.ChNo != 0 {
		delete(ChannelList[user.Data.ChNo].members, user.Data.ID)
		user.Data.ChNo = 0
	}
	sResult.Err = nil
	result <- sResult
}

func listCh(result chan *CmdResult) {
	sResult := <-result
	var monitor string
	for i := 0; i < len(ChannelList); i++ {
		if ChannelList[uint32(i)] == nil {
			log.Println("monitor empty " + strconv.Itoa(i))
			continue
		}

		monitor += fmt.Sprintln("<p>Channel No: " + strconv.Itoa(int(ChannelList[uint32(i)].data.ChNo)) + " Type: " +
			ChTypeName[ChannelList[uint32(i)].data.ChType] + " Limit: " +
			strconv.Itoa(int(ChannelList[uint32(i)].data.ChLimit)) + "</p>")

		for _, ur := range ChannelList[uint32(i)].members {
			monitor += "<p><blockquote>"
			monitor += fmt.Sprintf("User: %v", ur.Data)
			monitor += "</blockquote>"
		}
	}
	sResult.Data = monitor
	sResult.Err = nil
	result <- sResult
}
