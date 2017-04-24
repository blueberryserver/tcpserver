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
	Cmd     string `json:"cmd"`
	No      uint32 `json:"no"`
	Result  error  `json:"result"`
	User    *User  `json:"user"`
	UserID  uint32 `json:"userid"`
	Monitor string `json:"monitor"`
}

// go routine by channel commuity
func ChannelProcFunc() {
	ChCmd = make(chan *ChCmdData)
	for {
		select {
		case cmd := <-ChCmd:

			switch cmd.Cmd {
			case "LoadCh":
				cmd.Result = loadCh()
			case "EnterCh":
				cmd.Result = enterCh(cmd.No, cmd.User)
			case "LeaveCh":
				cmd.Result = leaveCh(cmd.User)
			case "ListCh":
				cmd.Result = listCh(&cmd.Monitor)
			}
			ChCmd <- cmd
		}
	}
}

//
func loadCh() error {
	ChannelList = make(map[uint32]*Channel)

	var cursor uint64
	var outputs []string

	outputs, cursor, err := rmchRedisClient.HScan("blue_server.ch.json", cursor, "", 10).Result()
	if err != nil {
		log.Println(err)
		return err
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
	return nil
}

func enterCh(no uint32, user *User) error {
	if int(no) > len(ChannelList) || no < 0 {
		return errors.New("invalid channel number")
	}
	log.Println("Enter channel no:", no, "user:", user.Data.Name)
	ChannelList[no].members[user.Data.ID] = user

	if no != 0 {
		user.Data.ChNo = no
	}
	return nil
}

func leaveCh(user *User) error {
	log.Println("Leave channel no:", user.Data.ChNo, "user:", user.Data.Name, "member count:", len(ChannelList[0].members))

	//leave defualt channel
	delete(ChannelList[0].members, user.Data.ID)
	log.Println("Remind channel no: 0 member count:", len(ChannelList[0].members))

	// leave current channel
	if user.Data.ChNo != 0 {
		delete(ChannelList[user.Data.ChNo].members, user.Data.ID)
		user.Data.ChNo = 0
	}
	return nil
}

func listCh(monitor *string) error {
	for i := 0; i < len(ChannelList); i++ {
		if ChannelList[uint32(i)] == nil {
			log.Println("monitor empty " + strconv.Itoa(i))
			continue
		}

		*monitor += fmt.Sprintln("<p>Channel No: " + strconv.Itoa(int(ChannelList[uint32(i)].data.ChNo)) + " Type: " +
			ChTypeName[ChannelList[uint32(i)].data.ChType] + " Limit: " +
			strconv.Itoa(int(ChannelList[uint32(i)].data.ChLimit)) + "</p>")

		for _, ur := range ChannelList[uint32(i)].members {
			*monitor += "<p><blockquote>"
			*monitor += fmt.Sprintf("User: %v", ur.Data)
			*monitor += "</blockquote>"
		}
	}
	return nil
}
