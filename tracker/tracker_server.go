package main

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/config"
	"github.com/omnilaboratory/obd/tracker/router"
	"github.com/omnilaboratory/obd/tracker/rpc"
	"github.com/omnilaboratory/obd/tracker/service"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func initTrackerLog() {
	_dir := "log"
	_ = tool.PathExistsAndCreate(_dir)
	path := "log/tracker"
	writer, err := rotatelogs.New(
		path+".%Y%m%d%H%M.log",
		rotatelogs.WithMaxAge(30*24*time.Hour),
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
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// gox -os "windows linux darwin" -arch amd64
// gox -os "linux" -arch amd64
func main() {
	initTrackerLog()

	err := rpc.NewClient().CheckVersion()
	if err != nil {
		log.Println(err)
		log.Println("because get wrong omniCore version, tracker fail to start")
		return
	}

	routersInit := router.InitRouter()
	if routersInit == nil {
		log.Println("fail to start tracker")
		return
	}
	addr := ":" + strconv.Itoa(cfg.TrackerServerPort)
	server := &http.Server{
		Addr:           addr,
		Handler:        routersInit,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	service.Start(service.ChannelService.BtcChainType)

	log.Println("tracker " + tool.GetTrackerNodeId() + " start at port: " + strconv.Itoa(cfg.TrackerServerPort) + " in " + service.ChannelService.BtcChainType)
	log.Fatal(server.ListenAndServe())
}
