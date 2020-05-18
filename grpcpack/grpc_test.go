package grpcpack

import (
	"context"
	"github.com/omnilaboratory/obd/config"
	pb "github.com/omnilaboratory/obd/grpcpack/pb"
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
func TestBtcRpcManager_CreateMultiSig(t *testing.T) {
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(config.GrpcPort), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	client := pb.NewBtcServiceClient(conn)
	var keys = make([]string, 2)
	keys[0] = "n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn"
	keys[1] = "n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA"
	reply, err := client.CreateMultiSig(context.Background(), &pb.CreateMultiSigRequest{MinSignNum: 1, Keys: keys})

	if err != nil {
		log.Fatal(err)
	}
	log.Println(reply)
}
