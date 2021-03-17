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
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
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

func TestPendingChannels(t *testing.T) {
	client, conn := getClient()
	defer conn.Close()
	resp, err := client.PendingChannels(context.Background(), &proxy.PendingChannelsRequest{
		PageIndex: 1,
		PageSize:  10,
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
		ChannelId: "343a94dd76703596b6b001a7751abfaa6afe27af196259b5e419ae17928aefdb",
		PageSize:  20,
		PageIndex: 1,
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(resp)
	log.Println(string(marshal))
}

func TestChannelBalance(t *testing.T) {

	client, conn := getClient()
	defer conn.Close()

	resp, err := client.ChannelBalance(context.Background(), &proxy.ChannelBalanceRequest{})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(resp)
	log.Println(string(marshal))
}
