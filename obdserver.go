package main

import (
	"github.com/lestrrat-go/file-rotatelogs"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/proxy/rpc"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/lightclient"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"

	grpc "google.golang.org/grpc"
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

	go startGrpcServer()

	log.Println("obd " + tool.GetObdNodeId() + " start in " + config.ChainNodeType)
	log.Println("wsAddress: " + bean.CurrObdNodeInfo.WebsocketLink)
	log.Fatal(server.ListenAndServe())
}

func startGrpcServer() {

	log.Println("startGrpcServer")
	rpc.ConnToObd()

	address := "localhost:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	log.Printf("grpc Server is listening on %v ...", address)

	s := grpc.NewServer()
	proxy.RegisterLightningServer(s, &rpc.RpcServer{})
	s.Serve(lis)
}
