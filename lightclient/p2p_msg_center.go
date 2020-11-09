package lightclient

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/service"
)

func routerOfP2PNode(msg bean.RequestMessage, data string, client *Client) (retData string, retErr error) {
	defaultErr := errors.New("fail to deal msg in the inter node")
	status := false
	msgType := msg.Type
	switch msgType {
	case enum.MsgType_ChannelOpen_32:
		err := service.ChannelService.BeforeBobOpenChannelAtBobSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_ChannelAccept_33:
		node, err := service.ChannelService.AfterBobAcceptChannelAtAliceSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
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
			return string(retData), nil
		} else {
			defaultErr = err
		}
	case enum.MsgType_CommitmentTx_CommitmentTransactionCreated_351:
		node, err := service.CommitmentTxSignedService.BeforeBobSignCommitmentTransactionAtBobSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_CommitmentTxSigned_ToAliceSign_352:
		node, needNoticeAlice, err := service.CommitmentTxService.OnGetBobC2bPartialSignTxAtAliceSide(msg, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		} else {
			if needNoticeAlice {
				client.sendToMyself(enum.MsgType_CommitmentTxSigned_RecvRevokeAndAcknowledgeCommitmentTransaction_352, true, string(err.Error()))
			}
		}
		defaultErr = err
	case enum.MsgType_CommitmentTxSigned_SecondToBobSign_353:
		node, err := service.CommitmentTxSignedService.OnGetAliceSignC2bTransactionAtBobSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_CloseChannelRequest_38:
		node, err := service.ChannelService.BeforeBobSignCloseChannelAtBobSide(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_CloseChannelSign_39:
		node, err := service.ChannelService.AfterBobSignCloseChannelAtAliceSide(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_AddHTLC_40:
		node, err := service.HtlcForwardTxService.BeforeBobSignPayerAddHtlcRequestAtBobSide_40(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			status = true
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayerSignC3b_42:
		node, _, err := service.HtlcForwardTxService.AfterBobSignAddHtlcAtAliceSide_42(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayeeCreateHTRD1a_43:
		node, _, err := service.HtlcForwardTxService.AfterAliceSignAddHtlcAtBobSide_43(data, *client.User)
		if err == nil {
			//给收款方的信息用sendToMyself发送
			tempData, _ := json.Marshal(node["bobData"])
			client.sendToMyself(enum.MsgType_HTLC_SendAddHTLCSigned_41, true, string(tempData))
			retData, _ := json.Marshal(node["aliceData"])
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayerSignHTRD1a_44:
		node, _, err := service.HtlcForwardTxService.AfterBobCreateHTRDAtAliceSide_44(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_VerifyR_45:
		responseData, _, err := service.HtlcBackwardTxService.BeforeSendRInfoToPayerAtAliceSide_Step2(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_SendHerdHex_47:
		responseData, err := service.HtlcBackwardTxService.SignHed1aAndUpdate_Step4(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_SignHedHex_48:
		responseData, err := service.HtlcBackwardTxService.CheckHed1aHex_Step5(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_RequestCloseCurrTx_49:
		responseData, err := service.HtlcCloseTxService.BeforeBobSignCloseHtlcAtBobSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_CloseHtlcRequestSignBR_51:
		responseData, _, err := service.HtlcCloseTxService.AfterBobCloseHTLCSigned_AtAliceSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_CloseHtlcUpdateCnb_52:
		node, err := service.HtlcCloseTxService.AfterAliceSignCloseHTLCAtBobSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_Atomic_Swap_80:
		node, err := service.AtomicSwapService.BeforeSignAtomicSwapAtBobSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_Atomic_SwapAccept_81:
		node, err := service.AtomicSwapService.BeforeSignAtomicSwapAcceptedAtAliceSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	default:
		status = true
	}
	if status {
		defaultErr = nil
	}
	return "", defaultErr
}
