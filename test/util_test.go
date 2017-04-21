package test

import (
	"testing"

	"fmt"

	"github.com/blueberryserver/tcpserver/util"
)

// hash test
func TestHashMD5(t *testing.T) {
	hashStr := util.HashMD5("noom")
	if hashStr == "" {
		t.Error("null string")
	}
	fmt.Println("TestHashMD5 result:", hashStr)
}

//rand test
func TestRandStr(t *testing.T) {
	randStr := util.RandStr(15)
	if randStr == "" {
		t.Error("null string")
	}
	fmt.Println("TestRandStr result:", randStr)
}

func TestLoadConfig(t *testing.T) {
	config := util.LoadConfig("conf.json")
	if config == nil {
		t.Error("config file loading fail")
	}
	fmt.Println("TestLoadConfig result:", config)
}
