package main

import (
	"context"
	"encoding/json"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func getRsmcClient() (proxy.RsmcClient, *grpc.ClientConn) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return proxy.NewRsmcClient(conn), conn
}

func TestRsmcPayment(t *testing.T) {

	client, conn := getRsmcClient()
	defer conn.Close()

	payment, err := client.RsmcPayment(context.Background(), &proxy.RsmcPaymentRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmZPzUh7Q6PQg6gXB4XheaoZMMhHA9JNeCrJsp3FWjFrAF",
			RecipientUserPeerId: "a5f24dc5d5414d961bba98c98624b87222da3984b324bcab7cfd7fd63aee33b3"},
		ChannelId: "04281cba5a1179cf47501ae3cc687bd58aeeae848fc4daa3690351292ad2eb7a",
		Amount:    0.01,
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(payment)
	log.Println(string(marshal))
}
