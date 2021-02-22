package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func TestRsmcPayment(t *testing.T) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	ctxb := context.Background()
	client := proxy.NewLightningClient(conn)

	payment, err := client.RsmcPayment(ctxb, &proxy.RsmcPaymentRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
		ChannelId: "180b07fb0b8855740ee51a4db142da2b1695730a3346f20beae12e3efe9f37d8",
		Amount:    0.001,
	})
	if err != nil {
		log.Println(err)
	}
	log.Println(payment)

	select {}
}
