package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"LightningOnOmni/tool"
	"encoding/json"
	"log"
	"strconv"
)

func (client *Client) fundingTransactionModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_FundingCreate_BtcCreate_N3400:
		node, _, err := service.FundingTransactionService.BTCFundingCreated(msg, client.User)
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
		if status {
			_ = client.sendDataToSomeone(msg, status, data)
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone

	case enum.MsgType_FundingCreate_AssetFundingCreated_N34:
		//check target whether is online
		fundingInfo := &bean.FundingCreated{}
		err := json.Unmarshal([]byte(msg.Data), fundingInfo)
		if err != nil {
			data = err.Error()
		} else {
			node, err := service.FundingTransactionService.AssetFundingCreated(msg.Data, client.User)
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
			if node != nil && status {
				peerId := node.PeerIdA
				if peerId == client.User.PeerId {
					peerId = node.PeerIdB
				}
				_ = client.sendToSomeone(msg.Type, status, peerId, data)
			}
		}

		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_FundingCreate_ItemByTempId_N3401:
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
	case enum.MsgType_FundingCreate_ItemById_N3402:
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
	case enum.MsgType_FundingCreate_ALlItem_N3403:
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
	case enum.MsgType_FundingCreate_DelById_N3405:
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
	case enum.MsgType_FundingCreate_Count_N3404:
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
	case enum.MsgType_FundingSign_BtcSign_N3500:
		node, funder, err := service.FundingTransactionService.FundingBtcTxSigned(msg.Data, client.User)
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

		if tool.CheckIsString(&funder) {
			_ = client.sendDataToSomeone(msg, status, data)
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_FundingSign_AssetFundingSigned_N35: //get openChannelReq from funder then send to fundee  create a funding tx
		node, err := service.FundingTransactionService.AssetFundingSigned(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		}

		bytes, err := json.Marshal(node)
		if err != nil {
			data = err.Error()
		}
		if len(data) == 0 {
			data = string(bytes)
			status = true
		}

		if node != nil && status {
			peerId := node.PeerIdA
			if peerId == client.User.PeerId {
				peerId = node.PeerIdB
			}
			_ = client.sendToSomeone(msg.Type, status, peerId, data)
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}
