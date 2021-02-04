package lightclient

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/agent"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"github.com/shopspring/decimal"
	"log"
	"strconv"
)

func (client *Client) fundingTransactionModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToSomeone
	var data string
	switch msg.Type {
	case enum.MsgType_Funding_134:
		channelFund(*client, msg)
	case enum.MsgType_FundingCreate_SendBtcFundingCreated_340:
		node, targetUser, err := service.FundingTransactionService.BtcFundingCreated(msg, client.User)
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
			if targetUser != client.User.PeerId {
				msg.Type = enum.MsgType_FundingCreate_BtcFundingCreated_340
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		msg.Type = enum.MsgType_FundingCreate_SendBtcFundingCreated_340
		client.SendToMyself(msg.Type, status, data)

		if status {
			if client.User.IsAdmin {
				signedData, err := agent.AliceFirstSignFundBtcRedeemTx(node, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signedData)
					msg.Data = string(marshal)
					node, _, err := service.FundingTransactionService.OnAliceSignBtcFundingMinerFeeRedeemTx(msg.Data, client.User)
					if err == nil {
						bytes, _ := json.Marshal(node)
						data = string(bytes)
						msg.Type = enum.MsgType_FundingCreate_BtcFundingCreated_340
						err = client.sendDataToP2PUser(msg, status, data)
						if err != nil {
							status = false
							data = err.Error()
						}
					} else {
						status = false
						data = err.Error()
					}
				} else {
					status = false
					data = err.Error()
				}
				if status == false {
					msg.Type = enum.MsgType_ClientSign_Duplex_BtcFundingMinerRDTx_341
					client.SendToMyself(msg.Type, status, data)
				}
			}
		}

	case enum.MsgType_ClientSign_Duplex_BtcFundingMinerRDTx_341:
		node, _, err := service.FundingTransactionService.OnAliceSignBtcFundingMinerFeeRedeemTx(msg.Data, client.User)
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
			err = client.sendDataToP2PUser(msg, status, data)
			if err != nil {
				data = err.Error()
				status = false
			}
		}
		msg.Type = enum.MsgType_ClientSign_Duplex_BtcFundingMinerRDTx_341
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_FundingCreate_Btc_AllItem_3104:
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
		client.SendToMyself(msg.Type, status, data)

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
		client.SendToMyself(msg.Type, status, data)

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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_FundingCreate_Btc_ItemByChannelId_3111:
		node, err := service.FundingTransactionService.BtcFundingItemByChannelId(msg.Data, *client.User)
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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_FundingCreate_Btc_RDAllItem_3107:
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
		client.SendToMyself(msg.Type, status, data)

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
		client.SendToMyself(msg.Type, status, data)

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
		client.SendToMyself(msg.Type, status, data)

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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_FundingCreate_SendAssetFundingCreated_34:

		if client.User.IsAdmin {
			agent.AliceCreateTempWalletForC1a(&msg, client.User)
		}
		node, needSign, err := service.FundingTransactionService.AssetFundingCreated(msg, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				if needSign == false {
					msg.Type = enum.MsgType_FundingCreate_AssetFundingCreated_34
					err = client.sendDataToP2PUser(msg, status, data)
					if err != nil {
						data = err.Error()
						status = false
					}
				}
			}
		}

		if client.User.IsAdmin && needSign {
			signedData, err := agent.AliceSignC1a(node, client.User)
			if err != nil {
				data = err.Error()
				status = false
			}
			marshal, _ := json.Marshal(signedData)
			msg.Data = string(marshal)
			p2pData, err := service.FundingTransactionService.OnAliceSignC1a(msg, client.User)
			if err != nil {
				data = err.Error()
				status = false
			} else {
				msg.Type = enum.MsgType_FundingCreate_AssetFundingCreated_34
				marshal, _ = json.Marshal(p2pData)
				msg.Data = string(marshal)
				err = client.sendDataToP2PUser(msg, status, msg.Data)
				if err != nil {
					data = err.Error()
					status = false
				}
				if status == false {
					client.SendToMyself(msg.Type, status, data)
				}
			}
		} else {
			msg.Type = enum.MsgType_FundingCreate_SendAssetFundingCreated_34
			client.SendToMyself(msg.Type, status, data)
		}

	case enum.MsgType_ClientSign_AssetFunding_AliceSignC1a_1034:
		node, err := service.FundingTransactionService.OnAliceSignC1a(msg, client.User)
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

		msg.Type = enum.MsgType_ClientSign_AssetFunding_AliceSignC1a_1034
		client.SendToMyself(msg.Type, status, data)

		if status {
			msg.Type = enum.MsgType_FundingCreate_AssetFundingCreated_34
			err = client.sendDataToP2PUser(msg, status, data)
			if err != nil {
				data = err.Error()
				msg.Type = enum.MsgType_FundingCreate_AssetFundingCreated_34
				client.SendToMyself(msg.Type, status, data)
			}
		}

	case enum.MsgType_ClientSign_AssetFunding_AliceSignRD_1134:
		node, err := service.FundingTransactionService.OnAliceSignedRdAtAliceSide(msg.Data, client.User)
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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_FundingCreate_Asset_AllItem_3100:
		node, err := service.FundingTransactionService.AssetFundingAllItem(*client.User)
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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_FundingCreate_Asset_ItemById_3101:
		id, err := strconv.Atoi(msg.Data)
		if err != nil {
			log.Println(err)
			data = err.Error()
		} else {
			node, err := service.FundingTransactionService.AssetFundingItemById(id, *client.User)
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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_FundingCreate_Asset_ItemByChannelId_3102:
		node, err := service.FundingTransactionService.AssetFundingItemByChannelId(msg.Data, *client.User)
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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_FundingCreate_Asset_Count_3103:
		count, err := service.FundingTransactionService.AssetFundingTotalCount(*client.User)
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(count)
			status = true
		}
		client.SendToMyself(msg.Type, status, data)

	default:
		sendType = enum.SendTargetType_SendToNone
	}
	return sendType, []byte(data), status
}
func (client *Client) fundingSignModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToSomeone
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

		msg.Type = enum.MsgType_FundingSign_SendBtcSign_350
		client.SendToMyself(msg.Type, status, data)

		if tool.CheckIsString(&funder) {
			if status {
				msg.Type = enum.MsgType_FundingSign_BtcSign_350
			} else {
				// 如果状态出错了，需要把错误信息发给funder:Alice
				msg.Type = enum.MsgType_FundingSign_RecvBtcSign_350
			}
			err = client.sendDataToP2PUser(msg, status, data)
			if err != nil {
				data = err.Error()
				status = false
			}
			if status == false {
				client.SendToMyself(msg.Type, status, data)
			}
		}
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
		msg.Type = enum.MsgType_FundingSign_SendAssetFundingSigned_35
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_ClientSign_AssetFunding_RdAndBr_1035:
		aliceData, bobData, err := service.FundingTransactionService.OnBobSignedRDAndBR(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(aliceData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
			if status {
				msg.Type = enum.MsgType_FundingSign_AssetFundingSigned_35
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}

			if status {
				bytes, err = json.Marshal(bobData)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		msg.Type = enum.MsgType_ClientSign_AssetFunding_RdAndBr_1035
		client.SendToMyself(msg.Type, status, data)
	default:
		sendType = enum.SendTargetType_SendToNone
	}
	return sendType, []byte(data), status
}

func channelFund(client Client, msg bean.RequestMessage) {
	status := true
	data := ""

	channelInfo, err := service.FundingTransactionService.CheckChannelFund(msg, client.User)
	if err != nil {
		status = false
		data = err.Error()
	} else {
		funding := &bean.SendRequestFunding{}
		_ = json.Unmarshal([]byte(msg.Data), funding)
		minerFee := 0.00001
		btcAmount, _ := decimal.NewFromFloat(funding.BtcAmount).Div(decimal.NewFromFloat(3.0)).Sub(decimal.NewFromFloat(minerFee)).Round(8).Float64()
		for i := 0; i < 3; i++ {
			resp, err := omnicore.BtcCreateRawTransaction(channelInfo.FundingAddress, []bean.TransactionOutputItem{{channelInfo.ChannelAddress, btcAmount}}, minerFee, 0, nil)
			if err != nil {
				status = false
				data = err.Error()
			} else {
				sendInfo := &bean.FundingBtc{}
				sendInfo.FromAddress = channelInfo.FundingAddress
				sendInfo.ToAddress = channelInfo.ChannelAddress
				marshal, _ := json.Marshal(sendInfo)
				msg.Data = string(marshal)
				resp, _ = agent.AliceSignFundBtc(msg, resp, client.User)

				msg.Type = enum.MsgType_FundingCreate_SendBtcFundingCreated_340
				fundingBtc := bean.SendRequestFundingBtc{}
				fundingBtc.TemporaryChannelId = channelInfo.TemporaryChannelId
				fundingBtc.FundingTxHex = resp["hex"].(string)
				bytes, _ := json.Marshal(fundingBtc)
				msg.Data = string(bytes)
				client.fundingTransactionModule(msg)
			}
		}

		respNode, err := omnicore.OmniCreateRawTransaction(channelInfo.FundingAddress, channelInfo.ChannelAddress, funding.PropertyId, funding.AssetAmount, minerFee)
		if err != nil {
			data = err.Error()
			status = false
		} else {
			sendInfo := &bean.FundingBtc{}
			sendInfo.FromAddress = channelInfo.FundingAddress
			sendInfo.ToAddress = channelInfo.ChannelAddress
			marshal, _ := json.Marshal(sendInfo)
			msg.Data = string(marshal)
			respNode, _ = agent.AliceSignFundAsset(msg, respNode, client.User)

			msg.Type = enum.MsgType_FundingCreate_SendAssetFundingCreated_34
			fundingBtc := bean.SendRequestFundingBtc{}
			fundingBtc.TemporaryChannelId = channelInfo.TemporaryChannelId
			fundingBtc.FundingTxHex = respNode["hex"].(string)
			bytes, _ := json.Marshal(fundingBtc)
			msg.Data = string(bytes)
			client.fundingTransactionModule(msg)
		}
	}
	if status == false {
		client.SendToMyself(msg.Type, status, data)
	}
}
