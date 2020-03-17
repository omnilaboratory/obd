package main

import (
	"LightningOnOmni/config"
	"LightningOnOmni/lightclient"
	"LightningOnOmni/service"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"strconv"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
}

// gox compile  https://blog.csdn.net/han0373/article/details/81391455
// gox -os "windows linux darwin" -arch amd64
func main() {

	// grpc
	//go grpcpack.Server()
	//conn := startupGRPCClient()
	//defer conn.Close()
	//routersInit := routers.InitRouter(conn)

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
