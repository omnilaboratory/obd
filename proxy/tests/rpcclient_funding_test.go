package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"log"
	"testing"
)

func TestFundChannel(t *testing.T) {
	client, conn := getClient()
	defer conn.Close()
	fundChannel, err := client.FundChannel(context.Background(), &proxy.FundChannelRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
		TemplateChannelId: "cec5a46b7a45673afa17eaf56a6a22f0cb73f1cc70813193ed18f271a8adb4a2",
		BtcAmount:         0.0004,
		PropertyId:        137,
		AssetAmount:       10,
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(fundChannel.ChannelId)
}
