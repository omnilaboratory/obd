package grpcpack

import (
	pb "LightningOnOmni/grpcpack/pb"
	"context"
	"google.golang.org/grpc"
	"log"
	"testing"
)

//https://cloud.tencent.com/developer/article/1417947
// how to get data from grpc  remember not rpc
func TestBtcRpcManager_GetNewAddress(t *testing.T) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	client := pb.NewBtcServiceClient(conn)
	reply, err := client.GetNewAddress(context.Background(), &pb.AddressRequest{Label: "Finish App"})

	if err != nil {
		log.Fatal("Problem editing ToDo: ", err)
	}
	log.Println(reply)
}
