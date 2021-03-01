package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/tool"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func TestNextAddr(t *testing.T) {
	client, conn := getWalletClient()
	defer conn.Close()
	resp, err := client.NextAddr(context.Background(), &proxy.AddrRequest{})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(resp)
}

func TestEstimateFee(t *testing.T) {
	client, conn := getWalletClient()
	defer conn.Close()
	resp, err := client.EstimateFee(context.Background(), &proxy.EstimateFeeRequest{})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(resp)
}

func TestGenSeed(t *testing.T) {
	client, conn := getWalletClient()
	defer conn.Close()
	seed, err := client.GenSeed(context.Background(), &proxy.GenSeedRequest{})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(seed)
}

func TestLogin(t *testing.T) {
	client, conn := getWalletClient()
	defer conn.Close()

	login, err := client.Login(context.Background(), &proxy.LoginRequest{
		Mnemonic:   "dawn enter attitude merry cliff stone rely convince team warfare wasp whisper",
		LoginToken: tool.SignMsgWithMd5([]byte("mjgwhdzx")),
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(login)
}
func TestChangePassword(t *testing.T) {

	client, conn := getWalletClient()
	defer conn.Close()

	token, err := client.ChangePassword(context.Background(), &proxy.ChangePasswordRequest{
		CurrentPassword: "mjgwhdzx",
		NewPassword:     "mjgwhdzx",
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(token)
}
func TestLogout(t *testing.T) {
	client, conn := getWalletClient()
	defer conn.Close()

	logout, err := client.Logout(context.Background(), &proxy.LogoutRequest{})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(logout)
}

func getWalletClient() (proxy.WalletClient, *grpc.ClientConn) {
	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return proxy.NewWalletClient(conn), conn
}

func getClient() (proxy.LightningClient, *grpc.ClientConn) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return proxy.NewLightningClient(conn), conn
}
