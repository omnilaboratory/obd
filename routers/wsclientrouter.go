package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/rpc"
	"LightningOnOmni/service"
	"encoding/json"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
)

func (c *Client) userModule(msg bean.RequestMessage, sendType enum.SendTargetType, dataReq []byte, data *string, status bool, receiver bean.User, err error) (enum.SendTargetType, []byte, bool) {
	switch msg.Type {
	case enum.MsgType_UserLogin:
		if c.User != nil {
			c.sendToMyself(msg.Type, true, "already login")
			sendType = enum.SendTargetType_SendToSomeone
		} else {
			var user bean.User
			json.Unmarshal([]byte(msg.Data), &user)
			if len(user.PeerId) > 0 {
				c.User = &user
				service.UserService.UserLogin(&user)
				dataReq = []byte((c.User.PeerId + " login"))
				sendType = enum.SendTargetType_SendToAll
				status = true
			} else {
				c.sendToMyself(msg.Type, true, "error peer_id")
				sendType = enum.SendTargetType_SendToSomeone
			}
		}
	case enum.MsgType_UserLogout:
		if c.User != nil {
			dataReq = []byte(c.User.PeerId + " logout")
			//c.sendToMyself(msg.Type, true, "logout success")
			c.User = nil
			sendType = enum.SendTargetType_SendToAll
			status = true
		} else {
			c.sendToMyself(msg.Type, true, "please login")
			sendType = enum.SendTargetType_SendToSomeone
		}
	}
	return sendType, dataReq, status
}

func (c *Client) omniCoreModule(msg bean.RequestMessage, sendType enum.SendTargetType, dataReq []byte, data *string, status bool, receiver bean.User, err error) (enum.SendTargetType, []byte, bool) {
	tempData := ""
	switch msg.Type {
	case enum.MsgType_Core_GetNewAddress:
		var label = msg.Data
		client := rpc.NewClient()
		address, err := client.GetNewAddress(label)
		if err != nil {
			tempData = err.Error()
		} else {
			tempData = address
			status = true
		}
		c.sendToMyself(msg.Type, status, tempData)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetMiningInfo:
		client := rpc.NewClient()
		result, err := client.GetMiningInfo()
		if err != nil {
			tempData = err.Error()
		} else {
			tempData = result
			status = true
		}
		c.sendToMyself(msg.Type, status, tempData)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetNetworkInfo:
		client := rpc.NewClient()
		result, err := client.GetNetworkInfo()
		if err != nil {
			tempData = err.Error()
		} else {
			tempData = result
			status = true
		}
		c.sendToMyself(msg.Type, status, tempData)
		sendType = enum.SendTargetType_SendToSomeone
	}
	data = &tempData

	return sendType, dataReq, status
}
func (c *Client) channelModule(msg bean.RequestMessage, sendType enum.SendTargetType, dataReq []byte, status bool, receiver bean.User, err error) (enum.SendTargetType, []byte, bool, string) {
	data := ""
	switch msg.Type {
	//get openChannelReq from funder then send to fundee
	case enum.MsgType_ChannelOpen:
		node, err := service.ChannelService.OpenChannel(msg.Data, c.User.PeerId)
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
		c.sendToSomeone(msg.Type, status, &receiver, data)
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
		node, err := service.ChannelService.AcceptChannel(msg.Data, c.User.PeerId)
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
		c.sendToSomeone(msg.Type, status, &receiver, data)
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
		// create a funding tx
	}
	return sendType, dataReq, status, data
}

func (c *Client) fundCreateModule(msg bean.RequestMessage, sendType enum.SendTargetType, dataReq []byte, status bool, receiver bean.User, err error) (enum.SendTargetType, []byte, bool, string) {
	data := ""
	switch msg.Type {
	//get openChannelReq from funder then send to fundee
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
	}
	return sendType, dataReq, status, data
}
func (c *Client) fundSignModule(msg bean.RequestMessage, sendType enum.SendTargetType, dataReq []byte, status bool, receiver bean.User, err error) (enum.SendTargetType, []byte, bool, string) {
	data := ""
	switch msg.Type {
	//get openChannelReq from funder then send to fundee
	// create a funding tx
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
	}
	return sendType, dataReq, status, data
}

func (c *Client) commitmentTxModule(msg bean.RequestMessage, sendType enum.SendTargetType, dataReq []byte, status bool, receiver bean.User, err error) (enum.SendTargetType, []byte, bool, string) {
	data := ""
	switch msg.Type {
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
	}
	return sendType, dataReq, status, data
}
func (c *Client) commitmentTxSignModule(msg bean.RequestMessage, sendType enum.SendTargetType, dataReq []byte, status bool, receiver bean.User, err error) (enum.SendTargetType, []byte, bool, string) {
	data := ""
	switch msg.Type {
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
	}
	return sendType, dataReq, status, data
}

func (c *Client) otherModule(msg bean.RequestMessage, sendType enum.SendTargetType, dataReq []byte, status bool, receiver bean.User, err error) (enum.SendTargetType, []byte, bool, string) {
	data := ""
	switch msg.Type {
	case enum.MsgType_GetBalanceRequest:
	case enum.MsgType_GetBalanceRespond:
	default:
	}
	return sendType, dataReq, status, data
}
