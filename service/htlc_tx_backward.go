package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
	trackerBean "github.com/omnilaboratory/obd/tracker/bean"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type htlcBackwardTxManager struct {
	operationFlag              sync.Mutex
	tempDataSendTo45PAtBobSide map[string]bean.NeedAliceSignHerdTxOfC3bP2p
	tempDataFrom45PAtAliceSide map[string]bean.NeedAliceSignHerdTxOfC3bP2p
}

// HTLC Reverse pass the R (Preimage R)
var HtlcBackwardTxService htlcBackwardTxManager

// step 1 bob -100045 收款方收到R，发送R到obd进行验证
func (service *htlcBackwardTxManager) SendRToPreviousNodeAtBobSide(msg bean.RequestMessage, user bean.User) (retData interface{}, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New(enum.Tips_common_empty + "msg data")
	}

	reqData := &bean.HtlcBobSendR{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	// region Check data inputed from websocket client of sender.
	if tool.CheckIsString(&reqData.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id ")
		log.Println(err)
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).First(channelInfo)
	if err != nil {
		err = errors.New(enum.Tips_common_notFound + "channelInfo by " + reqData.ChannelId)
		log.Println(err)
		return nil, err
	}

	payerPeerId := channelInfo.PeerIdA
	payeeChannelAddress := channelInfo.AddressB
	payerChannelAddress := channelInfo.AddressA
	payerChannelPubkey := channelInfo.PubKeyA
	if user.PeerId == channelInfo.PeerIdA {
		payeeChannelAddress = channelInfo.AddressA
		payerChannelAddress = channelInfo.AddressB
		payerChannelPubkey = channelInfo.PubKeyB
		payerPeerId = channelInfo.PeerIdB
	}

	if payerPeerId != msg.RecipientUserPeerId {
		return nil, errors.New(enum.Tips_rsmc_notTargetUser)
	}

	err = findUserIsOnline(msg.RecipientNodePeerId, payerPeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.R) == false {
		err = errors.New(enum.Tips_common_empty + "r")
		log.Println(err)
		return nil, err
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, reqData.ChannelId, user.PeerId)
	if err != nil {
		err = errors.New(enum.Tips_common_notFound + "latestCommitmentTxInfo")
		log.Println(err)
		return nil, err
	}
	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Htlc_GetH {
		err = errors.New(enum.Tips_rsmc_errorCommitmentTxState + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, err
	}

	if latestCommitmentTxInfo.HtlcSender != msg.RecipientUserPeerId {
		err = errors.New(enum.Tips_htlc_notTheHltcSender)
		log.Println(err)
		return nil, err
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.R, latestCommitmentTxInfo.HtlcH)
	if err != nil {
		return nil, errors.New(enum.Tips_htlc_wrongRForH)
	}

	latestCommitmentTxInfo.HtlcR = reqData.R

	// endregion

	currBlockHeight, err := rpcClient.GetBlockCount()
	if err != nil {
		return nil, errors.New(enum.Tips_htlc_failToGetBlockHeight)
	}

	htlcTimeOut := latestCommitmentTxInfo.HtlcCltvExpiry
	maxHeight := latestCommitmentTxInfo.BeginBlockHeight + htlcTimeOut
	if strings.Contains(config.ChainNode_Type, "main") {
		if currBlockHeight > maxHeight {
			return nil, errors.New(enum.Tips_htlc_timeOut)
		}
	}

	he1b, err := signHe1bAtPayeeSide_at45(tx, *channelInfo, latestCommitmentTxInfo.Id, *reqData, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	dataNeedBobSign := bean.HtlcBobSendRResult{}
	dataSendTo45P := bean.NeedAliceSignHerdTxOfC3bP2p{}
	dataSendTo45P.ChannelId = channelInfo.ChannelId
	dataNeedBobSign.ChannelId = channelInfo.ChannelId
	dataSendTo45P.R = reqData.R
	dataSendTo45P.PayeePeerId = msg.SenderUserPeerId
	dataSendTo45P.PayeeNodeAddress = msg.SenderNodePeerId

	//region 1  he的rd
	heHex := he1b.RSMCTxHex
	heOutputs, err := getInputsForNextTxByParseTxHashVout(heHex, he1b.RSMCMultiAddress, he1b.RSMCMultiAddressScriptPubKey, he1b.RSMCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	heRdTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		he1b.RSMCMultiAddress,
		heOutputs,
		payeeChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		latestCommitmentTxInfo.AmountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		1000,
		&he1b.RSMCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create HTD1b for C3b")
	}

	c3eHeRdRawData := bean.NeedClientSignTxData{}
	c3eHeRdRawData.Hex = heRdTx["hex"].(string)
	c3eHeRdRawData.Inputs = heRdTx["inputs"]
	c3eHeRdRawData.IsMultisig = true
	c3eHeRdRawData.PubKeyA = he1b.RSMCTempAddressPubKey
	c3eHeRdRawData.PubKeyB = payerChannelPubkey

	_, err = createHerd1bAtPayeeSide(tx, *channelInfo, he1b, c3eHeRdRawData.Hex, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	dataSendTo45P.C3bHtlcHerdPartialSignedData = c3eHeRdRawData
	dataSendTo45P.HeCompleteSignedHex = heHex
	dataSendTo45P.C3bHtlcTempAddressForHePubKey = he1b.RSMCTempAddressPubKey
	dataNeedBobSign.C3bHtlcHerdRawData = c3eHeRdRawData
	//endregion

	//region 2  he的br
	heBrTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		he1b.RSMCMultiAddress,
		heOutputs,
		payerChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		latestCommitmentTxInfo.AmountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		0,
		&he1b.RSMCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create HTD1b for C3b")
	}

	c3bHeBrRawData := bean.NeedClientSignTxData{}
	c3bHeBrRawData.Hex = heBrTx["hex"].(string)
	c3bHeBrRawData.Inputs = heBrTx["inputs"]
	c3bHeBrRawData.IsMultisig = true
	c3bHeBrRawData.PubKeyA = he1b.RSMCTempAddressPubKey
	c3bHeBrRawData.PubKeyB = payerChannelPubkey
	dataSendTo45P.C3bHtlcHebrRawData = c3bHeBrRawData

	_ = tx.Update(latestCommitmentTxInfo)
	_ = tx.Commit()

	if service.tempDataSendTo45PAtBobSide == nil {
		service.tempDataSendTo45PAtBobSide = make(map[string]bean.NeedAliceSignHerdTxOfC3bP2p)
	}
	service.tempDataSendTo45PAtBobSide[user.PeerId+"_"+channelInfo.ChannelId] = dataSendTo45P
	return dataNeedBobSign, nil
}

// step 2 bob -100106 bob签名HeRd的结果 推送45号协议
func (service *htlcBackwardTxManager) OnBobSignedHeRdAtBobSide(msg bean.RequestMessage, user bean.User) (retData interface{}, err error) {
	bobSignedData := bean.BobSignHerdForC3b{}
	_ = json.Unmarshal([]byte(msg.Data), &bobSignedData)

	if tool.CheckIsString(&bobSignedData.ChannelId) == false {
		return nil, errors.New("error channel_id")
	}
	dataSendTo45P := service.tempDataSendTo45PAtBobSide[user.PeerId+"_"+bobSignedData.ChannelId]
	if len(dataSendTo45P.ChannelId) == 0 {
		return nil, errors.New("error channel_id")
	}

	if pass, _ := rpcClient.CheckMultiSign(false, bobSignedData.C3bHtlcHerdPartialSignedHex, 1); pass == false {
		return nil, errors.New("error sign c3b_htlc_herd_partial_signed_hex")
	}
	dataSendTo45P.C3bHtlcHerdPartialSignedData.Hex = bobSignedData.C3bHtlcHerdPartialSignedHex
	service.tempDataSendTo45PAtBobSide[user.PeerId+"_"+bobSignedData.ChannelId] = dataSendTo45P
	return dataSendTo45P, nil
}

// step 3 alice p2p -45 推送待签名的herd
func (service *htlcBackwardTxManager) OnGetHeSubTxDataAtAliceObdAtAliceSide(msg string, user bean.User) (retData interface{}, err error) {
	dataFrom45P := bean.NeedAliceSignHerdTxOfC3bP2p{}
	_ = json.Unmarshal([]byte(msg), &dataFrom45P)

	if tool.CheckIsString(&dataFrom45P.ChannelId) == false {
		return nil, errors.New("error channel_id")
	}

	if tool.CheckIsString(&dataFrom45P.R) == false {
		return nil, errors.New(enum.Tips_common_empty + "R")
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	latestCommitmentTx, err := getLatestCommitmentTxUseDbTx(tx, dataFrom45P.ChannelId, user.PeerId)
	if err != nil {
		return nil, err
	}
	if _, err = tool.GetPubKeyFromWifAndCheck(dataFrom45P.R, latestCommitmentTx.HtlcH); err != nil {
		return nil, errors.New(enum.Tips_htlc_wrongRForH)
	}
	tx.Commit()

	if tool.CheckIsString(&dataFrom45P.C3bHtlcTempAddressForHePubKey) == false {
		return nil, errors.New(enum.Tips_common_empty + "c3b_htlc_temp_address_for_he_pub_key")
	}

	if tool.CheckIsString(&dataFrom45P.HeCompleteSignedHex) == false {
		return nil, errors.New(enum.Tips_common_empty + "he_complete_signed_hex")
	}

	if service.tempDataFrom45PAtAliceSide == nil {
		service.tempDataFrom45PAtAliceSide = make(map[string]bean.NeedAliceSignHerdTxOfC3bP2p)
	}
	service.tempDataFrom45PAtAliceSide[user.PeerId+"_"+dataFrom45P.ChannelId] = dataFrom45P
	return dataFrom45P, nil
}

// step 4 alice  -46 Alice完成herd签名 保存hebr 推送46 herd
func (service *htlcBackwardTxManager) OnAliceSignedHeRdAtAliceSide(msg bean.RequestMessage, user bean.User) (toAlice, toBob interface{}, err error) {
	herdSignedResult := bean.AliceSignHerdTxOfC3e{}
	_ = json.Unmarshal([]byte(msg.Data), herdSignedResult)

	if tool.CheckIsString(&herdSignedResult.ChannelId) == false {
		return nil, nil, errors.New("error channel_id")
	}

	dataFrom45P := service.tempDataFrom45PAtAliceSide[user.PeerId+"_"+herdSignedResult.ChannelId]
	if len(dataFrom45P.ChannelId) == 0 {
		return nil, nil, errors.New("error channel_id")
	}

	if pass, _ := rpcClient.CheckMultiSign(false, herdSignedResult.C3bHtlcHerdCompleteSignedHex, 2); pass == false {
		return nil, nil, errors.New("error sign c3b_htlc_herd_complete_signed_hex")
	}
	dataFrom45P.C3bHtlcHerdPartialSignedData.Hex = herdSignedResult.C3bHtlcHerdCompleteSignedHex

	if pass, _ := rpcClient.CheckMultiSign(false, herdSignedResult.C3bHtlcHebrPartialSignedHex, 1); pass == false {
		return nil, nil, errors.New("error sign c3b_htlc_hebr_partial_signed_hex")
	}
	dataFrom45P.C3bHtlcHebrRawData.Hex = herdSignedResult.C3bHtlcHebrPartialSignedHex

	dataTo46P := bean.AliceSignedHerdTxOfC3bP2p{}
	dataTo46P.ChannelId = herdSignedResult.ChannelId
	dataTo46P.C3bHtlcHerdCompleteSignedHex = herdSignedResult.C3bHtlcHerdCompleteSignedHex

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	// region Check data inputed from websocket client of sender.
	if tool.CheckIsString(&herdSignedResult.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id ")
		log.Println(err)
		return nil, nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", herdSignedResult.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).First(channelInfo)
	if err != nil {
		err = errors.New(enum.Tips_common_notFound + "channelInfo by " + herdSignedResult.ChannelId)
		log.Println(err)
		return nil, nil, err
	}
	latestCommitment, err := getLatestCommitmentTxUseDbTx(tx, herdSignedResult.ChannelId, user.PeerId)

	//签名Hed
	latestCommitment.HtlcR = dataFrom45P.R
	err = signHed1a(tx, *channelInfo, *latestCommitment, user)
	if err != nil {
		return nil, nil, err
	}

	aliceChannelPubKey := channelInfo.PubKeyA
	aliceChannelAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		aliceChannelPubKey = channelInfo.PubKeyB
		aliceChannelAddress = channelInfo.AddressB
	}

	he1bMultiAddress, he1bRedeemScript, he1bScriptPubKey, err := createMultiSig(dataFrom45P.C3bHtlcTempAddressForHePubKey, aliceChannelPubKey)
	if err != nil {
		return nil, nil, err
	}

	he1bOutputs, err := getInputsForNextTxByParseTxHashVout(dataFrom45P.HeCompleteSignedHex, he1bMultiAddress, he1bScriptPubKey, he1bRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	tempOtherSideCommitmentTx := &dao.CommitmentTransaction{}
	tempOtherSideCommitmentTx.Id = latestCommitment.Id
	tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
	tempOtherSideCommitmentTx.RSMCTempAddressPubKey = dataFrom45P.C3bHtlcTempAddressForHePubKey
	tempOtherSideCommitmentTx.RSMCMultiAddress = he1bMultiAddress
	tempOtherSideCommitmentTx.RSMCRedeemScript = he1bRedeemScript
	tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = he1bScriptPubKey
	tempOtherSideCommitmentTx.RSMCTxHex = dataFrom45P.HeCompleteSignedHex
	tempOtherSideCommitmentTx.RSMCTxid = he1bOutputs[0].Txid
	tempOtherSideCommitmentTx.AmountToRSMC = latestCommitment.AmountToHtlc
	err = createCurrCommitmentTxPartialSignedBR(tx, dao.BRType_HE1b, channelInfo, tempOtherSideCommitmentTx, he1bOutputs, aliceChannelAddress, herdSignedResult.C3bHtlcHebrPartialSignedHex, user)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	latestCommitment.CurrState = dao.TxInfoState_Htlc_GetR
	tx.Update(latestCommitment)

	tx.Commit()
	return latestCommitment, dataTo46P, nil
}

// step 5 bob  -110046 bob保存herd
func (service *htlcBackwardTxManager) OnGetHeRdDataAtBobObd(msg string, user bean.User) (retData interface{}, err error) {
	dataFrom46P := bean.AliceSignedHerdTxOfC3bP2p{}
	_ = json.Unmarshal([]byte(msg), dataFrom46P)

	if tool.CheckIsString(&dataFrom46P.ChannelId) == false {
		return nil, errors.New("error channel_id")
	}

	if pass, _ := rpcClient.CheckMultiSign(false, dataFrom46P.C3bHtlcHerdCompleteSignedHex, 2); pass == false {
		return nil, errors.New("error sign c3b_htlc_herd_complete_signed_hex")
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", dataFrom46P.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).First(&channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	latestCommitment, err := getLatestCommitmentTxUseDbTx(tx, dataFrom46P.ChannelId, user.PeerId)

	_, _ = updateHerd1bAtPayeeSide(tx, channelInfo, latestCommitment.Id, dataFrom46P.C3bHtlcHerdCompleteSignedHex, user)

	latestCommitment.CurrState = dao.TxInfoState_Htlc_GetR
	_ = tx.Update(latestCommitment)
	//endregion
	_ = tx.Commit()

	if channelInfo.IsPrivate == false {
		//update htlc state on tracker
		txStateRequest := trackerBean.UpdateHtlcTxStateRequest{}
		txStateRequest.Path = latestCommitment.HtlcRoutingPacket
		txStateRequest.H = latestCommitment.HtlcH
		if strings.HasSuffix(latestCommitment.HtlcRoutingPacket, channelInfo.ChannelId) {
			txStateRequest.R = latestCommitment.HtlcR
		}
		txStateRequest.DirectionFlag = trackerBean.HtlcTxState_ConfirmPayMoney
		txStateRequest.CurrChannelId = channelInfo.ChannelId
		sendMsgToTracker(enum.MsgType_Tracker_UpdateHtlcTxState_352, txStateRequest)
	}
	return latestCommitment, nil
}

// -45 at payer side
//func (service *htlcBackwardTxManager) BeforeSendRInfoToPayerAtAliceSide_Step2(msgData string, user bean.User) (data *bean.BobSendROfWs, needNoticeOtherSide bool, err error) {
//	bobSendRInfo := &bean.BobSendROfP2p{}
//	_ = json.Unmarshal([]byte(msgData), bobSendRInfo)
//	channelId := bobSendRInfo.ChannelId
//
//	tx, err := user.Db.Begin(true)
//	if err != nil {
//		log.Println(err)
//		return nil, false, err
//	}
//	defer tx.Rollback()
//
//	channelInfo := &dao.ChannelInfo{}
//	err = tx.Select(
//		q.Eq("ChannelId", channelId),
//		q.Or(
//			q.Eq("PeerIdA", user.PeerId),
//			q.Eq("PeerIdB", user.PeerId))).
//		First(channelInfo)
//	if channelInfo == nil {
//		return nil, true, errors.New(enum.Tips_funding_notFoundChannelByChannelId + channelId)
//	}
//
//	senderPeerId := channelInfo.PeerIdB
//	if user.PeerId == channelInfo.PeerIdB {
//		senderPeerId = channelInfo.PeerIdA
//	}
//	messageHash := messageService.saveMsgUseTx(tx, senderPeerId, user.PeerId, msgData)
//	returnData := &bean.BobSendROfWs{}
//	_ = tx.Commit()
//
//	returnData.BobSendROfP2p = *bobSendRInfo
//	returnData.MsgHash = messageHash
//	return returnData, false, nil
//}

// -46 at Payer side
//func (service *htlcBackwardTxManager) VerifyRAndCreateTxs_Step3(msg bean.RequestMessage, user bean.User) (responseData *bean.HtlcRPayerVerifyRInfoToPayee, err error) {
//	if tool.CheckIsString(&msg.Data) == false {
//		return nil, errors.New("empty json responseData")
//	}
//	reqData := &bean.HtlcCheckRAndCreateTx{}
//	err = json.Unmarshal([]byte(msg.Data), reqData)
//	if err != nil {
//		log.Println(err.Error())
//		return nil, err
//	}
//
//	if tool.CheckIsString(&reqData.MsgHash) == false {
//		err = errors.New("empty msg_hash")
//		log.Println(err)
//		return nil, err
//	}
//
//	tx, err := user.Db.Begin(true)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	defer tx.Rollback()
//
//	message, err := messageService.getMsgUseTx(tx, reqData.MsgHash)
//	if err != nil {
//		return nil, errors.New("wrong msg_hash")
//	}
//	if message.Receiver != user.PeerId {
//		return nil, errors.New("you are not the operator")
//	}
//	senderSendROfP2p := &bean.BobSendROfP2p{}
//	_ = json.Unmarshal([]byte(message.Data), senderSendROfP2p)
//
//	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
//		err = errors.New("channel_address_private_key is empty")
//		log.Println(err)
//		return nil, err
//	}
//	channelInfo := &dao.ChannelInfo{}
//	err = tx.Select(
//		q.Eq("ChannelId", reqData.ChannelId),
//		q.Or(
//			q.Eq("PeerIdA", user.PeerId),
//			q.Eq("PeerIdB", user.PeerId))).First(channelInfo)
//	if err != nil {
//		err = errors.New("not found the channel " + reqData.ChannelId)
//		log.Println(err)
//		return nil, err
//	}
//
//	payerChannelPubKey := channelInfo.PubKeyA
//	payerChannelAddress := channelInfo.AddressA
//	payeePeerId := channelInfo.PeerIdB
//	if user.PeerId == channelInfo.PeerIdB {
//		payerChannelPubKey = channelInfo.PubKeyB
//		payerChannelAddress = channelInfo.AddressB
//		payeePeerId = channelInfo.PeerIdA
//	}
//
//	if msg.RecipientUserPeerId != payeePeerId {
//		return nil, errors.New("error recipient_user_peer_id")
//	}
//
//	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, payerChannelPubKey)
//	if err != nil {
//		err = errors.New("channel_address_private_key is wrong for " + payerChannelPubKey)
//		log.Println(err)
//		return nil, err
//	}
//	tempAddrPrivateKeyMap[payerChannelPubKey] = reqData.ChannelAddressPrivateKey
//
//	if tool.CheckIsString(&reqData.R) == false {
//		err = errors.New("r is empty")
//		log.Println(err)
//		return nil, err
//	}
//	if senderSendROfP2p.R != reqData.R {
//		err = errors.New("your r not equal payee's r")
//		log.Println(err)
//		return nil, err
//	}
//
//	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, reqData.ChannelId, user.PeerId)
//	if err != nil {
//		err = errors.New("fail to find latestCommitmentTxInfo")
//		log.Println(err)
//		return nil, err
//	}
//
//	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Htlc_GetH {
//		err = errors.New("wrong latestCommitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
//		log.Println(err)
//		return nil, err
//	}
//
//	if latestCommitmentTxInfo.HtlcSender != user.PeerId {
//		err = errors.New("you are not the HtlcSender")
//		log.Println(err)
//		return nil, err
//	}
//	_, err = tool.GetPubKeyFromWifAndCheck(reqData.R, latestCommitmentTxInfo.HtlcH)
//	if err != nil {
//		return nil, errors.New("r is wrong")
//	}
//	latestCommitmentTxInfo.HtlcR = reqData.R
//	// endregion
//
//	//region 1 根据R创建HED1a的hex
//	hlockTx, hed1aHex, err := createHed1aHexAtPayerSide_at46(tx, *channelInfo, latestCommitmentTxInfo.Id, *reqData, user)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	//endregion
//
//	//region 2 签名herd1b
//	signedHerd1bHex, err := signHerd1bAtPayerSide_at46(*senderSendROfP2p, payerChannelPubKey, *reqData)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	//endregion
//
//	//region  3 创建HEBR1b for payer
//	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Htlc_GetH {
//		he1bTempPubKey := senderSendROfP2p.He1bTempPubKey
//		helbOutAddress, helbOutAddressRedeemScript, helbOutAddressScriptPubKey, err := createMultiSig(he1bTempPubKey, payerChannelPubKey)
//		he1bTxHex := senderSendROfP2p.He1bTxHex
//		he1bOutputs, err := getInputsForNextTxByParseTxHashVout(he1bTxHex, helbOutAddress, helbOutAddressScriptPubKey, helbOutAddressRedeemScript)
//		if err != nil || len(he1bOutputs) == 0 {
//			log.Println(err)
//			return nil, err
//		}
//
//		tempOtherSideCommitmentTx := &dao.CommitmentTransaction{}
//		tempOtherSideCommitmentTx.Id = latestCommitmentTxInfo.Id
//		tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
//		tempOtherSideCommitmentTx.RSMCTempAddressPubKey = he1bTempPubKey
//		tempOtherSideCommitmentTx.RSMCMultiAddress = helbOutAddress
//		tempOtherSideCommitmentTx.RSMCRedeemScript = helbOutAddressRedeemScript
//		tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = helbOutAddressScriptPubKey
//		tempOtherSideCommitmentTx.RSMCTxHex = he1bTxHex
//		tempOtherSideCommitmentTx.RSMCTxid = he1bOutputs[0].Txid
//		tempOtherSideCommitmentTx.AmountToRSMC = latestCommitmentTxInfo.AmountToHtlc
//		err = createCurrCommitmentTxBR(tx, dao.BRType_HE1b, channelInfo, tempOtherSideCommitmentTx, he1bOutputs, payerChannelAddress, reqData.ChannelAddressPrivateKey, user)
//		if err != nil {
//			log.Println(err)
//			return nil, err
//		}
//	}
//	//endregion
//
//	latestCommitmentTxInfo.CurrState = dao.TxInfoState_Htlc_GetR
//	_ = tx.Update(latestCommitmentTxInfo)
//
//	_ = messageService.updateMsgStateUseTx(tx, message)
//
//	_ = tx.Commit()
//
//	responseData = &bean.HtlcRPayerVerifyRInfoToPayee{}
//	responseData.ChannelId = channelInfo.ChannelId
//	responseData.PayerHlockTxHex = hlockTx.TxHex
//	responseData.PayerHed1aHex = hed1aHex //需要让收款方签名，支付给收款方，是从H+收款方地址的多签地址的支出
//	responseData.PayeeSignedHerd1bHex = signedHerd1bHex
//	return responseData, nil
//}

// -47 at Payee side
func (service *htlcBackwardTxManager) SignHed1aAndUpdate_Step4(msgData string, user bean.User) (responseData map[string]interface{}, err error) {
	jsonObjFromPayer := &bean.HtlcRPayerVerifyRInfoToPayee{}
	_ = json.Unmarshal([]byte(msgData), jsonObjFromPayer)

	channelId := jsonObjFromPayer.ChannelId
	payerHlockTxHex := jsonObjFromPayer.PayerHlockTxHex
	payerHed1aHex := jsonObjFromPayer.PayerHed1aHex
	signedHerd1bHex := jsonObjFromPayer.PayeeSignedHerd1bHex

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()
	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	payeeChannelPubKey := channelInfo.PubKeyB
	if user.PeerId == channelInfo.PeerIdA {
		payeeChannelPubKey = channelInfo.PubKeyA
	}
	commitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//region 1 签名hed1a
	hlockMultiAddress, hlockRedeemScript, hlockMultiAddressScriptPubKey, err := createMultiSig(commitmentTxInfo.HtlcH, payeeChannelPubKey)
	payerHLockOutputs, err := getInputsForNextTxByParseTxHashVout(payerHlockTxHex, hlockMultiAddress, hlockMultiAddressScriptPubKey, hlockRedeemScript)
	if err != nil || len(payerHLockOutputs) == 0 {
		log.Println(err)
		return nil, err
	}

	_, signedHed1aHex, err := rpcClient.OmniSignRawTransactionForUnsend(payerHed1aHex, payerHLockOutputs, tempAddrPrivateKeyMap[payeeChannelPubKey])
	if err != nil {
		return nil, err
	}
	result, err := rpcClient.TestMemPoolAccept(signedHed1aHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
			return nil, errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
		}
	}
	//endregion

	//region 2 验证回传的Herd
	err = checkSignedHerdHexAtPayeeSide_at47(tx, signedHerd1bHex, *channelInfo, *commitmentTxInfo, user)
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.CurrState = dao.TxInfoState_Htlc_GetR
	_ = tx.Update(commitmentTxInfo)
	//endregion
	_ = tx.Commit()

	if channelInfo.IsPrivate == false {
		//update htlc state on tracker
		txStateRequest := trackerBean.UpdateHtlcTxStateRequest{}
		txStateRequest.Path = commitmentTxInfo.HtlcRoutingPacket
		txStateRequest.H = commitmentTxInfo.HtlcH
		if strings.HasSuffix(commitmentTxInfo.HtlcRoutingPacket, channelInfo.ChannelId) {
			txStateRequest.R = commitmentTxInfo.HtlcR
		}
		txStateRequest.DirectionFlag = trackerBean.HtlcTxState_ConfirmPayMoney
		txStateRequest.CurrChannelId = channelInfo.ChannelId
		sendMsgToTracker(enum.MsgType_Tracker_UpdateHtlcTxState_352, txStateRequest)
	}

	responseData = make(map[string]interface{})
	payerData := bean.HtlcRPayeeSignHed1aToPayer{}
	payerData.ChannelId = channelId
	payerData.PayerSignedHed1aHex = signedHed1aHex

	responseData["payerData"] = payerData
	responseData["payeeData"] = commitmentTxInfo
	return responseData, nil
}

// -48 at Payer side
func (service *htlcBackwardTxManager) CheckHed1aHex_Step5(msgData string, user bean.User) (responseData interface{}, err error) {
	jsonObjFromPayee := &bean.HtlcRPayeeSignHed1aToPayer{}
	_ = json.Unmarshal([]byte(msgData), jsonObjFromPayee)

	channelId := jsonObjFromPayee.ChannelId
	signedHed1aHex := jsonObjFromPayee.PayerSignedHed1aHex

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//保存hed1a
	err = createAndSaveHed1a_at48(tx, signedHed1aHex, *channelInfo, *latestCommitmentTxInfo, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_ = tx.Commit()
	return latestCommitmentTxInfo, nil
}

//45 签名He1b
func signHe1bAtPayeeSide_at45(tx storm.Node, channelInfo dao.ChannelInfo, commitmentId int, reqData bean.HtlcBobSendR, user bean.User) (he1b *dao.HTLCTimeoutTxForAAndExecutionForB, err error) {
	he1b = &dao.HTLCTimeoutTxForAAndExecutionForB{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentId),
		q.Eq("Owner", user.PeerId)).First(he1b)
	if he1b.Id == 0 {
		return he1b, nil
	}

	hlockOutputs, err := getInputsForNextTxByParseTxHashVout(he1b.InputHex, he1b.RSMCMultiAddress, he1b.RSMCMultiAddressScriptPubKey, he1b.RSMCRedeemScript)
	if err != nil || len(hlockOutputs) == 0 {
		log.Println(err)
		return nil, err
	}

	txId, signedHex, err := rpcClient.OmniSignRawTransactionForUnsend(he1b.RSMCTxHex, hlockOutputs, reqData.R)
	if err != nil {
		return nil, err
	}
	he1b.RSMCTxHex = signedHex
	he1b.RSMCTxid = txId
	he1b.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Update(he1b)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return he1b, nil
}

//45 创建HERD for payee
func createHerd1bAtPayeeSide(tx storm.Node, channelInfo dao.ChannelInfo, he1b *dao.HTLCTimeoutTxForAAndExecutionForB, bobHerdHex string, user bean.User) (herd *dao.RevocableDeliveryTransaction, err error) {
	herd = &dao.RevocableDeliveryTransaction{}
	_ = tx.Select(
		q.Eq("ChannelId", he1b.ChannelId),
		q.Eq("CommitmentTxId", he1b.Id),
		q.Eq("RDType", 1),
		q.Eq("Owner", user.PeerId)).First(herd)
	if herd.Id > 0 {
		return herd, nil
	}

	payeeChannelAddress := channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdA {
		payeeChannelAddress = channelInfo.AddressA
	}

	herd.ChannelId = channelInfo.ChannelId
	herd.CommitmentTxId = he1b.Id
	herd.PeerIdA = channelInfo.PeerIdA
	herd.PeerIdB = channelInfo.PeerIdB
	herd.PropertyId = channelInfo.PropertyId
	herd.Owner = user.PeerId
	herd.RDType = 1

	//input
	herd.InputTxHex = he1b.RSMCTxHex
	herd.InputTxid = rpcClient.GetTxId(he1b.RSMCTxHex)
	herd.InputVout = 0
	herd.InputAmount = he1b.RSMCOutAmount
	//output
	herd.OutputAddress = payeeChannelAddress
	herd.Sequence = 1000
	herd.Amount = he1b.RSMCOutAmount

	herd.TxHex = bobHerdHex
	herd.CurrState = dao.TxInfoState_Init

	herd.CreateBy = user.PeerId
	herd.CreateAt = time.Now()
	err = tx.Save(herd)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return herd, nil
}
func updateHerd1bAtPayeeSide(tx storm.Node, channelInfo dao.ChannelInfo, commitmentId int, herdHex string, user bean.User) (herd *dao.RevocableDeliveryTransaction, err error) {

	he1b := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentId),
		q.Eq("Owner", user.PeerId)).First(he1b)
	if he1b.Id == 0 {
		return nil, errors.New("not found he1b")
	}

	herd = &dao.RevocableDeliveryTransaction{}
	_ = tx.Select(
		q.Eq("ChannelId", he1b.ChannelId),
		q.Eq("CommitmentTxId", he1b.Id),
		q.Eq("RDType", 1),
		q.Eq("Owner", user.PeerId)).First(herd)
	if herd.Id == 0 {
		return nil, errors.New("not found herd")
	}

	herd.TxHex = herdHex
	herd.Txid = rpcClient.GetTxId(herdHex)
	herd.CurrState = dao.TxInfoState_CreateAndSign

	herd.CreateBy = user.PeerId
	herd.CreateAt = time.Now()
	err = tx.Save(herd)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return herd, nil
}

// 46 签名收款方的HERD1b
//func signHerd1bAtPayerSide_at46(senderSendROfP2p bean.BobSendROfP2p, payerChannelPubKey string, reqData bean.HtlcCheckRAndCreateTx) (signedHerd1bHex string, err error) {
//
//	herd1bHex := senderSendROfP2p.Herd1bTxHex
//	he1bTempPubKey := senderSendROfP2p.He1bTempPubKey
//
//	helbOutAddress, he1bRedeemScript, helbOutAddressScriptPubKey, err := createMultiSig(he1bTempPubKey, payerChannelPubKey)
//	he1bTxHex := senderSendROfP2p.He1bTxHex
//	he1bOutputs, err := getInputsForNextTxByParseTxHashVout(he1bTxHex, helbOutAddress, helbOutAddressScriptPubKey, he1bRedeemScript)
//	if err != nil || len(he1bOutputs) == 0 {
//		log.Println(err)
//		return "", err
//	}
//	_, signedHerd1bHex, err = rpcClient.OmniSignRawTransactionForUnsend(herd1bHex, he1bOutputs, reqData.ChannelAddressPrivateKey)
//	if err != nil {
//		return "", err
//	}
//	result, err := rpcClient.TestMemPoolAccept(signedHerd1bHex)
//	if err != nil {
//		return "", err
//	}
//	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
//		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
//			return "", errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
//		}
//	}
//	return signedHerd1bHex, nil
//}

// 46的子交易 付款方创建HED1a的hex，需要收款方签名
func createHed1aHexAtPayerSide_at46(tx storm.Node, channelInfo dao.ChannelInfo, latestCommitmentTxInfoId int, reqData bean.HtlcCheckRAndCreateTx, user bean.User) (hlockTx *dao.HtlcLockTxByH, hed1aHex string, err error) {
	hlockTx = &dao.HtlcLockTxByH{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", latestCommitmentTxInfoId)).First(hlockTx)
	if err != nil {
		err = errors.New("not found the hLockTx")
		log.Println(err)
		return nil, "", err
	}

	hlockOutputs, err := getInputsForNextTxByParseTxHashVout(hlockTx.TxHex, hlockTx.OutputAddress, hlockTx.ScriptPubKey, hlockTx.RedeemScript)
	if err != nil || len(hlockOutputs) == 0 {
		log.Println(err)
		return nil, "", err
	}

	payeeChannelAddress := channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdB {
		payeeChannelAddress = channelInfo.AddressA
	}
	_, hed1aHex, err = rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		hlockTx.OutputAddress,
		[]string{
			reqData.R,
		},
		hlockOutputs,
		payeeChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		hlockTx.OutAmount,
		getBtcMinerAmount(channelInfo.BtcAmount),
		0,
		&hlockTx.RedeemScript)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	return hlockTx, hed1aHex, nil
}

func checkSignedHerdHexAtPayeeSide_at47(tx storm.Node, signedHerd1bHex string, channelInfo dao.ChannelInfo, commitmentTxInfo dao.CommitmentTransaction, user bean.User) (err error) {
	he1b := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTxInfo.Id),
		q.Eq("Owner", user.PeerId)).First(he1b)
	if err != nil {
		log.Println(err)
		return err
	}

	herd := &dao.RevocableDeliveryTransaction{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", he1b.Id),
		q.Eq("RDType", 1),
		q.Eq("Owner", user.PeerId)).First(herd)
	if err != nil {
		log.Println(err)
		return err
	}

	result, err := rpcClient.TestMemPoolAccept(signedHerd1bHex)
	if err != nil {
		return err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
			return errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
		}
	}

	he1bOutputs, err := getInputsForNextTxByParseTxHashVout(he1b.RSMCTxHex, he1b.RSMCMultiAddress, he1b.RSMCMultiAddressScriptPubKey, he1b.RSMCRedeemScript)
	if err != nil || len(he1bOutputs) == 0 {
		log.Println(err)
		return err
	}

	result, err = rpcClient.OmniDecodeTransactionWithPrevTxs(signedHerd1bHex, he1bOutputs)
	if err != nil {
		log.Println(err)
		return err
	}

	hexJsonObj := gjson.Parse(result)
	if he1b.RSMCMultiAddress != hexJsonObj.Get("sendingaddress").String() {
		err = errors.New("wrong inputAddress at payerHt1aHex  at 41 protocol")
		log.Println(err)
		return err
	}

	if herd.OutputAddress != hexJsonObj.Get("referenceaddress").String() {
		err = errors.New("wrong outputAddress at payerHt1aHex  at 41 protocol")
		log.Println(err)
		return err
	}
	if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
		err = errors.New("wrong propertyId at payerHt1aHex  at 41 protocol")
		log.Println(err)
		return err
	}
	if commitmentTxInfo.AmountToHtlc != hexJsonObj.Get("amount").Float() {
		err = errors.New("wrong amount at payerHt1aHex  at 41 protocol")
		log.Println(err)
		return err
	}

	herd.TxHex = signedHerd1bHex
	herd.Txid = hexJsonObj.Get("txid").Str
	herd.CurrState = dao.TxInfoState_CreateAndSign
	herd.SignAt = time.Now()
	_ = tx.Update(herd)
	return nil
}

func createAndSaveHed1a_at48(tx storm.Node, signedHed1aHex string, channelInfo dao.ChannelInfo, commitmentTxInfo dao.CommitmentTransaction, user bean.User) (err error) {
	hlockTx := &dao.HtlcLockTxByH{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTxInfo.Id)).First(hlockTx)
	if err != nil {
		err = errors.New("not found the hLockTx")
		log.Println(err)
		return err
	}

	hed1a := &dao.HTLCExecutionDeliveryOfR{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTxInfo.Id),
		q.Eq("HLockTxId", hlockTx.Id)).First(hed1a)
	if hed1a.Id == 0 {
		payeeChannelAddress := channelInfo.AddressB
		payeePeerId := channelInfo.PeerIdB
		if user.PeerId == channelInfo.PeerIdB {
			payeeChannelAddress = channelInfo.AddressA
			payeePeerId = channelInfo.PeerIdA
		}
		decodeHed1aHex, err := rpcClient.DecodeRawTransaction(signedHed1aHex)
		if err != nil {
			return err
		}

		hed1a.ChannelId = channelInfo.ChannelId
		hed1a.CommitmentTxId = commitmentTxInfo.Id
		hed1a.HLockTxId = hlockTx.Id

		hed1a.InputAmount = hlockTx.OutAmount
		hed1a.InputTxid = hlockTx.Txid
		hed1a.InputHex = hlockTx.TxHex
		hed1a.HtlcR = commitmentTxInfo.HtlcR

		hed1a.OutputAddress = payeeChannelAddress
		hed1a.TxHex = signedHed1aHex
		hed1a.Txid = gjson.Get(decodeHed1aHex, "txid").Str
		hed1a.OutAmount = hlockTx.OutAmount

		hed1a.CurrState = dao.TxInfoState_CreateAndSign
		hed1a.Owner = payeePeerId
		hed1a.CreateAt = time.Now()
		hed1a.CreateBy = user.PeerId
		_ = tx.Save(hed1a)
	}
	return nil
}

func createHed1a(tx storm.Node, signedHed1aHex string, channelInfo dao.ChannelInfo, commitmentTxInfo dao.CommitmentTransaction, user bean.User) (err error) {
	hlockTx := &dao.HtlcLockTxByH{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTxInfo.Id)).First(hlockTx)
	if err != nil {
		err = errors.New("not found the hLockTx")
		log.Println(err)
		return err
	}

	hed1a := &dao.HTLCExecutionDeliveryOfR{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTxInfo.Id),
		q.Eq("HLockTxId", hlockTx.Id)).First(hed1a)
	if hed1a.Id == 0 {
		payeeChannelAddress := channelInfo.AddressB
		payeePeerId := channelInfo.PeerIdB
		if user.PeerId == channelInfo.PeerIdB {
			payeeChannelAddress = channelInfo.AddressA
			payeePeerId = channelInfo.PeerIdA
		}
		decodeHed1aHex, err := rpcClient.DecodeRawTransaction(signedHed1aHex)
		if err != nil {
			return err
		}

		hed1a.ChannelId = channelInfo.ChannelId
		hed1a.CommitmentTxId = commitmentTxInfo.Id
		hed1a.HLockTxId = hlockTx.Id

		hed1a.InputAmount = hlockTx.OutAmount
		hed1a.InputTxid = hlockTx.Txid
		hed1a.InputHex = hlockTx.TxHex
		hed1a.HtlcR = commitmentTxInfo.HtlcR

		hed1a.OutputAddress = payeeChannelAddress
		hed1a.TxHex = signedHed1aHex
		hed1a.Txid = gjson.Get(decodeHed1aHex, "txid").Str
		hed1a.OutAmount = hlockTx.OutAmount

		hed1a.CurrState = dao.TxInfoState_Create
		hed1a.Owner = payeePeerId
		hed1a.CreateAt = time.Now()
		hed1a.CreateBy = user.PeerId
		_ = tx.Save(hed1a)
	}
	return nil
}

func signHed1a(tx storm.Node, channelInfo dao.ChannelInfo, commitmentTxInfo dao.CommitmentTransaction, user bean.User) (err error) {
	hlockTx := &dao.HtlcLockTxByH{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTxInfo.Id)).First(hlockTx)
	if err != nil {
		err = errors.New("not found the hLockTx")
		log.Println(err)
		return err
	}

	hed1a := &dao.HTLCExecutionDeliveryOfR{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTxInfo.Id),
		q.Eq("HLockTxId", hlockTx.Id)).First(hed1a)
	if hed1a.Id > 0 {
		payeeChannelPubKey := channelInfo.PubKeyB
		if user.PeerId == channelInfo.PeerIdB {
			payeeChannelPubKey = channelInfo.PubKeyA
		}

		c3aHlockMultiAddress, c3aHlockRedeemScript, c3aHlockAddrScriptPubKey, err := createMultiSig(commitmentTxInfo.HtlcH, payeeChannelPubKey)
		if err != nil {
			return err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(hed1a.InputHex, c3aHlockMultiAddress, c3aHlockAddrScriptPubKey, c3aHlockRedeemScript)
		if err != nil || len(inputs) == 0 {
			log.Println(err)
			return err
		}

		txId, hex, err := rpcClient.OmniSignRawTransactionForUnsend(hed1a.TxHex, inputs, commitmentTxInfo.HtlcR)
		if err != nil {
			return err
		}
		hed1a.TxHex = hex
		hed1a.Txid = txId
		_ = tx.Update(hed1a)
	}
	return errors.New("fail to sign")
}
