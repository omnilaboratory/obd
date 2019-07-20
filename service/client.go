package service

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

type Client struct {
	Id           string
	User         User
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
		_, _order, err := c.Socket.ReadMessage()
		if err != nil {
			Global_manager.Unregister <- c
			c.Socket.Close()
			break
		}

		println("receive data ", string(_order))
		/*
		   这里要先写入消息系统，而不是直接放入channel，此处只用来测试性能
		*/
		var sendBroadcast bool = true
		var msg Message
		json.Unmarshal(_order, &msg)
		var sender, recipient User
		json.Unmarshal([]byte(msg.Sender), &sender)
		json.Unmarshal([]byte(msg.Recipient), &recipient)
		switch msg.Type {
		case 1:
			var data User
			json.Unmarshal([]byte(msg.Data), &data)
			if len(data.Email) > 0 {
				c.User = data
				User_service.UserLogin(&data)
			}
			break
		case -32:
			var data OpenChannelData
			json.Unmarshal([]byte(msg.Data), &data)

			if err := Channel_Service.OpenChannel(&data); err == nil {
				sendBroadcast = false
			}

		case -33:
			var data OpenChannelData
			json.Unmarshal([]byte(msg.Data), &data)
			sendBroadcast = false
		}

		if sendBroadcast == true {
			jsonMessage, _ := json.Marshal(&MessageBody{Sender: c.Id, Data: string(_order)})
			Global_manager.Broadcast <- jsonMessage
		} else {
			c.sendToSomeone(recipient, _order)
		}
	}
}

func (c *Client) sendToSomeone(recipient User, _order []byte) {
	for client := range Global_manager.Clients_map {
		if client.User.Email == recipient.Email {
			jsonMessage, _ := json.Marshal(&MessageBody{Sender: c.Id, Data: string(_order)})
			client.Send_channel <- jsonMessage
			break
		}
	}
}
