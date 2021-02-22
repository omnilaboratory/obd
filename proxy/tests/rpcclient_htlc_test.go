package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func TestAddInvoice(t *testing.T) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	ctxb := context.Background()
	client := proxy.NewLightningClient(conn)

	invoice, err := client.AddInvoice(ctxb, &proxy.Invoice{
		CltvExpiry: "2021-08-15",
		Value:      0.001,
		PropertyId: 137,
		Private:    false,
	})
	log.Println(invoice.PaymentRequest)
}

func TestSendPayment(t *testing.T) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	ctxb := context.Background()
	client := proxy.NewLightningClient(conn)

	htlcPayment, err := client.SendPayment(ctxb, &proxy.SendPaymentRequest{
		PaymentRequest: "obtb100000s1pqzyfnpwQmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3Luzq63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36hzz03c80af7d0c2707b8bc9902d9e74d5cd16310735dea792ceda7eb050ad80b51b26xq8ps306yqtqp0dqtdescription355",
	})
	if err != nil {
		log.Println(err)
	}
	log.Println(htlcPayment)

	select {}
}
