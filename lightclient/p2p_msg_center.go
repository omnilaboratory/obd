package lightclient

import (
	"encoding/json"
	"errors"
	"obd/bean/enum"
	"obd/service"
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
		_, err := service.ChannelService.AfterBobAcceptChannelAtAliceSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingCreate_BtcCreate_N3400:
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
		node, err := service.CommitmentTxService.AfterBobSignCommitmentTranctionAtAliceSide(data, client.User)
		if err == nil {
			status = true
			retAliceData, _ := json.Marshal(node["aliceData"])
			retBobData, _ := json.Marshal(node["bobData"])
			client.sendToMyself(enum.MsgType_CommitmentTxSigned_RevokeAndAcknowledgeCommitmentTransaction_N352, true, string(retAliceData))
			return string(retBobData), nil
		}
		defaultErr = err
	case enum.MsgType_CommitmentTxSigned_SecondToBobSign_N354:
		node, err := service.CommitmentTxSignedService.AfterAliceSignCommitmentTranctionAtBobSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			client.sendToMyself(enum.MsgType_CommitmentTxSigned_RevokeAndAcknowledgeCommitmentTransaction_N352, true, string(retData))
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
