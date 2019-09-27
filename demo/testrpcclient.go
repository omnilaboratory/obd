package main

import (
	"LightningOnOmni/demo/beans"
	"log"
	"net"
	"net/rpc/jsonrpc"
	"os/exec"
)

func main() {

	cmd := exec.Command("msconfig")
	run := cmd.Run()
	log.Println(run)

	return

	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	client := jsonrpc.NewClient(conn)
	var user = beans.User{}
	//curl -i -H "Content-Type:application/json" -d "{\"jsonrpc\":\"1.0\", \"method\":\"GetUser\", \"params\":[1], \"id\":\"200\"}" http://127.0.0.1:8080
	err = client.Call("UserService.GetUser", 1, &user)
	if err != nil {
		panic(err)
	}
	log.Println(user)
}
