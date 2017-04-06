package main

import (
	"fmt"
	"time"

	"github.com/blueberryserver/tcpserver/msg"
	"github.com/blueberryserver/tcpserver/network"
	"github.com/golang/protobuf/proto"
)

func main() {
	protobufTest()
	netTest()
	// 입력 대기
	var s string
	fmt.Scanf("%s", &s)
}

func netTest() {
	fmt.Printf("net test\r\n")
	// networt test
	// 서버 시작 리슨 요청
	server := network.NewServer("tcp", ":20202")
	err := server.Listen()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 10초간 대기
	time.Sleep(1 * time.Second)

	// 클라이언트 접속 요청
	client := network.NewClient()
	err = client.Connect("tcp", ":20202")
	if err != nil {
		fmt.Println(err)
		return
	}

}

func protobufTest() {
	fmt.Printf("protobuf test\r\n")
	// proto test
	// enum type setting
	smalltype := msg.TestMessage_SmallType(msg.TestMessage_HARD)
	testtype := msg.TestType(msg.TestType_TYPE_1)

	message := &msg.TestMessage{
		TestString:    proto.String("Test String"),
		TestUint32:    proto.Uint32(100),
		TestSmallType: &smalltype,
		TestTestType:  &testtype,
		TestBool:      proto.Bool(false),
		TestInt32:     proto.Int32(1000),
		TestUint64:    proto.Uint64(10384),
		TestFloat:     proto.Float32(2398.45),
	}

	data, err := proto.Marshal(message)
	if err != nil {
		fmt.Println(err)
		return
	}

	newMessage := &msg.TestMessage{}
	err = proto.Unmarshal(data, newMessage)
	if err != nil {
		fmt.Println(err)
		return
	}

	if message.GetTestString() != newMessage.GetTestString() {
		fmt.Printf("%s %s\r\n", message.GetTestString(), newMessage.GetTestString())
		return
	}

	fmt.Printf("msaage: %v\r\n", message)
}
