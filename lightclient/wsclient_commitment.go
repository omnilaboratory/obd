package lightclient

import (
	"encoding/json"
	"log"
	"obd/bean"
	"obd/bean/enum"
	"obd/service"
	"strconv"

	"github.com/tidwall/gjson"
)

func (client *Client) commitmentTxModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_CommitmentTx_CommitmentTransactionCreated_N351:
		retData, err := service.CommitmentTxService.CommitmentTransactionCreated(msg, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(retData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		if status {
			_ = client.sendDataToP2PUser(msg, status, data)
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTx_ItemsByChanId_N35101:
		nodes, count, err := service.CommitmentTxService.GetItemsByChannelId(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			page := make(map[string]interface{})
			page["count"] = len(nodes)
			page["totalCount"] = count
			page["body"] = nodes
			bytes, err := json.Marshal(page)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTx_ItemById_N35102:
		node, err := service.CommitmentTxService.GetItemById(int(gjson.Parse(msg.Data).Int()))
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
	case enum.MsgType_CommitmentTx_Count_N35103:
		count, err := service.CommitmentTxService.TotalCount()
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(count)
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTx_LatestCommitmentTxByChanId_N35104:
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
	case enum.MsgType_CommitmentTx_LatestRDByChanId_N35105:
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
	case enum.MsgType_CommitmentTx_AllRDByChanId_N35108:
		node, err := service.CommitmentTxService.GetLatestAllRDByChannelId(msg.Data, client.User)
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
	case enum.MsgType_CommitmentTx_LatestBRByChanId_N35106:
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
	case enum.MsgType_CommitmentTx_AllBRByChanId_N35109:
		node, err := service.CommitmentTxService.GetLatestAllBRByChannelId(msg.Data, client.User)
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

	case enum.MsgType_SendBreachRemedyTransaction_N35107:
		node, err := service.ChannelService.SendBreachRemedyTransaction(msg.Data, client.User)
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
	case enum.MsgType_CommitmentTx_GetBroadcastCommitmentTx_N35110:
		node, err := service.CommitmentTxService.GetBroadcastCommitmentTxByChannelId(msg.Data, client.User)
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
	case enum.MsgType_CommitmentTx_GetBroadcastRDTx_N35111:
		node, err := service.CommitmentTxService.GetBroadcastRDTxByChannelId(msg.Data, client.User)
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
	case enum.MsgType_CommitmentTx_GetBroadcastBRTx_N35112:
		node, err := service.CommitmentTxService.GetBroadcastCommitmentTxByChannelId(msg.Data, client.User)
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
	case enum.MsgType_CommitmentTxSigned_RevokeAndAcknowledgeCommitmentTransaction_N352:
		retData, targetUser, err := service.CommitmentTxSignedService.RevokeAndAcknowledgeCommitmentTransaction(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(retData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
			if targetUser != "" {
				_ = client.sendDataToP2PUser(msg, status, data)
			}
		} else {
			//如果数据处理成功，就需要通过353，把数据传给alice所在的节点
			msg.Type = enum.MsgType_CommitmentTxSigned_ToAliceSign_N353
			_ = client.sendDataToP2PUser(msg, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTxSigned_ItemByChanId_N35201:
		nodes, count, err := service.CommitmentTxSignedService.GetItemsByChannelId(msg.Data)
		log.Println(*count)
		if err != nil {
			data = err.Error()
		} else {
			page := make(map[string]interface{})
			page["count"] = len(nodes)
			page["totalCount"] = count
			page["body"] = nodes
			bytes, err := json.Marshal(page)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTxSigned_ItemById_N35202:
		node, err := service.CommitmentTxSignedService.GetItemById(int(gjson.Parse(msg.Data).Int()))
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
	case enum.MsgType_CommitmentTxSigned_Count_N35203:
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
	case enum.MsgType_CommitmentTxSigned_ToAliceSign_N353:
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
		client.sendToMyself(enum.MsgType_CommitmentTxSigned_SecondToBobSign_N354, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	default:
	}

	return sendType, []byte(data), status
}
