package grpcpack

import (
	pb "LightningOnOmni/grpcpack/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

type BtcRpcManager struct{}

func Server() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Println("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBtcServiceServer(s, &BtcRpcManager{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Println("failed to serve: %v", err)
	}
}
