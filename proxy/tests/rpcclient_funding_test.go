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
			RecipientNodePeerId: "QmZPzUh7Q6PQg6gXB4XheaoZMMhHA9JNeCrJsp3FWjFrAF",
			RecipientUserPeerId: "a5f24dc5d5414d961bba98c98624b87222da3984b324bcab7cfd7fd63aee33b3"},
		TemplateChannelId: "ac98aef33e317a0063036c33ffd9b3703fd3e630ea22ffc2ba0b05e7da475ab8",
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
