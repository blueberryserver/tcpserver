package main

import (
	"fmt"
	_ "strconv"
	"time"

	redis "gopkg.in/redis.v4"

	"github.com/blueberryserver/tcpserver/contents"
	"github.com/blueberryserver/tcpserver/msg"
	"github.com/blueberryserver/tcpserver/network"
	"github.com/golang/protobuf/proto"
)

func main() {

	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	network.SetRedisClient(client)

	//protobufTest()
	netTest()
	//redisTest()
	// 입력 대기
	var s string
	fmt.Scanf("%s", &s)
}

func netTest() {
	fmt.Printf("net test\r\n")
	// networt test
	// 서버 시작 리슨 요청
	server := network.NewServer("tcp", ":20202")

	// ping req 처리 핸들러 등록
	server.AddMsgHandler(msg.Msg_Id_value["Ping_Req"], network.GetHandlerReqPing(msg.Msg_Id_value["Ping_Req"]))
	server.AddMsgHandler(msg.Msg_Id_value["Login_Req"], network.GetHandlerReqLogin(msg.Msg_Id_value["Login_Req"]))
	server.AddMsgHandler(msg.Msg_Id_value["Relay_Req"], network.GetHandlerReqRelay(msg.Msg_Id_value["Relay_Req"]))

	err := server.Listen()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 1초간 대기
	time.Sleep(1 * time.Second)

	// 클라이언트 접속 요청
	client := network.NewClient()
	// pong Ans 처리 핸들러 등록
	client.AddMsgHandler(msg.Msg_Id_value["Pong_Ans"], network.GetHandlerAnsPong(msg.Msg_Id_value["Pong_Ans"]))
	client.AddMsgHandler(msg.Msg_Id_value["Login_Ans"], network.GetHandlerAnsLogin(msg.Msg_Id_value["Login_Ans"]))
	client.AddMsgHandler(msg.Msg_Id_value["Relay_Ans"], network.GetHandlerAnsRelay(msg.Msg_Id_value["Relay_Ans"]))
	client.AddMsgHandler(msg.Msg_Id_value["Relay_Not"], network.GetHandlerNotRelay(msg.Msg_Id_value["Relay_Not"]))

	// 연결 시도
	err = client.Connect("tcp", ":20202")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 1초간 대기
	{
		time.Sleep(1 * time.Second)

		buff := make([]byte, 4096)
		str := "hello world"
		copy(buff, str)
		client.SendPacket(msg.Msg_Id_value["Ping_Req"], buff[:len(str)], uint16(len(str)))
	}

	{
		m := &msg.LoginReq{
			Id: proto.String("noom"),
		}
		data, err := proto.Marshal(m)
		if err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(1 * time.Second)
		client.SendPacket(msg.Msg_Id_value["Login_Req"], data, uint16(len(data)))
	}

	{
		// relay data
		m := &msg.RelayReq{
			RmNo: proto.Uint32(1),
			Data: proto.String("data:1723849, contens: 387438"),
		}

		data, err := proto.Marshal(m)
		if err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(1 * time.Second)
		client.SendPacket(msg.Msg_Id_value["Relay_Req"], data, uint16(len(data)))
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

func redisTest() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	vals, err := client.HGetAll("cart.user:1").Result()
	fmt.Println(vals, err)

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	// add user, room dummy data for redis
	// selectdb 1
	// pipe := client.Pipeline()
	// defer pipe.Close()
	// pipe.Select(1)
	// _, _ = pipe.Exec()

	// // user obj info
	// fid := 1234
	// result, err := client.HSet("blue_server.user.name", strconv.Itoa(fid), "noom").Result()
	// _, _ = client.HSet("blue_server.user.hashkey", strconv.Itoa(fid), "1234%^&").Result()
	// _, _ = client.HSet("blue_server.user.create.time", strconv.Itoa(fid), "2017-04-10 17:00:30").Result()
	// _, _ = client.HSet("blue_server.user.platform", strconv.Itoa(fid), "android").Result()
	// _, _ = client.HSet("blue_server.user.login.status", strconv.Itoa(fid), "logon").Result()
	// _, _ = client.HSet("blue_server.user.login.time", strconv.Itoa(fid), "2017-04-10 17:00:30").Result()
	// _, _ = client.HSet("blue_server.user.vc.gem", strconv.Itoa(fid), "1000").Result()
	// _, _ = client.HSet("blue_server.user.vc.gold", strconv.Itoa(fid), "99999").Result()
	// fmt.Println(result, err)

	// // selectdb 2
	// pipe = client.Pipeline()
	// pipe.Select(2)
	// _, _ = pipe.Exec()

	// rid := 1
	// result, err = client.HSet("blue_server.room.type", strconv.Itoa(rid), "NORMAL").Result()
	// _, _ = client.HSet("blue_server.room.status", strconv.Itoa(rid), "NONE").Result()
	// _, _ = client.HSet("blue_server.room.create.time", strconv.Itoa(rid), "2017-04-11 11:46:12").Result()
	// _, _ = client.HSet("blue_server.room.member", strconv.Itoa(rid), "[1234]").Result()
	// fmt.Println(result, err)

	// room obj info
	user, err := contents.LoadUser(1234, client)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(user.ToString())
	room, err := contents.LoadRoom(1, client)
	if err == nil {
		room.AddMember(user)
		fmt.Println(room.ToString())
	}
}
