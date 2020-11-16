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
}

// htlc 关闭当前htlc交易
var HtlcCloseTxService htlcCloseTxManager

// step1 Alice 100049 请求关闭htlc交易 -100049 request close htlc
func (service *htlcCloseTxManager) RequestCloseHtlc(msg bean.RequestMessage, user bean.User) (data interface{}, needSign bool, err error) {
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

	if tool.CheckIsString(&reqData.CurrRsmcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_rsmc_temp_address_pub_key")
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
		if reqData.CurrRsmcTempAddressPubKey != latestCommitmentTxInfo.RSMCTempAddressPubKey {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, reqData.CurrRsmcTempAddressPubKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
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
	dataTo49P.CurrTempAddressPubKey = reqData.CurrRsmcTempAddressPubKey
	dataTo49P.CloserNodeAddress = msg.SenderNodePeerId
	dataTo49P.CloserPeerId = msg.SenderUserPeerId

	needSign = false
	//如果是第一次请求
	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Htlc {
		//创建c2a omni的交易不能一个输入，多个输出，所以就是两个交易
		reqTempData := &bean.RequestCreateCommitmentTx{}
		reqTempData.CurrTempAddressIndex = reqData.CurrRsmcTempAddressIndex
		reqTempData.CurrTempAddressPubKey = reqData.CurrRsmcTempAddressPubKey
		reqTempData.Amount = 0
		newCommitmentTxInfo, err := createCommitmentTxHex(tx, true, reqTempData, channelInfo, latestCommitmentTxInfo, user)
		if err != nil {
			return nil, false, err
		}

		newCommitmentTxInfo.CurrState = dao.TxInfoState_Init
		_ = tx.UpdateField(newCommitmentTxInfo, "CurrState", dao.TxInfoState_Init)

		dataTo49P.CommitmentTxHash = newCommitmentTxInfo.CurrHash
		dataTo49P.RsmcPartialSignedData = newCommitmentTxInfo.RsmcRawTxData
		dataTo49P.CounterpartyPartialSignedData = newCommitmentTxInfo.ToCounterpartyRawTxData
		needSign = true

	} else { //	上一次的请求出现异常，再次发起请求
		dataTo49P.CommitmentTxHash = latestCommitmentTxInfo.CurrHash
		dataTo49P.RsmcPartialSignedData = latestCommitmentTxInfo.RsmcRawTxData
		dataTo49P.CounterpartyPartialSignedData = latestCommitmentTxInfo.ToCounterpartyRawTxData

		if len(latestCommitmentTxInfo.RSMCTxid) == 0 {
			retSignData.ChannelId = channelInfo.ChannelId
			retSignData.C4aRsmcRawData = latestCommitmentTxInfo.RsmcRawTxData
			retSignData.C4aCounterpartyRawData = latestCommitmentTxInfo.ToCounterpartyRawTxData
			needSign = true
		}
	}
	_ = tx.Commit()

	if needSign {
		if service.tempDataSendTo49PAtAliceSide == nil {
			service.tempDataSendTo49PAtAliceSide = make(map[string]bean.AliceRequestCloseHtlcCurrTxOfP2p)
		}
		service.tempDataSendTo49PAtAliceSide[user.PeerId+"_"+dataTo49P.ChannelId] = dataTo49P
		return retSignData, true, nil
	}

	return dataTo49P, false, nil
}

// step2 Alice 100100 Alice签名Cxa 并推送49号协议
func (service *htlcCloseTxManager) OnAliceSignedCxa(msg bean.RequestMessage, user bean.User) (toALice, toBob interface{}, err error) {
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

	if tool.CheckIsString(&signedData.CounterpartyPartialSignedHex) == false {
		err = errors.New(enum.Tips_common_empty + "counterparty_partial_signed_hex")
		log.Println(err)
		return nil, nil, err
	}

	if pass, _ := rpcClient.CheckMultiSign(true, signedData.CounterpartyPartialSignedHex, 1); pass == false {
		err = errors.New(enum.Tips_common_wrong + "counterparty_signed_hex")
		log.Println(err)
		return nil, nil, err
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
		latestCommitmentTxInfo.RsmcRawTxData.Hex = signedData.RsmcPartialSignedHex
		latestCommitmentTxInfo.RSMCTxHex = signedData.CounterpartyPartialSignedHex
		latestCommitmentTxInfo.RSMCTxid = rpcClient.GetTxId(signedData.CounterpartyPartialSignedHex)
	}

	if len(latestCommitmentTxInfo.ToCounterpartyTxHex) > 0 {
		//封装好的签名数据，给bob的客户端签名使用
		latestCommitmentTxInfo.ToCounterpartyRawTxData.Hex = signedData.CounterpartyPartialSignedHex
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedData.CounterpartyPartialSignedHex
		latestCommitmentTxInfo.ToCounterpartyTxid = rpcClient.GetTxId(signedData.CounterpartyPartialSignedHex)
	}

	p2pData.RsmcPartialSignedData = latestCommitmentTxInfo.RsmcRawTxData
	p2pData.CounterpartyPartialSignedData = latestCommitmentTxInfo.ToCounterpartyRawTxData

	latestCommitmentTxInfo.RsmcRawTxData = bean.NeedClientSignTxData{Hex: "", Inputs: nil, IsMultisig: false}
	latestCommitmentTxInfo.ToCounterpartyRawTxData = bean.NeedClientSignTxData{Hex: "", Inputs: nil, IsMultisig: false}
	latestCommitmentTxInfo.CurrState = dao.TxInfoState_Create
	_ = tx.Update(latestCommitmentTxInfo)
	tx.Commit()

	toAliceResult := bean.AliceSignedRsmcDataForC2aResult{}
	toAliceResult.ChannelId = p2pData.ChannelId
	toAliceResult.CurrTempAddressPubKey = p2pData.CurrTempAddressPubKey
	toAliceResult.CommitmentTxHash = p2pData.CommitmentTxHash
	toAliceResult.Amount = latestCommitmentTxInfo.AmountToHtlc
	return toAliceResult, p2pData, nil
}

// step3 obd 110049 推送p2p消息给bob
//func (service *htlcCloseTxManager) OnObdOfBobGet49PData(msg string, user bean.User) (toBob interface{}, err error) {
//
//}
//
//// step4 bob 100050 响应bob对这次关闭htlc交易的签收及他对Cxa的签名
//func (service *htlcCloseTxManager) OnBobSignCloseHtlcRequest(msg bean.RequestMessage, user bean.User) (toBob interface{}, err error) {
//
//}
//
//// step5 bob 100111 响应bob对Cxb的签名，并推送50号协议
//func (service *htlcCloseTxManager) OnBobSignedCxb(msg bean.RequestMessage, user bean.User) (toAlice, toBob interface{}, err error) {
//
//}
//
//// step6 obd 110050 推送p2p消息给Alice
//func (service *htlcCloseTxManager) OnObdOfAliceGet50PData(msg string, user bean.User) (toAlice interface{}, err error) {
//
//}
//
//// step7 alice 100112 Alice完成对Cxb的签名
//func (service *htlcCloseTxManager) OnAliceSignedCxb(msg bean.RequestMessage, user bean.User) (toAlice interface{}, err error) {
//
//}
//
//// step8 alice 100113 Alice完成对Cxb的Rd和Br的签名 并推送51号协议
//func (service *htlcCloseTxManager) OnAliceSignedCxbBubTx(msg bean.RequestMessage, user bean.User) (toAlice, toBob interface{}, err error) {
//
//}
//
//// step9 obd 110051 推送51号协议消息到bob
//func (service *htlcCloseTxManager) OnObdOfBobGet51PData(msg string, user bean.User) (toBob interface{}, err error) {
//
//}
//
//// step10 bob 110114 bob完成Cxb的Rd的签名
//func (service *htlcCloseTxManager) OnBobSignedCxbSubTx(msg bean.RequestMessage, user bean.User) (toBob interface{}, err error) {
//
//}

// -49 接收方节点的信息缓存处理
func (service *htlcCloseTxManager) BeforeBobSignCloseHtlcAtBobSide(data string, user *bean.User) (retData *bean.RequestCloseHtlcTxOfWs, err error) {

	closeHtlcTxOfP2p := &bean.RequestCloseHtlcTxOfP2p{}
	_ = json.Unmarshal([]byte(data), closeHtlcTxOfP2p)

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

	closeHtlcTxOfWs := &bean.RequestCloseHtlcTxOfWs{}
	closeHtlcTxOfWs.RequestCloseHtlcTxOfP2p = *closeHtlcTxOfP2p
	closeHtlcTxOfWs.MsgHash = messageHash

	return closeHtlcTxOfWs, nil
}

// -50 sign close htlc
func (service *htlcCloseTxManager) CloseHTLCSigned(msg bean.RequestMessage, user bean.User) (retData *bean.HtlcCloseCloseeSignedInfoToCloser, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New(enum.Tips_common_empty + "msg data")
	}
	reqData := &bean.HtlcSignCloseCurrTx{}
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
	closeHtlcTxOfP2p := &bean.RequestCloseHtlcTxOfP2p{}
	_ = json.Unmarshal([]byte(message.Data), closeHtlcTxOfP2p)

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "channel_address_private_key")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.CurrRsmcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_rsmc_temp_address_pub_key")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&reqData.CurrRsmcTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_rsmc_temp_address_private_key")
		log.Println(err)
		return nil, err
	}
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.CurrRsmcTempAddressPrivateKey, reqData.CurrRsmcTempAddressPubKey)
	if err != nil {
		return nil, err
	}
	tempAddrPrivateKeyMap[reqData.CurrRsmcTempAddressPubKey] = reqData.CurrRsmcTempAddressPrivateKey

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
		q.Eq("ChannelId", closeHtlcTxOfP2p.ChannelId),
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
	if senderPeerId != msg.RecipientUserPeerId {
		return nil, errors.New(enum.Tips_common_userNotInTx)
	}

	err = findUserIsOnline(msg.RecipientNodePeerId, senderPeerId)
	if err != nil {
		return nil, err
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, signerPubKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongChannelPrivateKey, reqData.ChannelAddressPrivateKey, signerPubKey))
	}
	tempAddrPrivateKeyMap[signerPubKey] = reqData.ChannelAddressPrivateKey

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
		if reqData.CurrRsmcTempAddressPubKey != latestCommitmentTxInfo.RSMCTempAddressPubKey {
			return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, reqData.CurrRsmcTempAddressPubKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
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

	retData = &bean.HtlcCloseCloseeSignedInfoToCloser{}
	if he1b.Id > 0 {
		_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressForHtnxPrivateKey, he1b.RSMCTempAddressPubKey)
		if err != nil {
			return nil, err
		}
		retData.CloseeLastHtlcTempAddressForHtnxPrivateKey = reqData.LastHtlcTempAddressForHtnxPrivateKey
	}

	retData.ChannelId = channelInfo.ChannelId
	retData.CloseeLastRsmcTempAddressPrivateKey = reqData.LastRsmcTempAddressPrivateKey
	retData.CloseeLastHtlcTempAddressPrivateKey = reqData.LastHtlcTempAddressPrivateKey
	retData.CloseeCurrRsmcTempAddressPubKey = reqData.CurrRsmcTempAddressPubKey

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
	if tool.CheckIsString(&closeHtlcTxOfP2p.RsmcHex) {

		aliceRsmcTxId, signedRsmcHex, err = rpcClient.BtcSignRawTransaction(closeHtlcTxOfP2p.RsmcHex, reqData.ChannelAddressPrivateKey)
		if err != nil {
			return nil, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "rsmc hex"))
		}
		testResult, err := rpcClient.TestMemPoolAccept(signedRsmcHex)
		if err != nil {
			return nil, err
		}
		if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
			return nil, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
		}

		//  根据alice的临时地址+bob的通道address,获取alice2+bob的多签地址，并得到AliceSignedRsmcHex签名后的交易的input，为创建alice的RD和bob的BR做准备
		aliceRsmcMultiAddress, aliceRsmcRedeemScript, aliceRsmcMultiAddressScriptPubKey, err = createMultiSig(closeHtlcTxOfP2p.CurrRsmcTempAddressPubKey, signerPubKey)
		if err != nil {
			return nil, err
		}

		aliceRsmcOutputs, err = getInputsForNextTxByParseTxHashVout(signedRsmcHex, aliceRsmcMultiAddress, aliceRsmcMultiAddressScriptPubKey, aliceRsmcRedeemScript)
		if err != nil || len(aliceRsmcOutputs) == 0 {
			log.Println(err)
			return nil, err
		}
	}
	retData.CloserSignedRsmcHex = signedRsmcHex

	//  签名对方传过来的toOtherHex
	_, signedToOtherHex, err := rpcClient.BtcSignRawTransaction(closeHtlcTxOfP2p.ToCounterpartyTxHex, reqData.ChannelAddressPrivateKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "to_counterparty_tx_hex"))
	}
	testResult, err := rpcClient.TestMemPoolAccept(signedToOtherHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
		return nil, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
	}
	retData.CloserSignedToCounterpartyTxHex = signedToOtherHex
	//endregion

	//第一次签名，不是失败后的重试
	var amountToCounterparty = 0.0
	if latestCommitmentTxInfo.TxType == dao.CommitmentTransactionType_Htlc {
		//region 3、根据对方传过来的上一个交易的临时rsmc私钥，签名最近的BR交易，保证对方确实放弃了上一个承诺交易
		err = signLastBR(tx, dao.BRType_Rmsc, *channelInfo, user.PeerId, closeHtlcTxOfP2p.LastRsmcTempAddressPrivateKey, latestCommitmentTxInfo.Id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = signLastBR(tx, dao.BRType_Htlc, *channelInfo, user.PeerId, closeHtlcTxOfP2p.LastHtlcTempAddressPrivateKey, latestCommitmentTxInfo.Id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Htlc_GetR {
			err = signLastBR(tx, dao.BRType_Ht1a, *channelInfo, user.PeerId, closeHtlcTxOfP2p.LastHtlcTempAddressForHtnxPrivateKey, latestCommitmentTxInfo.Id)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
		//endregion

		//region 4、创建C3b
		commitmentTxRequest := &bean.RequestCreateCommitmentTx{}
		commitmentTxRequest.ChannelId = channelInfo.ChannelId
		commitmentTxRequest.Amount = 0
		commitmentTxRequest.CurrTempAddressIndex = reqData.CurrRsmcTempAddressIndex
		commitmentTxRequest.CurrTempAddressPubKey = reqData.CurrRsmcTempAddressPubKey
		newCommitmentTxInfo, err := createCommitmentTxHex(tx, false, commitmentTxRequest, channelInfo, latestCommitmentTxInfo, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		amountToCounterparty = newCommitmentTxInfo.AmountToCounterparty

		retData.CloseeRsmcHex = newCommitmentTxInfo.RSMCTxHex
		retData.CloseeToCounterpartyTxHex = newCommitmentTxInfo.ToCounterpartyTxHex
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
			senderCommitmentTx.RSMCTempAddressPubKey = closeHtlcTxOfP2p.CurrRsmcTempAddressPubKey
			senderCommitmentTx.RSMCMultiAddress = aliceRsmcMultiAddress
			senderCommitmentTx.RSMCRedeemScript = aliceRsmcRedeemScript
			senderCommitmentTx.RSMCMultiAddressScriptPubKey = aliceRsmcMultiAddressScriptPubKey
			senderCommitmentTx.RSMCTxHex = signedRsmcHex
			senderCommitmentTx.RSMCTxid = aliceRsmcTxId
			senderCommitmentTx.AmountToRSMC = newCommitmentTxInfo.AmountToCounterparty
			err = createCurrCommitmentTxBR(tx, dao.BRType_Rmsc, channelInfo, senderCommitmentTx, aliceRsmcOutputs, myAddress, reqData.ChannelAddressPrivateKey, user)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
		//endregion
	} else {
		retData.CloseeRsmcHex = latestCommitmentTxInfo.RSMCTxHex
		retData.CloseeToCounterpartyTxHex = latestCommitmentTxInfo.ToCounterpartyTxHex
		amountToCounterparty = latestCommitmentTxInfo.AmountToCounterparty
	}

	//region 6、根据签名后的AliceRsmc创建alice的RD create RD tx for alice
	if len(aliceRsmcOutputs) > 0 {
		outputAddress := channelInfo.AddressA
		if user.PeerId == channelInfo.PeerIdA {
			outputAddress = channelInfo.AddressB
		}
		_, senderRdhex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
			aliceRsmcMultiAddress,
			[]string{
				reqData.ChannelAddressPrivateKey,
			},
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
		retData.CloserRsmcRdHex = senderRdhex
	}
	retData.CloserCommitmentTxHash = closeHtlcTxOfP2p.CommitmentTxHash
	//endregion create RD tx for alice

	_ = messageService.updateMsgStateUseTx(tx, message)
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return retData, nil
}

// -51 alice根据得到的bob的临时私钥签名
func (service *htlcCloseTxManager) AfterBobCloseHTLCSigned_AtAliceSide(data string, user *bean.User) (retData map[string]interface{}, needNoticeAlice bool, err error) {
	jsonObj := &bean.HtlcCloseCloseeSignedInfoToCloser{}
	_ = json.Unmarshal([]byte(data), jsonObj)

	//region 检测传入数据
	var channelId = jsonObj.ChannelId
	if tool.CheckIsString(&channelId) == false {
		err = errors.New("wrong channelId")
		log.Println(err)
		return nil, false, err
	}
	var commitmentTxHash = jsonObj.CloserCommitmentTxHash
	if tool.CheckIsString(&commitmentTxHash) == false {
		err = errors.New("wrong commitmentHash")
		log.Println(err)
		return nil, false, err
	}

	var signedRsmcHex = jsonObj.CloserSignedRsmcHex
	var rsmcTxid string
	if tool.CheckIsString(&signedRsmcHex) {
		rsmcTxid, err = rpcClient.TestMemPoolAccept(signedRsmcHex)
		if err != nil {
			err = errors.New("wrong signedRsmcHex")
			log.Println(err)
			return nil, false, err
		}
	}

	var signedToCounterpartyHex = jsonObj.CloserSignedToCounterpartyTxHex
	if tool.CheckIsString(&signedToCounterpartyHex) == false {
		err = errors.New("wrong signedToCounterpartyHex")
		log.Println(err)
		return nil, false, err
	}
	toCounterpartyTxid, err := rpcClient.TestMemPoolAccept(signedToCounterpartyHex)
	if err != nil {
		err = errors.New("wrong signedToCounterpartyHex")
		log.Println(err)
		return nil, false, err
	}

	var aliceRdHex = jsonObj.CloserRsmcRdHex
	if tool.CheckIsString(&aliceRdHex) {
		_, err = rpcClient.TestMemPoolAccept(aliceRdHex)
		if err != nil {
			err = errors.New("wrong senderRdHex")
			log.Println(err)
			return nil, false, err
		}
	}

	var bobRsmcHex = jsonObj.CloseeRsmcHex
	if tool.CheckIsString(&bobRsmcHex) == false {
		err = errors.New("wrong rsmcHex")
		log.Println(err)
		return nil, false, err
	}
	_, err = rpcClient.TestMemPoolAccept(bobRsmcHex)
	if err != nil {
		err = errors.New("wrong rsmcHex")
		log.Println(err)
		return nil, false, err
	}

	var bobCurrTempAddressPubKey = jsonObj.CloseeCurrRsmcTempAddressPubKey
	if tool.CheckIsString(&bobCurrTempAddressPubKey) == false {
		err = errors.New("wrong currTempAddressPubKey")
		log.Println(err)
		return nil, false, err
	}
	var bobToOtherHex = jsonObj.CloseeToCounterpartyTxHex
	//endregion

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}
	defer tx.Rollback()

	bobData := bean.HtlcCloseCloserSignTxInfoToClosee{}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", jsonObj.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if err != nil {
		return nil, true, errors.New("not found channelInfo at targetSide")
	}

	fundingTransaction := getFundingTransactionByChannelId(tx, channelId, user.PeerId)
	if fundingTransaction == nil {
		return nil, true, errors.New("not found fundingTransaction at targetSide")
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find sender's commitmentTxInfo")
		log.Println(err)
		return nil, true, err
	}

	if latestCommitmentTxInfo.CurrHash != commitmentTxHash {
		err = errors.New("wrong request hash")
		log.Println(err)
		return nil, false, err
	}

	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
		err = errors.New("wrong commitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, false, err
	}

	var myChannelPubKey = channelInfo.PubKeyA
	var myChannelAddress = channelInfo.AddressA
	var partnerChannelAddress = channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdB {
		myChannelAddress = channelInfo.AddressB
		myChannelPubKey = channelInfo.PubKeyB
		partnerChannelAddress = channelInfo.AddressA
	}
	var myChannelPrivateKey = tempAddrPrivateKeyMap[myChannelPubKey]

	//region 根据对方传过来的上一个交易的临时rsmc私钥，签名上一次的BR交易，保证对方确实放弃了上一个承诺交易
	var bobLastRsmcTempAddressPrivateKey = jsonObj.CloseeLastRsmcTempAddressPrivateKey
	var bobLastHtlcTempAddressPrivateKey = jsonObj.CloseeLastHtlcTempAddressPrivateKey
	var bobLastHtlcTempAddressForHtnxPrivateKey = jsonObj.CloseeLastHtlcTempAddressForHtnxPrivateKey
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
	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Htlc_GetR {
		err = signLastBR(tx, dao.BRType_HE1b, *channelInfo, user.PeerId, bobLastHtlcTempAddressForHtnxPrivateKey, latestCommitmentTxInfo.LastCommitmentTxId)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}
	}
	//endregion

	// region 对自己的RD 二次签名
	if tool.CheckIsString(&rsmcTxid) {
		latestCommitmentTxInfo.RSMCTxHex = signedRsmcHex
		latestCommitmentTxInfo.RSMCTxid = gjson.Parse(rsmcTxid).Array()[0].Get("txid").Str
		err = signRdTx(tx, channelInfo, signedRsmcHex, aliceRdHex, latestCommitmentTxInfo, myChannelAddress, user)
		if err != nil {
			return nil, true, err
		}
	}
	// endregion

	//更新alice的当前承诺交易
	latestCommitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	latestCommitmentTxInfo.ToCounterpartyTxHex = signedToCounterpartyHex
	latestCommitmentTxInfo.ToCounterpartyTxid = gjson.Parse(toCounterpartyTxid).Array()[0].Get("txid").Str
	latestCommitmentTxInfo.SignAt = time.Now()

	bytes, err := json.Marshal(latestCommitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	latestCommitmentTxInfo.CurrHash = msgHash
	_ = tx.Update(latestCommitmentTxInfo)

	lastCommitmentTxInfo := dao.CommitmentTransaction{}
	err = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, &lastCommitmentTxInfo)
	if err == nil {
		lastCommitmentTxInfo.CurrState = dao.TxInfoState_Abord
		_ = tx.Update(lastCommitmentTxInfo)
	}

	//处理对方的数据
	//签名对方传过来的rsmcHex
	bobRsmcTxid, bobSignedRsmcHex, err := rpcClient.BtcSignRawTransaction(bobRsmcHex, myChannelPrivateKey)
	if err != nil {
		return nil, false, errors.New("fail to sign rsmc hex ")
	}
	testResult, err := rpcClient.TestMemPoolAccept(bobSignedRsmcHex)
	if err != nil {
		return nil, false, err
	}
	if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
		return nil, false, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
	}
	err = checkBobRemcData(bobSignedRsmcHex, latestCommitmentTxInfo)
	if err != nil {
		return nil, false, err
	}
	bobData.CloseeSignedRsmcHex = bobSignedRsmcHex

	//region create RD tx for bob
	bobMultiAddr, err := rpcClient.CreateMultiSig(2, []string{bobCurrTempAddressPubKey, myChannelPubKey})
	if err != nil {
		return nil, false, err
	}
	bobRsmcMultiAddress := gjson.Get(bobMultiAddr, "address").String()
	bobRsmcRedeemScript := gjson.Get(bobMultiAddr, "redeemScript").String()
	addressJson, err := rpcClient.GetAddressInfo(bobRsmcMultiAddress)
	if err != nil {
		return nil, false, err
	}
	bobRsmcMultiAddressScriptPubKey := gjson.Get(addressJson, "scriptPubKey").String()

	inputs, err := getInputsForNextTxByParseTxHashVout(bobSignedRsmcHex, bobRsmcMultiAddress, bobRsmcMultiAddressScriptPubKey, bobRsmcRedeemScript)
	if err != nil || len(inputs) == 0 {
		log.Println(err)
		return nil, false, err
	}

	_, bobRdhex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		bobRsmcMultiAddress,
		[]string{
			myChannelPrivateKey,
		},
		inputs,
		partnerChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		latestCommitmentTxInfo.AmountToCounterparty,
		getBtcMinerAmount(channelInfo.BtcAmount),
		1000,
		&bobRsmcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, false, errors.New("fail to create rd")
	}
	bobData.CloseeRsmcRdHex = bobRdhex
	//endregion create RD tx for alice

	//region 根据对方的Rsmc签名，生成惩罚对方，自己获益BR
	bobCommitmentTx := &dao.CommitmentTransaction{}
	bobCommitmentTx.Id = latestCommitmentTxInfo.Id
	bobCommitmentTx.PropertyId = channelInfo.PropertyId
	bobCommitmentTx.RSMCTempAddressPubKey = bobCurrTempAddressPubKey
	bobCommitmentTx.RSMCMultiAddress = bobRsmcMultiAddress
	bobCommitmentTx.RSMCRedeemScript = bobRsmcRedeemScript
	bobCommitmentTx.RSMCMultiAddressScriptPubKey = bobRsmcMultiAddressScriptPubKey
	bobCommitmentTx.RSMCTxHex = bobSignedRsmcHex
	bobCommitmentTx.RSMCTxid = bobRsmcTxid
	bobCommitmentTx.AmountToRSMC = latestCommitmentTxInfo.AmountToCounterparty
	err = createCurrCommitmentTxBR(tx, dao.BRType_Rmsc, channelInfo, bobCommitmentTx, inputs, myChannelAddress, myChannelPrivateKey, *user)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	//endregion

	//签名对方传过来的toOtherHex
	if tool.CheckIsString(&bobToOtherHex) {
		_, bobSignedToOtherHex, err := rpcClient.BtcSignRawTransaction(bobToOtherHex, myChannelPrivateKey)
		if err != nil {
			return nil, false, errors.New("fail to sign toOther hex ")
		}
		testResult, err = rpcClient.TestMemPoolAccept(bobSignedToOtherHex)
		if err != nil {
			return nil, false, err
		}
		if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
			return nil, false, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
		}
		bobData.CloseeSignedToCounterpartyTxHex = bobSignedToOtherHex
	}

	channelInfo.CurrState = dao.ChannelState_CanUse
	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	//同步通道信息到tracker
	sendChannelStateToTracker(*channelInfo, *latestCommitmentTxInfo)

	bobData.ChannelId = channelId
	retData = make(map[string]interface{})
	retData["aliceData"] = latestCommitmentTxInfo
	retData["bobData"] = bobData
	return retData, true, nil
}

//52 bob保存更新自己的交易
func (service *htlcCloseTxManager) AfterAliceSignCloseHTLCAtBobSide(data string, user *bean.User) (retData interface{}, err error) {
	jsonObj := &bean.HtlcCloseCloserSignTxInfoToClosee{}
	_ = json.Unmarshal([]byte(data), jsonObj)

	var channelId = jsonObj.ChannelId
	var signedRsmcHex = jsonObj.CloseeSignedRsmcHex
	var signedToOtherHex = jsonObj.CloseeSignedToCounterpartyTxHex
	var rdHex = jsonObj.CloseeRsmcRdHex

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", jsonObj.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if err != nil {
		return nil, errors.New("not found channelInfo at targetSide")
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find sender's commitmentTxInfo")
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
		decodeRsmcHex, err := rpcClient.OmniDecodeTransaction(signedRsmcHex)
		if err != nil {
			return nil, err
		}
		latestCommitmentTxInfo.RSMCTxHex = signedRsmcHex
		latestCommitmentTxInfo.RSMCTxid = gjson.Get(decodeRsmcHex, "txid").Str
		err = signRdTx(tx, channelInfo, signedRsmcHex, rdHex, latestCommitmentTxInfo, myChannelAddress, user)
		if err != nil {
			return nil, err
		}
	}

	if tool.CheckIsString(&signedToOtherHex) {
		decodeSignedToOtherHex, err := rpcClient.OmniDecodeTransaction(signedToOtherHex)
		if err != nil {
			return nil, err
		}
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedToOtherHex
		latestCommitmentTxInfo.RSMCTxid = gjson.Get(decodeSignedToOtherHex, "txid").Str
	}

	// update bob's latestCommitmentTxInfo
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

	channelInfo.CurrState = dao.ChannelState_CanUse
	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	return latestCommitmentTxInfo, nil
}
