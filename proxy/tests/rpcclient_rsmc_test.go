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
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
		ChannelId: "037136dd08e5daffad209ac214f3939d0a2fc9109202df07b2df0406b4fc4c51",
		Amount:    0.01,
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(payment)
	log.Println(string(marshal))
}
