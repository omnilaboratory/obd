package main

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/tool"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func TestLogin(t *testing.T) {
	client, conn := getClient()
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

	client, conn := getClient()
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
	client, conn := getClient()
	defer conn.Close()

	logout, err := client.Logout(context.Background(), &proxy.LogoutRequest{})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(logout)
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
