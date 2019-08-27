package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"github.com/tidwall/gjson"
)

func (c *Client) omniCoreModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	var dataOut []byte
	data := ""
	client := rpc.NewClient()
	switch msg.Type {
	case enum.MsgType_Core_GetNewAddress:
		var label = msg.Data
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
		result, err := client.GetNetworkInfo()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_SignMessageWithPrivKey:
		privkey := gjson.Get(msg.Data, "privkey").String()
		message := gjson.Get(msg.Data, "message").String()
		if tool.CheckIsString(&privkey) && tool.CheckIsString(&message) {
			result, err := client.SignMessageWithPrivKey(privkey, message)
			if err != nil {
				data = err.Error()
			} else {
				data = result
				status = true
			}
		} else {
			data = "error data"
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_VerifyMessage:
		address := gjson.Get(msg.Data, "address").String()
		signature := gjson.Get(msg.Data, "signature").String()
		message := gjson.Get(msg.Data, "message").String()
		if tool.CheckIsString(&address) && tool.CheckIsString(&signature) && tool.CheckIsString(&message) {
			ok, err := client.VerifyMessage(address, signature, message)
			if err != nil {
				data = err.Error()
			} else {
				data = ok
				status = true
			}
		} else {
			data = "error data"
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_DumpPrivKey:
		address := gjson.Get(msg.Data, "address").String()
		if tool.CheckIsString(&address) {
			ok, err := client.DumpPrivKey(address)
			if err != nil {
				data = err.Error()
			} else {
				data = ok
				status = true
			}
		} else {
			data = "error address"
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	dataOut = []byte(data)
	return sendType, dataOut, status
}
