package main

import (
	context "context"
	fmt "fmt"
	"log"
	"net"

	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/proxy/rpc"
	grpc "google.golang.org/grpc"
)

type rpcServer struct{}

// for testing
func (r *rpcServer) Hello(ctx context.Context,
	in *proxy.HelloRequest) (*proxy.HelloResponse, error) {

	log.Println("hello " + in.GetSayhi())
	resp, err := rpc.Hello(in.Sayhi)
	if err != nil {
		return nil, err
	}

	return &proxy.HelloResponse{Resp: resp}, nil
}

func startServer() {

	address := "localhost:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	fmt.Printf("Server is listening on %v ...", address)

	s := grpc.NewServer()
	proxy.RegisterProxyServer(s, &rpcServer{})
	proxy.RegisterUserServer(s, &rpc.UserRpc{})

	s.Serve(lis)
}

func main() {
	startServer()
}
