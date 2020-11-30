package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"sync"
	"time"
)

//close htlc or close channel
type htlcCloseTxManager struct {
	operationFlag sync.Mutex
	//在步骤1，缓存需要发往47号协议的信息
	tempDataSendTo49PAtAliceSide map[string]bean.AliceRequestCloseHtlcCurrTxOfP2p
	tempDataSendTo50PAtBobSide   map[string]bean.CloseeSignCloseHtlcTxOfP2p
	tempDataFromTo50PAtAliceSide map[string]bean.CloseeSignCloseHtlcTxOfP2p
	tempDataFromTo51PAtBobSide   map[string]bean.AliceSignedC4bTxDataP2p
}

// htlc 关闭当前htlc交易
var HtlcCloseTxService htlcCloseTxManager

// step1 Alice 100049 请求关闭htlc交易 -100049 request close htlc
func (service *htlcCloseTxManager) RequestCloseHtlc(msg bean.RequestMessage, user bean.User) (data interface{}, needSign bool, err error) {
	log.Println("htlc close step 1 begin", time.Now())

	if tool.CheckIsString(&msg.Data) == false {
		return nil, false, errors.New(enum.Tips_common_empty + "msg data")
	}

	reqData := &bean.HtlcCloseRequestCurrTx{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	// region check data
	if tool.CheckIsString(&reqData.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id")
		log.Println(err)
		return nil, false, err
	}

	if tool.CheckIsString(&reqData.CurrTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_temp_address_pub_key")
		log.Println(err)
		return nil, false, err
	}
	if tool.CheckIsString(&reqData.LastHtlcTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "last_htlc_temp_address_private_key")
		log.Println(err)
		return nil, false, err
	}
	if tool.CheckIsString(&reqData.LastRsmcTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "last_rsmc_temp_address_private_key")
		log.Println(err)
		return nil, false, err
	}
	// endregion

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	if channelInfo.CurrState != dao.ChannelState_HtlcTx {
		return nil, false, errors.New(fmt.Sprintf(enum.Tips_htlc_wrongChannelState, channelInfo.CurrState, dao.ChannelState_HtlcTx))
	}

	targetUser := channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	}
	if msg.RecipientUserPeerId != targetUser {
		return nil, false, errors.New(enum.Tips_rsmc_notTargetUser)
	}

	if err := findUserIsOnline(msg.RecipientNodePeerId, targetUser); err != nil {
		return nil, false, err
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, reqData.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Init {
		tx.DeleteStruct(latestCommitmentTxInfo)
		latestCommitmentTxInfo, err = getLatestCommitmentTxUseDbTx(tx, reqData.ChannelId, user.PeerId)
		if err != nil {
			return nil, false, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
		}
	}

	//如果是第一次发起的关闭请求
	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Htlc {
		if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Htlc_GetH &&
			latestCommitmentTxInfo.CurrState != dao.TxInfoState_Htlc_GetR {
			return nil, false, errors.New(enum.Tips_channel_wrongLatestCommitmentTxState + ": should be 11 or 12")
		}
	}

	//如果是第二次发起的请求，前面的请求失败了
	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Rsmc {
		if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
			return nil, false, errors.New(enum.Tips_channel_wrongLatestCommitmentTxState + ": should be 5")
		}
	}

	ht1aOrHe1b := dao.HTLCTimeoutTxForAAndExecutionForB{}
	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Htlc {
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastRsmcTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey)
		if err != nil {
			return nil, false, err
		}
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressPrivateKey, latestCommitmentTxInfo.HTLCTempAddressPubKey)
		if err != nil {
			return nil, false, err
		}

		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Htlc_GetR {
			err = tx.Select(
				q.Eq("ChannelId", channelInfo.ChannelId),
				q.Eq("CommitmentTxId", latestCommitmentTxInfo.Id),
				q.Eq("Owner", user.PeerId)).
				First(&ht1aOrHe1b)
			if err != nil {
				log.Println(err)
				return nil, false, err
			}
		}
	} else {
		if reqData.CurrTempAddressPubKey != latestCommitmentTxInfo.RSMCTempAddressPubKey {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, reqData.CurrTempAddressPubKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
		}

		lastCommitmentTxInfo := &dao.CommitmentTransaction{}
		err = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, lastCommitmentTxInfo)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastRsmcTempAddressPrivateKey, lastCommitmentTxInfo.RSMCTempAddressPubKey)
		if err != nil {
			return nil, false, err
		}
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressPrivateKey, lastCommitmentTxInfo.HTLCTempAddressPubKey)
		if err != nil {
			return nil, false, err
		}
		if lastCommitmentTxInfo.CurrState == dao.TxInfoState_Htlc_GetR {
			err = tx.Select(
				q.Eq("ChannelId", channelInfo.ChannelId),
				q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
				q.Eq("Owner", user.PeerId)).
				First(&ht1aOrHe1b)
			if err != nil {
				log.Println(err)
				return nil, false, err
			}
		}
	}

	retSignData := bean.NeedAliceSignRsmcDataForC4a{}

	dataTo49P := bean.AliceRequestCloseHtlcCurrTxOfP2p{}
	if ht1aOrHe1b.Id > 0 {
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressForHtnxPrivateKey, ht1aOrHe1b.RSMCTempAddressPubKey)
		if err != nil {
			return nil, false, err
		}
		dataTo49P.LastHtlcTempAddressForHtnxPrivateKey = reqData.LastHtlcTempAddressForHtnxPrivateKey
	}

	dataTo49P.ChannelId = channelInfo.ChannelId
	dataTo49P.LastRsmcTempAddressPrivateKey = reqData.LastRsmcTempAddressPrivateKey
	dataTo49P.LastHtlcTempAddressPrivateKey = reqData.LastHtlcTempAddressPrivateKey
	dataTo49P.CurrTempAddressPubKey = reqData.CurrTempAddressPubKey
	dataTo49P.SenderNodeAddress = msg.SenderNodePeerId
	dataTo49P.SenderPeerId = msg.SenderUserPeerId

	needSign = false
	//如果是第一次请求
	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Htlc {
		//创建c2a omni的交易不能一个输入，多个输出，所以就是两个交易
		reqTempData := &bean.RequestCreateCommitmentTx{}
		reqTempData.CurrTempAddressIndex = reqData.CurrTempAddressIndex
		reqTempData.CurrTempAddressPubKey = reqData.CurrTempAddressPubKey
		reqTempData.Amount = 0
		newCommitmentTxInfo, rawTx, err := createCommitmentTxHex(tx, true, reqTempData, channelInfo, latestCommitmentTxInfo, user)
		if err != nil {
			return nil, false, err
		}

		newCommitmentTxInfo.CurrState = dao.TxInfoState_Init
		_ = tx.UpdateField(newCommitmentTxInfo, "CurrState", dao.TxInfoState_Init)

		dataTo49P.CommitmentTxHash = newCommitmentTxInfo.CurrHash
		dataTo49P.RsmcPartialSignedData = rawTx.RsmcRawTxData
		dataTo49P.CounterpartyPartialSignedData = rawTx.ToCounterpartyRawTxData
		needSign = true

	} else { //	上一次的请求出现异常，再次发起请求
		dataTo49P.CommitmentTxHash = latestCommitmentTxInfo.CurrHash

		if len(latestCommitmentTxInfo.RSMCTxid) == 0 {
			rawTx := &dao.CommitmentTxRawTx{}
			tx.Select(q.Eq("CommitmentTxId", latestCommitmentTxInfo.Id)).First(rawTx)
			if rawTx.Id == 0 {
				return nil, false, errors.New("not found rawTx")
			}
			dataTo49P.RsmcPartialSignedData = rawTx.RsmcRawTxData
			dataTo49P.CounterpartyPartialSignedData = rawTx.ToCounterpartyRawTxData
			needSign = true
		}
	}
	_ = tx.Commit()

	if needSign {

		retSignData.ChannelId = channelInfo.ChannelId
		retSignData.C4aRsmcRawData = dataTo49P.RsmcPartialSignedData
		retSignData.C4aCounterpartyRawData = dataTo49P.CounterpartyPartialSignedData

		if service.tempDataSendTo49PAtAliceSide == nil {
			service.tempDataSendTo49PAtAliceSide = make(map[string]bean.AliceRequestCloseHtlcCurrTxOfP2p)
		}
		service.tempDataSendTo49PAtAliceSide[user.PeerId+"_"+dataTo49P.ChannelId] = dataTo49P

		log.Println("htlc close step 1 end", time.Now())
		return retSignData, true, nil
	}

	return dataTo49P, false, nil
}

// step2 Alice 100100 Alice签名Cxa 并推送49号协议
func (service *htlcCloseTxManager) OnAliceSignedCxa(msg bean.RequestMessage, user bean.User) (toALice, toBob interface{}, err error) {
	log.Println("htlc close step 2 begin", time.Now())
	if tool.CheckIsString(&msg.Data) == false {
		err = errors.New(enum.Tips_common_empty + "msg.data")
		log.Println(err)
		return nil, nil, err
	}
	signedData := bean.AliceSignedRsmcDataForC4a{}
	_ = json.Unmarshal([]byte(msg.Data), &signedData)

	if tool.CheckIsString(&signedData.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id")
		log.Println(err)
		return nil, nil, err
	}

	p2pData := service.tempDataSendTo49PAtAliceSide[user.PeerId+"_"+signedData.ChannelId]
	if &p2pData == nil {
		return nil, nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}

	if tool.CheckIsString(&signedData.RsmcPartialSignedHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedData.RsmcPartialSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "rsmc_partial_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	if tool.CheckIsString(&signedData.CounterpartyPartialSignedHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedData.CounterpartyPartialSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "counterparty_partial_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, signedData.ChannelId, user.PeerId)
	if err != nil {
		return nil, nil, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
	}

	if len(latestCommitmentTxInfo.RSMCTxHex) > 0 {
		//封装好的签名数据，给bob的客户端签名使用
		latestCommitmentTxInfo.RSMCTxHex = signedData.CounterpartyPartialSignedHex
		latestCommitmentTxInfo.RSMCTxid = rpcClient.GetTxId(signedData.CounterpartyPartialSignedHex)
	}

	if len(latestCommitmentTxInfo.ToCounterpartyTxHex) > 0 {
		//封装好的签名数据，给bob的客户端签名使用
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedData.CounterpartyPartialSignedHex
		latestCommitmentTxInfo.ToCounterpartyTxid = rpcClient.GetTxId(signedData.CounterpartyPartialSignedHex)
	}

	p2pData.RsmcPartialSignedData.Hex = signedData.RsmcPartialSignedHex
	p2pData.CounterpartyPartialSignedData.Hex = signedData.CounterpartyPartialSignedHex

	latestCommitmentTxInfo.CurrState = dao.TxInfoState_Create
	_ = tx.Update(latestCommitmentTxInfo)

	_ = tx.Commit()

	toAliceResult := bean.AliceSignedRsmcDataForC2aResult{}
	toAliceResult.ChannelId = p2pData.ChannelId
	toAliceResult.CurrTempAddressPubKey = p2pData.CurrTempAddressPubKey
	toAliceResult.CommitmentTxHash = p2pData.CommitmentTxHash
	toAliceResult.Amount = latestCommitmentTxInfo.AmountToHtlc
	log.Println("htlc close step 2 end", time.Now())
	return toAliceResult, p2pData, nil
}

//step3 obd 110049 推送p2p消息给bob
func (service *htlcCloseTxManager) OnObdOfBobGet49PData(data string, user bean.User) (toBob interface{}, err error) {
	log.Println("htlc close step 3 begin", time.Now())
	closeHtlcTxOfP2p := bean.AliceRequestCloseHtlcCurrTxOfP2p{}
	_ = json.Unmarshal([]byte(data), &closeHtlcTxOfP2p)

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", closeHtlcTxOfP2p.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if err != nil {
		return nil, err
	}
	if channelInfo == nil {
		return nil, errors.New("not found channelInfo at targetSide")
	}

	channelInfo.CurrState = dao.ChannelState_NewTx
	_ = tx.Update(channelInfo)

	senderPeerId := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		senderPeerId = channelInfo.PeerIdB
	}
	messageHash := messageService.saveMsgUseTx(tx, senderPeerId, user.PeerId, data)
	_ = tx.Commit()

	closeHtlcTxOfWs := &bean.AliceRequestCloseHtlcCurrTxOfP2pToBobClient{}
	closeHtlcTxOfWs.C4aRsmcPartialSignedData = closeHtlcTxOfP2p.RsmcPartialSignedData
	closeHtlcTxOfWs.C4aCounterpartyPartialSignedData = closeHtlcTxOfP2p.CounterpartyPartialSignedData
	closeHtlcTxOfWs.ChannelId = channelInfo.ChannelId
	closeHtlcTxOfWs.MsgHash = messageHash
	closeHtlcTxOfWs.SenderNodeAddress = closeHtlcTxOfP2p.SenderNodeAddress
	closeHtlcTxOfWs.SenderPeerId = closeHtlcTxOfP2p.SenderPeerId
	log.Println("htlc close step 3 end", time.Now())
	return closeHtlcTxOfWs, nil
}

// step4 bob 100050 响应bob对这次关闭htlc交易的签收及他对Cxa的签名
func (service *htlcCloseTxManager) OnBobSignCloseHtlcRequest(msg bean.RequestMessage, user bean.User) (toBob interface{}, err error) {
	log.Println("htlc close step 4 begin", time.Now())
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New(enum.Tips_common_empty + "msg data")
	}
	reqData := &bean.HtlcBobSignCloseCurrTx{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// region check data
	if tool.CheckIsString(&reqData.MsgHash) == false {
		err = errors.New(enum.Tips_common_empty + "msg_hash")
		log.Println(err)
		return nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	message, err := messageService.getMsgUseTx(tx, reqData.MsgHash)
	if err != nil {
		return nil, errors.New(enum.Tips_common_wrong + "msg_hash")
	}
	if message.Receiver != user.PeerId {
		return nil, errors.New(enum.Tips_rsmc_notTargetUser)
	}

	dataFrom49POfP2p := bean.AliceRequestCloseHtlcCurrTxOfP2p{}
	_ = json.Unmarshal([]byte(message.Data), &dataFrom49POfP2p)

	if tool.CheckIsString(&reqData.CurrTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_temp_address_pub_key")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.LastHtlcTempAddressForHtnxPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "last_htlc_temp_address_for_htnx_private_key")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&reqData.LastHtlcTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "last_htlc_temp_address_private_key")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&reqData.LastRsmcTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "last_rsmc_temp_address_private_key")
		log.Println(err)
		return nil, err
	}
	// endregion

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", dataFrom49POfP2p.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if err != nil {
		return nil, err
	}

	signerPubKey := channelInfo.PubKeyB
	senderPeerId := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		senderPeerId = channelInfo.PeerIdB
		signerPubKey = channelInfo.PubKeyA
	}

	err = findUserIsOnline(msg.RecipientNodePeerId, senderPeerId)
	if err != nil {
		return nil, err
	}

	latestCommitmentTxInfo, _ := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//如果是第一次收到的关闭请求
	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Htlc {
		if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Htlc_GetH &&
			latestCommitmentTxInfo.CurrState != dao.TxInfoState_Htlc_GetR {
			return nil, errors.New(enum.Tips_channel_wrongLatestCommitmentTxState + ": should be 11 or 12")
		}
	}

	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Rsmc {
		if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
			//如果是第二次收到的请求，前面的请求有异常了
			return nil, errors.New(enum.Tips_channel_wrongLatestCommitmentTxState + ": should be 5")
		}
	}

	he1b := dao.HTLCTimeoutTxForAAndExecutionForB{}
	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Htlc {
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastRsmcTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey)
		if err != nil {
			return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongPrivateKeyForLast, reqData.LastRsmcTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
		}
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressPrivateKey, latestCommitmentTxInfo.HTLCTempAddressPubKey)
		if err != nil {
			return nil, err
		}
		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Htlc_GetR {
			err = tx.Select(
				q.Eq("ChannelId", channelInfo.ChannelId),
				q.Eq("CommitmentTxId", latestCommitmentTxInfo.Id),
				q.Eq("Owner", user.PeerId)).
				First(&he1b)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
	} else {
		if reqData.CurrTempAddressPubKey != latestCommitmentTxInfo.RSMCTempAddressPubKey {
			return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, reqData.CurrTempAddressPubKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
		}
		lastCommitTxInfo := dao.CommitmentTransaction{}
		err = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, &lastCommitTxInfo)
		if err != nil {
			return nil, err
		}
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastRsmcTempAddressPrivateKey, lastCommitTxInfo.RSMCTempAddressPubKey)
		if err != nil {
			return nil, err
		}
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressPrivateKey, lastCommitTxInfo.HTLCTempAddressPubKey)
		if err != nil {
			return nil, err
		}

		if lastCommitTxInfo.CurrState == dao.TxInfoState_Htlc_GetR {
			err = tx.Select(
				q.Eq("ChannelId", channelInfo.ChannelId),
				q.Eq("CommitmentTxId", lastCommitTxInfo.Id),
				q.Eq("Owner", user.PeerId)).
				First(&he1b)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
	}

	dataSendToBob := bean.NeedBobSignRawDataForC4b{}
	dataSendToBob.ChannelId = channelInfo.ChannelId
	dataSendToBob.SenderNodeAddress = user.P2PLocalPeerId
	dataSendToBob.SenderPeerId = user.PeerId

	dataSendTo50P := bean.CloseeSignCloseHtlcTxOfP2p{}
	if he1b.Id > 0 {
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressForHtnxPrivateKey, he1b.RSMCTempAddressPubKey)
		if err != nil {
			return nil, err
		}
		dataSendTo50P.CloseeLastHtlcTempAddressForHtnxPrivateKey = reqData.LastHtlcTempAddressForHtnxPrivateKey
	}

	dataSendTo50P.ChannelId = channelInfo.ChannelId
	dataSendTo50P.SendeeNodeAddress = user.P2PLocalPeerId
	dataSendTo50P.SendeePeerId = user.PeerId
	dataSendTo50P.CloserCommitmentTxHash = dataFrom49POfP2p.CommitmentTxHash
	dataSendTo50P.CloseeLastRsmcTempAddressPrivateKey = reqData.LastRsmcTempAddressPrivateKey
	dataSendTo50P.CloseeLastHtlcTempAddressPrivateKey = reqData.LastHtlcTempAddressPrivateKey
	dataSendTo50P.CloseeCurrRsmcTempAddressPubKey = reqData.CurrTempAddressPubKey

	// get the funding transaction
	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, user.PeerId)
	if fundingTransaction == nil {
		return nil, errors.New(enum.Tips_funding_notFoundFundAssetTx)
	}

	// region 签名requester的承诺交易
	// 签名对方传过来的rsmcHex
	var signedRsmcHex, aliceRsmcTxId string
	var aliceRsmcMultiAddress, aliceRsmcRedeemScript, aliceRsmcMultiAddressScriptPubKey string
	var aliceRsmcOutputs []rpc.TransactionInputItem
	if tool.CheckIsString(&dataFrom49POfP2p.RsmcPartialSignedData.Hex) {
		signedRsmcHex = reqData.C4aRsmcCompleteSignedHex
		if pass, _ := rpcClient.CheckMultiSign(true, signedRsmcHex, 2); pass == false {
			return nil, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "c4a_rsmc_complete_signed_hex"))
		}
		aliceRsmcTxId = rpcClient.GetTxId(signedRsmcHex)

		//  根据alice的临时地址+bob的通道address,获取alice2+bob的多签地址，并得到AliceSignedRsmcHex签名后的交易的input，为创建alice的RD和bob的BR做准备
		aliceRsmcMultiAddress, aliceRsmcRedeemScript, aliceRsmcMultiAddressScriptPubKey, err = createMultiSig(dataFrom49POfP2p.CurrTempAddressPubKey, signerPubKey)
		if err != nil {
			return nil, err
		}

		aliceRsmcOutputs, err = getInputsForNextTxByParseTxHashVout(signedRsmcHex, aliceRsmcMultiAddress, aliceRsmcMultiAddressScriptPubKey, aliceRsmcRedeemScript)
		if err != nil || len(aliceRsmcOutputs) == 0 {
			log.Println(err)
			return nil, err
		}
	}
	dataSendTo50P.C4aRsmcCompleteSignedHex = signedRsmcHex

	//  签名对方传过来的toOtherHex
	if tool.CheckIsString(&dataFrom49POfP2p.CounterpartyPartialSignedData.Hex) {
		signedToOtherHex := reqData.C4aCounterpartyCompleteSignedHex
		if pass, _ := rpcClient.CheckMultiSign(true, signedToOtherHex, 2); pass == false {
			return nil, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "c4a_counterparty_complete_signed_hex"))
		}
		dataSendTo50P.C4aCounterpartyCompleteSignedHex = signedToOtherHex
	}
	//endregion

	//第一次签名，不是失败后的重试
	var amountToCounterparty = 0.0
	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Htlc {
		//region 3、根据对方传过来的上一个交易的临时rsmc私钥，签名最近的BR交易，保证对方确实放弃了上一个承诺交易
		err = signLastBR(tx, dao.BRType_Rmsc, *channelInfo, user.PeerId, dataFrom49POfP2p.LastRsmcTempAddressPrivateKey, latestCommitmentTxInfo.Id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = signLastBR(tx, dao.BRType_Htlc, *channelInfo, user.PeerId, dataFrom49POfP2p.LastHtlcTempAddressPrivateKey, latestCommitmentTxInfo.Id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = signLastBR(tx, dao.BRType_Ht1a, *channelInfo, user.PeerId, dataFrom49POfP2p.LastHtlcTempAddressForHtnxPrivateKey, latestCommitmentTxInfo.Id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		//endregion

		//region 4、创建C3b
		commitmentTxRequest := &bean.RequestCreateCommitmentTx{}
		commitmentTxRequest.ChannelId = channelInfo.ChannelId
		commitmentTxRequest.Amount = 0
		commitmentTxRequest.CurrTempAddressIndex = reqData.CurrTempAddressIndex
		commitmentTxRequest.CurrTempAddressPubKey = reqData.CurrTempAddressPubKey
		newCommitmentTxInfo, rawTx, err := createCommitmentTxHex(tx, false, commitmentTxRequest, channelInfo, latestCommitmentTxInfo, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		amountToCounterparty = newCommitmentTxInfo.AmountToCounterparty

		dataSendTo50P.C4bRsmcPartialSignedData = rawTx.RsmcRawTxData
		dataSendTo50P.C4bCounterpartyPartialSignedData = rawTx.ToCounterpartyRawTxData
		//endregion

		// region 5、根据alice的Rsmc，创建对应的BR,为下一个交易做准备，create BR2b tx  for bob
		if len(aliceRsmcOutputs) > 0 {
			var myAddress = channelInfo.AddressB
			if user.PeerId == channelInfo.PeerIdA {
				myAddress = channelInfo.AddressA
			}
			senderCommitmentTx := &dao.CommitmentTransaction{}
			senderCommitmentTx.Id = newCommitmentTxInfo.Id
			senderCommitmentTx.PropertyId = fundingTransaction.PropertyId
			senderCommitmentTx.RSMCTempAddressPubKey = dataFrom49POfP2p.CurrTempAddressPubKey
			senderCommitmentTx.RSMCMultiAddress = aliceRsmcMultiAddress
			senderCommitmentTx.RSMCRedeemScript = aliceRsmcRedeemScript
			senderCommitmentTx.RSMCMultiAddressScriptPubKey = aliceRsmcMultiAddressScriptPubKey
			senderCommitmentTx.RSMCTxHex = signedRsmcHex
			senderCommitmentTx.RSMCTxid = aliceRsmcTxId
			senderCommitmentTx.AmountToRSMC = newCommitmentTxInfo.AmountToCounterparty
			cnaBrHexData, err := createCurrCommitmentTxRawBR(tx, dao.BRType_Rmsc, channelInfo, senderCommitmentTx, aliceRsmcOutputs, myAddress, user)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err != nil {
				log.Println(err)
				return nil, err
			}
			cnaBrRawData := bean.NeedClientSignRawBRTxData{}
			cnaBrRawData.Hex = cnaBrHexData["hex"].(string)
			cnaBrRawData.Inputs = cnaBrHexData["inputs"]
			cnaBrRawData.BrId = cnaBrHexData["br_id"].(int)
			cnaBrRawData.IsMultisig = true
			cnaBrRawData.PubKeyA = dataFrom49POfP2p.CurrTempAddressPubKey
			cnaBrRawData.PubKeyB = signerPubKey
			dataSendToBob.C4aBrRawData = cnaBrRawData
		}
		//endregion
	} else {
		rawTx := &dao.CommitmentTxRawTx{}
		tx.Select(q.Eq("CommitmentTxId", latestCommitmentTxInfo.Id)).First(rawTx)
		if rawTx.Id == 0 {
			return nil, errors.New("not found rawTx")
		}
		dataSendTo50P.C4bRsmcPartialSignedData = rawTx.RsmcRawTxData
		dataSendTo50P.C4bCounterpartyPartialSignedData = rawTx.ToCounterpartyRawTxData
		amountToCounterparty = latestCommitmentTxInfo.AmountToCounterparty
	}

	//region 6、根据签名后的AliceRsmc创建alice的RD create RD tx for alice
	if len(aliceRsmcOutputs) > 0 {
		outputAddress := channelInfo.AddressA
		if user.PeerId == channelInfo.PeerIdA {
			outputAddress = channelInfo.AddressB
		}
		senderRdTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
			aliceRsmcMultiAddress,
			aliceRsmcOutputs,
			outputAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			amountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&aliceRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_failToCreate, "rd raw transaction"))
		}
		c2aRdRawData := bean.NeedClientSignTxData{}
		c2aRdRawData.Hex = senderRdTx["hex"].(string)
		c2aRdRawData.Inputs = senderRdTx["inputs"]
		c2aRdRawData.PubKeyA = dataFrom49POfP2p.CurrTempAddressPubKey
		c2aRdRawData.PubKeyB = signerPubKey
		c2aRdRawData.IsMultisig = true
		dataSendTo50P.C4aRdPartialSignedData = c2aRdRawData
		dataSendToBob.C4aRdRawData = c2aRdRawData
	}
	//endregion create RD tx for alice

	channelInfo.CurrState = dao.ChannelState_NewTx
	_ = tx.Update(channelInfo)

	_ = messageService.updateMsgStateUseTx(tx, message)
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// 缓存数据
	if service.tempDataSendTo50PAtBobSide == nil {
		service.tempDataSendTo50PAtBobSide = make(map[string]bean.CloseeSignCloseHtlcTxOfP2p)
	}
	service.tempDataSendTo50PAtBobSide[user.PeerId+"_"+channelInfo.ChannelId] = dataSendTo50P

	dataSendToBob.C4bRsmcRawData = dataSendTo50P.C4bRsmcPartialSignedData
	dataSendToBob.C4bCounterpartyRawData = dataSendTo50P.C4bCounterpartyPartialSignedData
	log.Println("htlc close step 4 end", time.Now())
	return dataSendToBob, nil
}

// step5 bob 100111 响应bob对Cxb的签名，并推送50号协议
func (service *htlcCloseTxManager) OnBobSignedCxb(msg bean.RequestMessage, user bean.User) (toAlice, toBob interface{}, err error) {
	log.Println("htlc close step 5 begin", time.Now())
	if tool.CheckIsString(&msg.Data) == false {
		err = errors.New(enum.Tips_common_empty + "msg.data")
		log.Println(err)
		return nil, nil, err
	}

	signedData := bean.BobSignedRsmcDataForC4b{}
	_ = json.Unmarshal([]byte(msg.Data), &signedData)

	if tool.CheckIsString(&signedData.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id")
		log.Println(err)
		return nil, nil, err
	}

	//得到step4缓存的数据
	p2pData := service.tempDataSendTo50PAtBobSide[user.PeerId+"_"+signedData.ChannelId]
	if len(p2pData.ChannelId) == 0 {
		return nil, nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}

	if tool.CheckIsString(&p2pData.C4bRsmcPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedData.C4bRsmcPartialSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c4b_rsmc_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	if tool.CheckIsString(&p2pData.C4bCounterpartyPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedData.C4bCounterpartyPartialSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c2b_counterparty_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	if tool.CheckIsString(&p2pData.C4aRdPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(false, signedData.C4aRdPartialSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c4a_rd_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	if tool.CheckIsString(&signedData.C4aBrPartialSignedHex) {
		if pass, _ := rpcClient.CheckMultiSign(false, signedData.C4aBrPartialSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "c2a_br_signed_hex")
			log.Println(err)
			return nil, nil, err
		}

		if signedData.C4aBrId == 0 {
			err = errors.New(enum.Tips_common_wrong + "c4a_br_id")
			log.Println(err)
			return nil, nil, err
		}
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, signedData.ChannelId, user.PeerId)
	if err != nil {
		return nil, nil, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
	}

	senderPeerId := latestCommitmentTxInfo.PeerIdA
	if user.PeerId == latestCommitmentTxInfo.PeerIdA {
		senderPeerId = latestCommitmentTxInfo.PeerIdB
	}
	if senderPeerId != msg.RecipientUserPeerId {
		return nil, nil, errors.New(enum.Tips_common_userNotInTx)
	}

	if len(signedData.C4bRsmcPartialSignedHex) > 0 {
		latestCommitmentTxInfo.RSMCTxHex = signedData.C4bRsmcPartialSignedHex
		latestCommitmentTxInfo.RSMCTxid = rpcClient.GetTxId(signedData.C4bRsmcPartialSignedHex)
	}

	if len(signedData.C4bCounterpartyPartialSignedHex) > 0 {
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedData.C4bCounterpartyPartialSignedHex
		latestCommitmentTxInfo.ToCounterpartyTxid = rpcClient.GetTxId(signedData.C4bCounterpartyPartialSignedHex)
	}

	if len(signedData.C4aBrPartialSignedHex) > 0 {
		err = updateCurrCommitmentTxRawBR(tx, signedData.C4aBrId, signedData.C4aBrPartialSignedHex, user)
		if err != nil {
			return nil, nil, err
		}
	}
	_ = tx.Update(latestCommitmentTxInfo)

	_ = tx.Commit()

	p2pData.C4bRsmcPartialSignedData.Hex = signedData.C4bRsmcPartialSignedHex
	p2pData.C4bCounterpartyPartialSignedData.Hex = signedData.C4bCounterpartyPartialSignedHex
	p2pData.C4aRdPartialSignedData.Hex = signedData.C4aRdPartialSignedHex

	toBobData := bean.BobSignedRsmcDataForC4bResult{}
	toBobData.ChannelId = p2pData.ChannelId
	toBobData.CommitmentTxHash = latestCommitmentTxInfo.CurrHash
	log.Println("htlc close step 5 end", time.Now())
	return p2pData, toBobData, nil
}

// step6 obd 110050 推送p2p消息给Alice
func (service *htlcCloseTxManager) OnObdOfAliceGet50PData(data string, user bean.User) (toAlice interface{}, needNoticeBob bool, err error) {
	log.Println("htlc close step 6 begin", time.Now())
	dataFrom50P := bean.CloseeSignCloseHtlcTxOfP2p{}
	_ = json.Unmarshal([]byte(data), &dataFrom50P)

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", dataFrom50P.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	if channelInfo.CurrState != dao.ChannelState_HtlcTx {
		return nil, false, errors.New(fmt.Sprintf(enum.Tips_htlc_wrongChannelState, channelInfo.CurrState, dao.ChannelState_HtlcTx))
	}

	channelInfo.CurrState = dao.ChannelState_NewTx
	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	if service.tempDataFromTo50PAtAliceSide == nil {
		service.tempDataFromTo50PAtAliceSide = make(map[string]bean.CloseeSignCloseHtlcTxOfP2p)
	}
	service.tempDataFromTo50PAtAliceSide[user.PeerId+"_"+dataFrom50P.ChannelId] = dataFrom50P

	needAliceSignData := bean.NeedAliceSignRsmcTxForC4b{}
	needAliceSignData.ChannelId = dataFrom50P.ChannelId
	needAliceSignData.C4aRdPartialSignedData = dataFrom50P.C4aRdPartialSignedData
	needAliceSignData.C4bRsmcPartialSignedData = dataFrom50P.C4bRsmcPartialSignedData
	needAliceSignData.C4bCounterpartyPartialSignedData = dataFrom50P.C4bCounterpartyPartialSignedData
	needAliceSignData.SendeeNodeAddress = dataFrom50P.SendeeNodeAddress
	needAliceSignData.SendeePeerId = dataFrom50P.SendeePeerId
	log.Println("htlc close step 6 end", time.Now())
	return needAliceSignData, false, nil
}

// step7 alice 100112 Alice完成对Cxb的签名
func (service *htlcCloseTxManager) OnAliceSignedCxb(msg bean.RequestMessage, user bean.User) (toAlice interface{}, err error) {
	log.Println("htlc close step 7 begin", time.Now())
	aliceSignedData := bean.AliceSignedRsmcTxForC4b{}
	_ = json.Unmarshal([]byte(msg.Data), &aliceSignedData)

	dataFromP2p50P := service.tempDataFromTo50PAtAliceSide[user.PeerId+"_"+aliceSignedData.ChannelId]
	if len(dataFromP2p50P.ChannelId) == 0 {
		return nil, errors.New(enum.Tips_common_empty + "channel_id")
	}

	if tool.CheckIsString(&dataFromP2p50P.C4bRsmcPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, aliceSignedData.C4bRsmcCompleteSignedHex, 2); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c4b_rsmc_complete_signed_hex")
			log.Println(err)
			return nil, err
		}
	}
	dataFromP2p50P.C4bRsmcPartialSignedData.Hex = aliceSignedData.C4bRsmcCompleteSignedHex

	if tool.CheckIsString(&dataFromP2p50P.C4bCounterpartyPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, aliceSignedData.C4bCounterpartyCompleteSignedHex, 2); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed C4bCounterpartyCompleteSignedHex")
			log.Println(err)
			return nil, err
		}
	}
	dataFromP2p50P.C4bCounterpartyPartialSignedData.Hex = aliceSignedData.C4bCounterpartyCompleteSignedHex

	if tool.CheckIsString(&dataFromP2p50P.C4aRdPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedData.C4aRdCompleteSignedHex, 2); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c4a_rd_complete_signed_hex")
			log.Println(err)
			return nil, err
		}
	}
	dataFromP2p50P.C4aRdPartialSignedData.Hex = aliceSignedData.C4aRdCompleteSignedHex

	service.tempDataFromTo50PAtAliceSide[user.PeerId+"_"+dataFromP2p50P.ChannelId] = dataFromP2p50P

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	needAliceSignRdTxForC4b := bean.NeedAliceSignRdTxForC4b{}

	channelInfo := getChannelInfoByChannelId(tx, dataFromP2p50P.ChannelId, user.PeerId)
	if channelInfo == nil {
		return nil, errors.New("not found channelInfo at targetSide")
	}
	needAliceSignRdTxForC4b.ChannelId = channelInfo.ChannelId

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find sender's commitmentTxInfo")
		log.Println(err)
		return nil, err
	}

	if latestCommitmentTxInfo.CurrHash != dataFromP2p50P.CloserCommitmentTxHash {
		err = errors.New("wrong request hash, Please notice closee,")
		log.Println(err)
		return nil, err
	}

	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
		err = errors.New("wrong commitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, err
	}

	var myChannelPubKey = channelInfo.PubKeyA
	var myChannelAddress = channelInfo.AddressA
	var partnerChannelAddress = channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdB {
		myChannelAddress = channelInfo.AddressB
		myChannelPubKey = channelInfo.PubKeyB
		partnerChannelAddress = channelInfo.AddressA
	}

	//处理对方的数据
	//签名对方传过来的rsmcHex
	bobSignedRsmcHex := aliceSignedData.C4bRsmcCompleteSignedHex

	//region create RD tx for bob
	bobMultiAddr, err := rpcClient.CreateMultiSig(2, []string{dataFromP2p50P.CloseeCurrRsmcTempAddressPubKey, myChannelPubKey})
	if err != nil {
		return nil, err
	}
	cnbRsmcMultiAddress := gjson.Get(bobMultiAddr, "address").String()
	cnbRsmcRedeemScript := gjson.Get(bobMultiAddr, "redeemScript").String()
	cnbRsmcMultiAddressScriptPubKey := gjson.Get(bobMultiAddr, "scriptPubKey").String()

	if tool.CheckIsString(&bobSignedRsmcHex) {
		rsmcOutputs, err := getInputsForNextTxByParseTxHashVout(
			bobSignedRsmcHex,
			cnbRsmcMultiAddress,
			cnbRsmcMultiAddressScriptPubKey,
			cnbRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		c2bRdHexData, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
			cnbRsmcMultiAddress,
			rsmcOutputs,
			partnerChannelAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			latestCommitmentTxInfo.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&cnbRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, errors.New("fail to create rd")
		}
		rdRawData := bean.NeedClientSignTxData{}
		rdRawData.Hex = c2bRdHexData["hex"].(string)
		rdRawData.Inputs = c2bRdHexData["inputs"]
		rdRawData.IsMultisig = true
		rdRawData.PubKeyA = dataFromP2p50P.CloseeCurrRsmcTempAddressPubKey
		rdRawData.PubKeyB = myChannelPubKey
		needAliceSignRdTxForC4b.C4bRdRawData = rdRawData
		//endregion create RD tx for bob

		//region 根据对对方的Rsmc签名，生成惩罚对方，自己获益BR
		bobCommitmentTx := &dao.CommitmentTransaction{}
		bobCommitmentTx.Id = latestCommitmentTxInfo.Id
		bobCommitmentTx.PropertyId = channelInfo.PropertyId
		bobCommitmentTx.RSMCTempAddressPubKey = dataFromP2p50P.CloseeCurrRsmcTempAddressPubKey
		bobCommitmentTx.RSMCMultiAddress = cnbRsmcMultiAddress
		bobCommitmentTx.RSMCRedeemScript = cnbRsmcRedeemScript
		bobCommitmentTx.RSMCMultiAddressScriptPubKey = cnbRsmcMultiAddressScriptPubKey
		bobCommitmentTx.RSMCTxHex = bobSignedRsmcHex
		bobCommitmentTx.RSMCTxid = rpcClient.GetTxId(dataFromP2p50P.C4bRsmcPartialSignedData.Hex)
		bobCommitmentTx.AmountToRSMC = latestCommitmentTxInfo.AmountToCounterparty
		c2bBrHexData, err := createCurrCommitmentTxRawBR(tx, dao.BRType_Rmsc, channelInfo, bobCommitmentTx, rsmcOutputs, myChannelAddress, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		brRawData := bean.NeedClientSignRawBRTxData{}
		brRawData.Hex = c2bBrHexData["hex"].(string)
		brRawData.Inputs = c2bBrHexData["inputs"]
		brRawData.BrId = c2bBrHexData["br_id"].(int)
		brRawData.IsMultisig = true
		brRawData.PubKeyA = dataFromP2p50P.CloseeCurrRsmcTempAddressPubKey
		brRawData.PubKeyB = myChannelPubKey
		needAliceSignRdTxForC4b.C4bBrRawData = brRawData
	}
	//endregion

	_ = tx.Commit()
	needAliceSignRdTxForC4b.SendeeNodeAddress = dataFromP2p50P.SendeeNodeAddress
	needAliceSignRdTxForC4b.SendeePeerId = dataFromP2p50P.SendeePeerId
	log.Println("htlc close step 7 end", time.Now())
	return needAliceSignRdTxForC4b, nil
}

// step8 alice 100113 Alice完成对Cxb的Rd和Br的签名 并推送51号协议
func (service *htlcCloseTxManager) OnAliceSignedCxbBubTx(msg bean.RequestMessage, user bean.User) (toAlice, toBob interface{}, err error) {
	log.Println("htlc close step 8 begin", time.Now())
	aliceSignedRdTxForCnb := bean.AliceSignedRdTxForC4b{}
	_ = json.Unmarshal([]byte(msg.Data), &aliceSignedRdTxForCnb)

	dataFromP2p50P := service.tempDataFromTo50PAtAliceSide[user.PeerId+"_"+aliceSignedRdTxForCnb.ChannelId]
	if len(dataFromP2p50P.ChannelId) == 0 {
		return nil, nil, errors.New(enum.Tips_common_empty + "channel_id")
	}

	//region 检测传入数据

	var channelId = dataFromP2p50P.ChannelId
	if tool.CheckIsString(&channelId) == false {
		err = errors.New("wrong channelId")
		log.Println(err)
		return nil, nil, err
	}

	if tool.CheckIsString(&dataFromP2p50P.CloserCommitmentTxHash) == false {
		err = errors.New("wrong commitmentTxHash")
		log.Println(err)
		return nil, nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	channelInfo := getChannelInfoByChannelId(tx, channelId, user.PeerId)
	if channelInfo == nil {
		return nil, nil, errors.New("not found channelInfo at targetSide")
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find sender's commitmentTxInfo")
		log.Println(err)
		return nil, nil, err
	}

	if latestCommitmentTxInfo.CurrHash != dataFromP2p50P.CloserCommitmentTxHash {
		err = errors.New("wrong request hash, Please notice payee,")
		log.Println(err)
		return nil, nil, err
	}

	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
		err = errors.New("wrong commitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, nil, err
	}

	aliceData := make(map[string]interface{})
	aliceData["channel_id"] = dataFromP2p50P.ChannelId

	var c2aSignedRsmcHex = dataFromP2p50P.C4aRsmcCompleteSignedHex
	if tool.CheckIsString(&c2aSignedRsmcHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, c2aSignedRsmcHex, 2); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c4a_rsmc_complete_signed_hex")
		}
	}

	var signedToCounterpartyHex = dataFromP2p50P.C4aCounterpartyCompleteSignedHex
	if tool.CheckIsString(&signedToCounterpartyHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedToCounterpartyHex, 2); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c4a_counterparty_complete_signed_hex")
		}
	}

	var aliceRdHex = dataFromP2p50P.C4aRdPartialSignedData.Hex
	if tool.CheckIsString(&aliceRdHex) {
		if pass, _ := rpcClient.CheckMultiSign(false, aliceRdHex, 2); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c4a_rd_signed_hex")
		}
	}

	var bobRsmcHex = dataFromP2p50P.C4bRsmcPartialSignedData.Hex
	if tool.CheckIsString(&bobRsmcHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, bobRsmcHex, 2); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c4b_rsmc_signed_hex")
		}
	}

	var bobCurrTempAddressPubKey = dataFromP2p50P.CloseeCurrRsmcTempAddressPubKey
	if tool.CheckIsString(&bobCurrTempAddressPubKey) == false {
		err = errors.New("wrong curr_temp_address_pub_key")
		log.Println(err)
		return nil, nil, err
	}

	var c2bToCounterpartyTxHex = dataFromP2p50P.C4bCounterpartyPartialSignedData.Hex
	if len(c2bToCounterpartyTxHex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, c2bToCounterpartyTxHex, 2); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c4b_counterparty_tx_data_hex")
		}
	}
	//endregion

	fundingTransaction := getFundingTransactionByChannelId(tx, channelId, user.PeerId)
	if fundingTransaction == nil {
		return nil, nil, errors.New("not found fundingTransaction at targetSide")
	}

	var myChannelPubKey = channelInfo.PubKeyA
	var myChannelAddress = channelInfo.AddressA
	var partnerChannelAddress = channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdB {
		myChannelAddress = channelInfo.AddressB
		myChannelPubKey = channelInfo.PubKeyB
		partnerChannelAddress = channelInfo.AddressA
	}

	//region 根据对方传过来的上一个交易的临时rsmc私钥，签名上一次的BR交易，保证对方确实放弃了上一个承诺交易
	var bobLastRsmcTempAddressPrivateKey = dataFromP2p50P.CloseeLastRsmcTempAddressPrivateKey
	var bobLastHtlcTempAddressPrivateKey = dataFromP2p50P.CloseeLastHtlcTempAddressPrivateKey
	var bobLastHtlcTempAddressForHtnxPrivateKey = dataFromP2p50P.CloseeLastHtlcTempAddressForHtnxPrivateKey
	err = signLastBR(tx, dao.BRType_Rmsc, *channelInfo, user.PeerId, bobLastRsmcTempAddressPrivateKey, latestCommitmentTxInfo.LastCommitmentTxId)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	err = signLastBR(tx, dao.BRType_Htlc, *channelInfo, user.PeerId, bobLastHtlcTempAddressPrivateKey, latestCommitmentTxInfo.LastCommitmentTxId)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	err = signLastBR(tx, dao.BRType_HE1b, *channelInfo, user.PeerId, bobLastHtlcTempAddressForHtnxPrivateKey, latestCommitmentTxInfo.LastCommitmentTxId)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	//endregion

	if tool.CheckIsString(&c2aSignedRsmcHex) {
		latestCommitmentTxInfo.RSMCTxHex = c2aSignedRsmcHex
		latestCommitmentTxInfo.RSMCTxid = rpcClient.GetTxId(c2aSignedRsmcHex)

		// 保存Rd交易
		err = saveRdTx(tx, channelInfo, c2aSignedRsmcHex, aliceRdHex, latestCommitmentTxInfo, myChannelAddress, &user)
		if err != nil {
			return nil, nil, err
		}
	}

	latestCommitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	latestCommitmentTxInfo.SignAt = time.Now()

	if tool.CheckIsString(&signedToCounterpartyHex) {
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedToCounterpartyHex
		latestCommitmentTxInfo.ToCounterpartyTxid = rpcClient.GetTxId(signedToCounterpartyHex)
	}

	//重新生成交易id
	bytes, err := json.Marshal(latestCommitmentTxInfo)
	latestCommitmentTxInfo.CurrHash = tool.SignMsgWithSha256(bytes)
	_ = tx.Update(latestCommitmentTxInfo)

	lastCommitmentTxInfo := dao.CommitmentTransaction{}
	err = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, &lastCommitmentTxInfo)
	if err == nil {
		lastCommitmentTxInfo.CurrState = dao.TxInfoState_Abord
		_ = tx.Update(lastCommitmentTxInfo)
	}

	channelInfo.CurrState = dao.ChannelState_CanUse
	_ = tx.Update(channelInfo)

	//返回给alice的数据
	aliceData["latest_commitment_tx_info"] = latestCommitmentTxInfo

	//处理对方的数据
	bobData := bean.AliceSignedC4bTxDataP2p{}
	bobData.C4aCommitmentTxHash = dataFromP2p50P.CloserCommitmentTxHash

	c2bMultiAddr, err := rpcClient.CreateMultiSig(2, []string{bobCurrTempAddressPubKey, myChannelPubKey})
	if err != nil {
		return nil, nil, err
	}
	cnbRsmcMultiAddress := gjson.Get(c2bMultiAddr, "address").String()
	cnbRsmcRedeemScript := gjson.Get(c2bMultiAddr, "redeemScript").String()
	cnbRsmcMultiAddressScriptPubKey := gjson.Get(c2bMultiAddr, "scriptPubKey").String()

	//Cnb rsmc hex
	cnbSignedRsmcHex := dataFromP2p50P.C4bRsmcPartialSignedData.Hex
	if len(cnbSignedRsmcHex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, cnbSignedRsmcHex, 2); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c4b_rsmc_tx_data_hex")
		}
		err = checkBobRemcData(cnbSignedRsmcHex, cnbRsmcMultiAddress, latestCommitmentTxInfo)
		if err != nil {
			return nil, nil, err
		}
	}
	bobData.C4bRsmcCompleteSignedHex = cnbSignedRsmcHex

	//region create RD tx for bob

	if len(cnbSignedRsmcHex) > 0 {
		cnbRsmcOutputs, err := getInputsForNextTxByParseTxHashVout(
			cnbSignedRsmcHex,
			cnbRsmcMultiAddress,
			cnbRsmcMultiAddressScriptPubKey,
			cnbRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}

		c2bRdHexData, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
			cnbRsmcMultiAddress,
			cnbRsmcOutputs,
			partnerChannelAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			latestCommitmentTxInfo.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&cnbRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, nil, errors.New("fail to create rd")
		}
		cnbRdRawData := bean.NeedClientSignTxData{}
		cnbRdRawData.Hex = aliceSignedRdTxForCnb.C4bRdPartialSignedHex
		cnbRdRawData.Inputs = c2bRdHexData["inputs"]
		cnbRdRawData.IsMultisig = true
		cnbRdRawData.PubKeyA = dataFromP2p50P.CloseeCurrRsmcTempAddressPubKey
		cnbRdRawData.PubKeyB = myChannelPubKey
		bobData.C4bRdPartialSignedData = cnbRdRawData
		//endregion create RD tx for alice

		//region 根据对对方的Rsmc签名，生成惩罚对方，自己获益BR
		err = updateCurrCommitmentTxRawBR(tx, aliceSignedRdTxForCnb.C4bBrId, aliceSignedRdTxForCnb.C4bRdPartialSignedHex, user)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}
	}

	//endregion
	_ = tx.Commit()

	bobData.C4bCounterpartyCompleteSignedHex = c2bToCounterpartyTxHex
	bobData.ChannelId = channelId

	log.Println("htlc close step 8 end", time.Now())
	return aliceData, bobData, nil
}

// step9 obd 51 推送110051号协议消息到bob
func (service *htlcCloseTxManager) OnObdOfBobGet51PData(data string, user bean.User) (toBob interface{}, err error) {
	log.Println("htlc close step 9 begin", time.Now())
	dataFrom51P := bean.AliceSignedC4bTxDataP2p{}
	err = json.Unmarshal([]byte(data), &dataFrom51P)
	if err != nil {
		return nil, err
	}

	if service.tempDataFromTo51PAtBobSide == nil {
		service.tempDataFromTo51PAtBobSide = make(map[string]bean.AliceSignedC4bTxDataP2p)
	}
	service.tempDataFromTo51PAtBobSide[user.PeerId+"_"+dataFrom51P.ChannelId] = dataFrom51P

	needBobSignRdTxData := bean.NeedBobSignRdTxForC4b{}
	needBobSignRdTxData.ChannelId = dataFrom51P.ChannelId
	needBobSignRdTxData.C4bRdPartialSignedData = dataFrom51P.C4bRdPartialSignedData
	log.Println("htlc close step 9 end", time.Now())
	return needBobSignRdTxData, nil
}

// step10 bob 110114 bob完成Cxb的Rd的签名
func (service *htlcCloseTxManager) OnBobSignedCxbSubTx(msg bean.RequestMessage, user bean.User) (toBob interface{}, err error) {
	log.Println("htlc close step 10 begin", time.Now())
	bobSignedRdData := bean.BobSignedRdTxForC4b{}
	_ = json.Unmarshal([]byte(msg.Data), &bobSignedRdData)

	dataFrom51P := service.tempDataFromTo51PAtBobSide[user.PeerId+"_"+bobSignedRdData.ChannelId]
	if len(dataFrom51P.ChannelId) == 0 {
		return nil, errors.New(enum.Tips_common_empty + "channel_id")
	}

	if tool.CheckIsString(&dataFrom51P.C4bRdPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(false, bobSignedRdData.C4bRdCompleteSignedHex, 2); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c4b_rd_complete_signed_hex")
			log.Println(err)
			return nil, err
		}
	}

	var channelId = dataFrom51P.ChannelId
	var signedRsmcHex = dataFrom51P.C4bRsmcCompleteSignedHex
	var signedToCounterpartyTxHex = dataFrom51P.C4bCounterpartyCompleteSignedHex
	var c2bSignedRdHex = bobSignedRdData.C4bRdCompleteSignedHex
	var c2aInitTxHash = dataFrom51P.C4aCommitmentTxHash

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := getChannelInfoByChannelId(tx, channelId, user.PeerId)
	if channelInfo == nil {
		return nil, errors.New("not found channelInfo at targetSide")
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find payer's commitmentTxInfo")
		log.Println(err)
		return nil, err
	}

	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
		err = errors.New("wrong commitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, err
	}

	myChannelAddress := channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdA {
		myChannelAddress = channelInfo.AddressA
	}

	if tool.CheckIsString(&signedRsmcHex) {
		latestCommitmentTxInfo.RSMCTxHex = signedRsmcHex
		latestCommitmentTxInfo.RSMCTxid = rpcClient.GetTxId(signedRsmcHex)
		err = saveRdTx(tx, channelInfo, signedRsmcHex, c2bSignedRdHex, latestCommitmentTxInfo, myChannelAddress, &user)
		if err != nil {
			return nil, err
		}
	}

	if tool.CheckIsString(&signedToCounterpartyTxHex) {
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedToCounterpartyTxHex
		latestCommitmentTxInfo.ToCounterpartyTxid = rpcClient.GetTxId(signedToCounterpartyTxHex)
	}

	latestCommitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	latestCommitmentTxInfo.SignAt = time.Now()

	bytes, err := json.Marshal(latestCommitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	latestCommitmentTxInfo.CurrHash = msgHash
	_ = tx.Update(latestCommitmentTxInfo)

	lastCommitmentTxInfo := dao.CommitmentTransaction{}
	_ = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, &lastCommitmentTxInfo)
	if lastCommitmentTxInfo.Id > 0 {
		lastCommitmentTxInfo.CurrState = dao.TxInfoState_Abord
		_ = tx.Update(lastCommitmentTxInfo)
	}

	payeeRevokeAndAcknowledgeCommitment := &dao.PayeeRevokeAndAcknowledgeCommitment{}
	payeeRevokeAndAcknowledgeCommitment.ChannelId = channelId
	payeeRevokeAndAcknowledgeCommitment.CommitmentTxHash = c2aInitTxHash
	payeeRevokeAndAcknowledgeCommitment.Approval = true
	_ = tx.Save(payeeRevokeAndAcknowledgeCommitment)

	channelInfo.CurrState = dao.ChannelState_CanUse
	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	//同步通道信息到tracker
	sendChannelStateToTracker(*channelInfo, *latestCommitmentTxInfo)

	log.Println("htlc close step 10 end", time.Now())
	return latestCommitmentTxInfo, nil
}
