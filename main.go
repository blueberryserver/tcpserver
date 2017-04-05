package main

import (
	"fmt"

	"github.com/blueberryserver/tcpserver/network"
)

func main() {
	fmt.Printf("server start\r\n")

	// 서버 시작 리슨 요청
	server := network.NewServer("tcp", "127.0.0.1:10011")
	err := server.Listen()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 클라이언트 접속 요청
	client := network.NewClient()
	err = client.Connect("tcp", "127.0.0.1:10011")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 입력 대기
	var s string
	fmt.Scanf("%s", &s)
}
