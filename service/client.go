package service

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

type SendTargetType int

const (
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
			fmt.Printf("send data to client %v \n", string(_order))
			c.Socket.WriteMessage(websocket.TextMessage, _order)
		}
	}
}

func (c *Client) Read() {
	defer func() {
		Global_manager.Unregister <- c
		c.Socket.Close()
		fmt.Println("socket closed after reading...")
	}()

	for {
		_, dataReq, err := c.Socket.ReadMessage()
		if err != nil {
			User_service.UserLogout(c.User)
			Global_manager.Unregister <- c
			c.Socket.Close()
			break
		}

		var msg Message
		json.Unmarshal(dataReq, &msg)
		var sender, recipient User
		json.Unmarshal([]byte(msg.Sender), &sender)
		json.Unmarshal([]byte(msg.Recipient), &recipient)

		var sendBroadcast SendTargetType = SendToAll
		switch msg.Type {
		//user login
		case 1:
			if c.User != nil {
				c.sendToMyself("already login")
				sendBroadcast = SendToSomeone
			} else {
				var data User
				json.Unmarshal([]byte(msg.Data), &data)
				if len(data.Email) > 0 {
					c.User = &data
					User_service.UserLogin(&data)
				}
				sendBroadcast = SendToExceptMe
			}
		//user logout
		case 2:
			if c.User != nil {
				c.sendToMyself("logout success")
				c.User = nil
			} else {
				c.sendToMyself("please login")
			}
			sendBroadcast = SendToSomeone
		//get openChannelReq from funder then send to fundee
		case -32:
			var data OpenChannelInfo
			json.Unmarshal([]byte(msg.Data), &data)
			if err := Channel_Service.OpenChannel(&data); err != nil {
				fmt.Println(err)
			} else {
				bytes, err := json.Marshal(data)
				if err == nil {
					c.sendToSomeone(recipient, string(bytes))
				}
				sendBroadcast = SendToSomeone
			}
		//get acceptChannelReq from fundee then send to funder
		case -33:
			var data AcceptChannelInfo
			json.Unmarshal([]byte(msg.Data), &data)
			bytes, err := json.Marshal(data)
			if err == nil {
				c.sendToSomeone(recipient, string(bytes))
			}
			sendBroadcast = SendToSomeone
		}

		//broadcast except me
		if sendBroadcast == SendToExceptMe {
			for client := range Global_manager.Clients_map {
				if client != c {
					jsonMessage, _ := json.Marshal(&MessageBody{Sender: client.Id, Data: string(dataReq)})
					client.Send_channel <- jsonMessage
				}
			}
			//broadcast to all
		} else if sendBroadcast == SendToAll {
			jsonMessage, _ := json.Marshal(&MessageBody{Sender: c.Id, Data: string(dataReq)})
			Global_manager.Broadcast <- jsonMessage
		}
	}
}

func (c *Client) sendToSomeone(recipient User, data string) {
	for client := range Global_manager.Clients_map {
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
