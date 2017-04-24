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
	Result  error            `json:"result"`
	User    *User            `json:"user"`
	Session *network.Session `json:"session"`
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
				cmd.Result = addUser(cmd.ID, cmd.User)
			case "DelUser":
				cmd.Result = delUser(cmd.ID)
			case "ListUser":
				cmd.Result = listUser()
			case "CheckUser":
				cmd.Result = checkUser()
			case "FindUserByID":
				cmd.User, cmd.Result = findUserByID(cmd.ID)
			case "FindUser":
				cmd.User, cmd.Result = findUser(cmd.Session)
			}
			UserCmd <- cmd
		}
	}
}

// add user
func addUser(id uint32, data *User) error {
	UserList[id] = data
	return nil
}

// add user
func listUser() error {
	var str string
	for _, ur := range UserList {
		data, _ := json.Marshal(ur)
		str += string(data) + "\r\n"
	}
	fmt.Println(str)
	return nil
}

// add user
func delUser(id uint32) error {
	delete(UserList, id)
	return nil
}

// chech
func checkUser() error {
	for _, ur := range UserList {
		if (ur.Data.Status == _LogOff && time.Now().After(ur.Data.LogoutTime.Add(30*time.Second))) ||
			(time.Now().After(ur.KeepaliveTime.Add(300 * time.Second))) {
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
	return nil
}

func findUserByID(id uint32) (*User, error) {
	for _, ur := range UserList {
		if ur.Data.ID == id {
			return ur, nil
		}
	}
	return nil, errors.New("not found user id")
}

func findUser(session *network.Session) (*User, error) {
	for _, ur := range UserList {
		if ur.Session == session {
			return ur, nil
		}
	}
	return nil, errors.New("not find user session")
}
