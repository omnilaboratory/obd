package service

import (
	"encoding/json"
	"fmt"
)

type ClientManager struct {
	Clients_map map[*Client]bool
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
}

func (client_manager *ClientManager) Start() {
	for {
		select {
		case conn := <-client_manager.Register:
			client_manager.Clients_map[conn] = true
			jsonMessage, _ := json.Marshal(&MessageBody{Data: "/A new socket has connected."})
			fmt.Println("new socket has connected.")

			client_manager.Send(jsonMessage, conn)
		case conn := <-client_manager.Unregister:
			if _, ok := client_manager.Clients_map[conn]; ok {
				close(conn.Send_channel)
				delete(client_manager.Clients_map, conn)
				jsonMessage, _ := json.Marshal(&MessageBody{Data: "/A socket has disconnected."})
				fmt.Println("socket has disconnected.")

				client_manager.Send(jsonMessage, conn)
			}
		case order_message := <-client_manager.Broadcast:
			for conn := range client_manager.Clients_map {
				select {
				case conn.Send_channel <- order_message:
				default:
					close(conn.Send_channel)
					delete(client_manager.Clients_map, conn)
				}
			}
		}
	}
}

func (client_manager *ClientManager) Send(message []byte, myself *Client) {
	for conn := range client_manager.Clients_map {
		if conn != myself {
			conn.Send_channel <- message
		}
	}
}
