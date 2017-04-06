package main

import (
	"fmt"
	"time"

	"github.com/blueberryserver/tcpserver/network"
)

func main() {
	fmt.Printf("server start\r\n")

	// 서버 시작 리슨 요청
	server := network.NewServer("tcp", ":20202")
	err := server.Listen()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 10초간 대기
	time.Sleep(10 * time.Second)

	// 클라이언트 접속 요청
	client := network.NewClient()
	err = client.Connect("tcp", "13.124.76.58:20202")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 입력 대기
	var s string
	fmt.Scanf("%s", &s)
}
