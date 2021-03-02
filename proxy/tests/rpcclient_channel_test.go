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
