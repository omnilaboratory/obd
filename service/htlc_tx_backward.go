package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
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
	operationFlag sync.Mutex
}

// HTLC Reverse pass the R (Preimage R)
var HtlcBackwardTxService htlcBackwardTxManager

// step 1 bob -100045 收款方收到R，发送R到obd进行验证
func (service *htlcBackwardTxManager) SendRToPreviousNodeAtBobSide(msg bean.RequestMessage, user bean.User) (retData interface{}, err error) {

	log.Println("htlc step 13 begin", time.Now())

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

	_, err = omnicore.GetPubKeyFromWifAndCheck(reqData.R, latestCommitmentTxInfo.HtlcH)
	if err != nil {
		return nil, errors.New(enum.Tips_htlc_wrongRForH)
	}

	latestCommitmentTxInfo.HtlcR = reqData.R

	// endregion

	currBlockHeight := conn2tracker.GetBlockCount()

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
	heRdTx, err := omnicore.OmniCreateRawTransactionUseUnsendInput(
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
	heBrTx, err := omnicore.OmniCreateRawTransactionUseUnsendInput(
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

	cacheDataForTx := &dao.CacheDataForTx{}
	cacheDataForTx.KeyName = user.PeerId + "_htlcBack_" + channelInfo.ChannelId
	_ = tx.Select(q.Eq("KeyName", cacheDataForTx.KeyName)).First(cacheDataForTx)
	if cacheDataForTx.Id != 0 {
		_ = tx.DeleteStruct(cacheDataForTx)
	}
	bytes, _ := json.Marshal(&dataSendTo45P)
	cacheDataForTx.Data = bytes
	_ = tx.Save(cacheDataForTx)

	_ = tx.Commit()

	log.Println("htlc step 13 end", time.Now())
	return dataNeedBobSign, nil
}

// step 2 bob -100106 bob签名HeRd的结果 推送45号协议
func (service *htlcBackwardTxManager) OnBobSignedHeRdAtBobSide(msg bean.RequestMessage, user bean.User) (retData interface{}, err error) {
	log.Println("htlc step 14 begin", time.Now())
	bobSignedData := bean.BobSignHerdForC3b{}
	_ = json.Unmarshal([]byte(msg.Data), &bobSignedData)

	if tool.CheckIsString(&bobSignedData.ChannelId) == false {
		return nil, errors.New("error channel_id")
	}

	cacheDataForTx := &dao.CacheDataForTx{}
	_ = user.Db.Select(q.Eq("KeyName", user.PeerId+"_htlcBack_"+bobSignedData.ChannelId)).First(cacheDataForTx)
	if cacheDataForTx.Id == 0 {
		return nil, errors.New("error channel_id")
	}

	dataSendTo45P := &bean.NeedAliceSignHerdTxOfC3bP2p{}
	_ = json.Unmarshal(cacheDataForTx.Data, dataSendTo45P)
	if len(dataSendTo45P.ChannelId) == 0 {
		return nil, errors.New("error channel_id")
	}

	if pass, _ := omnicore.CheckMultiSign(bobSignedData.C3bHtlcHerdPartialSignedHex, 1); pass == false {
		return nil, errors.New("error sign c3b_htlc_herd_partial_signed_hex")
	}

	dataSendTo45P.C3bHtlcHerdPartialSignedData.Hex = bobSignedData.C3bHtlcHerdPartialSignedHex
	_ = user.Db.Update(dataSendTo45P)

	log.Println("htlc step 14 end", time.Now())
	return dataSendTo45P, nil
}

// step 3 alice p2p -45 推送待签名的herd
func (service *htlcBackwardTxManager) OnGetHeSubTxDataAtAliceObdAtAliceSide(msg string, user bean.User) (retData interface{}, err error) {
	log.Println("htlc step 15 begin", time.Now())
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
	if _, err = omnicore.GetPubKeyFromWifAndCheck(dataFrom45P.R, latestCommitmentTx.HtlcH); err != nil {
		return nil, errors.New(enum.Tips_htlc_wrongRForH)
	}

	if tool.CheckIsString(&dataFrom45P.C3bHtlcTempAddressForHePubKey) == false {
		return nil, errors.New(enum.Tips_common_empty + "c3b_htlc_temp_address_for_he_pub_key")
	}

	if tool.CheckIsString(&dataFrom45P.HeCompleteSignedHex) == false {
		return nil, errors.New(enum.Tips_common_empty + "he_complete_signed_hex")
	}

	cacheDataForTx := &dao.CacheDataForTx{}
	cacheDataForTx.KeyName = user.PeerId + "_htlcBack_" + dataFrom45P.ChannelId
	_ = tx.Select(q.Eq("KeyName", cacheDataForTx.KeyName)).First(cacheDataForTx)
	if cacheDataForTx.Id != 0 {
		_ = tx.DeleteStruct(cacheDataForTx)
	}
	bytes, _ := json.Marshal(&dataFrom45P)
	cacheDataForTx.Data = bytes
	_ = tx.Save(cacheDataForTx)

	_ = tx.Commit()

	log.Println("htlc step 15 end", time.Now())
	return dataFrom45P, nil
}

// step 4 alice  -46 Alice完成herd签名 保存hebr 推送46 herd
func (service *htlcBackwardTxManager) OnAliceSignedHeRdAtAliceSide(msg bean.RequestMessage, user bean.User) (toAlice, toBob interface{}, err error) {
	log.Println("htlc step 16 begin", time.Now())
	herdSignedResult := bean.AliceSignHerdTxOfC3e{}
	_ = json.Unmarshal([]byte(msg.Data), &herdSignedResult)

	if tool.CheckIsString(&herdSignedResult.ChannelId) == false {
		return nil, nil, errors.New("error channel_id")
	}

	cacheDataForTx := &dao.CacheDataForTx{}
	_ = user.Db.Select(q.Eq("KeyName", user.PeerId+"_htlcBack_"+herdSignedResult.ChannelId)).First(cacheDataForTx)
	if cacheDataForTx.Id == 0 {
		return nil, nil, errors.New("error channel_id")
	}

	dataFrom45P := &bean.NeedAliceSignHerdTxOfC3bP2p{}
	_ = json.Unmarshal(cacheDataForTx.Data, dataFrom45P)
	if len(dataFrom45P.ChannelId) == 0 {
		return nil, nil, errors.New("error channel_id")
	}

	if pass, _ := omnicore.CheckMultiSign(herdSignedResult.C3bHtlcHerdCompleteSignedHex, 2); pass == false {
		return nil, nil, errors.New("error sign c3b_htlc_herd_complete_signed_hex")
	}
	dataFrom45P.C3bHtlcHerdPartialSignedData.Hex = herdSignedResult.C3bHtlcHerdCompleteSignedHex

	if pass, _ := omnicore.CheckMultiSign(herdSignedResult.C3bHtlcHebrPartialSignedHex, 1); pass == false {
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
	log.Println("htlc step 16 end", time.Now())
	return latestCommitment, dataTo46P, nil
}

// step 5 bob  -110046 bob保存herd
func (service *htlcBackwardTxManager) OnGetHeRdDataAtBobObd(msg string, user bean.User) (retData interface{}, err error) {
	log.Println("htlc step 17 begin", time.Now())
	dataFrom46P := bean.AliceSignedHerdTxOfC3bP2p{}
	_ = json.Unmarshal([]byte(msg), &dataFrom46P)

	if tool.CheckIsString(&dataFrom46P.ChannelId) == false {
		return nil, errors.New("error channel_id")
	}

	if pass, _ := omnicore.CheckMultiSign(dataFrom46P.C3bHtlcHerdCompleteSignedHex, 2); pass == false {
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
	log.Println("htlc step 17 end", time.Now())
	return latestCommitment, nil
}

//45 签名He1b
func signHe1bAtPayeeSide_at45(tx storm.Node, channelInfo dao.ChannelInfo, commitmentId int, reqData bean.HtlcBobSendR, user bean.User) (he1b *dao.HTLCTimeoutTxForAAndExecutionForB, err error) {
	hlock := &dao.HtlcLockTxByH{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentId),
		q.Eq("Owner", user.PeerId)).First(hlock)
	if hlock.Id == 0 {
		return nil, errors.New("fail to find hlock")
	}
	he1b = &dao.HTLCTimeoutTxForAAndExecutionForB{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentId),
		q.Eq("Owner", user.PeerId)).First(he1b)
	if he1b.Id == 0 {
		return nil, errors.New("fail to find he1b")
	}

	hlockOutputs, err := getInputsForNextTxByParseTxHashVout(he1b.InputHex, hlock.OutputAddress, hlock.ScriptPubKey, hlock.RedeemScript)
	if err != nil || len(hlockOutputs) == 0 {
		log.Println(err)
		return nil, err
	}

	txId, signedHex, err := omnicore.OmniSignRawTransactionForUnsend(he1b.RSMCTxHex, hlockOutputs, reqData.R)
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
	herd.InputTxid = omnicore.GetTxId(he1b.RSMCTxHex)
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
	herd.Txid = omnicore.GetTxId(herdHex)
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
		decodeHed1aHex, err := omnicore.DecodeBtcRawTransaction(signedHed1aHex)
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

		c3bHlockMultiAddress, c3bHlockRedeemScript, c3bHlockAddrScriptPubKey, err := createMultiSig(commitmentTxInfo.HtlcH, payeeChannelPubKey)
		if err != nil {
			return err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(hed1a.InputHex, c3bHlockMultiAddress, c3bHlockAddrScriptPubKey, c3bHlockRedeemScript)
		if err != nil || len(inputs) == 0 {
			log.Println(err)
			return err
		}

		txId, hex, err := omnicore.OmniSignRawTransactionForUnsend(hed1a.TxHex, inputs, commitmentTxInfo.HtlcR)
		if err != nil {
			return err
		}
		hed1a.TxHex = hex
		hed1a.Txid = txId
		return tx.Update(hed1a)
	}
	return errors.New("fail to sign")
}
