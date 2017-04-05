package main

import (
	"fmt"

	"github.com/blueberryserver/tcpserver/network"
)

func main() {
	fmt.Printf("server start\r\n")

	// 서버 시작 리슨 요청
	server := network.NewNetServer("tcp", "127.0.0.1:10011", nil, nil)
	go server.Listen()

	// 클라이언트 접속 요청
	client := network.NewNetClient(nil, nil)
	go client.Connect("tcp", "127.0.0.1:10011")

	fmt.Println("server ready ..........")

	// 입력 대기
	var s string
	fmt.Scanf("%s", &s)
}
