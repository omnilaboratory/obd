package main

import (
	"context"
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

func TestParseInvoice(t *testing.T) {
	client, conn := getHtlcClient()
	defer conn.Close()
	invoice, err := client.ParseInvoice(context.Background(), &proxy.ParseInvoiceRequest{
		PaymentRequest: "obtb100000s1pqzyfnpwQmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3Luzq63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36hzz0202c3aac5755ec024a102da30bcfc80658a7969a5448f145bf0e986606b459d1dxq8ps306yqtqp0dqtdescription3wh",
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(invoice)
}

func TestSendPayment(t *testing.T) {

	client, conn := getHtlcClient()
	defer conn.Close()

	htlcPayment, err := client.SendPayment(context.Background(), &proxy.SendPaymentRequest{
		PaymentRequest: "obtb100000s1pqzyfnpwQmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3Luzq63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36hzz0202c3aac5755ec024a102da30bcfc80658a7969a5448f145bf0e986606b459d1dxq8ps306yqtqp0dqtdescription3wh",
		InvoiceDetail: &proxy.ParseInvoiceResponse{
			PropertyId:          137,
			Value:               0.001,
			Memo:                "description",
			CltvExpiry:          "2021-08-15",
			H:                   "0202c3aac5755ec024a102da30bcfc80658a7969a5448f145bf0e986606b459d1d",
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36",
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(htlcPayment)
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
