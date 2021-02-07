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
	response, err := proxyClient.Hello(ctxb, &proxy.HelloRequest{Sayhi: "Test  obd grpc 你好"})
	if err != nil {
		log.Println(err)
	}
	log.Println(response)

	userClient := proxy.NewUserClient(conn)
	login, err := userClient.Login(ctxb, &proxy.LoginRequest{Mnemonic: "dawn enter attitude merry cliff stone rely convince team warfare wasp whisper"})

	log.Println(login)
}
