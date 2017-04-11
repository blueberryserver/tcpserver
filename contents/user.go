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
	ID         uint32
	Name       string
	Platform   UserPlatform
	VcGem      uint32
	VcGold     uint32
	Key        string
	Status     UserStatus
	LoginTime  time.Time
	CreateTime time.Time
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
		ID: 0}
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

	return User{ID: id,
		Name:       name,
		Platform:   iPlatform,
		VcGem:      uint32(iGem),
		VcGold:     uint32(iGold),
		Key:        hashkey,
		Status:     iStatus,
		LoginTime:  login,
		CreateTime: create}, nil
}

// save redis user
func (u User) Save(client *redis.Client) error {
	pipe := client.Pipeline()
	defer pipe.Close()

	pipe.Select(1)
	_, _ = pipe.Exec()

	id := strconv.Itoa(int(u.ID))
	result, err := client.HSet("blue_server.user.name", id, u.Name).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.hashkey", id, u.Key).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.platform", id, strconv.Itoa(int(u.Platform))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.login.status", id, strconv.Itoa(int(u.Status))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.vc.gem", id, strconv.Itoa(int(u.VcGem))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.vc.gold", id, strconv.Itoa(int(u.VcGold))).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.create.time", id, u.CreateTime.Format("2006-01-02 15:04:05")).Result()
	if err != nil {
		return err
	}

	result, err = client.HSet("blue_server.user.login.time", id, u.LoginTime.Format("2006-01-02 15:04:05")).Result()
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

	u.ID = id
}

// to string
func (u User) ToString() string {
	return fmt.Sprintf("%d %s %s %s %d %d %s %s", u.ID, UserPlatformName[u.Platform], u.Name, UserStatusName[u.Status], u.VcGem, u.VcGold,
		u.CreateTime.Format("2006-01-02 15:04:05"), u.LoginTime.Format("2006-01-02 15:04:05"))
}
