package grpcpack

import (
	pb "LightningOnOmni/grpcpack/pb"
	"LightningOnOmni/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strconv"
)

type btcRpcManager struct{}

func (s *btcRpcManager) GetNewAddress(ctx context.Context, in *pb.AddressRequest) (reply *pb.AddressReply, err error) {
	client := rpc.NewClient()
	result, err := client.GetNewAddress(in.GetLabel())
	if err != nil {
		log.Println(err)
	}
	return &pb.AddressReply{Address: result}, nil
}
func (s *btcRpcManager) GetBlockCount(ctx context.Context, in *pb.EmptyRequest) (reply *pb.BlockCountReply, err error) {
	client := rpc.NewClient()
	result, err := client.GetBlockCount()
	if err != nil {
		log.Println(err)
	}
	count, err := strconv.Atoi(result)
	return &pb.BlockCountReply{Count: int32(count)}, nil
}

func Server() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Println("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBtcServiceServer(s, &btcRpcManager{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Println("failed to serve: %v", err)
	}
}
