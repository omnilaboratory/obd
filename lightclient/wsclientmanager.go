package lightclient

import (
	"LightningOnOmni/bean"
	"encoding/json"
	"log"
)

type ClientManager struct {
	Clients_map  map[*Client]bool
	Broadcast    chan []byte
	Connected    chan *Client
	Disconnected chan *Client
}

var GlobalWsClientManager = ClientManager{
	Broadcast:    make(chan []byte),
	Connected:    make(chan *Client),
	Disconnected: make(chan *Client),
	Clients_map:  make(map[*Client]bool),
}

func (clientManager *ClientManager) Start() {
	for {
		select {
		case conn := <-clientManager.Connected:
			clientManager.Clients_map[conn] = true
			jsonMessage, _ := json.Marshal(&bean.RequestMessage{Data: "A new socket has connected."})
			log.Println("new socket has connected.")
			clientManager.Send(jsonMessage, conn)
		case conn := <-clientManager.Disconnected:
			if _, ok := clientManager.Clients_map[conn]; ok {
				close(conn.SendChannel)
				delete(clientManager.Clients_map, conn)
				jsonMessage, _ := json.Marshal(&bean.RequestMessage{Data: "A socket has disconnected."})
				log.Println("socket has disconnected.")
				clientManager.Send(jsonMessage, conn)
			}
		case order_message := <-clientManager.Broadcast:
			for conn := range clientManager.Clients_map {
				select {
				case conn.SendChannel <- order_message:
				default:
					close(conn.SendChannel)
					delete(clientManager.Clients_map, conn)
				}
			}
		}
	}
}

func (clientManager *ClientManager) Send(message []byte, myself *Client) {
	for conn := range clientManager.Clients_map {
		if conn != myself {
			conn.SendChannel <- message
		}
	}
}
