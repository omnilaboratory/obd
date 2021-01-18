package proxy

import (
	context "context"
)

type rpcServer struct {}

// for testing
func (r *rpcServer) HelloProxy(ctx context.Context,
	in *HelloProxyRequest) (*HelloProxyResponse, error) {

	resp, err := HelloProxy(in.Sayhi)
	if err != nil {
		return nil, err
	}

	return &HelloProxyResponse{Resp: resp}, nil
}