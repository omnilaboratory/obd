package rpc

import (
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

func StartGrpcServer() {

	log.Println("startGrpcServer")
	ConnToObd()

	address := "localhost:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	log.Printf("grpc Server is listening on %v ...", address)

	s := grpc.NewServer()
	proxy.RegisterLightningServer(s, &RpcServer{})
	proxy.RegisterWalletServer(s, &RpcServer{})
	proxy.RegisterHtlcServer(s, &RpcServer{})
	s.Serve(lis)
}
