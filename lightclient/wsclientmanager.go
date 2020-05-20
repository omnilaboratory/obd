package lightclient

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"log"
)

type Client struct {
	Id          string
	User        *bean.User
	Socket      *websocket.Conn
	SendChannel chan []byte
}

type clientManager struct {
	Broadcast     chan []byte
	Connected     chan *Client
	Disconnected  chan *Client
	P2PData       chan []byte
	ClientsMap    map[*Client]bool
	OnlineUserMap map[string]*Client
}

var globalWsClientManager = clientManager{
	Broadcast:     make(chan []byte),
	Connected:     make(chan *Client),
	Disconnected:  make(chan *Client),
	P2PData:       make(chan []byte),
	ClientsMap:    make(map[*Client]bool),
	OnlineUserMap: make(map[string]*Client),
}

func (clientManager *clientManager) Start() {
	for {
		select {
		case conn := <-clientManager.Connected:
			clientManager.ClientsMap[conn] = true
			jsonMessage, _ := json.Marshal(&bean.RequestMessage{Data: "A new socket has connected."})
			log.Println("new socket has connected.")
			clientManager.Send(jsonMessage, conn)

		case conn := <-clientManager.Disconnected:
			if _, ok := clientManager.ClientsMap[conn]; ok {
				close(conn.SendChannel)
				delete(clientManager.ClientsMap, conn)
				jsonMessage, _ := json.Marshal(&bean.RequestMessage{Data: "A socket has disconnected."})
				log.Println("socket has disconnected.")
				clientManager.Send(jsonMessage, conn)
			}
		case P2PData := <-clientManager.P2PData:
			log.Println(string(P2PData))
		case orderMessage := <-clientManager.Broadcast:
			for conn := range clientManager.ClientsMap {
				select {
				case conn.SendChannel <- orderMessage:
				default:
					close(conn.SendChannel)
					delete(clientManager.ClientsMap, conn)
				}
			}
		}
	}
}

func (clientManager *clientManager) Send(message []byte, myself *Client) {
	for conn := range clientManager.ClientsMap {
		if conn == myself {
			conn.SendChannel <- message
		}
	}
}
