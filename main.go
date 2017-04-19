package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	_ "strconv"
	"time"

	redis "gopkg.in/redis.v4"

	"log"

	"github.com/blueberryserver/tcpserver/contents"
	"github.com/blueberryserver/tcpserver/msg"
	"github.com/blueberryserver/tcpserver/network"
)

func main() {
	//file, err := os.OpenFile("pprof.csv", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	//recorder := pprof.NewTimeRecorder()
	//contents.SetTimeRecorder(recorder)
	//summary := pprof.GCSummary()

	runtime.GOMAXPROCS(runtime.NumCPU())
	//runtime.LockOSThread()

	logfile := "log_" + time.Now().Format("2006_01_02_15") + ".txt"
	fileLog, err := os.OpenFile(logfile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	defer fileLog.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	mutiWriter := io.MultiWriter(fileLog, os.Stdout)
	log.SetOutput(mutiWriter)

	userClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})
	defer userClient.Close()

	rmchClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       2,  // use default DB
	})
	defer rmchClient.Close()

	// set redis client
	contents.SetUserRedisClient(userClient)
	contents.SetRmChRedisClient(rmchClient)

	// generate channel list
	//contents.NewChannel()
	contents.LoadChannel()
	contents.LoadRoom()

	// monitoring
	monitorID := make(chan int)
	go monitor(monitorID)
	monitorID <- 1

	updateID := make(chan int)
	go update(updateID)
	updateID <- 2

	// server start
	ServerStart()

	// wait 1 second
	time.Sleep(1 * time.Second)
}

// server start
func ServerStart() {
	log.Printf("server start\r\n")
	server := network.NewServer("tcp", ":20202", contents.CloseHandler)
	network.SetGlobalNetServer(server)

	// regist server handler
	server.AddMsgHandler(msg.Msg_Id_value["Ping_Req"], contents.GetHandlerReqPing())
	server.AddMsgHandler(msg.Msg_Id_value["Login_Req"], contents.GetHandlerReqLogin())
	server.AddMsgHandler(msg.Msg_Id_value["Relay_Req"], contents.GetHandlerReqRelay())
	server.AddMsgHandler(msg.Msg_Id_value["Enter_Ch_Req"], contents.GetHandlerReqEnterCh())
	server.AddMsgHandler(msg.Msg_Id_value["Enter_Rm_Req"], contents.GetHandlerReqEnterRm())
	server.AddMsgHandler(msg.Msg_Id_value["Leave_Rm_Req"], contents.GetHandlerReqLeaveRm())
	server.AddMsgHandler(msg.Msg_Id_value["Regist_Req"], contents.GetHandlerReqRegist())
	server.AddMsgHandler(msg.Msg_Id_value["List_Rm_Req"], contents.GetHandlerReqListRm())

	c := make(chan bool)
	server.Listen(&c)
	_ = <-c

	//var s string
	//fmt.Scanf("%s", &s)
	//server.Stop()

}

func monitor(c chan int) {
	id := <-c
	for {
		time.Sleep(10 * time.Second)
		log.Println(id, "monitoring ..........................")
		contents.MonitorChannel(id)
		log.Println(id, "......................................")
	}
}

func update(c chan int) {
	id := <-c
	for {
		time.Sleep(10 * time.Second)
		contents.UpdateChannel(id)
		contents.UpdateManager(id)
	}
}
