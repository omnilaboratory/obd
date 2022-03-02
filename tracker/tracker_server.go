package main

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/config"
	"github.com/omnilaboratory/obd/tracker/router"
	"github.com/omnilaboratory/obd/tracker/rpc"
	"github.com/omnilaboratory/obd/tracker/service"
	"github.com/omnilaboratory/obd/tracker/tkrpc"
	"github.com/omnilaboratory/obd/tracker/tkrpc/inforpc"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
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
	service.Start(cfg.ChainNode_Type)

	go service.StartP2PNode()
	go StartGrpc()

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

	log.Println("tracker " + service.GetTrackerNodeId() + " start at port: " + strconv.Itoa(cfg.TrackerServerPort) + " in " + cfg.ChainNode_Type)
	log.Fatal(server.ListenAndServe())
}
func StartGrpc(){
	log.Println("StartGrpc")
	lis, err := net.Listen("tcp", "0.0.0.0:"+cfg.TrackerServerGrpcPort)
	if err != nil {
		log.Fatalf("无法绑定grpc地址: %v", err)
	}
	log.Println("grpc server start at: 0.0.0.0:"+cfg.TrackerServerGrpcPort)
	s := grpc.NewServer()
	//reflection.Register(s)
	tkrpc.RegisterInfoTrackerServer(s,&inforpc.ImpInfoServer{})
	s.Serve(lis)

}
