package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"log"
	"testing"
)

func TestOpenChannel(t *testing.T) {

	client, conn := getClient()
	defer conn.Close()

	channelResponse, err := client.OpenChannel(context.Background(), &proxy.OpenChannelRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmZPzUh7Q6PQg6gXB4XheaoZMMhHA9JNeCrJsp3FWjFrAF",
			RecipientUserPeerId: "a5f24dc5d5414d961bba98c98624b87222da3984b324bcab7cfd7fd63aee33b3"},
		NodePubkeyString: "023769b549838e48db217c4d2a8bbeb199c5dbf63dfa38649b6bc2bb18261d7454",
		NodePubkeyIndex:  1,
		Private:          false,
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(channelResponse.TemplateChannelId)
}

func TestFundChannel(t *testing.T) {

	client, conn := getClient()
	defer conn.Close()

	fundChannel, err := client.FundChannel(context.Background(), &proxy.FundChannelRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmZPzUh7Q6PQg6gXB4XheaoZMMhHA9JNeCrJsp3FWjFrAF",
			RecipientUserPeerId: "a5f24dc5d5414d961bba98c98624b87222da3984b324bcab7cfd7fd63aee33b3"},
		TemplateChannelId: "2c665f11dfa1d5daba57ad98ec2036d4b92cd7ae503b698ead5504d2b5ab1850",
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
