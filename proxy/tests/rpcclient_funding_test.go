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
		TemplateChannelId: "8abc7d1042de7d9e43b8b64acb2e65de4cacd05f9ea5955ef36a4db3f3547a1b",
		BtcAmount:         0.0004,
		PropertyId:        137,
		AssetAmount:       1,
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(fundChannel.ChannelId)
}
