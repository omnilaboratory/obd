package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"google.golang.org/grpc"
	"log"
)

func main() {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	ctxb := context.Background()
	proxyClient := proxy.NewProxyClient(conn)
	//response, err := proxyClient.Hello(ctxb, &proxy.HelloRequest{Sayhi: "Test  obd grpc 你好"})
	//if err != nil {
	//	log.Println(err)
	//}
	//log.Println(response)

	login, err := proxyClient.Login(ctxb, &proxy.LoginRequest{
		Mnemonic:   "dawn enter attitude merry cliff stone rely convince team warfare wasp whisper",
		LoginToken: "mvgcnx",
	})

	log.Println(login)

	//token, err := proxyClient.ChangePassword(ctxb, &proxy.ChangePasswordRequest{
	//	CurrentPassword: "mvgcnx",
	//	NewPassword:     "mvgcnx",
	//})
	//log.Println(token)

	//channelResponse, err := proxyClient.OpenChannel(ctxb, &proxy.OpenChannelRequest{
	//	RecipientInfo: &proxy.RecipientNodeInfo{
	//		RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
	//		RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
	//	NodePubkeyString: "03c384b8d9c65edea28ce205537bb58dc0096bc618e9e553455e1db1f36cc25642",
	//	NodePubkeyIndex:  1,
	//	Private:          false,
	//})
	//if err != nil {
	//	log.Println(err)
	//}
	//log.Println(channelResponse.TemplateChannelId)
	//
	//fundChannel, err := proxyClient.FundChannel(ctxb, &proxy.FundChannelRequest{
	//	RecipientInfo: &proxy.RecipientNodeInfo{
	//		RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
	//		RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
	//	TemplateChannelId: channelResponse.TemplateChannelId,
	//	BtcAmount:         0.0004,
	//	PropertyId:        137,
	//	AssetAmount:       1,
	//})
	//log.Println(fundChannel.ChannelId)

	payment, err := proxyClient.RsmcPayment(ctxb, &proxy.RsmcPaymentRequest{
		RecipientInfo: &proxy.RecipientNodeInfo{
			RecipientNodePeerId: "QmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3L",
			RecipientUserPeerId: "63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36"},
		ChannelId: "d91836282058158348162e37b62a6dfae4eedd3739960111632a6ec8fe92162b",
		Amount:    0.001,
	})
	if err != nil {
		log.Println(err)
	}
	log.Println(payment)

	//invoice, err := proxyClient.AddInvoice(ctxb, &proxy.Invoice{
	//	CltvExpiry: "2021-08-15",
	//	Value:      0.001,
	//	PropertyId: 137,
	//	Private:    false,
	//})
	//log.Println(invoice.PaymentRequest)

	htlcPayment, err := proxyClient.SendPayment(ctxb, &proxy.SendPaymentRequest{
		PaymentRequest: "obtb100000s1pqzyfnpwQmccE4s2uhEXrJXE778NChn1ed8NyWNyAHH23mP7f9NM3Luzq63167817c979ade9e42f3204404c1513a4b1b4e9eea654c9498ed9cc920dbb36hzz02a56bedfb2aa9772fd984a0a6e83f25713a2cc8db7d9a29c95b7d9d62041306c2xq8ps306yqtqp0dqtdescription34x",
	})
	if err != nil {
		log.Println(err)
	}
	log.Println(htlcPayment)

	//logout, err := proxyClient.Logout(ctxb, &proxy.LogoutRequest{})
	//log.Println(logout)
}
