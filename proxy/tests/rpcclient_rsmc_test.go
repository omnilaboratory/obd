package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"log"
	"testing"
)

func TestRsmcPayment(t *testing.T) {

	client, conn := getClient()
	defer conn.Close()

	payment, err := client.RsmcPayment(context.Background(), &proxy.RsmcPaymentRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
		ChannelId: "a32a69ba51319d47a5185428639feddcdf87485b025b0fb87cf1d6e767bdb1e7",
		Amount:    0.001,
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(payment)
}
