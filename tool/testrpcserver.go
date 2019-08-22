package main

import (
	"LightningOnOmni/tool/beans"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

type UserService struct {
}

func (UserService) GetUser(userId int, result *beans.User) error {
	result.UserId = userId
	result.Name = "Jack"
	log.Println(result)
	return nil
}

func main() {
	rpc.Register(UserService{})
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(conn)
		go jsonrpc.ServeConn(conn)
	}

}
