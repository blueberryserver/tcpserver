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

	user.LoginTime = time.Now()

	data, _ := json.Marshal(user)
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
	user := contents.User{}
	json.Unmarshal([]byte(jsonData), &user)
	fmt.Println(user)
	fmt.Println(user.LoginTime.Format("2006-01-02 15:04:05"))
	client.Close()
}
