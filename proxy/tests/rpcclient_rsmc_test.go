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

	payment, err := client.LatestRsmcTx(context.Background(), &proxy.LatestRsmcTxRequest{
		ChannelId: "0caf45d8b014dbb84b557f671bb10981af13f7b9ffd317f56abcd9d77a45bf87",
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(payment)
}

func TestRsmcPayment(t *testing.T) {

	client, conn := getRsmcClient()
	defer conn.Close()

	payment, err := client.RsmcPayment(context.Background(), &proxy.RsmcPaymentRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
		ChannelId: "3b5cd6f2cc9a011158431935259670898e1500387e3586482abf2abcf9648e3d",
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
		ChannelId: "3b5cd6f2cc9a011158431935259670898e1500387e3586482abf2abcf9648e3d",
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
