package proxy

import (
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/obrpc"
	rpcNew "github.com/omnilaboratory/obd/obrpc/rpc"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/proxy/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strconv"
)

func StartGrpcServer() {

	log.Println("startGrpcServer")
	err := rpc.ConnToObd()
	if err != nil {
		log.Println(err)
		return
	}

	address := "0.0.0.0:" + strconv.Itoa(config.GrpcServerPort)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	log.Printf("grpc Server is listening on %v ...", address)

	s := grpc.NewServer()
	reflection.Register(s)
	proxy.RegisterLightningServer(s, &rpc.RpcServer{})
	proxy.RegisterWalletServer(s, &rpc.RpcServer{})
	proxy.RegisterRsmcServer(s, &rpc.RpcServer{})
	proxy.RegisterHtlcServer(s, &rpc.RpcServer{})


	//unlocker:=rpcNew.NewInstance(config.ChainNodeType,config.DataDirectory)
	unlock:=&rpcNew.UnlockerService{}
	obrpc.RegisterWalletUnlockerServer(s,unlock)
	s.Serve(lis)
}
