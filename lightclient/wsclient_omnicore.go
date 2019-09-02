package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"encoding/json"

	"github.com/tidwall/gjson"
)

func (client *Client) omniCoreModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetMiningInfo:
		result, err := client.GetMiningInfo()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetNetworkInfo:
		result, err := client.GetNetworkInfo()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
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
		client.sendToMyself(msg.Type, status, data)
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
		client.sendToMyself(msg.Type, status, data)
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_ListUnspent:
		address := gjson.Get(msg.Data, "address").String()
		if tool.CheckIsString(&address) {
			ok, err := client.ListUnspent(address)
			if err != nil {
				data = err.Error()
			} else {
				data = ok
				status = true
			}
		} else {
			data = "error address"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_BalanceByAddress:
		address := gjson.Get(msg.Data, "address").String()
		if tool.CheckIsString(&address) {
			balance, err := client.GetBalanceByAddress(address)
			if err != nil {
				data = err.Error()
			} else {
				data = balance.String()
				status = true
			}
		} else {
			data = "error address"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_BtcCreateAndSignRawTransaction:
		fromBitCoinAddress := gjson.Get(msg.Data, "fromBitCoinAddress").String()
		toBitCoinAddress := gjson.Get(msg.Data, "toBitCoinAddress").String()
		amount := gjson.Get(msg.Data, "amount").Float()
		minerFee := gjson.Get(msg.Data, "minerFee").Float()
		if tool.CheckIsString(&fromBitCoinAddress) &&
			tool.CheckIsString(&toBitCoinAddress) {
			txid, hex, err := client.BtcCreateAndSignRawTransaction(fromBitCoinAddress, nil, []rpc.TransactionOutputItem{{toBitCoinAddress, amount}}, minerFee, 0)
			node := make(map[string]interface{})
			node["txid"] = txid
			node["hex"] = hex

			if err != nil {
				data = err.Error()
			} else {
				bytes, _ := json.Marshal(node)
				data = string(bytes)
				status = true
			}
		} else {
			data = "error address"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone

	}
	return sendType, []byte(data), status
}
