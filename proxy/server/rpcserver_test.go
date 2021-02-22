package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func TestClient(t *testing.T) {

	//go startServer()

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	ctxb := context.Background()
	proxyClient := proxy.NewProxyClient(conn)

	login, err := proxyClient.Login(ctxb, &proxy.LoginRequest{
		Mnemonic:   "dawn enter attitude merry cliff stone rely convince team warfare wasp whisper",
		LoginToken: "mjgwhdzx",
	})
	log.Println(login)

	//token, err := proxyClient.ChangePassword(ctxb, &proxy.ChangePasswordRequest{
	//	CurrentPassword: "mvgcnx",
	//	NewPassword:     "mvgcnx",
	//})
	//log.Println(token)

	if false {
		channelResponse, err := proxyClient.OpenChannel(ctxb, &proxy.OpenChannelRequest{
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

		fundChannel, err := proxyClient.FundChannel(ctxb, &proxy.FundChannelRequest{
			RecipientInfo: &proxy.RecipientNodeInfo{
				RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
				RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
			TemplateChannelId: channelResponse.TemplateChannelId,
			BtcAmount:         0.0004,
			PropertyId:        137,
			AssetAmount:       1,
		})
		log.Println(fundChannel.ChannelId)

		payment, err := proxyClient.RsmcPayment(ctxb, &proxy.RsmcPaymentRequest{
			RecipientInfo: &proxy.RecipientNodeInfo{
				RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
				RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
			//ChannelId: "180b07fb0b8855740ee51a4db142da2b1695730a3346f20beae12e3efe9f37d8",
			ChannelId: fundChannel.ChannelId,
			Amount:    0.001,
		})
		if err != nil {
			log.Println(err)
		}
		log.Println(payment)
	}

	//payment, err := proxyClient.RsmcPayment(ctxb, &proxy.RsmcPaymentRequest{
	//	RecipientInfo: &proxy.RecipientNodeInfo{
	//		RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
	//		RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
	//	ChannelId: "a89f2b4b5f007fbea0c58af5cd9a9cc1d9c50689c66271bfbaaefe61c973e15e",
	//	Amount:    0.001,
	//})
	//if err != nil {
	//	log.Println(err)
	//}
	//log.Println(payment)

	//invoice, err := proxyClient.AddInvoice(ctxb, &proxy.Invoice{
	//	CltvExpiry: "2021-08-15",
	//	Value:      0.001,
	//	PropertyId: 137,
	//	Private:    false,
	//})
	//log.Println(invoice.PaymentRequest)

	htlcPayment, err := proxyClient.SendPayment(ctxb, &proxy.SendPaymentRequest{
		PaymentRequest: "obtb100000s1pqzyfnpwQmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3Luzq63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36hzz03c80af7d0c2707b8bc9902d9e74d5cd16310735dea792ceda7eb050ad80b51b26xq8ps306yqtqp0dqtdescription355",
	})
	if err != nil {
		log.Println(err)
	}
	log.Println(htlcPayment)

	//logout, err := proxyClient.Logout(ctxb, &proxy.LogoutRequest{})
	//if err != nil {
	//	log.Println(err)
	//}
	//log.Println(logout)
	select {}
}
