package lightclient

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/agent"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/service"
	"github.com/tidwall/gjson"
)

func routerOfP2PNode(msg bean.RequestMessage, data string, client *Client) (retData string, isGoOn bool, retErr error) {
	defaultErr := errors.New("fail to deal msg in the inter node")
	status := false
	msgType := msg.Type
	switch msgType {
	case enum.MsgType_ChannelOpen_32:
		err := service.ChannelService.BeforeBobOpenChannelAtBobSide(data, client.User)
		if err == nil {
			status = true
			// when bob get the request for open channel
			if client.User.IsAdmin {
				msg.Data = data
				acceptOpenChannelInfo, err := agent.BeforeBobAcceptOpenChannel(&msg, client.User)
				if err == nil {
					signOpenChannelMsg := bean.RequestMessage{}
					marshal, _ := json.Marshal(acceptOpenChannelInfo)
					signOpenChannelMsg.Type = enum.MsgType_SendChannelAccept_33
					signOpenChannelMsg.RecipientNodePeerId = msg.SenderNodePeerId
					signOpenChannelMsg.RecipientUserPeerId = msg.SenderUserPeerId
					signOpenChannelMsg.Data = string(marshal)
					signedData, err := service.ChannelService.BobAcceptChannel(signOpenChannelMsg, client.User)
					if err == nil {
						signedOpenChannelMsg := bean.RequestMessage{}
						signedOpenChannelMsg.Type = enum.MsgType_ChannelAccept_33
						signedOpenChannelMsg.RecipientNodePeerId = signOpenChannelMsg.RecipientNodePeerId
						signedOpenChannelMsg.RecipientUserPeerId = signOpenChannelMsg.RecipientUserPeerId
						signedOpenChannelMsg.SenderNodePeerId = client.User.P2PLocalPeerId
						signedOpenChannelMsg.SenderUserPeerId = client.User.PeerId
						marshal, _ = json.Marshal(signedData)
						signedOpenChannelMsg.Data = string(marshal)
						_ = client.sendDataToP2PUser(signedOpenChannelMsg, true, signedOpenChannelMsg.Data)
						return "", false, nil
					}
				}
			}
		} else {
			defaultErr = err
		}
	case enum.MsgType_ChannelAccept_33:
		node, err := service.ChannelService.AfterBobAcceptChannelAtAliceSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingCreate_BtcFundingCreated_340:
		_, err := service.FundingTransactionService.BeforeSignBtcFundingCreatedAtBobSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingSign_BtcSign_350:
		_, err := service.FundingTransactionService.AfterBobSignBtcFundingAtAliceSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingCreate_AssetFundingCreated_34:
		_, err := service.FundingTransactionService.BeforeSignAssetFundingCreateAtBobSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingSign_AssetFundingSigned_35:
		node, err := service.FundingTransactionService.OnGetBobSignedMsgAndSendDataToAlice(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		} else {
			defaultErr = err
		}
	case enum.MsgType_CommitmentTx_CommitmentTransactionCreated_351:
		node, err := service.CommitmentTxSignedService.BeforeBobSignCommitmentTransactionAtBobSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_CommitmentTxSigned_ToAliceSign_352:
		node, needNoticeAlice, err := service.CommitmentTxService.OnGetBobC2bPartialSignTxAtAliceSide(msg, data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		} else {
			if needNoticeAlice {
				client.SendToMyself(enum.MsgType_CommitmentTxSigned_RecvRevokeAndAcknowledgeCommitmentTransaction_352, true, string(err.Error()))
			}
		}
		defaultErr = err
	case enum.MsgType_CommitmentTxSigned_SecondToBobSign_353:
		node, err := service.CommitmentTxSignedService.OnGetAliceSignC2bTransactionAtBobSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_CloseChannelRequest_38:
		node, err := service.ChannelService.BeforeBobSignCloseChannelAtBobSide(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_CloseChannelSign_39:
		node, err := service.ChannelService.AfterBobSignCloseChannelAtAliceSide(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_AddHTLC_40:
		node, err := service.HtlcForwardTxService.BeforeBobSignAddHtlcRequestAtBobSide_40(data, *client.User)
		if client.User.IsAdmin {
			//TODO 代理模式的自动化模式
			signedData, err := agent.BobSignAddHtlcRequestAtBobSide_40(*node, client.User)
			if err == nil {
				marshal, _ := json.Marshal(signedData)
				service.HtlcForwardTxService.BobSignedAddHtlcAtBobSide(string(marshal), *client.User)
				//TODO bob签名C3b
				agent.BobSignC3b()
			}
		}
		if err == nil {
			retData, _ := json.Marshal(node)
			status = true
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_NeedPayerSignC3b_41:
		node, _, err := service.HtlcForwardTxService.AfterBobSignAddHtlcAtAliceSide_41(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayeeCreateHTRD1a_42:
		node, err := service.HtlcForwardTxService.OnGetNeedBobSignC3bSubTxAtBobSide(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayerSignHTRD1a_43:
		node, err := service.HtlcForwardTxService.OnGetHtrdTxDataFromBobAtAliceSide_43(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_VerifyR_45:
		responseData, err := service.HtlcBackwardTxService.OnGetHeSubTxDataAtAliceObdAtAliceSide(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_SendHerdHex_46:
		responseData, err := service.HtlcBackwardTxService.OnGetHeRdDataAtBobObd(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_Close_RequestCloseCurrTx_49:
		responseData, err := service.HtlcCloseTxService.OnObdOfBobGet49PData(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_CloseHtlcRequestSignBR_50:
		responseData, _, err := service.HtlcCloseTxService.OnObdOfAliceGet50PData(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_CloseHtlcUpdateCnb_51:
		node, err := service.HtlcCloseTxService.OnObdOfBobGet51PData(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_Atomic_Swap_80:
		node, err := service.AtomicSwapService.BeforeSignAtomicSwapAtBobSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_Atomic_SwapAccept_81:
		node, err := service.AtomicSwapService.BeforeSignAtomicSwapAcceptedAtAliceSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	default:
		status = true
	}
	if status {
		defaultErr = nil
	}
	return "", true, defaultErr
}

func p2pMiddleNodeTransferData(msg *bean.RequestMessage, itemClient Client, data string, retData string) string {
	if msg.Type == enum.MsgType_ChannelOpen_32 {
		msg.Type = enum.MsgType_RecvChannelOpen_32
	}

	if msg.Type == enum.MsgType_ChannelAccept_33 {
		msg.Type = enum.MsgType_RecvChannelAccept_33
	}

	if msg.Type == enum.MsgType_FundingCreate_AssetFundingCreated_34 {
		msg.Type = enum.MsgType_FundingCreate_RecvAssetFundingCreated_34
	}

	if msg.Type == enum.MsgType_FundingSign_AssetFundingSigned_35 {
		msg.Type = enum.MsgType_FundingSign_RecvAssetFundingSigned_35
	}

	if msg.Type == enum.MsgType_FundingCreate_BtcFundingCreated_340 {
		msg.Type = enum.MsgType_FundingCreate_RecvBtcFundingCreated_340
	}

	if msg.Type == enum.MsgType_FundingSign_BtcSign_350 {
		msg.Type = enum.MsgType_FundingSign_RecvBtcSign_350
	}

	if msg.Type == enum.MsgType_CommitmentTx_CommitmentTransactionCreated_351 {
		msg.Type = enum.MsgType_CommitmentTx_RecvCommitmentTransactionCreated_351
	}

	if msg.Type == enum.MsgType_CloseChannelRequest_38 {
		msg.Type = enum.MsgType_RecvCloseChannelRequest_38
	}

	if msg.Type == enum.MsgType_CloseChannelSign_39 {
		msg.Type = enum.MsgType_RecvCloseChannelSign_39
	}

	if msg.Type == enum.MsgType_HTLC_AddHTLC_40 {
		msg.Type = enum.MsgType_HTLC_RecvAddHTLC_40
	}

	if msg.Type == enum.MsgType_CommitmentTxSigned_ToAliceSign_352 {
		//发给alice
		msg.Type = enum.MsgType_CommitmentTxSigned_RecvRevokeAndAcknowledgeCommitmentTransaction_352
		payerData := gjson.Parse(retData).String()
		data = payerData
	}

	//当353处理完成，就改成110353 推送给bob的客户端
	if msg.Type == enum.MsgType_CommitmentTxSigned_SecondToBobSign_353 {
		msg.Type = enum.MsgType_ClientSign_BobC2b_Rd_353
	}

	if msg.Type == enum.MsgType_HTLC_NeedPayerSignC3b_41 {
		msg.Type = enum.MsgType_HTLC_RecvAddHTLCSigned_41
	}

	//当42处理完成
	if msg.Type == enum.MsgType_HTLC_PayeeCreateHTRD1a_42 {
		msg.Type = enum.MsgType_HTLC_BobSignC3bSubTx_42
	}

	if msg.Type == enum.MsgType_HTLC_PayerSignHTRD1a_43 {
		msg.Type = enum.MsgType_HTLC_FinishTransferH_43
	}

	if msg.Type == enum.MsgType_HTLC_VerifyR_45 {
		msg.Type = enum.MsgType_HTLC_RecvVerifyR_45
	}

	//当47处理完成，发送48号协议给收款方
	if msg.Type == enum.MsgType_HTLC_SendHerdHex_46 {
		msg.Type = enum.MsgType_HTLC_RecvSignVerifyR_46
	}

	if msg.Type == enum.MsgType_HTLC_Close_RequestCloseCurrTx_49 {
		msg.Type = enum.MsgType_HTLC_Close_RecvRequestCloseCurrTx_49
	}

	if msg.Type == enum.MsgType_HTLC_CloseHtlcRequestSignBR_50 {
		msg.Type = enum.MsgType_HTLC_RecvCloseSigned_50
	}

	if msg.Type == enum.MsgType_HTLC_CloseHtlcUpdateCnb_51 {
		msg.Type = enum.MsgType_HTLC_Close_ClientSign_Bob_C4bSub_51
	}

	if msg.Type == enum.MsgType_Atomic_Swap_80 {
		msg.Type = enum.MsgType_Atomic_RecvSwap_80
	}

	if msg.Type == enum.MsgType_Atomic_SwapAccept_81 {
		msg.Type = enum.MsgType_Atomic_RecvSwapAccept_81
	}

	return data
}
