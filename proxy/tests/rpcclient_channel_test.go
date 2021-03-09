package main

import (
	"context"
	"encoding/json"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"log"
	"testing"
)

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
		ChannelId: "1973a103c683f6d8d5b015e3ce4927bf38b7295ccb4f27487b3fc478f2a118cc",
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
