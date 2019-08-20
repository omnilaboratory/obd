package main

import (
	"LightningOnOmni/config"
	"LightningOnOmni/grpcpack"
	"LightningOnOmni/routers"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"strconv"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	//service.ScheduleService.StartSchudule()

	go grpcpack.Server()

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("did not connect: %v", err)
	}
	defer conn.Close()

	routersInit := routers.InitRouter(conn)
	addr := ":" + strconv.Itoa(config.ServerPort)
	server := &http.Server{
		Addr:           addr,
		Handler:        routersInit,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(server.ListenAndServe())

}
