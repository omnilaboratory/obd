package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/rpc"
)

func (c *Client) omniCoreModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	var dataOut []byte
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
	dataOut = []byte(tempData)
	return sendType, dataOut, status
}
