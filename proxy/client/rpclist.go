package client

import (

)

// for testing
func Hello(sayhi string) (string, error) {
	returnMsg := "You sent: [" + sayhi + "]. We're testing proxy mode of OBD."
	return returnMsg, nil
}

// func Start() (string, error) {

// 	address := "0.0.0.0:50051"
// 	lis, err := net.Listen("tcp", address)
// 	if err != nil {
// 		log.Fatalf("Error %v", err)
// 	}
// 	fmt.Printf("Server is listening on %v ...", address)

// 	s := grpc.NewServer()
// 	proxy.RegisterProxyServer(s, &rpcServer{})

// 	s.Serve(lis)

// 	returnMsg := "Starting gRPC server done."
// 	return returnMsg, nil
// }