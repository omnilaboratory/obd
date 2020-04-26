package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"log"
	"obd/bean"
	"obd/config"
	"obd/dao"
	"obd/tool"
	"strconv"
	"sync"
	"time"
)

type htlcBackwardTxManager struct {
	operationFlag sync.Mutex
}

// HTLC Reverse pass the R (Preimage R)
var HtlcBackwardTxService htlcBackwardTxManager

// 45 at payee side
func (service *htlcBackwardTxManager) SendRToPreviousNode_Step1(msg bean.RequestMessage, user bean.User) (responseData map[string]interface{}, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New("empty json data")
	}

	reqData := &bean.HtlcSendR{}
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
		err = errors.New("channel_id is empty")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("channel_address_private_key is empty")
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
		err = errors.New("fail to find channelInfo")
		log.Println(err)
		return nil, err
	}

	payeeChannelPubKey := channelInfo.PubKeyB
	payerPeerId := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		payeeChannelPubKey = channelInfo.PubKeyA
		payerPeerId = channelInfo.PeerIdB
	}
	if payerPeerId != msg.RecipientUserPeerId {
		return nil, errors.New("recipientUserPeerId is wrong")
	}
	if msg.RecipientNodePeerId == P2PLocalPeerId {
		err = FindUserIsOnline(payerPeerId)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, payeeChannelPubKey)
	if err != nil {
		return nil, errors.New("ChannelAddressPrivateKey is wrong")
	}
	tempAddrPrivateKeyMap[payeeChannelPubKey] = reqData.ChannelAddressPrivateKey

	if tool.CheckIsString(&reqData.R) == false {
		err = errors.New("r is empty")
		log.Println(err)
		return nil, err
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, reqData.ChannelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find latestCommitmentTxInfo")
		log.Println(err)
		return nil, err
	}
	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Htlc_GetH {
		err = errors.New("wrong latestCommitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, err
	}

	if latestCommitmentTxInfo.HtlcSender != msg.RecipientUserPeerId {
		err = errors.New("wrong HtlcSender")
		log.Println(err)
		return nil, err
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.R, latestCommitmentTxInfo.HtlcH)
	if err != nil {
		return nil, errors.New("r is wrong")
	}
	latestCommitmentTxInfo.HtlcR = reqData.R

	if tool.CheckIsString(&reqData.CurrHtlcTempAddressForHE1bPubKey) == false {
		err = errors.New("curr_htlc_temp_address_for_he1b_pub_key is empty")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&reqData.CurrHtlcTempAddressForHE1bPrivateKey) == false {
		err = errors.New("curr_htlc_temp_address_for_he1b_private_key is empty")
		log.Println(err)
		return nil, err
	}
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.CurrHtlcTempAddressForHE1bPrivateKey, reqData.CurrHtlcTempAddressForHE1bPubKey)
	if err != nil {
		return nil, errors.New("CurrHtlcTempAddressForHE1bPrivateKey is wrong")
	}
	// endregion

	currBlockHeight, err := rpcClient.GetBlockCount()
	if err != nil {
		return nil, errors.New("fail to get blockHeight ,please try again later")
	}

	htlcTimeOut := getHtlcTimeout(latestCommitmentTxInfo.HtlcChannelPath, channelInfo.ChannelId)
	maxHeight := latestCommitmentTxInfo.BeginBlockHeight + htlcTimeOut
	if config.ChainNode_Type == "mainnet" {
		if currBlockHeight > maxHeight {
			return nil, errors.New("timeout, can't transfer the R")
		}
	}
	tempAddrPrivateKeyMap[reqData.CurrHtlcTempAddressForHE1bPubKey] = reqData.CurrHtlcTempAddressForHE1bPrivateKey

	he1b, err := createHe1bAtPayeeSide_at45(tx, *channelInfo, latestCommitmentTxInfo.Id, *reqData, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	herd, err := createHerd1bAtPayeeSide_at45(tx, *channelInfo, he1b, *reqData, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_ = tx.Update(latestCommitmentTxInfo)
	_ = tx.Commit()

	responseData = make(map[string]interface{})
	responseData["channelId"] = reqData.ChannelId
	responseData["r"] = reqData.R
	responseData["he1bHex"] = he1b.TxHex
	responseData["he1bTempPubKey"] = reqData.CurrHtlcTempAddressForHE1bPubKey
	responseData["herd1bHex"] = herd.TxHex

	return responseData, nil
}

// 45 at payer side
func (service *htlcBackwardTxManager) BeforeSendRInfoToPayerAtAliceSide_Step2(msgData string, user bean.User) (data map[string]interface{}, needNoticeOtherSide bool, err error) {

	jsonObjFromPayee := gjson.Parse(msgData)
	channelId := jsonObjFromPayee.Get("channelId").String()

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if channelInfo == nil {
		return nil, true, errors.New("not found channel info")
	}

	senderPeerId := channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		senderPeerId = channelInfo.PeerIdA
	}
	messageHash := MessageService.saveMsgUseTx(tx, senderPeerId, user.PeerId, msgData)
	returnData := make(map[string]interface{})
	_ = tx.Commit()

	returnData["channelId"] = channelId
	returnData["r"] = jsonObjFromPayee.Get("r").String()
	returnData["msgHash"] = messageHash

	return returnData, false, nil
}

// -46 at Payer side
func (service *htlcBackwardTxManager) VerifyRAndCreateTxs_Step3(msgData string, user bean.User) (responseData map[string]interface{}, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, errors.New("empty json responseData")
	}
	reqData := &bean.HtlcCheckRAndCreateTx{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	if tool.CheckIsString(&reqData.RequestHash) == false {
		err = errors.New("empty request_hash")
		log.Println(err)
		return nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	message, err := MessageService.getMsgUseTx(tx, reqData.RequestHash)
	if err != nil {
		return nil, errors.New("wrong request_hash")
	}
	if message.Receiver != user.PeerId {
		return nil, errors.New("you are not the operator")
	}
	jsonDataFromPayee := gjson.Parse(message.Data)

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("channel_address_private_key is empty")
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
		err = errors.New("not found the channel " + reqData.ChannelId)
		log.Println(err)
		return nil, err
	}

	payerChannelPubKey := channelInfo.PubKeyA
	payerChannelAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		payerChannelPubKey = channelInfo.PubKeyB
		payerChannelAddress = channelInfo.AddressB
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, payerChannelPubKey)
	if err != nil {
		err = errors.New("channel_address_private_key is wrong")
		log.Println(err)
		return nil, err
	}
	tempAddrPrivateKeyMap[payerChannelPubKey] = reqData.ChannelAddressPrivateKey

	if tool.CheckIsString(&reqData.R) == false {
		err = errors.New("channel_address_private_key is empty")
		log.Println(err)
		return nil, err
	}
	if jsonDataFromPayee.Get("r").String() != reqData.R {
		err = errors.New("your r not equal payee's r")
		log.Println(err)
		return nil, err
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, reqData.ChannelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find latestCommitmentTxInfo")
		log.Println(err)
		return nil, err
	}
	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Htlc_GetH {
		err = errors.New("wrong latestCommitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, err
	}

	if latestCommitmentTxInfo.HtlcSender != user.PeerId {
		err = errors.New("you are not the HtlcSender")
		log.Println(err)
		return nil, err
	}
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.R, latestCommitmentTxInfo.HtlcH)
	if err != nil {
		return nil, errors.New("r is wrong")
	}
	latestCommitmentTxInfo.HtlcR = reqData.R
	// endregion

	//region 1 根据R创建HED1a的hex
	hlockTx, hed1aHex, err := createHed1aHexAtPayerSide_at46(tx, *channelInfo, latestCommitmentTxInfo.Id, *reqData, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//endregion

	//region 2 签名herd1b
	signedHerd1bHex, err := signHerd1bAtPayerSide_at46(jsonDataFromPayee, payerChannelPubKey, *reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//endregion

	//region  3 创建HEBR1b for payer
	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Htlc_GetH {
		he1bTempPubKey := jsonDataFromPayee.Get("he1bTempPubKey").Str
		helbOutAddress, helbOutAddressRedeemScript, helbOutAddressScriptPubKey, err := createMultiSig(he1bTempPubKey, payerChannelPubKey)
		he1bTxHex := jsonDataFromPayee.Get("he1bHex").String()
		he1bOutputs, err := getInputsForNextTxByParseTxHashVout(he1bTxHex, helbOutAddress, helbOutAddressScriptPubKey)
		if err != nil || len(he1bOutputs) == 0 {
			log.Println(err)
			return nil, err
		}

		tempOtherSideCommitmentTx := &dao.CommitmentTransaction{}
		tempOtherSideCommitmentTx.Id = latestCommitmentTxInfo.Id
		tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
		tempOtherSideCommitmentTx.RSMCTempAddressPubKey = he1bTempPubKey
		tempOtherSideCommitmentTx.RSMCMultiAddress = helbOutAddress
		tempOtherSideCommitmentTx.RSMCRedeemScript = helbOutAddressRedeemScript
		tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = helbOutAddressScriptPubKey
		tempOtherSideCommitmentTx.RSMCTxHex = he1bTxHex
		tempOtherSideCommitmentTx.RSMCTxid = he1bOutputs[0].Txid
		tempOtherSideCommitmentTx.AmountToRSMC = latestCommitmentTxInfo.AmountToHtlc
		err = createCurrCommitmentTxBR(tx, dao.BRType_HE1b, channelInfo, tempOtherSideCommitmentTx, he1bOutputs, payerChannelAddress, reqData.ChannelAddressPrivateKey, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	//endregion

	latestCommitmentTxInfo.CurrState = dao.TxInfoState_Htlc_GetR
	_ = tx.Update(latestCommitmentTxInfo)

	_ = tx.Commit()

	responseData = make(map[string]interface{})
	responseData["channelId"] = channelInfo.ChannelId
	responseData["payerHlockTxHex"] = hlockTx.TxHex
	responseData["payerHed1aHex"] = hed1aHex //需要让收款方签名，支付给收款方，是从H+收款方地址的多签地址的支出
	responseData["signedHerd1bHex"] = signedHerd1bHex
	return responseData, nil
}

//47 at Payee side
func (service *htlcBackwardTxManager) SignHed1aAndUpdate_Step4(msgData string, user bean.User) (responseData map[string]interface{}, err error) {
	jsonObjFromPayer := gjson.Parse(msgData)

	channelId := jsonObjFromPayer.Get("channelId").Str
	payerHlockTxHex := jsonObjFromPayer.Get("payerHlockTxHex").Str
	payerHed1aHex := jsonObjFromPayer.Get("payerHed1aHex").Str
	signedHerd1bHex := jsonObjFromPayer.Get("signedHerd1bHex").Str

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
	hlockMultiAddress, _, hlockMultiAddressScriptPubKey, err := createMultiSig(commitmentTxInfo.HtlcH, payeeChannelPubKey)
	payerHLockOutputs, err := getInputsForNextTxByParseTxHashVout(payerHlockTxHex, hlockMultiAddress, hlockMultiAddressScriptPubKey)
	if err != nil {
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
	//endregion
	_ = tx.Commit()

	responseData = make(map[string]interface{})
	payerData := make(map[string]interface{})
	payerData["channelId"] = channelId
	payerData["signedHed1aHex"] = signedHed1aHex

	payeeData := make(map[string]interface{})
	payeeData["commitmentTxInfo"] = commitmentTxInfo

	responseData["payerData"] = payerData
	responseData["payeeData"] = payeeData
	return responseData, nil
}

//48 at Payer side
func (service *htlcBackwardTxManager) CheckHed1aHex_Step5(msgData string, user bean.User) (responseData map[string]interface{}, err error) {
	jsonObjFromPayee := gjson.Parse(msgData)

	channelId := jsonObjFromPayee.Get("channelId").Str
	signedHed1aHex := jsonObjFromPayee.Get("signedHed1aHex").Str

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
	commitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//保存hed1a
	err = createAndSaveHed1a_at48(tx, signedHed1aHex, *channelInfo, *commitmentTxInfo, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_ = tx.Commit()
	responseData = make(map[string]interface{})
	responseData["commitmentTxInfo"] = commitmentTxInfo
	return responseData, nil
}

//45 创建He1b
func createHe1bAtPayeeSide_at45(tx storm.Node, channelInfo dao.ChannelInfo, latestCommitmentTxInfoId int, reqData bean.HtlcSendR, user bean.User) (he1b *dao.HTLCExecutionB, err error) {
	he1b = &dao.HTLCExecutionB{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", latestCommitmentTxInfoId),
		q.Eq("Owner", user.PeerId)).First(he1b)
	if he1b.Id > 0 {
		return he1b, nil
	}

	hlockTx := &dao.HtlcLockTxByH{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", latestCommitmentTxInfoId)).First(hlockTx)
	if err != nil {
		err = errors.New("not found the hLockTx")
		log.Println(err)
		return nil, err
	}

	hlockOutputs, err := getInputsForNextTxByParseTxHashVout(hlockTx.TxHex, hlockTx.OutputAddress, hlockTx.ScriptPubKey)
	if err != nil || len(hlockOutputs) == 0 {
		log.Println(err)
		return nil, err
	}

	payerChannelPubKey := channelInfo.PubKeyA
	if user.PeerId == channelInfo.PeerIdA {
		payerChannelPubKey = channelInfo.PubKeyB
	}

	he1bMultiAddress, he1bRedeemScript, he1bScriptPubKey, err := createMultiSig(reqData.CurrHtlcTempAddressForHE1bPubKey, payerChannelPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bobHe1bTxid, bobHe1bHex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		hlockTx.OutputAddress,
		[]string{
			reqData.R,
			reqData.ChannelAddressPrivateKey,
		},
		hlockOutputs,
		he1bMultiAddress,
		he1bMultiAddress,
		channelInfo.PropertyId,
		hlockTx.OutAmount,
		0,
		0,
		&hlockTx.RedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	he1b.InputHex = hlockTx.TxHex
	he1b.InputTxid = hlockOutputs[0].Txid
	he1b.InputAmount = hlockTx.OutAmount

	he1b.HtlcTempAddressForHE1bPubKey = reqData.CurrHtlcTempAddressForHE1bPubKey
	he1b.OutputAddress = he1bMultiAddress
	he1b.OutAddressRedeemScript = he1bRedeemScript
	he1b.OutAddressScriptPubKey = he1bScriptPubKey
	he1b.OutAmount = hlockTx.OutAmount
	he1b.HtlcR = reqData.R
	he1b.TxHex = bobHe1bHex
	he1b.Txid = bobHe1bTxid

	he1b.ChannelId = channelInfo.ChannelId
	he1b.CommitmentTxId = latestCommitmentTxInfoId
	he1b.Owner = user.PeerId
	he1b.CreateBy = user.PeerId
	he1b.CreateAt = time.Now()
	he1b.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(he1b)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return he1b, nil
}

//45 创建HERD for payee
func createHerd1bAtPayeeSide_at45(tx storm.Node, channelInfo dao.ChannelInfo, he1b *dao.HTLCExecutionB, reqData bean.HtlcSendR, user bean.User) (herd *dao.RevocableDeliveryTransaction, err error) {
	herd = &dao.RevocableDeliveryTransaction{}
	_ = tx.Select(
		q.Eq("ChannelId", he1b.ChannelId),
		q.Eq("CommitmentTxId", he1b.Id),
		q.Eq("RDType", 1),
		q.Eq("Owner", user.PeerId)).First(herd)
	if herd.Id > 0 {
		return herd, nil
	}

	he1bOutputs, err := getInputsForNextTxByParseTxHashVout(he1b.TxHex, he1b.OutputAddress, he1b.OutAddressScriptPubKey)
	if err != nil || len(he1bOutputs) == 0 {
		log.Println(err)
		return nil, err
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
	herd.InputTxHex = he1b.TxHex
	herd.InputTxid = he1bOutputs[0].Txid
	herd.InputVout = 0
	herd.InputAmount = he1b.OutAmount
	//output
	herd.OutputAddress = payeeChannelAddress
	herd.Sequence = 1000
	herd.Amount = he1b.OutAmount

	bobHerdTxid, bobHerdHex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		he1b.OutputAddress,
		[]string{
			reqData.CurrHtlcTempAddressForHE1bPrivateKey,
		},
		he1bOutputs,
		herd.OutputAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		he1b.OutAmount,
		0,
		herd.Sequence,
		&he1b.OutAddressRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	herd.TxHex = bobHerdHex
	herd.Txid = bobHerdTxid
	herd.CurrState = dao.TxInfoState_Create

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
func signHerd1bAtPayerSide_at46(jsonDataFromPayee gjson.Result, payerChannelPubKey string, reqData bean.HtlcCheckRAndCreateTx) (signedHerd1bHex string, err error) {

	herd1bHex := jsonDataFromPayee.Get("herd1bHex").Str

	he1bTempPubKey := jsonDataFromPayee.Get("he1bTempPubKey").Str
	helbOutAddress, _, helbOutAddressScriptPubKey, err := createMultiSig(he1bTempPubKey, payerChannelPubKey)
	he1bTxHex := jsonDataFromPayee.Get("he1bHex").String()
	he1bOutputs, err := getInputsForNextTxByParseTxHashVout(he1bTxHex, helbOutAddress, helbOutAddressScriptPubKey)
	if err != nil || len(he1bOutputs) == 0 {
		log.Println(err)
		return "", err
	}
	_, signedHerd1bHex, err = rpcClient.OmniSignRawTransactionForUnsend(herd1bHex, he1bOutputs, reqData.ChannelAddressPrivateKey)
	if err != nil {
		return "", err
	}
	result, err := rpcClient.TestMemPoolAccept(signedHerd1bHex)
	if err != nil {
		return "", err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
			return "", errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
		}
	}
	return signedHerd1bHex, nil
}

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

	hlockOutputs, err := getInputsForNextTxByParseTxHashVout(hlockTx.TxHex, hlockTx.OutputAddress, hlockTx.ScriptPubKey)
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
		0,
		0,
		&hlockTx.RedeemScript)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	return hlockTx, hed1aHex, nil
}

func checkSignedHerdHexAtPayeeSide_at47(tx storm.Node, signedHerd1bHex string, channelInfo dao.ChannelInfo, commitmentTxInfo dao.CommitmentTransaction, user bean.User) (err error) {
	he1b := &dao.HTLCExecutionB{}
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

	he1bOutputs, err := getInputsForNextTxByParseTxHashVout(he1b.TxHex, he1b.OutputAddress, he1b.OutAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return err
	}

	result, err = rpcClient.OmniDecodeTransactionWithPrevTxs(signedHerd1bHex, he1bOutputs)
	if err != nil {
		log.Println(err)
		return err
	}

	hexJsonObj := gjson.Parse(result)
	if he1b.OutputAddress != hexJsonObj.Get("sendingaddress").String() {
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

	hed1a := dao.HTLCExecutionDeliveryOfR{}
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
