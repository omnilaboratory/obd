package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"encoding/json"
	"log"
	"strconv"
)

func (c *Client) fundingTransactionModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_FundingCreate_Create:
		//check target whether is online
		fundingInfo := &bean.FundingCreated{}
		err := json.Unmarshal([]byte(msg.Data), fundingInfo)
		if err != nil {
			data = err.Error()
		} else {
			node, err := service.ChannelService.GetChannelByTemporaryChanIdArray(fundingInfo.TemporaryChannelId)
			if node != nil {
				peerId := node.PeerIdA
				if peerId == c.User.PeerId {
					peerId = node.PeerIdB
				}
				_, err := c.FindUser(&peerId)
				if err != nil {
					data = err.Error()
				}
			} else {
				data = err.Error()
			}
		}

		if data == "" {
			node, err := service.FundingTransactionService.CreateFundingTx(msg.Data, c.User)
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
			if node != nil {
				peerId := node.PeerIdA
				if peerId == c.User.PeerId {
					peerId = node.PeerIdB
				}
				c.sendToSomeone(msg.Type, status, peerId, data)
			}
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_FundingCreate_ItemByTempId:
		node, err := service.FundingTransactionService.ItemByTempId(msg.Data)
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
		node, err := service.FundingTransactionService.ItemById(id)
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
	case enum.MsgType_FundingCreate_ALlItem:
		node, err := service.FundingTransactionService.AllItem(c.User.PeerId)
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
	case enum.MsgType_FundingCreate_DelById:
		id, err := strconv.Atoi(msg.Data)
		for {
			if err != nil {
				data = err.Error()
				break
			}
			err = service.FundingTransactionService.Del(id)
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
		count, err := service.FundingTransactionService.TotalCount(c.User.PeerId)
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(count)
			status = true
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}
func (c *Client) fundingSignModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	//get openChannelReq from funder then send to fundee
	// create a funding tx
	case enum.MsgType_FundingSign_Edit:
		signed, err := service.FundingTransactionService.FundingTransactionSign(msg.Data, c.User)
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
	}
	return sendType, []byte(data), status
}
