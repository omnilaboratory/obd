package lightclient

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	. "github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/tool"

	"github.com/tidwall/gjson"
)

func (client *Client) omniCoreModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	rpcClient := rpc.NewClient()
	switch msg.Type {
	case enum.MsgType_Core_GetNewAddress_2101:
		var label = msg.Data
		address, err := GetNewAddress(label)
		if err != nil {
			data = err.Error()
		} else {
			data = address
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetMiningInfo_2102:
		result, err := GetMiningInfo()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetNetworkInfo_2103:
		result, err := GetNetworkInfo()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_ListProperties_2117:
		result, err := OmniListProperties()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_GetProperty_2119:
		propertyId := gjson.Get(msg.Data, "propertyId").Int()
		if propertyId == 0 {
			data = "error propertyId"
		} else {
			result, err := OmniGetProperty(propertyId)
			if err != nil {
				data = err.Error()
			} else {
				data = result
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_Send_2121:
		fromAddress := gjson.Get(msg.Data, "from_address").String()
		toAddress := gjson.Get(msg.Data, "to_address").String()
		propertyId := gjson.Get(msg.Data, "property_id").Int()
		amount := gjson.Get(msg.Data, "amount").Float()
		result, err := OmniSend(fromAddress, toAddress, int(propertyId), amount)
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_GetTransaction_2118:
		txid := gjson.Get(msg.Data, "txid").String()
		if tool.CheckIsString(&txid) {
			result, err := OmniGetTransaction(txid)
			if err != nil {
				data = err.Error()
			} else {
				data = result
				status = true
			}
		} else {
			data = "error txid"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_ListUnspent_2107:
		address := gjson.Get(msg.Data, "address").String()
		if tool.CheckIsString(&address) {
			data = ListUnspent(address)
			status = true
		} else {
			data = "error address"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_BalanceByAddress_2108:
		address := gjson.Get(msg.Data, "address").String()
		if tool.CheckIsString(&address) {
			balance, err := GetBalanceByAddress(address)
			if err != nil {
				data = err.Error()
			} else {
				data = tool.FloatToString(balance, 8)
				status = true
			}
		} else {
			data = "error address"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_GetBalance_2112:
		address := gjson.Get(msg.Data, "address").String()
		if tool.CheckIsAddress(address) {
			result := OmniGetAllBalancesByAddress(address)
			if result == "" {
				data = "empty result"
			} else {
				data = result
				status = true
			}
		} else {
			data = "error address"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_CreateNewTokenFixed_2113:
		if tool.CheckIsString(&msg.Data) {
			reqData := &bean.OmniSendIssuanceFixed{}
			err := json.Unmarshal([]byte(msg.Data), reqData)
			if err != nil {
				data = err.Error()
			} else {
				result, err := OmniSendIssuanceFixed(reqData.FromAddress, reqData.Ecosystem, reqData.DivisibleType, reqData.Name, reqData.Data, reqData.Amount)
				if err != nil {
					data = err.Error()
				} else {
					data = result
					status = true
				}
			}
		} else {
			data = "error data"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_CreateNewTokenManaged_2114:
		if tool.CheckIsString(&msg.Data) {
			reqData := &bean.OmniSendIssuanceManaged{}
			err := json.Unmarshal([]byte(msg.Data), reqData)
			if err != nil {
				data = err.Error()
			} else {
				result, err := OmniSendIssuanceManaged(reqData.FromAddress, reqData.Ecosystem, reqData.DivisibleType, reqData.Name, reqData.Data)
				if err != nil {
					data = err.Error()
				} else {
					data = result
					status = true
				}
			}
		} else {
			data = "error data"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_GrantNewUnitsOfManagedToken_2115:
		if tool.CheckIsString(&msg.Data) {
			reqData := &bean.OmniSendGrant{}
			err := json.Unmarshal([]byte(msg.Data), reqData)
			if err != nil {
				data = err.Error()
			} else {
				result, err := OmniSendGrant(reqData.FromAddress, reqData.PropertyId, reqData.Amount, reqData.Memo)
				if err != nil {
					data = err.Error()
				} else {
					data = result
					status = true
				}
			}
		} else {
			data = "error data"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_RevokeUnitsOfManagedToken_2116:
		if tool.CheckIsString(&msg.Data) {
			reqData := &bean.OmniSendRevoke{}
			err := json.Unmarshal([]byte(msg.Data), reqData)
			if err != nil {
				data = err.Error()
			} else {
				result, err := OmniSendRevoke(reqData.FromAddress, reqData.PropertyId, reqData.Amount, reqData.Memo)
				if err != nil {
					data = err.Error()
				} else {
					data = result
					status = true
				}
			}
		} else {
			data = "error data"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetTransactionByTxid_2122:
		txid := gjson.Get(msg.Data, "txid").String()
		if tool.CheckIsString(&txid) {
			GetTransactionById(txid)
			data = GetTransactionById(txid)
			status = true
		} else {
			data = "error txid"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_SignRawTransaction_2123:
		result, err := BtcSignRawTransactionFromJson(msg.Data)
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_FundingBTC_2109:
		sendInfo := &bean.FundingBtc{}
		err := json.Unmarshal([]byte(msg.Data), sendInfo)
		if err != nil {
			data = "error data: " + err.Error()
		} else {
			if tool.CheckIsString(&sendInfo.FromAddress) &&
				tool.CheckIsString(&sendInfo.ToAddress) &&
				sendInfo.Amount > 0 {
				resp, err := rpcClient.BtcCreateRawTransaction(sendInfo.FromAddress, []bean.TransactionOutputItem{{sendInfo.ToAddress, sendInfo.Amount}}, sendInfo.MinerFee, 0, nil)
				if err != nil {
					data = err.Error()
				} else {
					resp["is_multisig"] = false
					bytes, _ := json.Marshal(resp)
					data = string(bytes)
					status = true
				}
			} else {
				data = "error address"
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_BtcCreateMultiSig_2110:
		reqData := &bean.CreateMultiSigRequest{}
		err := json.Unmarshal([]byte(msg.Data), reqData)
		if err != nil {
			data = "error data: " + err.Error()
		} else {
			result, err := omnicore.CreateMultiSig(reqData.MiniSignCount, reqData.PubKeys)
			if err == nil {
				parse := gjson.Parse(string(result))
				node := make(map[string]interface{})
				node["txid"] = parse.Get("address")
				node["hex"] = parse.Get("redeemScript")
				bytes, _ := json.Marshal(node)
				data = string(bytes)
				status = true
			} else {
				data = "error data"
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_FundingAsset_2120:
		sendInfo := &bean.FundingAsset{}
		err := json.Unmarshal([]byte(msg.Data), sendInfo)
		if err != nil {
			data = "error data"
		} else {
			if tool.CheckIsString(&sendInfo.FromAddress) &&
				tool.CheckIsString(&sendInfo.ToAddress) &&
				sendInfo.Amount > 0 {
				respNode, err := rpcClient.OmniCreateRawTransaction(sendInfo.FromAddress, sendInfo.ToAddress, sendInfo.PropertyId, sendInfo.Amount, sendInfo.MinerFee)
				if err != nil {
					data = err.Error()
				} else {
					respNode["is_multisig"] = false
					bytes, _ := json.Marshal(respNode)
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
