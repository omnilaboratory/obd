package main

import (
	"LightningOnOmni/service"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
)

func wsPage(res http.ResponseWriter, req *http.Request) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if error != nil {
		http.NotFound(res, req)
		return
	}
	uuid_str, _ := uuid.NewV4()
	client := &service.Client{Id: uuid_str.String(),
		Socket:       conn,
		Send_channel: make(chan []byte)}

	service.Global_manager.Register <- client
	go client.Write()
	client.Read()
}

func main() {
	service.Global_params.Interval = 1000
	service.Global_params.MaximumClients = 1024 * 1024
	service.Global_params.PoolSize = 4 * 1024

	fmt.Println("Starting application...")
	go service.Global_manager.Start()

	fmt.Println("Global client manager started...")
	http.HandleFunc("/ws", wsPage)

	log.Fatal(http.ListenAndServe(":60020", nil))

}
