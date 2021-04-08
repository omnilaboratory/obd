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
		NodePubkeyString: "03c384b8d9c65edea28ce205537bb58dc0096bc618e9e553455e1db1f36cc25642",
		Private:          false,
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(channelResponse.TemplateChannelId)
}

func TestCloseChannel(t *testing.T) {
	client, conn := getClient()
	defer conn.Close()
	resp, err := client.CloseChannel(context.Background(), &proxy.CloseChannelRequest{
		ChannelId: "e936e067a885ebf6fe94b4c75f7356aacb438ecdbc182172f55febe7d1e860fa",
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(resp)
	log.Println(string(marshal))
}

func TestGetChanInfo(t *testing.T) {
	client, conn := getClient()
	defer conn.Close()
	resp, err := client.GetChanInfo(context.Background(), &proxy.ChanInfoRequest{
		ChannelId: "e936e067a885ebf6fe94b4c75f7356aacb438ecdbc182172f55febe7d1e860fa",
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(resp)
	log.Println(string(marshal))
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
