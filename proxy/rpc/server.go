package rpc

import (
	"github.com/omnilaboratory/obd/config"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strconv"
)

func StartGrpcServer() {

	log.Println("startGrpcServer")
	err := ConnToObd()
	if err != nil {
		log.Println(err)
		return
	}

	address := "localhost:" + strconv.Itoa(config.GrpcServerPort)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	log.Printf("grpc Server is listening on %v ...", address)

	s := grpc.NewServer()
	reflection.Register(s)
	proxy.RegisterLightningServer(s, &RpcServer{})
	proxy.RegisterWalletServer(s, &RpcServer{})
	proxy.RegisterRsmcServer(s, &RpcServer{})
	proxy.RegisterHtlcServer(s, &RpcServer{})
	s.Serve(lis)
}
