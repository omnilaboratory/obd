package lightclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"log"
)

type clientManager struct {
	Broadcast       chan []byte
	Connected       chan *Client
	Disconnected    chan *Client
	ClientsMap      map[*Client]bool
	OnlineClientMap map[string]*Client
}

var GlobalWsClientManager = clientManager{
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
				SenderNodePeerId:    P2PLocalNodeId,
				RecipientNodePeerId: P2PLocalNodeId,
				RecipientUserPeerId: client.Id,
				Data:                "welcome to you."})
			log.Println(fmt.Sprintf("a new socket %s has connected.", client.Id))
			client.SendChannel <- jsonMessage

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
	if client.User != nil && client.User.IsAdmin {
		return
	}
	delete(clientManager.ClientsMap, client)
	if client.User != nil {
		_ = service.UserService.UserLogout(client.User)
		sendInfoOnUserStateChange(client.User.PeerId)
		delete(clientManager.OnlineClientMap, client.User.PeerId)
		delete(service.OnlineUserMap, client.User.PeerId)
		client.User = nil
	}
	close(client.SendChannel)
}

func FindUserOnLine(msg bean.RequestMessage) (*Client, error) {
	if tool.CheckIsString(&msg.RecipientUserPeerId) {
		itemClient := GlobalWsClientManager.OnlineClientMap[msg.RecipientUserPeerId]
		if itemClient != nil && itemClient.User != nil {
			return itemClient, nil
		}

		if msg.RecipientNodePeerId != P2PLocalNodeId {
			if conn2tracker.GetUserState(msg.RecipientNodePeerId, msg.RecipientUserPeerId) > 0 {
				return nil, nil
			}
		}
	}
	return nil, errors.New(fmt.Sprintf(enum.Tips_user_notExistOrOnline, msg.RecipientUserPeerId))
}
