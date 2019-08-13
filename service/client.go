package service

import (
	"LightningOnOmni/config"
	"LightningOnOmni/config/msgtype"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type SendTargetType int

const (
	SendToNone     SendTargetType = -1
	SendToAll      SendTargetType = 0
	SendToSomeone  SendTargetType = 1
	SendToExceptMe SendTargetType = 2
)

type Client struct {
	Id           string
	User         *User
	Socket       *websocket.Conn
	Send_channel chan []byte
}

func (c *Client) Write() {
	defer func() {
		c.Socket.Close()
		fmt.Println("socket closed after writing...")
	}()

	for {
		select {
		case _order, ok := <-c.Send_channel:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			fmt.Printf("send data: %v \n", string(_order))
			c.Socket.WriteMessage(websocket.TextMessage, _order)
		}
	}
}

func (c *Client) Read() {
	defer func() {
		User_service.UserLogout(c.User)
		GlobalWsClientManager.Unregister <- c
		c.Socket.Close()
		fmt.Println("socket closed after reading...")
	}()

	for {
		_, dataReq, err := c.Socket.ReadMessage()
		if err != nil {
			log.Println(err)

			break
		}

		var msg config.Message
		log.Println(string(dataReq))
		err = json.Unmarshal(dataReq, &msg)
		if err != nil {
			log.Println(err)
			continue
		}
		var sender, receiver User
		json.Unmarshal([]byte(msg.Sender), &sender)
		json.Unmarshal([]byte(msg.Recipient), &receiver)

		var sendType = SendToNone
		switch msg.Type {
		case msgtype.UserLogin:
			if c.User != nil {
				c.sendToMyself("already login")
				sendType = SendToSomeone
			} else {
				var data User
				json.Unmarshal([]byte(msg.Data), &data)
				if len(data.Email) > 0 {
					c.User = &data
					User_service.UserLogin(&data)
				}
				sendType = SendToExceptMe
			}
		case msgtype.UserLogout:
			if c.User != nil {
				c.sendToMyself("logout success")
				c.User = nil
			} else {
				c.sendToMyself("please login")
			}
			sendType = SendToSomeone
		//get openChannelReq from funder then send to fundee
		case msgtype.ChannelOpen:
			var data OpenChannelInfo
			json.Unmarshal([]byte(msg.Data), &data)
			if err := Channel_Service.OpenChannel(&data); err != nil {
				fmt.Println(err)
			} else {
				bytes, err := json.Marshal(data)
				if err == nil {
					c.sendToSomeone(receiver, string(bytes))
				}
				sendType = SendToSomeone
			}
		//get acceptChannelReq from fundee then send to funder
		case msgtype.ChannelAccept:
			var data AcceptChannelInfo
			json.Unmarshal([]byte(msg.Data), &data)
			bytes, err := json.Marshal(data)
			if err == nil {
				c.sendToSomeone(receiver, string(bytes))
			}
			sendType = SendToSomeone
		// create a funding tx
		case msgtype.FundingCreated:
			node, err := FundingService.CreateFunding(msg.Data)
			if err != nil {
				log.Println(err)
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					log.Println(err)
				} else {
					c.sendToMyself(string(bytes))
					sendType = SendToSomeone
				}
			}
		case msgtype.FundingSigned:
			sendType = SendToAll
		case msgtype.CommitmentTx:
			sendType = SendToAll
		case msgtype.CommitmentTxSigned:
			sendType = SendToAll
		case msgtype.GetBalanceRequest:
			sendType = SendToAll
		case msgtype.GetBalanceRespond:
			sendType = SendToAll
		default:
			sendType = SendToAll

		}

		//broadcast except me
		if sendType == SendToExceptMe {
			for client := range GlobalWsClientManager.Clients_map {
				if client != c {
					jsonMessage, _ := json.Marshal(&MessageBody{Sender: client.Id, Data: string(dataReq)})
					client.Send_channel <- jsonMessage
				}
			}
			//broadcast to all
		} else if sendType == SendToAll {
			jsonMessage, _ := json.Marshal(&MessageBody{Sender: c.Id, Data: string(dataReq)})
			GlobalWsClientManager.Broadcast <- jsonMessage
		}
	}
}

func (c *Client) sendToSomeone(recipient User, data string) {
	for client := range GlobalWsClientManager.Clients_map {
		if client.User.Email == recipient.Email {
			jsonMessage, _ := json.Marshal(&MessageBody{Sender: c.Id, Data: data})
			client.Send_channel <- jsonMessage
			break
		}
	}
}

func (client *Client) sendToMyself(data string) {
	jsonMessage, _ := json.Marshal(&MessageBody{Sender: client.Id, Data: data})
	client.Send_channel <- jsonMessage
}
