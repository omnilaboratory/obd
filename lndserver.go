package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"lnd-server/modules"
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
	client := &modules.Client{Id: uuid_str.String(),
		Socket:       conn,
		Send_channel: make(chan []byte)}

	modules.Global_manager.Register <- client
	go client.Write()
	client.Read()
}

func main() {
	modules.Global_params.Interval = 1000
	modules.Global_params.MaximumClients = 1024 * 1024
	modules.Global_params.PoolSize = 4 * 1024

	fmt.Println("Starting application...")
	go modules.Global_manager.Start()

	fmt.Println("Global client manager started...")
	http.HandleFunc("/ws", wsPage)

	log.Fatal(http.ListenAndServe(":60020", nil))

}
