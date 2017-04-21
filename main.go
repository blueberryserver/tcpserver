package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	_ "strconv"
	"time"

	redis "gopkg.in/redis.v4"

	"log"

	"github.com/blueberryserver/tcpserver/contents"
	"github.com/blueberryserver/tcpserver/msg"
	"github.com/blueberryserver/tcpserver/network"
	"github.com/blueberryserver/tcpserver/util"
)

var serverConfig *util.Config

func main() {
	//file, err := os.OpenFile("pprof.csv", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	//recorder := pprof.NewTimeRecorder()
	//contents.SetTimeRecorder(recorder)
	//summary := pprof.GCSummary()

	serverConfig = util.LoadConfig("conf.json")
	if serverConfig == nil {
		fmt.Println("conf.json file loading fail")
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

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
		Addr:     serverConfig.REDISAddr,
		Password: "", // no password set
		DB:       1,  // use default DB
	})
	defer userClient.Close()

	rmchClient := redis.NewClient(&redis.Options{
		Addr:     serverConfig.REDISAddr,
		Password: "", // no password set
		DB:       2,  // use default DB
	})
	defer rmchClient.Close()

	// set redis client
	contents.SetUserRedisClient(userClient)
	contents.SetRmChRedisClient(rmchClient)

	// generate channel list
	contents.LoadChannel()
	contents.LoadRoom()

	// monitoring
	// monitorID := make(chan int)
	// go monitor(monitorID)
	// monitorID <- 1

	updateID := make(chan int)
	go update(updateID)
	updateID <- 2

	// wait 1 second
	time.Sleep(1 * time.Second)

	// http server
	go httpServer()

	// server start
	ServerStart()
}

// server start
func ServerStart() {
	log.Printf("server start\r\n")
	server := network.NewServer("tcp", serverConfig.TCPAddr, contents.CloseHandler)
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

func update(c chan int) {
	id := <-c
	for {
		time.Sleep(10 * time.Second)
		contents.UpdateChannel(id)
		contents.UpdateManager(id)
	}
}

func httpServer() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		//res.Write([]byte("Hello, world!")) // 웹 브라우저에 응답
		str := "<p>.......... ..........................</p>"
		str += contents.MonitorChannel()
		str += "<p>.....................................</p>"
		html := `
		<html>
		<head>
			<title>Montor</title>
			<meta http-equiv="refresh" content="10; url=/" />
		</head>
		<body>
			<span class="montor">` + str + `</span>
		</body>
		</html>
		`
		res.Header().Set("Content-Type", "text/html")
		res.Write([]byte(html))
	})

	log.Println("server httpstart")
	http.ListenAndServe(serverConfig.HTTPAddr, nil)
}
