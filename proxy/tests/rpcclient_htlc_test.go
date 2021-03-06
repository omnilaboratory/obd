package main

import (
	"context"
	"encoding/json"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func TestAddInvoice(t *testing.T) {

	client, conn := getHtlcClient()
	defer conn.Close()

	invoice, err := client.AddInvoice(context.Background(), &proxy.Invoice{
		CltvExpiry: "2021-08-15",
		Value:      0.001,
		PropertyId: 137,
		Private:    false,
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(invoice.PaymentRequest)
}

func TestListInvoices(t *testing.T) {

	client, conn := getHtlcClient()
	defer conn.Close()

	resp, err := client.ListInvoices(context.Background(), &proxy.ListInvoiceRequest{
		Reversed:       false,
		IndexOffset:    0,
		NumMaxInvoices: 10,
	})

	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(resp)
	log.Println(string(marshal))
}

func TestParseInvoice(t *testing.T) {
	client, conn := getHtlcClient()
	defer conn.Close()
	invoice, err := client.ParseInvoice(context.Background(), &proxy.ParseInvoiceRequest{
		PaymentRequest: "obtb100000s1pqzyfnpwQmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3Luzq63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36hzz03a0f341f5df657c288eae6c8f6256871b029489928302541267d4ec2e779983b2xq8ps306yqtqp0dqtdescription3gc",
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(invoice)
	log.Println(string(marshal))
}

func TestSendPayment(t *testing.T) {

	client, conn := getHtlcClient()
	defer conn.Close()

	htlcPayment, err := client.SendPayment(context.Background(), &proxy.SendRequest{
		PaymentRequest: "obtb100000s1pqzyfnpwQmZPzUh7Q6PQg6gXB4XheaoZMMhHA9JNeCrJsp3FWjFrAFuzqa5f24dc5d5414d961bba98c98624b87222da3984b324bcab7cfd7fd63aee33b3hzz023fef30a568d1c3c2f8af1d825785418c934a1a2250969dd0c6c745841d353aeexq8ps306yqtqp0dqtdescription3uf",
		InvoiceDetail: &proxy.ParseInvoiceResponse{
			PropertyId:          137,
			Value:               0.001,
			Memo:                "description",
			CltvExpiry:          "2021-08-15",
			H:                   "03cadca1eaca9fa258e07296cc947a5551c75bb3fd0d5769d603fe7733ded72992",
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36",
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
	marshal, _ := json.Marshal(htlcPayment)
	log.Println(string(marshal))
}

func getHtlcClient() (proxy.HtlcClient, *grpc.ClientConn) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return proxy.NewHtlcClient(conn), conn
}
