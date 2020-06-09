package lightclient

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"log"
	"strconv"
)

func (client *Client) fundingTransactionModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_FundingCreate_SendBtcFundingCreated_340:
		node, _, err := service.FundingTransactionService.BTCFundingCreated(msg, client.User)
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
		if status {
			msg.Type = enum.MsgType_FundingCreate_BtcFundingCreated_340
			err := client.sendDataToP2PUser(msg, status, data)
			if err != nil {
				data = err.Error()
				status = false
			}
		}
		msg.Type = enum.MsgType_FundingCreate_SendBtcFundingCreated_340
		client.sendToMyself(msg.Type, status, data)
	case enum.MsgType_FundingCreate_Btc_ALlItem_3104:
		node, err := service.FundingTransactionService.BtcFundingAllItem(*client.User)
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
	case enum.MsgType_FundingCreate_Btc_ItemById_3105:
		id, err := strconv.Atoi(msg.Data)
		if err != nil {
			log.Println(err)
			break
		}
		node, err := service.FundingTransactionService.BtcFundingItemById(id, *client.User)
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
	case enum.MsgType_FundingCreate_Btc_ItemByTempChannelId_3106:
		node, err := service.FundingTransactionService.BtcFundingItemByTempChannelId(msg.Data, *client.User)
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

	case enum.MsgType_FundingCreate_Btc_RDALlItem_3107:
		node, err := service.FundingTransactionService.BtcFundingRDAllItem(*client.User)
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
	case enum.MsgType_FundingCreate_Btc_ItemRDById_3108:
		id, err := strconv.Atoi(msg.Data)
		if err != nil {
			log.Println(err)
			break
		}
		node, err := service.FundingTransactionService.BtcFundingRDItemById(id, *client.User)
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
	case enum.MsgType_FundingCreate_Btc_ItemRDByTempChannelId_3109:
		node, err := service.FundingTransactionService.BtcFundingItemRDByTempChannelId(msg.Data, *client.User)
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
	case enum.MsgType_FundingCreate_Btc_ItemRDByTempChannelIdAndTxId_3110:
		node, err := service.FundingTransactionService.BtcFundingItemRDByTempChannelIdAndFundingTxid(msg.Data, *client.User)
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

	case enum.MsgType_FundingCreate_SendAssetFundingCreated_34:
		//check target whether is online
		fundingInfo := &bean.FundingCreated{}
		err := json.Unmarshal([]byte(msg.Data), fundingInfo)
		if err != nil {
			data = err.Error()
		} else {
			node, err := service.FundingTransactionService.AssetFundingCreated(msg.Data, client.User)
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
			if status {
				msg.Type = enum.MsgType_FundingCreate_AssetFundingCreated_34
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		msg.Type = enum.MsgType_FundingCreate_SendAssetFundingCreated_34
		client.sendToMyself(msg.Type, status, data)
	case enum.MsgType_FundingCreate_Asset_ALlItem_3100:
		node, err := service.FundingTransactionService.OmniFundingAllItem(*client.User)
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
	case enum.MsgType_FundingCreate_Asset_ItemById_3101:
		id, err := strconv.Atoi(msg.Data)
		if err != nil {
			log.Println(err)
			data = err.Error()
		} else {
			node, err := service.FundingTransactionService.OmniFundingItemById(id, *client.User)
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
		}
		client.sendToMyself(msg.Type, status, data)
	case enum.MsgType_FundingCreate_Asset_ItemByChannelId_3102:
		node, err := service.FundingTransactionService.OmniFundingItemByChannelId(msg.Data, *client.User)
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
	case enum.MsgType_FundingCreate_Asset_Count_3103:
		count, err := service.FundingTransactionService.OmniFundingTotalCount(*client.User)
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(count)
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
	}
	return sendType, []byte(data), status
}
func (client *Client) fundingSignModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_FundingSign_SendBtcSign_350:
		node, funder, err := service.FundingTransactionService.FundingBtcTxSigned(msg, client.User)
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

		if tool.CheckIsString(&funder) {
			msg.Type = enum.MsgType_FundingSign_BtcSign_350
			err = client.sendDataToP2PUser(msg, status, data)
			if err != nil {
				data = err.Error()
				status = false
			}
		}
		msg.Type = enum.MsgType_FundingSign_SendBtcSign_350
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_FundingSign_SendAssetFundingSigned_35: //get openChannelReq from funder then send to fundee  create a funding tx
		node, err := service.FundingTransactionService.AssetFundingSigned(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		}

		bytes, err := json.Marshal(node)
		if err != nil {
			data = err.Error()
		}
		if len(data) == 0 {
			data = string(bytes)
			status = true
		}

		if node != nil && status {
			msg.Type = enum.MsgType_FundingSign_AssetFundingSigned_35
			err := client.sendDataToP2PUser(msg, status, data)
			if err != nil {
				data = err.Error()
				status = false
			}
		}
		msg.Type = enum.MsgType_FundingSign_SendAssetFundingSigned_35
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}
