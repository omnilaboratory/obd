package main

import (
	fmt "fmt"
	"log"
	"net"

	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/proxy/rpc"
	grpc "google.golang.org/grpc"
)

func startServer() {

	address := "localhost:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	fmt.Printf("Server is listening on %v ...", address)

	s := grpc.NewServer()
	proxy.RegisterProxyServer(s, &rpc.RpcServer{})
	s.Serve(lis)
}

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	startServer()
}
