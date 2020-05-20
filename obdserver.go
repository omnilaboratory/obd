package main

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/lightclient"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"google.golang.org/grpc"
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

	log.Println("obd " + tool.GetObdNodeId() + " start at " + addr)
	log.Fatal(server.ListenAndServe())
}

func startupGRPCClient() *grpc.ClientConn {
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(config.GrpcPort), grpc.WithInsecure())
	if err != nil {
		log.Println("did not connect: ", err)
	}
	return conn
}
