package test

import (
	"testing"

	"fmt"

	redis "gopkg.in/redis.v4"
)

func TestRedisConnect(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer client.Close()

	if client == nil {
		t.Error("redis connect fail")
	}
	fmt.Println("TestRedisConnect result: success")
}
