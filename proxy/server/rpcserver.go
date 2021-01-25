package main

import (
	context "context"
	fmt "fmt"
	"log"
	"net"
	"os"

	proxy "github.com/omnilaboratory/obd/proxy/pb"
	client "github.com/omnilaboratory/obd/proxy/client"
	"github.com/urfave/cli"
	grpc "google.golang.org/grpc"
)

type rpcServer struct {}

// for testing
func (r *rpcServer) Hello(ctx context.Context,
	in *proxy.HelloRequest) (*proxy.HelloResponse, error) {

	resp, err := client.Hello(in.Sayhi)
	if err != nil {
		return nil, err
	}

	return &proxy.HelloResponse{Resp: resp}, nil
}


func startGRPCServer() (string, error) {

	address := "localhost:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	fmt.Printf("Server is listening on %v ...", address)

	s := grpc.NewServer()
	proxy.RegisterProxyServer(s, &rpcServer{})

	s.Serve(lis)

	returnMsg := "Starting gRPC server done."
	return returnMsg, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "obdcli"
	app.Version = "0.0.1-beta"
	app.Usage = "Control plane for your Omni Bolt Daemon (obd)"
	app.Commands = []cli.Command{
		client.HelloCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	
	startGRPCServer()
}