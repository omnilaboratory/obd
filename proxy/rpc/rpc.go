package rpc

import (
	"context"
	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"log"
)

type RpcServer struct{}

// for testing
func (r *RpcServer) Hello(ctx context.Context,
	in *proxy.HelloRequest) (*proxy.HelloResponse, error) {

	log.Println("hello " + in.GetSayhi())
	resp, err := Hello(in.Sayhi)
	if err != nil {
		return nil, err
	}

	return &proxy.HelloResponse{Resp: resp}, nil
}

// for testing
func Hello(sayhi string) (string, error) {
	returnMsg := "You sent: [" + sayhi + "]. We're testing proxy mode of OBD."
	return returnMsg, nil
}
