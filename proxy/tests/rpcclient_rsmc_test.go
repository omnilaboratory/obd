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

func TestLatestRsmcTx(t *testing.T) {
	client, conn := getRsmcClient()
	defer conn.Close()

	resp, err := client.LatestRsmcTx(context.Background(), &proxy.LatestRsmcTxRequest{
		ChannelId: "c299b48ed293ff8a36d959310a7bc698492fec9caee8fbe61c2c004a9921478e",
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(resp)
	log.Println(string(marshal))
}

func TestRsmcPayment(t *testing.T) {

	client, conn := getRsmcClient()
	defer conn.Close()

	payment, err := client.RsmcPayment(context.Background(), &proxy.RsmcPaymentRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
		ChannelId: "c299b48ed293ff8a36d959310a7bc698492fec9caee8fbe61c2c004a9921478e",
		Amount:    0.001,
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(payment)
	log.Println(string(marshal))
}

func TestTxListByChannelId(t *testing.T) {

	client, conn := getRsmcClient()
	defer conn.Close()

	resp, err := client.TxListByChannelId(context.Background(), &proxy.TxListRequest{
		ChannelId: "c299b48ed293ff8a36d959310a7bc698492fec9caee8fbe61c2c004a9921478e",
		PageSize:  10,
		PageIndex: 1,
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(resp)
	log.Println(string(marshal))
}
