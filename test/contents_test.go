package test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/blueberryserver/tcpserver/contents"
	redis "gopkg.in/redis.v4"
)

func TestUserToJson(t *testing.T) {
	// time zone setting
	//time.FixedZone("Asia/Seoul", 9*60*60)
	//fmt.Printf("Now: %s\r\n", time.Now().Format("2006-01-02 15:04:05"))

	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})

	contents.SetUserRedisClient(client)
	user, err := contents.LoadUser(1234)
	if err != nil {
		t.Error("user to json fail")
	}

	user.Data.LoginTime = time.Now()

	data, _ := json.Marshal(user.Data)
	fmt.Println(string(data))

	_, _ = client.HSet("blue_server.user.json", "1234", string(data)).Result()
	client.Close()
}

func TestJsonToUser(t *testing.T) {
	// time zone setting
	time.FixedZone("Asia/Seoul", 9*60*60)
	fmt.Printf("Now: %s\r\n", time.Now().Format("2006-01-02 15:04:05"))

	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})

	jsonData, _ := client.HGet("blue_server.user.json", "1234").Result()
	urData := contents.UrData{}
	json.Unmarshal([]byte(jsonData), &urData)
	fmt.Println(urData)
	fmt.Println(urData.LoginTime.Format("2006-01-02 15:04:05"))
	client.Close()
}

func TestUserProc(t *testing.T) {
	go contents.UserProcFunc()

	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})

	contents.SetUserRedisClient(client)
	user, err := contents.LoadUser(1234)
	if err != nil {
		t.Error("user to json fail")
	}

	cmd := &contents.UserCmdData{
		Cmd:  "AddUser",
		ID:   user.Data.ID,
		User: user}

	contents.UserCmd <- cmd
	cmd = <-contents.UserCmd

	fmt.Println("-----------------------------------------")
	fmt.Println(cmd.Result)

	cmd.Cmd = "ListUser"

	contents.UserCmd <- cmd
	cmd = <-contents.UserCmd

	fmt.Println("-----------------------------------------")
	fmt.Println(cmd.Result)

	cmd.Cmd = "DelUser"

	contents.UserCmd <- cmd
	cmd = <-contents.UserCmd

	fmt.Println("-----------------------------------------")
	fmt.Println(cmd.Result)

	cmd.Cmd = "ListUser"

	contents.UserCmd <- cmd
	cmd = <-contents.UserCmd

	fmt.Println("-----------------------------------------")
	fmt.Println(cmd.Result)
	fmt.Println("-----------------------------------------")
}
