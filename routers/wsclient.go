package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/rpc"
	"LightningOnOmni/service"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
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
		e := c.Socket.Close()
		if e != nil {
			log.Println(e)
		} else {
			log.Println("socket closed after writing...")
		}
	}()

	for {
		select {
		case _order, ok := <-c.SendChannel:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Printf("send data: %v \n", string(_order))
			c.Socket.WriteMessage(websocket.TextMessage, _order)
		}
	}
}

func (c *Client) Read() {
	defer func() {
		service.UserService.UserLogout(c.User)
		GlobalWsClientManager.Unregister <- c
		c.Socket.Close()
		log.Println("socket closed after reading...")
	}()

	for {
		_, dataReq, err := c.Socket.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var msg bean.RequestMessage
		log.Println(string(dataReq))
		parse := gjson.Parse(string(dataReq))

		if parse.Exists() == false {
			log.Println("wrong json input")
			continue
		}

		msg.Type = enum.MsgType(parse.Get("type").Int())
		msg.Data = parse.Get("data").String()
		msg.Sender = parse.Get("sender").String()
		msg.Recipient = parse.Get("recipient").String()

		var sender, receiver bean.User
		json.Unmarshal([]byte(msg.Sender), &sender)
		json.Unmarshal([]byte(msg.Recipient), &receiver)

		var sendType = enum.SendTargetType_SendToNone
		data := ""
		status := false
		switch msg.Type {
		case enum.MsgType_UserLogin:
			if c.User != nil {
				c.sendToMyself(msg.Type, true, "already login")
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
				c.sendToMyself(msg.Type, true, "logout success")
				c.User = nil
			} else {
				c.sendToMyself(msg.Type, true, "please login")
			}
			sendType = enum.SendTargetType_SendToSomeone

		case enum.MsgType_Core_GetNewAddress:
			var label = msg.Data
			client := rpc.NewClient()
			address, err := client.GetNewAddress(label)
			if err != nil {
				data = err.Error()
			} else {
				data = address
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_Core_GetMiningInfo:
			client := rpc.NewClient()
			result, err := client.GetMiningInfo()
			if err != nil {
				data = err.Error()
			} else {
				data = result
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_Core_GetNetworkInfo:
			client := rpc.NewClient()
			result, err := client.GetNetworkInfo()
			if err != nil {
				data = err.Error()
			} else {
				data = result
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		//get openChannelReq from funder then send to fundee
		case enum.MsgType_ChannelOpen:
			node, err := service.ChannelService.OpenChannel(msg.Data)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_ChannelOpen_ItemByTempId:
			node, err := service.ChannelService.GetChannelByTemporaryChanId(msg.Data)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_ChannelOpen_DelItemByTempId:
			node, err := service.ChannelService.DelChannelByTemporaryChanId(msg.Data)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_ChannelOpen_AllItem:
			nodes, err := service.ChannelService.AllItem()
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(nodes)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_ChannelOpen_Count:
			node, err := service.ChannelService.TotalCount()
			if err != nil {
				data = err.Error()
			} else {
				data = strconv.Itoa(node)
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		//get acceptChannelReq from fundee then send to funder
		case enum.MsgType_ChannelAccept:
			var vo bean.AcceptChannelInfo
			json.Unmarshal([]byte(msg.Data), &vo)
			bytes, err := json.Marshal(vo)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
			c.sendToSomeone(msg.Type, status, receiver, data)
			sendType = enum.SendTargetType_SendToSomeone
		// create a funding tx
		case enum.MsgType_FundingCreate_Edit:
			node, err := service.FundingCreateService.Edit(msg.Data)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_FundingCreate_ItemById:
			id, err := strconv.Atoi(msg.Data)
			if err != nil {
				log.Println(err)
				break
			}
			node, err := service.FundingCreateService.Item(id)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_FundingCreate_DelAll:
			err := service.FundingCreateService.DelAll()
			if err != nil {
				data = err.Error()
			} else {
				data = "del success"
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_FundingCreate_DelById:
			id, err := strconv.Atoi(msg.Data)
			for {
				if err != nil {
					data = err.Error()
					break
				}
				err = service.FundingCreateService.Del(id)
				if err != nil {
					data = err.Error()
				} else {
					data = "del success"
					status = true
				}
				break
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_FundingCreate_Count:
			count, err := service.FundingCreateService.TotalCount()
			if err != nil {
				data = err.Error()
			} else {
				data = strconv.Itoa(count)
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_FundingSign_Edit:
			signed, err := service.FundingSignService.Edit(msg.Data)
			if err != nil {
				data = err.Error()
			}

			bytes, err := json.Marshal(signed)
			if err != nil {
				data = err.Error()
			}
			if len(data) == 0 {
				data = string(bytes)
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_FundingSign_ItemById:
			id, err := strconv.Atoi(msg.Data)
			if err != nil {
				data = err.Error()
			} else {
				signed, err := service.FundingSignService.Item(id)
				if err != nil {
					data = err.Error()
				} else {
					bytes, err := json.Marshal(signed)
					if err != nil {
						data = err.Error()
					} else {
						data = string(bytes)
						status = true
					}
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_FundingSign_Count:
			count, err := service.FundingSignService.TotalCount()
			if err != nil {
				data = err.Error()
			} else {
				data = strconv.Itoa(count)
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone

		case enum.MsgType_FundingSign_DelById:
			id, err := strconv.Atoi(msg.Data)
			if err != nil {
				data = err.Error()
			} else {
				signed, err := service.FundingSignService.Del(id)
				if err != nil {
					data = err.Error()
				} else {
					bytes, err := json.Marshal(signed)
					if err != nil {
						data = err.Error()
					} else {
						data = string(bytes)
						status = true
					}
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_FundingSign_DelAll:
			err = service.FundingSignService.DelAll()
			if err != nil {
				data = err.Error()
			} else {
				data = "del success"
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone

		case enum.MsgType_CommitmentTx_Edit:
			node, err := service.CommitTxService.Edit(msg.Data)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_CommitmentTx_ItemByChanId:
			nodes, count, err := service.CommitTxService.GetItemsByChannelId(msg.Data)
			log.Println(*count)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(nodes)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_CommitmentTx_ItemById:
			nodes, err := service.CommitTxService.GetItemById(int(gjson.Parse(msg.Data).Int()))
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(nodes)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_CommitmentTx_Count:
			count, err := service.CommitTxService.TotalCount()
			if err != nil {
				data = err.Error()
			} else {
				data = strconv.Itoa(count)
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone

		case enum.MsgType_CommitmentTxSigned_Edit:
			node, err := service.CommitTxSignedService.Edit(msg.Data)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_CommitmentTxSigned_ItemByChanId:
			nodes, count, err := service.CommitTxSignedService.GetItemsByChannelId(msg.Data)
			log.Println(*count)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(nodes)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_CommitmentTxSigned_ItemById:
			nodes, err := service.CommitTxSignedService.GetItemById(int(gjson.Parse(msg.Data).Int()))
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(nodes)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		case enum.MsgType_CommitmentTxSigned_Count:
			count, err := service.CommitTxSignedService.TotalCount()
			if err != nil {
				data = err.Error()
			} else {
				data = strconv.Itoa(count)
				status = true
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone

		case enum.MsgType_GetBalanceRequest:
		case enum.MsgType_GetBalanceRespond:
		default:
		}

		//broadcast except me
		if sendType == enum.SendTargetType_SendToExceptMe {
			for client := range GlobalWsClientManager.Clients_map {
				if client != c {
					jsonMessage := getReplyObj(string(dataReq), msg.Type, status, c)
					client.SendChannel <- jsonMessage
				}
			}
		}
		//broadcast to all
		if sendType == enum.SendTargetType_SendToAll {
			jsonMessage := getReplyObj(string(dataReq), msg.Type, status, c)
			GlobalWsClientManager.Broadcast <- jsonMessage
		}
	}
}

func (c *Client) sendToSomeone(msgType enum.MsgType, status bool, recipient bean.User, data string) {
	for client := range GlobalWsClientManager.Clients_map {
		if client.User.Email == recipient.Email {
			jsonMessage := getReplyObj(data, msgType, status, c)
			client.SendChannel <- jsonMessage
			break
		}
	}
}

func getReplyObj(data string, msgType enum.MsgType, status bool, client *Client) []byte {
	var jsonMessage []byte
	parse := gjson.Parse(data)
	if parse.Exists() {
		jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, Sender: client.Id, Result: parse.Value()})
	} else {
		jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, Sender: client.Id, Result: data})
	}
	return jsonMessage
}

func (c *Client) sendToMyself(msgType enum.MsgType, status bool, data string) {
	parse := gjson.Parse(data)
	var jsonMessage []byte
	if parse.Exists() {
		jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, Sender: c.Id, Result: parse.Value()})
	} else {
		jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, Sender: c.Id, Result: data})
	}
	c.SendChannel <- jsonMessage
}
