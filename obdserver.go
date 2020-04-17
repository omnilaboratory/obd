package main

import (
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
	"obd/config"
	"obd/lightclient"
	"obd/rpc"
	"obd/service"
	"obd/tool"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	_dir := "log"
	_ = tool.PathExistsAndCreate(_dir)
	file := _dir + "/logFile" + strings.Split(time.Now().String(), " ")[0] + ".log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	writers := []io.Writer{
		logFile,
		os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	log.SetOutput(fileAndStdoutWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
}

// gox compile  https://blog.csdn.net/han0373/article/details/81391455
// gox -os "windows linux darwin" -arch amd64
// gox -os "windows" -arch amd64
func main() {

	// grpc
	//go grpcpack.Server()
	//conn := startupGRPCClient()
	//defer conn.Close()
	//routersInit := routers.InitRouter(conn)

	err := rpc.NewClient().CheckVersion()
	if err != nil {
		log.Println(err)
		log.Println("obd fail to start")
		return
	}

	lightclient.StartP2PServer()
	routersInit := lightclient.InitRouter(nil)
	addr := ":" + strconv.Itoa(config.ServerPort)
	server := &http.Server{
		Addr:           addr,
		Handler:        routersInit,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Timer
	service.ScheduleService.StartSchedule()

	log.Fatal(server.ListenAndServe())

}

func startupGRPCClient() *grpc.ClientConn {
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(config.GrpcPort), grpc.WithInsecure())
	if err != nil {
		log.Println("did not connect: ", err)
	}
	return conn
}
