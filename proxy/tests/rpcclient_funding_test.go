package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func TestOpenChannel(t *testing.T) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	ctxb := context.Background()
	client := proxy.NewLightningClient(conn)

	channelResponse, err := client.OpenChannel(ctxb, &proxy.OpenChannelRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
		NodePubkeyString: "03c384b8d9c65edea28ce205537bb58dc0096bc618e9e553455e1db1f36cc25642",
		NodePubkeyIndex:  1,
		Private:          false,
	})
	if err != nil {
		log.Println(err)
	}
	log.Println(channelResponse.TemplateChannelId)

	select {}
}

func TestFundChannel(t *testing.T) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	ctxb := context.Background()
	client := proxy.NewLightningClient(conn)

	fundChannel, err := client.FundChannel(ctxb, &proxy.FundChannelRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
		TemplateChannelId: "",
		BtcAmount:         0.0004,
		PropertyId:        137,
		AssetAmount:       1,
	})
	log.Println(fundChannel.ChannelId)
	select {}
}
