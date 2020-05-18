package lightclient

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
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
	case enum.MsgType_Core_GetNewAddress_1001:
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
	case enum.MsgType_Core_GetMiningInfo_1002:
		result, err := rpcClient.GetMiningInfo()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_GetNetworkInfo_1003:
		result, err := rpcClient.GetNetworkInfo()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_ListProperties_1205:
		result, err := rpcClient.OmniListProperties()
		if err != nil {
			data = err.Error()
		} else {
			data = result
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_GetProperty_1207:
		propertyId := gjson.Get(msg.Data, "propertyId").Int()
		if propertyId == 0 {
			data = "error propertyId"
		} else {
			result, err := rpcClient.OmniGetProperty(propertyId)
			if err != nil {
				data = err.Error()
			} else {
				data = result
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_SignMessageWithPrivKey_1004:
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
	case enum.MsgType_Core_VerifyMessage_1005:
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
	case enum.MsgType_Core_DumpPrivKey_1006:
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
	case enum.MsgType_Core_Omni_GetTransaction_1206:
		txid := gjson.Get(msg.Data, "txid").String()
		if tool.CheckIsString(&txid) {
			result, err := rpcClient.OmniGettransaction(txid)
			if err != nil {
				data = err.Error()
			} else {
				data = result
				status = true
			}
		} else {
			data = "error address"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Btc_ImportPrivKey_1011:
		privkey := gjson.Get(msg.Data, "privkey").String()
		if tool.CheckIsString(&privkey) {
			outData, err := rpcClient.ImportPrivKey(privkey)
			if err != nil {
				data = err.Error()
			} else {
				data = outData
				status = true
			}
		} else {
			data = "error privkey"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_ListUnspent_1007:
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
	case enum.MsgType_Core_BalanceByAddress_1008:
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
	case enum.MsgType_Core_Omni_Getbalance_1200:
		address := gjson.Get(msg.Data, "address").String()
		if tool.CheckIsString(&address) {
			result, err := rpcClient.OmniGetAllBalancesForAddress(address)
			if err != nil {
				data = err.Error()
			} else {
				data = result
				status = true
			}
		} else {
			data = "error address"
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_Omni_CreateNewTokenFixed_1201:
		if tool.CheckIsString(&msg.Data) {
			reqData := &bean.OmniSendIssuanceFixed{}
			err := json.Unmarshal([]byte(msg.Data), reqData)
			if err != nil {
				data = err.Error()
			} else {
				result, err := rpcClient.OmniSendIssuanceFixed(reqData.FromAddress, reqData.Ecosystem, reqData.DivisibleType, reqData.Name, reqData.Data, reqData.Amount)
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
	case enum.MsgType_Core_Omni_CreateNewTokenManaged_1202:
		if tool.CheckIsString(&msg.Data) {
			reqData := &bean.OmniSendIssuanceManaged{}
			err := json.Unmarshal([]byte(msg.Data), reqData)
			if err != nil {
				data = err.Error()
			} else {
				result, err := rpcClient.OmniSendIssuanceManaged(reqData.FromAddress, reqData.Ecosystem, reqData.DivisibleType, reqData.Name, reqData.Data)
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
	case enum.MsgType_Core_Omni_GrantNewUnitsOfManagedToken_1203:
		if tool.CheckIsString(&msg.Data) {
			reqData := &bean.OmniSendGrant{}
			err := json.Unmarshal([]byte(msg.Data), reqData)
			if err != nil {
				data = err.Error()
			} else {
				result, err := rpcClient.OmniSendGrant(reqData.FromAddress, reqData.PropertyId, reqData.Amount, reqData.Memo)
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
	case enum.MsgType_Core_Omni_RevokeUnitsOfManagedToken_1204:
		if tool.CheckIsString(&msg.Data) {
			reqData := &bean.OmniSendRevoke{}
			err := json.Unmarshal([]byte(msg.Data), reqData)
			if err != nil {
				data = err.Error()
			} else {
				result, err := rpcClient.OmniSendRevoke(reqData.FromAddress, reqData.PropertyId, reqData.Amount, reqData.Memo)
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
	case enum.MsgType_Core_FundingBTC_1009:
		sendInfo := &bean.FundingBtc{}
		err := json.Unmarshal([]byte(msg.Data), sendInfo)
		if err != nil {
			data = "error data: " + err.Error()
		} else {
			privKeys := make([]string, 0)
			if tool.CheckIsString(&sendInfo.FromAddressPrivateKey) {
				privKeys = append(privKeys, sendInfo.FromAddressPrivateKey)
			}
			if tool.CheckIsString(&sendInfo.FromAddress) &&
				tool.CheckIsString(&sendInfo.ToAddress) &&
				sendInfo.Amount > 0 {
				txid, hex, err := rpcClient.BtcCreateAndSignRawTransaction(sendInfo.FromAddress, privKeys, []rpc.TransactionOutputItem{{sendInfo.ToAddress, sendInfo.Amount}}, sendInfo.MinerFee, 0, nil)
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
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_Core_BtcCreateMultiSig_1010:
		reqData := &bean.CreateMultiSigRequest{}
		err := json.Unmarshal([]byte(msg.Data), reqData)
		if err != nil {
			data = "error data: " + err.Error()
		} else {
			result, err := rpcClient.CreateMultiSig(reqData.MiniSignCount, reqData.PubKeys)
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
	case enum.MsgType_Core_Omni_FundingAsset_2001:
		sendInfo := &bean.FundingAsset{}
		err := json.Unmarshal([]byte(msg.Data), sendInfo)
		if err != nil {
			data = "error data"
		} else {
			privKeys := make([]string, 0)
			if tool.CheckIsString(&sendInfo.FromAddressPrivateKey) {
				privKeys = append(privKeys, sendInfo.FromAddressPrivateKey)
			}
			_, err := rpcClient.OmniGetProperty(sendInfo.PropertyId)
			if err != nil {
				data = err.Error()
			} else {
				if tool.CheckIsString(&sendInfo.FromAddress) &&
					tool.CheckIsString(&sendInfo.ToAddress) &&
					sendInfo.Amount > 0 {
					_, hex, err := rpcClient.OmniCreateAndSignRawTransaction(sendInfo.FromAddress, privKeys, sendInfo.ToAddress, sendInfo.PropertyId, sendInfo.Amount, sendInfo.MinerFee, 0)
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
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}
