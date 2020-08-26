package lightclient

import (
	"encoding/json"
	"fmt"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/service"
	"log"
)

type clientManager struct {
	Broadcast       chan []byte
	Connected       chan *Client
	Disconnected    chan *Client
	ClientsMap      map[*Client]bool
	OnlineClientMap map[string]*Client
}

var globalWsClientManager = clientManager{
	Broadcast:       make(chan []byte),
	Connected:       make(chan *Client),
	Disconnected:    make(chan *Client),
	ClientsMap:      make(map[*Client]bool),
	OnlineClientMap: make(map[string]*Client),
}

func (clientManager *clientManager) Start() {
	for {
		select {
		case client := <-clientManager.Connected:
			clientManager.ClientsMap[client] = true
			jsonMessage, _ := json.Marshal(&bean.RequestMessage{
				SenderUserPeerId:    client.Id,
				SenderNodePeerId:    P2PLocalPeerId,
				RecipientNodePeerId: P2PLocalPeerId,
				RecipientUserPeerId: client.Id,
				Data:                "welcome to you."})
			log.Println(fmt.Sprintf("a new socket %s has connected.", client.Id))
			clientManager.sendToSomeConn(jsonMessage, client)

		case client := <-clientManager.Disconnected:
			if _, ok := clientManager.ClientsMap[client]; ok {
				log.Println(fmt.Sprintf("socket %s has disconnected.", client.Id))
				clientManager.cleanConn(client)
			}

		case msg := <-clientManager.Broadcast:
			for client := range clientManager.ClientsMap {
				select {
				case client.SendChannel <- msg:
				default:
					clientManager.cleanConn(client)
				}
			}
		}
	}
}

func (clientManager *clientManager) cleanConn(client *Client) {
	delete(clientManager.ClientsMap, client)
	if client.User != nil {
		delete(clientManager.OnlineClientMap, client.User.PeerId)
		delete(service.OnlineUserMap, client.User.PeerId)
		client.User = nil
	}
	close(client.SendChannel)
}

func (clientManager *clientManager) sendToSomeConn(message []byte, myself *Client) {
	for conn := range clientManager.ClientsMap {
		if conn == myself {
			conn.SendChannel <- message
		}
	}
}
