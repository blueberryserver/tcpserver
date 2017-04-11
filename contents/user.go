package contents

import (
	"errors"
	"strconv"
	"time"

	"fmt"

	"gopkg.in/redis.v4"
)

// user obj
type User struct {
	uID        uint32
	name       string
	platform   UserPlatform
	vcGem      uint32
	vcGold     uint32
	key        string
	status     UserStatus
	loginTime  time.Time
	createTime time.Time
}

// user status
type UserStatus uint32

// user status
const (
	_Logon  UserStatus = 1
	_LogOff UserStatus = 2
)

// room status
var UserStatusName = map[UserStatus]string{
	1: "LOGON",
	2: "LOGOFF",
}

// room status
var UserStatusValue = map[string]UserStatus{
	"LOGON":  1,
	"LOGOFF": 2,
}

// user status
type UserPlatform uint32

// user status
const (
	_IOS     UserPlatform = 0
	_Android UserPlatform = 1
)

// room status
var UserPlatformName = map[UserPlatform]string{
	1: "ANDROID",
	0: "IOS",
}

// room status
var UserPlatformValue = map[string]UserPlatform{
	"ANDROID": 1,
	"IOS":     0,
}

// create new User
func NewUser() User {
	return User{
		uID: 0}
}

// loading user info from redis
func LoadUser(id uint32, client *redis.Client) (User, error) {

	// redis slelct db 1(user)
	pipe := client.Pipeline()
	defer pipe.Close()

	pipe.Select(1)
	_, _ = pipe.Exec()

	// hget
	userID := strconv.Itoa(int(id))
	name, err := client.HGet("blue_server.user.name", userID).Result()
	if err != nil {
		return User{}, err
	}

	hashkey, err := client.HGet("blue_server.user.hashkey", userID).Result()
	if err != nil {
		return User{}, err
	}
	createTime, err := client.HGet("blue_server.user.create.time", userID).Result()
	if err != nil {
		return User{}, err
	}
	platform, err := client.HGet("blue_server.user.platform", userID).Result()
	if err != nil {
		return User{}, err
	}
	loginStatus, err := client.HGet("blue_server.user.login.status", userID).Result()
	if err != nil {
		return User{}, err
	}
	loginTime, err := client.HGet("blue_server.user.login.time", userID).Result()
	if err != nil {
		return User{}, err
	}
	gem, err := client.HGet("blue_server.user.vc.gem", userID).Result()
	if err != nil {
		return User{}, err
	}
	gold, err := client.HGet("blue_server.user.vc.gold", userID).Result()
	if err != nil {
		return User{}, err
	}

	iPlatform := UserPlatformValue[platform]
	if err != nil {
		return User{}, err
	}
	iGem, err := strconv.Atoi(gem)
	if err != nil {
		return User{}, err
	}
	iGold, err := strconv.Atoi(gold)
	if err != nil {
		return User{}, err
	}
	login, err := time.Parse("2006-01-02 15:04:05", loginTime)
	if err != nil {
		return User{}, err
	}
	create, err := time.Parse("2006-01-02 15:04:05", createTime)
	if err != nil {
		return User{}, err
	}
	iStatus := UserStatusValue[loginStatus]

	return User{uID: id,
		name:       name,
		platform:   iPlatform,
		vcGem:      uint32(iGem),
		vcGold:     uint32(iGold),
		key:        hashkey,
		status:     iStatus,
		loginTime:  login,
		createTime: create}, nil
}

// save redis user
func (u User) Save(client *redis.Client) error {
	pipe := client.Pipeline()
	defer pipe.Close()

	pipe.Select(1)
	_, _ = pipe.Exec()

	id := strconv.Itoa(int(u.uID))
	result, err := client.HSet("blue_server.user.name", id, u.name).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.hashkey", id, u.key).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.platform", id, strconv.Itoa(int(u.platform))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.login.status", id, strconv.Itoa(int(u.status))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.vc.gem", id, strconv.Itoa(int(u.vcGem))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.vc.gold", id, strconv.Itoa(int(u.vcGold))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.create.time", id, u.createTime.Format("2006-01-02 15:04:05")).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.login.time", id, u.loginTime.Format("2006-01-02 15:04:05")).Result()
	if err != nil {
		return err
	}

	if result == false {
		return errors.New("already set data")
	}

	return nil
}

// setting user id
func (u User) SetID(id uint32) {

	u.uID = id
}

// to string
func (u User) ToString() string {
	return fmt.Sprintf("%d %s %s %s %d %d %s %s", u.uID, UserPlatformName[u.platform], u.name, UserStatusName[u.status], u.vcGem, u.vcGold,
		u.createTime.Format("2006-01-02 15:04:05"), u.loginTime.Format("2006-01-02 15:04:05"))
}
