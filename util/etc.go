package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

// md5 hash
func HashMD5(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// generator random letter
func RandStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// config struct
type Config struct {
	TCPAddr   string `json:"tcpaddr"`
	HTTPAddr  string `json:"httpaddr"`
	REDISAddr string `json:"redisaddr"`
}

// load config
func LoadConfig(fileName string) *Config {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var config Config

	jsonParser := json.NewDecoder(file)
	if err := jsonParser.Decode(&config); err != nil {
		fmt.Println(err)
		return nil
	}

	//fmt.Println(config)
	return &config
}
