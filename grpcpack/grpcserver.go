package grpcpack

import (
	"github.com/omnilaboratory/obd/config"
	pb "github.com/omnilaboratory/obd/grpcpack/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strconv"
)

type BtcRpcManager struct{}

func Server() {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(config.GrpcPort))
	if err != nil {
		log.Println(err)
	}
	s := grpc.NewServer()
	pb.RegisterBtcServiceServer(s, &BtcRpcManager{})

	// Connected reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Println(err)
	}
}
