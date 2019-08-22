package grpcpack

import (
	"LightningOnOmni/config"
	pb "LightningOnOmni/grpcpack/pb"
	"context"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"testing"
)

//https://cloud.tencent.com/developer/article/1417947
// how to get data from grpc  remember not rpc
func TestBtcRpcManager_GetNewAddress(t *testing.T) {
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(config.GrpcPort), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	client := pb.NewBtcServiceClient(conn)
	reply, err := client.GetNewAddress(context.Background(), &pb.AddressRequest{Label: "Finish App"})

	if err != nil {
		log.Fatal(err)
	}
	log.Println(reply)
}
