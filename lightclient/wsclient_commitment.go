package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"encoding/json"
	"log"
	"strconv"

	"github.com/tidwall/gjson"
)

func (client *Client) commitmentTxModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_CommitmentTx_Create:
		node, targetUser, err := service.CommitmentTxService.CreateNewCommitmentTxRequest(msg.Data, client.User)
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
		if targetUser != nil && status {
			client.sendToSomeone(msg.Type, status, *targetUser, data)
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTx_ItemsByChanId:
		nodes, count, err := service.CommitmentTxService.GetItemsByChannelId(msg.Data)
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTx_ItemById:
		nodes, err := service.CommitmentTxService.GetItemById(int(gjson.Parse(msg.Data).Int()))
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTx_Count:
		count, err := service.CommitmentTxService.TotalCount()
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(count)
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTx_LatestCommitmentTxByChanId:
		node, err := service.CommitmentTxService.GetLatestCommitmentTxByChannelId(msg.Data, client.User)
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
	case enum.MsgType_CommitmentTx_LatestRDByChanId:
		node, err := service.CommitmentTxService.GetLatestRDTxByChannelId(msg.Data, client.User)
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
	case enum.MsgType_CommitmentTx_LatestBRByChanId:
		node, err := service.CommitmentTxService.GetLatestBRTxByChannelId(msg.Data, client.User)
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
	}
	return sendType, []byte(data), status
}
func (client *Client) commitmentTxSignModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""

	switch msg.Type {
	case enum.MsgType_CommitmentTxSigned_Sign:
		ca, cb, _, err := service.CommitmentTxSignedService.CommitmentTxSign(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(ca)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				client.sendToSomeone(msg.Type, status, ca.PeerIdA, data)
			}
			bytes, err = json.Marshal(cb)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				client.sendToSomeone(msg.Type, status, ca.PeerIdB, data)
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTxSigned_ItemByChanId:
		nodes, count, err := service.CommitmentTxSignedService.GetItemsByChannelId(msg.Data)
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTxSigned_ItemById:
		nodes, err := service.CommitmentTxSignedService.GetItemById(int(gjson.Parse(msg.Data).Int()))
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTxSigned_Count:
		count, err := service.CommitmentTxSignedService.TotalCount()
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

func (client *Client) otherModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_GetBalanceRequest:
	case enum.MsgType_GetBalanceRespond:
	default:
	}

	return sendType, []byte(data), status
}
