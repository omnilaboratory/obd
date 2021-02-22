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
	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	ctxb := context.Background()
	client := proxy.NewLightningClient(conn)

	login, err := client.Login(ctxb, &proxy.LoginRequest{
		Mnemonic:   "dawn enter attitude merry cliff stone rely convince team warfare wasp whisper",
		LoginToken: tool.SignMsgWithMd5([]byte("mjgwhdzx")),
	})
	log.Println(login)
}
func TestChangePassword(t *testing.T) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	ctxb := context.Background()
	client := proxy.NewLightningClient(conn)

	token, err := client.ChangePassword(ctxb, &proxy.ChangePasswordRequest{
		CurrentPassword: "mjgwhdzx",
		NewPassword:     "mjgwhdzx",
	})
	if err != nil {
		log.Println(err)
	}
	log.Println(token)
}
func TestLogout(t *testing.T) {

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	ctxb := context.Background()
	client := proxy.NewLightningClient(conn)

	logout, err := client.Logout(ctxb, &proxy.LogoutRequest{})
	if err != nil {
		log.Println(err)
	}
	log.Println(logout)
}
