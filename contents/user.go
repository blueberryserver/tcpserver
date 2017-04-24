package contents

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	redis "gopkg.in/redis.v4"

	"log"

	"github.com/blueberryserver/tcpserver/network"
)

// user db(1)
var userRedisClient *redis.Client

//
func SetUserRedisClient(client *redis.Client) {
	userRedisClient = client

	pipe := userRedisClient.Pipeline()
	defer pipe.Close()

	pipe.Select(1)
	_, _ = pipe.Exec()
}

// user data for json
type UrData struct {
	ID         uint32       `json:"id"`
	Name       string       `json:"name"`
	Platform   UserPlatform `json:"platform"`
	VcGem      uint32       `json:"gem"`
	VcGold     uint32       `json:"gold"`
	Key        string       `json:"key"`
	Status     UserStatus   `json:"status"`
	ChNo       uint32       `json:"chno"`
	RmNo       uint32       `json:"rmno"`
	LoginTime  time.Time    `json:"logintime"`
	LogoutTime time.Time    `json:"logouttime"`
	CreateTime time.Time    `json:"createtime"`
}

// user obj
type User struct {
	Data          UrData
	Session       *network.Session
	KeepaliveTime time.Time
}

// user login status
type UserStatus uint32

// user login status
const (
	_Logon  UserStatus = 1
	_LogOff UserStatus = 2
)

// user status
var UserStatusName = map[UserStatus]string{
	1: "LOGON",
	2: "LOGOFF",
}

// user status
var UserStatusValue = map[string]UserStatus{
	"LOGON":  1,
	"LOGOFF": 2,
}

// user platform
type UserPlatform uint32

// user platform
const (
	_IOS     UserPlatform = 0
	_Android UserPlatform = 1
)

// user platform
var UserPlatformName = map[UserPlatform]string{
	1: "ANDROID",
	0: "IOS",
}

// user platform
var UserPlatformValue = map[string]UserPlatform{
	"ANDROID": 1,
	"IOS":     0,
}

// create new User
func NewUser() *User {
	return &User{}
}

// loading user info from redis
func LoadUser(id uint32) (*User, error) {

	// hget
	userID := strconv.Itoa(int(id))
	jsonData, err := userRedisClient.HGet("blue_server.user.json", userID).Result()
	if err != nil {
		return &User{}, err
	}
	urdata := UrData{}
	json.Unmarshal([]byte(jsonData), &urdata)
	return &User{
		Data:          urdata,
		Session:       nil,
		KeepaliveTime: time.Now()}, nil
}

// loading user info from redis
func LoadUserByName(name string) (*User, error) {
	result, err := userRedisClient.HGet("blue_server.user.id", name).Result()
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(result)
	if err != nil {
		return nil, err
	}

	return LoadUser(uint32(id))
}

// save redis user
func (u User) Save() error {

	id := strconv.Itoa(int(u.Data.ID))
	result, err := userRedisClient.HSet("blue_server.user.id", u.Data.Name, id).Result()
	if err != nil {
		return err
	}

	data, _ := json.Marshal(u.Data)
	result, err = userRedisClient.HSet("blue_server.user.json", id, string(data)).Result()
	if result == false || err != nil {
		return errors.New("already set data")
	}
	return nil
}

// generate id
func UserGenID() uint32 {
	genID, _ := userRedisClient.Incr("blue_server.manager.user.genid").Result()
	return uint32(genID)
}

//
func UpdateManager(id int) {
	serverStatus, _ := userRedisClient.Get("blue_server.manager.server.running").Result()
	log.Println(id, "server manager status:", serverStatus)
	if serverStatus == "FALSE" {
		network.StopServer()
		_, _ = userRedisClient.Set("blue_server.manager.server.running", "TRUE", 0).Result()
	}
}

//
func FindUser(session *network.Session) (*User, error) {
	ucmd := &UserCmdData{
		Cmd:     "FindUser",
		Session: session,
	}
	UserCmd <- ucmd
	ucmd = <-UserCmd
	if ucmd.Result != nil {
		return nil, ucmd.Result
	}
	return ucmd.User, nil
}

//
func FindUserByID(id uint32) (*User, error) {
	ucmd := &UserCmdData{
		Cmd: "FindUserByID",
		ID:  id,
	}
	UserCmd <- ucmd
	ucmd = <-UserCmd
	if ucmd.Result != nil {
		return nil, ucmd.Result
	}
	return ucmd.User, nil
}

//
func CheckUser() error {
	ucmd := &UserCmdData{
		Cmd: "CheckUser",
	}
	UserCmd <- ucmd
	ucmd = <-UserCmd
	if ucmd.Result != nil {
		return ucmd.Result
	}
	return nil
}

//
func AddUser(user *User) error {
	ucmd := &UserCmdData{
		Cmd:  "AddUser",
		ID:   user.Data.ID,
		User: user,
	}
	UserCmd <- ucmd
	ucmd = <-UserCmd
	if ucmd.Result != nil {
		return ucmd.Result
	}
	return nil
}

//
func DelUser(user *User) error {
	ucmd := &UserCmdData{
		Cmd: "DelUser",
		ID:  user.Data.ID,
	}
	UserCmd <- ucmd
	ucmd = <-UserCmd
	if ucmd.Result != nil {
		return ucmd.Result
	}
	return nil
}
