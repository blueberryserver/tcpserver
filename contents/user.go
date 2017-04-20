package contents

import (
	"errors"
	"fmt"
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

// user obj
type User struct {
	ID         uint32
	Name       string
	Platform   UserPlatform
	VcGem      uint32
	VcGold     uint32
	Key        string
	Status     UserStatus
	ChNo       uint32
	RmNo       uint32
	LoginTime  time.Time
	LogoutTime time.Time
	CreateTime time.Time

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
	return &User{
		ID: 0}
}

// loading user info from redis
func LoadUser(id uint32) (*User, error) {

	// hget
	userID := strconv.Itoa(int(id))
	name, err := userRedisClient.HGet("blue_server.user.name", userID).Result()
	if err != nil {
		return &User{}, err
	}

	hashkey, err := userRedisClient.HGet("blue_server.user.hashkey", userID).Result()
	if err != nil {
		return &User{}, err
	}
	createTime, err := userRedisClient.HGet("blue_server.user.create.time", userID).Result()
	if err != nil {
		return &User{}, err
	}
	platform, err := userRedisClient.HGet("blue_server.user.platform", userID).Result()
	if err != nil {
		return &User{}, err
	}
	loginStatus, err := userRedisClient.HGet("blue_server.user.login.status", userID).Result()
	if err != nil {
		return &User{}, err
	}
	rmNo, err := userRedisClient.HGet("blue_server.user.room.no", userID).Result()
	if err != nil {
		return &User{}, err
	}
	loginTime, err := userRedisClient.HGet("blue_server.user.login.time", userID).Result()
	if err != nil {
		return &User{}, err
	}
	gem, err := userRedisClient.HGet("blue_server.user.vc.gem", userID).Result()
	if err != nil {
		return &User{}, err
	}
	gold, err := userRedisClient.HGet("blue_server.user.vc.gold", userID).Result()
	if err != nil {
		return &User{}, err
	}

	iPlatform := UserPlatformValue[platform]
	if err != nil {
		return &User{}, err
	}
	iGem, err := strconv.Atoi(gem)
	if err != nil {
		return &User{}, err
	}
	iGold, err := strconv.Atoi(gold)
	if err != nil {
		return &User{}, err
	}
	login, err := time.Parse("2006-01-02 15:04:05", loginTime)
	if err != nil {
		return &User{}, err
	}
	create, err := time.Parse("2006-01-02 15:04:05", createTime)
	if err != nil {
		return &User{}, err
	}
	iStatus := UserStatusValue[loginStatus]
	iRmNo, err := strconv.Atoi(rmNo)
	if err != nil {
		return &User{}, err
	}

	return &User{ID: id,
		Name:       name,
		Platform:   iPlatform,
		VcGem:      uint32(iGem),
		VcGold:     uint32(iGold),
		Key:        hashkey,
		Status:     iStatus,
		ChNo:       0,
		RmNo:       uint32(iRmNo),
		LoginTime:  login,
		CreateTime: create,
		Session:    nil}, nil
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

	id := strconv.Itoa(int(u.ID))
	result, err := userRedisClient.HSet("blue_server.user.id", u.Name, id).Result()
	if err != nil {
		return err
	}

	result, err = userRedisClient.HSet("blue_server.user.name", id, u.Name).Result()
	if err != nil {
		return err
	}

	result, err = userRedisClient.HSet("blue_server.user.hashkey", id, u.Key).Result()
	if err != nil {
		return err
	}

	result, err = userRedisClient.HSet("blue_server.user.platform", id, UserPlatformName[u.Platform]).Result()
	if err != nil {
		return err
	}

	result, err = userRedisClient.HSet("blue_server.user.login.status", id, UserStatusName[u.Status]).Result()
	if err != nil {
		return err
	}

	result, err = userRedisClient.HSet("blue_server.user.vc.gem", id, strconv.Itoa(int(u.VcGem))).Result()
	if err != nil {
		return err
	}

	result, err = userRedisClient.HSet("blue_server.user.vc.gold", id, strconv.Itoa(int(u.VcGold))).Result()
	if err != nil {
		return err
	}

	result, err = userRedisClient.HSet("blue_server.user.create.time", id, u.CreateTime.Format("2006-01-02 15:04:05")).Result()
	if err != nil {
		return err
	}

	result, err = userRedisClient.HSet("blue_server.user.login.time", id, u.LoginTime.Format("2006-01-02 15:04:05")).Result()
	if err != nil {
		return err
	}
	if u.Status == _LogOff {
		result, err = userRedisClient.HSet("blue_server.user.logout.time", id, u.LogoutTime.Format("2006-01-02 15:04:05")).Result()
		if err != nil {
			return err
		}
	}
	result, err = userRedisClient.HSet("blue_server.user.room.no", id, strconv.Itoa(int(u.RmNo))).Result()
	if err != nil {
		return err
	}

	if result == false {
		return errors.New("already set data")
	}

	return nil
}

// to string
func (u User) ToString() string {
	return fmt.Sprintf("ID:%d Platform:%s Name:%s Status:%s Gem:%d Gold:%d Create Time:%s Login Time:%s",
		u.ID, UserPlatformName[u.Platform], u.Name, UserStatusName[u.Status], u.VcGem, u.VcGold,
		u.CreateTime.Format("2006-01-02 15:04:05"), u.LoginTime.Format("2006-01-02 15:04:05"))
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
