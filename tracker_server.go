package main

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker"
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
	path := "log/tracker"
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
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
}

func main() {

	routersInit := tracker.InitRouter()
	addr := ":" + strconv.Itoa(config.TrackerServerPort)
	server := &http.Server{
		Addr:           addr,
		Handler:        routersInit,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("tracker " + tool.GetTrackerNodeId() + " start at " + addr)
	log.Fatal(server.ListenAndServe())
}
