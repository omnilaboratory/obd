package main

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/lightclient"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func init() {
	_dir := "log"
	_ = tool.PathExistsAndCreate(_dir)
	path := "log/obdServer"
	writer, err := rotatelogs.New(
		path+".%Y%m%d%H%M.log",
		rotatelogs.WithMaxAge(time.Duration(12)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(12)*time.Hour),
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
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// gox compile  https://blog.csdn.net/han0373/article/details/81391455
// gox -os "windows linux darwin" -arch amd64
// gox -os "windows" -arch amd64
func main() {

	err := rpc.NewClient().CheckVersion()
	if err != nil {
		log.Println(err)
		log.Println("because get wrong omnicore version, obd fail to start")
		return
	}
	//tracker
	lightclient.ConnectToTracker()
	log.Println("obd " + tool.GetObdNodeId() + " start at port: " + strconv.Itoa(config.ServerPort))

	//StartP2PServer
	err = lightclient.StartP2PServer()
	if err != nil {
		log.Println(err)
		log.Println("because fail to start P2PServer, obd fail to start")
		return
	}

	routersInit := lightclient.InitRouter(nil)
	addr := ":" + strconv.Itoa(config.ServerPort)
	server := &http.Server{
		Addr:           addr,
		Handler:        routersInit,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	service.Start()

	//synData to tracker
	go lightclient.SynData()
	// Timer
	service.ScheduleService.StartSchedule()
	log.Fatal(server.ListenAndServe())
}
