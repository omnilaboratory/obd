package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
)

type Client struct {
	Id          string
	User        *bean.User
	Socket      *websocket.Conn
	SendChannel chan []byte
}

func (c *Client) Write() {
	defer func() {
		c.Socket.Close()
		fmt.Println("socket closed after writing...")
	}()

	for {
		select {
		case _order, ok := <-c.SendChannel:
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
		service.UserService.UserLogout(c.User)
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

		var msg bean.Message
		log.Println(string(dataReq))
		err = json.Unmarshal(dataReq, &msg)
		if err != nil {
			log.Println(err)
			continue
		}
		var sender, receiver bean.User
		json.Unmarshal([]byte(msg.Sender), &sender)
		json.Unmarshal([]byte(msg.Recipient), &receiver)

		var sendType = enum.SendTargetType_SendToNone
		switch msg.Type {
		case enum.MsgType_UserLogin:
			if c.User != nil {
				c.sendToMyself("already login")
				sendType = enum.SendTargetType_SendToSomeone
			} else {
				var data bean.User
				json.Unmarshal([]byte(msg.Data), &data)
				if len(data.Email) > 0 {
					c.User = &data
					service.UserService.UserLogin(&data)
				}
				sendType = enum.SendTargetType_SendToExceptMe
			}
		case enum.MsgType_UserLogout:
			if c.User != nil {
				c.sendToMyself("logout success")
				c.User = nil
			} else {
				c.sendToMyself("please login")
			}
			sendType = enum.SendTargetType_SendToSomeone
		//get openChannelReq from funder then send to fundee
		case enum.MsgType_ChannelOpen:
			var data bean.OpenChannelInfo
			json.Unmarshal([]byte(msg.Data), &data)
			if err := service.ChannelService.OpenChannel(&data); err != nil {
				fmt.Println(err)
			} else {
				bytes, err := json.Marshal(data)
				if err == nil {
					c.sendToSomeone(receiver, string(bytes))
				}
				sendType = enum.SendTargetType_SendToSomeone
			}
		//get acceptChannelReq from fundee then send to funder
		case enum.MsgType_ChannelAccept:
			var data bean.AcceptChannelInfo
			json.Unmarshal([]byte(msg.Data), &data)
			bytes, err := json.Marshal(data)
			if err == nil {
				c.sendToSomeone(receiver, string(bytes))
			}
			sendType = enum.SendTargetType_SendToSomeone
		// create a funding tx
		case enum.MsgType_FundingCreated:
			node, err := service.FundingService.CreateFunding(msg.Data)
			if err != nil {
				log.Println(err)
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					log.Println(err)
				} else {
					c.sendToMyself(string(bytes))
					sendType = enum.SendTargetType_SendToSomeone
				}
			}
		case enum.MsgType_GetFundingCreated:
			id, err := strconv.Atoi(msg.Data)
			if err != nil {
				log.Println(err)
				break
			}
			node, err := service.FundingService.GetFundingTx(id)
			data := ""
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
				}
			}
			c.sendToMyself(data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_DelTableFundingCreated:
			err := service.FundingService.DeleteTable()
			data := ""
			if err != nil {
				data = err.Error()
			} else {
				data = "del success"
			}
			c.sendToMyself(data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_DelItemFundingCreated:
			id, err := strconv.Atoi(msg.Data)
			data := ""
			for {
				if err != nil {
					data = err.Error()
					break
				}
				err = service.FundingService.DeleteItem(id)
				if err != nil {
					data = err.Error()
				} else {
					data = "del success"
				}
				break
			}
			c.sendToMyself(data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_CountFundingCreated:
			data := ""
			count, err := service.FundingService.TotalCount()
			if err != nil {
				data = err.Error()
			} else {
				data = strconv.Itoa(count)
			}
			c.sendToMyself(data)
			sendType = enum.SendTargetType_SendToSomeone

		case enum.MsgType_FundingSigned:
			sendType = enum.SendTargetType_SendToAll
		case enum.MsgType_CommitmentTx:
			sendType = enum.SendTargetType_SendToAll
		case enum.MsgType_CommitmentTxSigned:
			sendType = enum.SendTargetType_SendToAll
		case enum.MsgType_GetBalanceRequest:
			sendType = enum.SendTargetType_SendToAll
		case enum.MsgType_GetBalanceRespond:
			sendType = enum.SendTargetType_SendToAll
		default:
			sendType = enum.SendTargetType_SendToAll
		}

		//broadcast except me
		if sendType == enum.SendTargetType_SendToExceptMe {
			for client := range GlobalWsClientManager.Clients_map {
				if client != c {
					jsonMessage, _ := json.Marshal(&bean.Message{Sender: client.Id, Data: string(dataReq)})
					client.SendChannel <- jsonMessage
				}
			}
			//broadcast to all
		} else if sendType == enum.SendTargetType_SendToAll {
			jsonMessage, _ := json.Marshal(&bean.Message{Sender: c.Id, Data: string(dataReq)})
			GlobalWsClientManager.Broadcast <- jsonMessage
		}
	}
}

func (c *Client) sendToSomeone(recipient bean.User, data string) {
	for client := range GlobalWsClientManager.Clients_map {
		if client.User.Email == recipient.Email {
			jsonMessage, _ := json.Marshal(&bean.Message{Sender: c.Id, Data: data})
			client.SendChannel <- jsonMessage
			break
		}
	}
}

func (client *Client) sendToMyself(data string) {
	jsonMessage, _ := json.Marshal(&bean.Message{Sender: client.Id, Data: data})
	client.SendChannel <- jsonMessage
}
