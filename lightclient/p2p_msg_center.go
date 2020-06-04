package lightclient

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/service"
)

func routerOfP2PNode(msgType enum.MsgType, data string, client *Client) (retData string, retErr error) {
	defaultErr := errors.New("fail to deal msg in the inter node")
	status := false
	switch msgType {
	case enum.MsgType_ChannelOpen_N32:
		err := service.ChannelService.BeforeBobOpenChannelAtBobSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_ChannelAccept_N33:
		node, err := service.ChannelService.AfterBobAcceptChannelAtAliceSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingCreate_BtcFundingCreated_N3400:
		_, err := service.FundingTransactionService.BeforeBobSignBtcFundingAtBobSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingSign_BtcSign_N3500:
		_, err := service.FundingTransactionService.AfterBobSignBtcFundingAtAliceSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingCreate_AssetFundingCreated_N34:
		_, err := service.FundingTransactionService.BeforeBobSignOmniFundingAtBobSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingSign_AssetFundingSigned_N35:
		node, err := service.FundingTransactionService.AfterBobSignOmniFundingAtAilceSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		} else {
			defaultErr = err
		}
	case enum.MsgType_CommitmentTx_CommitmentTransactionCreated_N351:
		node, err := service.CommitmentTxSignedService.BeforeBobSignCommitmentTranctionAtBobSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_CommitmentTxSigned_ToAliceSign_N353:
		node, _, err := service.CommitmentTxService.AfterBobSignCommitmentTranctionAtAliceSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_CommitmentTxSigned_SecondToBobSign_N354:
		node, err := service.CommitmentTxSignedService.AfterAliceSignCommitmentTranctionAtBobSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_CloseChannelRequest_N38:
		node, err := service.ChannelService.BeforeBobSignCloseChannelAtBobSide(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_CloseChannelSign_N39:
		node, err := service.ChannelService.AfterBobSignCloseChannelAtAliceSide(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_AddHTLC_N40:
		node, err := service.HtlcForwardTxService.BeforeBobSignPayerAddHtlcRequestAtBobSide_40(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			status = true
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayerSignC3b_N42:
		node, _, err := service.HtlcForwardTxService.AfterBobSignAddHtlcAtAliceSide_42(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayeeCreateHTRD1a_N43:
		node, _, err := service.HtlcForwardTxService.AfterAliceSignAddHtlcAtBobSide_43(data, *client.User)
		if err == nil {
			//给收款方的信息用sendToMyself发送
			tempData, _ := json.Marshal(node["bobData"])
			client.sendToMyself(enum.MsgType_HTLC_AddHTLCSigned_N41, true, string(tempData))
			retData, _ := json.Marshal(node["aliceData"])
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayerSignHTRD1a_N44:
		node, _, err := service.HtlcForwardTxService.AfterBobCreateHTRDAtAliceSide_44(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_SendR_N45:
		responseData, _, err := service.HtlcBackwardTxService.BeforeSendRInfoToPayerAtAliceSide_Step2(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_SendHerdHex_N47:
		responseData, err := service.HtlcBackwardTxService.SignHed1aAndUpdate_Step4(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_SignHedHex_N48:
		responseData, err := service.HtlcBackwardTxService.CheckHed1aHex_Step5(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_RequestCloseCurrTx_N49:
		responseData, err := service.HtlcCloseTxService.BeforeBobSignCloseHtlcAtBobSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_CloseHtlcRequestSignBR_N51:
		responseData, _, err := service.HtlcCloseTxService.AfterBobCloseHTLCSigned_AtAliceSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(responseData)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_CloseHtlcUpdateCnb_N52:
		node, err := service.HtlcCloseTxService.AfterAliceSignCloseHTLCAtBobSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_Atomic_Swap_N80:
		node, err := service.AtomicSwapService.BeforeSignAtomicSwapAtBobSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), nil
		}
		defaultErr = err
	case enum.MsgType_Atomic_Swap_Accept_N81:
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
