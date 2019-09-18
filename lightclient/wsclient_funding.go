package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"encoding/json"
	"log"
	"strconv"
)

func (client *Client) fundingTransactionModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_FundingCreate_BtcCreate:
		node, err := service.FundingTransactionService.CreateFundingBtcTxRequest(msg.Data, client.User)
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone

	case enum.MsgType_FundingCreate_OmniCreate:
		//check target whether is online
		fundingInfo := &bean.FundingCreated{}
		err := json.Unmarshal([]byte(msg.Data), fundingInfo)
		if err != nil {
			data = err.Error()
		} else {
			node, err := service.ChannelService.GetChannelByTemporaryChanIdArray(fundingInfo.TemporaryChannelId)
			if node != nil {
				peerId := node.PeerIdA
				if peerId == client.User.PeerId {
					peerId = node.PeerIdB
				}
				_, err := client.FindUser(&peerId)
				if err != nil {
					data = err.Error()
				}
			} else {
				data = err.Error()
			}
		}

		if data == "" {
			node, err := service.FundingTransactionService.CreateFundingOmniTxRequest(msg.Data, client.User)
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
				if peerId == client.User.PeerId {
					peerId = node.PeerIdB
				}
				client.sendToSomeone(msg.Type, status, peerId, data)
			}
		}
		client.sendToMyself(msg.Type, status, data)
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
		client.sendToMyself(msg.Type, status, data)
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_FundingCreate_ALlItem:
		node, err := service.FundingTransactionService.AllItem(client.User.PeerId)
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
		client.sendToMyself(msg.Type, status, data)
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_FundingCreate_Count:
		count, err := service.FundingTransactionService.TotalCount(client.User.PeerId)
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(count)
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}
func (client *Client) fundingSignModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_FundingSign_BtcSign:
		node, err := service.FundingTransactionService.FundingBtcTxSign(msg.Data, client.User)
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_FundingSign_OmniSign: //get openChannelReq from funder then send to fundee  create a funding tx
		signed, err := service.FundingTransactionService.FundingTxSign(msg.Data, client.User)
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}
