package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"encoding/json"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
)

func (c *Client) commitmentTxModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_CommitmentTx_Edit:
		node, err := service.CommitmentTxService.Edit(msg.Data)
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
		c.sendToMyself(msg.Type, status, data)
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
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTx_Count:
		count, err := service.CommitmentTxService.TotalCount()
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(count)
			status = true
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CommitmentTx_NewestRDByChanId:
		node, err := service.CommitmentTxService.GetNewestRDTxByChannelId(msg.Data, c.User)
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
	case enum.MsgType_CommitmentTx_NewestBRByChanId:
		count, err := service.CommitmentTxService.TotalCount()
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
func (c *Client) commitmentTxSignModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
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

	return sendType, []byte(data), status
}

func (c *Client) otherModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
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
