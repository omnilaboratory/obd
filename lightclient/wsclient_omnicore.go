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
	rpcClient := rpc.NewClient()
	switch msg.Type {
	case enum.MsgType_Core_GetNewAddress:
		var label = msg.Data
		address, err := rpcClient.GetNewAddress(label)
		if err != nil {
			data = err.Error()
		} else {
			data = address
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetMiningInfo:
		result, err := rpcClient.GetMiningInfo()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetNetworkInfo:
		result, err := rpcClient.GetNetworkInfo()
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
			result, err := rpcClient.SignMessageWithPrivKey(privkey, message)
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
			ok, err := rpcClient.VerifyMessage(address, signature, message)
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
			ok, err := rpcClient.DumpPrivKey(address)
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
			ok, err := rpcClient.ListUnspent(address)
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
			balance, err := rpcClient.GetBalanceByAddress(address)
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
		fromBitCoinAddressPrivKey := gjson.Get(msg.Data, "fromBitCoinAddressPrivKey").String()
		toBitCoinAddress := gjson.Get(msg.Data, "toBitCoinAddress").String()
		amount := gjson.Get(msg.Data, "amount").Float()
		minerFee := gjson.Get(msg.Data, "minerFee").Float()
		privKeys := make([]string, 0)
		if tool.CheckIsString(&fromBitCoinAddressPrivKey) {
			privKeys = append(privKeys, fromBitCoinAddressPrivKey)
		}
		if tool.CheckIsString(&fromBitCoinAddress) &&
			tool.CheckIsString(&toBitCoinAddress) {
			txid, hex, err := rpcClient.BtcCreateAndSignRawTransaction(fromBitCoinAddress, privKeys, []rpc.TransactionOutputItem{{toBitCoinAddress, amount}}, minerFee, 0, nil)
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
	case enum.MsgType_Core_Omni_CreateAndSignRawTransaction:
		fromBitCoinAddress := gjson.Get(msg.Data, "fromBitCoinAddress").String()
		fromBitCoinAddressPrivKey := gjson.Get(msg.Data, "fromBitCoinAddressPrivKey").String()
		toBitCoinAddress := gjson.Get(msg.Data, "toBitCoinAddress").String()
		amount := gjson.Get(msg.Data, "amount").Float()
		minerFee := gjson.Get(msg.Data, "minerFee").Float()
		propertyId := gjson.Get(msg.Data, "propertyId").Int()
		privKeys := make([]string, 0)
		if tool.CheckIsString(&fromBitCoinAddressPrivKey) {
			privKeys = append(privKeys, fromBitCoinAddressPrivKey)
		}
		_, err := rpcClient.OmniGetProperty(propertyId)
		if err != nil {
			data = err.Error()
		} else {
			if tool.CheckIsString(&fromBitCoinAddress) &&
				tool.CheckIsString(&toBitCoinAddress) {
				_, hex, err := rpcClient.OmniCreateAndSignRawTransaction(fromBitCoinAddress, privKeys, toBitCoinAddress, propertyId, amount, minerFee, 0)
				node := make(map[string]interface{})
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
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone

	}
	return sendType, []byte(data), status
}
