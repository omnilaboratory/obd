package main

import (
	"context"
	"encoding/json"
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

func TestListChannels(t *testing.T) {
	client, conn := getClient()
	defer conn.Close()
	resp, err := client.ListChannels(context.Background(), &proxy.ListChannelsRequest{
		ActiveOnly: true,
		PageIndex:  1,
		PageSize:   10,
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(resp)
	log.Println(string(marshal))
}

func TestLatestTransaction(t *testing.T) {
	client, conn := getClient()
	defer conn.Close()
	resp, err := client.LatestTransaction(context.Background(), &proxy.LatestTransactionRequest{
		ChannelId: "1973a103c683f6d8d5b015e3ce4927bf38b7295ccb4f27487b3fc478f2a118cc",
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(resp)
	log.Println(string(marshal))
}

func TestGetTransactionsByChannelId(t *testing.T) {

	client, conn := getClient()
	defer conn.Close()

	resp, err := client.GetTransactions(context.Background(), &proxy.GetTransactionsRequest{
		ChannelId: "ce7d6a2a15093b80bc9ce12d53c49070ff00bc9e522bf3b40d384cf0c5f5fcc3",
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
