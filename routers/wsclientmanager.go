package routers

import (
	"LightningOnOmni/bean"
	"encoding/json"
	"log"
)

type ClientManager struct {
	Clients_map map[*Client]bool
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
}

var GlobalWsClientManager = ClientManager{
	Broadcast:   make(chan []byte),
	Register:    make(chan *Client),
	Unregister:  make(chan *Client),
	Clients_map: make(map[*Client]bool),
}

func (client_manager *ClientManager) Start() {
	for {
		select {
		case conn := <-client_manager.Register:
			client_manager.Clients_map[conn] = true
			jsonMessage, _ := json.Marshal(&bean.Message{Data: "/A new socket has connected."})
			log.Println("new socket has connected.")
			client_manager.Send(jsonMessage, conn)
		case conn := <-client_manager.Unregister:
			if _, ok := client_manager.Clients_map[conn]; ok {
				close(conn.SendChannel)
				delete(client_manager.Clients_map, conn)
				jsonMessage, _ := json.Marshal(&bean.Message{Data: "/A socket has disconnected."})
				log.Println("socket has disconnected.")
				client_manager.Send(jsonMessage, conn)
			}
		case order_message := <-client_manager.Broadcast:
			for conn := range client_manager.Clients_map {
				select {
				case conn.SendChannel <- order_message:
				default:
					close(conn.SendChannel)
					delete(client_manager.Clients_map, conn)
				}
			}
		}
	}
}

func (client_manager *ClientManager) Send(message []byte, myself *Client) {
	for conn := range client_manager.Clients_map {
		if conn != myself {
			conn.SendChannel <- message
		}
	}
}
