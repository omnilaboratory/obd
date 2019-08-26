package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"encoding/json"
	"log"
	"strconv"
)

func (c *Client) fundCreateModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	var dataOut []byte
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
	dataOut = []byte(data)
	return sendType, dataOut, status
}
func (c *Client) fundSignModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	var dataOut []byte
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
		err := service.FundingSignService.DelAll()
		if err != nil {
			data = err.Error()
		} else {
			data = "del success"
			status = true
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	dataOut = []byte(data)

	return sendType, dataOut, status
}
