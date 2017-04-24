package contents

import (
	"log"

	redis "gopkg.in/redis.v4"
)

// room,  channel db(2)
var rmchRedisClient *redis.Client

//
func SetRmChRedisClient(client *redis.Client) {
	rmchRedisClient = client

	pipe := rmchRedisClient.Pipeline()
	defer pipe.Close()
	pipe.Select(2)
	_, _ = pipe.Exec()
}

// channel config data
type ChData struct {
	ChNo    uint32 `json:"chno"`
	ChType  ChType `json:"chtype"`
	ChLimit uint32 `json:"chlimit"`
}

// channel
type Channel struct {
	data    ChData
	members map[uint32]*User // channel user
}

// channel type
type ChType uint32

// user status
const (
	_ChDefault ChType = 0
	_ChNormal  ChType = 10
	_ChLevel1  ChType = 1
	_ChLevel2  ChType = 2
)

// room status
var ChTypeName = map[ChType]string{
	0:  "DEFAULT",
	10: "NORMAL",
	1:  "LEVEL1",
	2:  "LEVEL2",
}

// room status
var ChTypeValue = map[string]ChType{
	"DEFAULT": 0,
	"NORMAL":  10,
	"LEVEL1":  1,
	"LEVEL2":  2,
}

//
func EnterCh(no uint32, user *User) error {
	chcmd := &ChCmdData{
		Cmd:  "EnterCh",
		No:   no,
		User: user,
	}

	ChCmd <- chcmd
	chcmd = <-ChCmd
	if chcmd.Result != nil {
		return chcmd.Result
	}
	return chcmd.Result
}

//
func LeaveCh(no uint32, user *User) error {
	chcmd := &ChCmdData{
		Cmd:  "LeaveCh",
		No:   no,
		User: user,
	}

	ChCmd <- chcmd
	chcmd = <-ChCmd
	if chcmd.Result != nil {
		return chcmd.Result
	}
	return chcmd.Result
}

//
func LoadChannel() error {
	log.Println("loading channel info")
	chcmd := &ChCmdData{
		Cmd: "LoadCh",
	}

	ChCmd <- chcmd
	chcmd = <-ChCmd
	if chcmd.Result != nil {
		return chcmd.Result
	}
	return chcmd.Result
}

//
func ListChannel() string {
	chcmd := &ChCmdData{
		Cmd:     "ListCh",
		Monitor: "",
	}

	ChCmd <- chcmd
	chcmd = <-ChCmd
	if chcmd.Result != nil {
		return ""
	}
	return chcmd.Monitor
}
