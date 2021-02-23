package main

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/omnilaboratory/obd/proxy/rpc"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/lightclient"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
)

func initObdLog() {
	_dir := "log"
	_ = tool.PathExistsAndCreate(_dir)
	path := "log/obdServer"
	writer, err := rotatelogs.New(
		path+".%Y%m%d%H%M.log",
		rotatelogs.WithMaxAge(30*34*time.Hour),
		rotatelogs.WithRotationTime(4*time.Hour),
	)

	if err != nil {
		panic(err)
	}
	writers := []io.Writer{
		os.Stdout,
		writer,
	}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	log.SetOutput(fileAndStdoutWriter)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

// gox compile  https://blog.csdn.net/han0373/article/details/81391455
// gox -os "windows linux darwin" -arch amd64
// gox -os "linux" -arch amd64
func main() {
	config.Init()
	initObdLog()
	//tracker
	err := lightclient.ConnectToTracker()
	if err != nil {
		log.Println("because fail to connect to tracker, obd fail to start")
		return
	}

	routersInit := lightclient.InitRouter()
	addr := ":" + strconv.Itoa(config.ServerPort)
	server := &http.Server{
		Addr:           addr,
		Handler:        routersInit,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	service.Start()

	// Timer
	service.ScheduleService.StartSchedule()

	go rpc.StartGrpcServer()

	log.Println("obd " + tool.GetObdNodeId() + " start in " + config.ChainNodeType)
	log.Println("wsAddress: " + bean.CurrObdNodeInfo.WebsocketLink)
	log.Fatal(server.ListenAndServe())
}
