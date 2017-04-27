package contents

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/blueberryserver/tcpserver/network"
)

// user command channel
var UserCmd chan *UserCmdData

// user data list map
var UserList map[uint32]*User

// cmd
type UserCmdData struct {
	Cmd     string           `json:"cmd"`
	ID      uint32           `json:"id"`
	User    *User            `json:"user"`
	Session *network.Session `json:"session"`
	Result  chan *CmdResult  `json:"result"`
}

//
type CmdResult struct {
	Err  error       `json:"err"`
	Data interface{} `json:"data"`
}

// go routine by channel commuity
func UserProcFunc() {
	UserCmd = make(chan *UserCmdData)
	UserList = make(map[uint32]*User)
	for {
		select {
		case cmd := <-UserCmd:

			switch cmd.Cmd {
			case "AddUser":
				addUser(cmd.ID, cmd.User, cmd.Result)

			case "DelUser":
				delUser(cmd.ID, cmd.Result)

			case "ListUser":
				listUser(cmd.Result)

			case "CheckUser":
				checkUser(cmd.Result)

			case "FindUserByID":
				findUserByID(cmd.ID, cmd.Result)

			case "FindUser":
				findUser(cmd.Session, cmd.Result)
			}
		}
	}
}

// add user
func addUser(id uint32, data *User, result chan *CmdResult) {
	sResult := <-result
	sResult.Err = nil

	UserList[id] = data

	result <- sResult
}

// add user
func listUser(result chan *CmdResult) {
	sResult := <-result
	sResult.Err = nil

	var str string
	for _, ur := range UserList {
		data, _ := json.Marshal(ur)
		str += string(data) + "\r\n"
	}
	fmt.Println(str)

	result <- sResult
}

// add user
func delUser(id uint32, result chan *CmdResult) {
	sResult := <-result
	sResult.Err = nil

	delete(UserList, id)

	result <- sResult
}

// chech
func checkUser(result chan *CmdResult) {
	sResult := <-result
	sResult.Err = nil

	for _, ur := range UserList {
		if (ur.Data.Status == _LogOff && time.Now().After(ur.Data.LogoutTime.Add(30*time.Second))) ||
			(time.Now().After(ur.KeepaliveTime.Add(300 * time.Second))) {
			// session closing
			ur.Session.Close()
			/*
				// channel leave proc
				LeaveCh(ur.Data.ChNo, ur)

				// request room leave proc
				if ur.Data.RmNo != 0 {
					LeaveRm(ur.Data.RmNo, ur)
				}

				ur.Save()
				delete(UserList, ur.Data.ID)
			*/
		}
	}
	result <- sResult
}

func findUserByID(id uint32, result chan *CmdResult) {
	sResult := <-result
	sResult.Data = nil
	sResult.Err = errors.New("not find user id")

	for _, ur := range UserList {
		if ur.Data.ID == id {
			sResult.Data = ur
			sResult.Err = nil
			break
		}
	}
	result <- sResult
}

func findUser(session *network.Session, result chan *CmdResult) {
	sResult := <-result
	sResult.Data = nil
	sResult.Err = errors.New("not find user session")

	for _, ur := range UserList {
		if ur.Session == session {
			sResult.Data = ur
			sResult.Err = nil
			break
		}
	}
	result <- sResult
}
