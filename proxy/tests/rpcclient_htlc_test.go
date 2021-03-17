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
		PaymentRequest: "obtb100000s1pqzyfnpwQmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3Luzq63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36hzz03cadca1eaca9fa258e07296cc947a5551c75bb3fd0d5769d603fe7733ded72992xq8ps306yqtqp0dqtdescription3hz",
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
		PaymentRequest: "obtb100000s1pqzyfnpwQmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3Luzq63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36hzz03cadca1eaca9fa258e07296cc947a5551c75bb3fd0d5769d603fe7733ded72992xq8ps306yqtqp0dqtdescription3hz",
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
