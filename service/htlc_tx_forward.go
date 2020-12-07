package service

import (
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/shopspring/decimal"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type htlcForwardTxManager struct {
	operationFlag sync.Mutex

	//缓存数据
	//在步骤4，缓存需要发往41号协议的信息
	tempDataSendTo41PAtBobSide map[string]bean.NeedAliceSignHtlcTxOfC3bP2p
	//在步骤6，缓存来自41号协议的信息
	tempDataFrom41PAtAliceSide map[string]bean.NeedAliceSignHtlcTxOfC3bP2p
	//在步骤7，缓存需要发往42号协议的信息
	tempDataSendTo42PAtAliceSide map[string]bean.NeedBobSignHtlcSubTxOfC3bP2p
	//在步骤9，缓存来自42号协议的信息
	tempDataFrom42PAtBobSide map[string]bean.NeedBobSignHtlcSubTxOfC3bP2p
}

// htlc pay money  付款
var HtlcForwardTxService htlcForwardTxManager

func (service *htlcForwardTxManager) CreateHtlcInvoice(msg bean.RequestMessage) (data interface{}, err error) {

	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New(enum.Tips_common_empty + "msd data")
	}

	requestData := &bean.HtlcRequestInvoice{}
	err = json.Unmarshal([]byte(msg.Data), requestData)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	addr := ""
	//obbc,obtb,obcrt
	if strings.Contains(config.ChainNode_Type, "main") {
		addr = "obbc"
	}
	if strings.Contains(config.ChainNode_Type, "test") {
		addr = "obtb"
	}
	if strings.Contains(config.ChainNode_Type, "reg") {
		addr = "obcrt"
	}
	if requestData.Amount < config.GetOmniDustBtc() {
		return nil, errors.New(enum.Tips_common_wrong + "amount")
	} else {
		requestData.Amount *= 100000000
		temp := int(requestData.Amount)
		addr += strconv.Itoa(temp) + "s"
	}

	addr += "1"

	if requestData.PropertyId < 0 {
		return nil, errors.New(enum.Tips_common_wrong + "property_id")
	}
	propertyId := ""
	tool.ConvertNumToString(int(requestData.PropertyId), &propertyId)
	code, err := tool.GetMsgLengthFromInt(len(propertyId))
	if err != nil {
		return nil, err
	}
	addr += "p" + code + propertyId

	code, err = tool.GetMsgLengthFromInt(len(msg.SenderNodePeerId))
	if err != nil {
		return nil, err
	}
	addr += "n" + code + msg.SenderNodePeerId

	code, err = tool.GetMsgLengthFromInt(len(msg.SenderUserPeerId))
	if err != nil {
		return nil, err
	}
	addr += "u" + code + msg.SenderUserPeerId

	if tool.CheckIsString(&requestData.H) == false {
		return nil, errors.New(enum.Tips_common_wrong + "h")
	} else {
		//ph payment H
		code, err = tool.GetMsgLengthFromInt(len(requestData.H))
		if err != nil {
			return nil, err
		}
		addr += "h" + code + requestData.H
	}

	if time.Time(requestData.ExpiryTime).IsZero() {
		return nil, errors.New(enum.Tips_common_wrong + "expiry_time")
	} else {
		if time.Now().After(time.Time(requestData.ExpiryTime)) {
			return nil, errors.New(fmt.Sprintf(enum.Tips_htlc_expiryTimeAfterNow, "expiry_time"))
		}
		expiryTime := ""
		tool.ConvertNumToString(int(time.Time(requestData.ExpiryTime).Unix()), &expiryTime)
		code, err = tool.GetMsgLengthFromInt(len(expiryTime))
		if err != nil {
			return nil, err
		}
		addr += "x" + code + expiryTime
	}

	code, err = tool.GetMsgLengthFromInt(1)
	if err != nil {
		return nil, err
	}
	isPrivate := "0"
	if requestData.IsPrivate {
		isPrivate = "1"
	}
	addr += "t" + code + isPrivate

	if len(requestData.Description) > 0 {
		code, err = tool.GetMsgLengthFromInt(len(requestData.Description))
		if err != nil {
			return nil, err
		}
		addr += "d" + code + requestData.Description
	}

	bytes := []byte(addr)
	sum := 0
	for _, item := range bytes {
		sum += int(item)
	}
	checkSum := ""
	tool.ConvertNumToString(sum, &checkSum)

	addr += checkSum

	return addr, nil
}

// 401 find htlc find path
func (service *htlcForwardTxManager) PayerRequestFindPath(msgData string, user bean.User) (data interface{}, isPrivate bool, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, false, errors.New(enum.Tips_common_empty + "msg data")
	}

	requestData := &bean.HtlcRequestFindPath{}
	err = json.Unmarshal([]byte(msgData), requestData)
	if err != nil {
		log.Println(err.Error())
		return nil, false, err
	}

	var requestFindPathInfo bean.HtlcRequestFindPathInfo

	if tool.CheckIsString(&requestData.Invoice) {
		htlcRequestInvoice, err := tool.DecodeInvoiceObjFromCodes(requestData.Invoice)
		if err != nil {
			return nil, false, errors.New(enum.Tips_common_wrong + "invoice")
		}
		if err = findUserIsOnline(htlcRequestInvoice.RecipientNodePeerId, htlcRequestInvoice.RecipientUserPeerId); err != nil {
			return nil, requestFindPathInfo.IsPrivate, err
		}
		requestFindPathInfo = htlcRequestInvoice.HtlcRequestFindPathInfo
	} else {
		requestFindPathInfo = requestData.HtlcRequestFindPathInfo
		if tool.CheckIsString(&requestFindPathInfo.RecipientNodePeerId) == false {
			return nil, requestFindPathInfo.IsPrivate, errors.New(enum.Tips_common_wrong + "recipient_node_peer_id")
		}
		if tool.CheckIsString(&requestFindPathInfo.RecipientUserPeerId) == false {
			return nil, requestFindPathInfo.IsPrivate, errors.New(enum.Tips_common_wrong + "recipient_user_peer_id")
		}

		if err = findUserIsOnline(requestFindPathInfo.RecipientNodePeerId, requestFindPathInfo.RecipientUserPeerId); err != nil {
			return nil, requestFindPathInfo.IsPrivate, err
		}
	}

	if requestFindPathInfo.RecipientUserPeerId == user.PeerId && requestFindPathInfo.RecipientNodePeerId == user.P2PLocalPeerId {
		return nil, requestFindPathInfo.IsPrivate, errors.New("recipient_user_peer_id can not be yourself")
	}

	if requestFindPathInfo.PropertyId < 0 {
		return nil, requestFindPathInfo.IsPrivate, errors.New(enum.Tips_common_wrong + "property_id")
	}

	if requestFindPathInfo.Amount < config.GetOmniDustBtc() {
		return nil, requestFindPathInfo.IsPrivate, errors.New(enum.Tips_common_wrong + "amount")
	}

	if time.Now().After(time.Time(requestFindPathInfo.ExpiryTime)) {
		return nil, requestFindPathInfo.IsPrivate, errors.New(fmt.Sprintf(enum.Tips_htlc_expiryTimeAfterNow, "expiry_time"))
	}

	if requestFindPathInfo.IsPrivate == false {
		//tracker find path
		pathRequest := trackerBean.HtlcPathRequest{}
		pathRequest.H = requestFindPathInfo.H
		pathRequest.PropertyId = requestFindPathInfo.PropertyId
		pathRequest.Amount = requestFindPathInfo.Amount
		pathRequest.RealPayerPeerId = user.PeerId
		pathRequest.PayerObdNodeId = tool.GetObdNodeId()
		pathRequest.PayeePeerId = requestFindPathInfo.RecipientUserPeerId

		cacheDataForTx := &dao.CacheDataForTx{}
		cacheDataForTx.KeyName = user.PeerId + "_" + pathRequest.H
		err = user.Db.Select(q.Eq("KeyName", cacheDataForTx.KeyName)).First(cacheDataForTx)
		if cacheDataForTx.Id != 0 {
			user.Db.DeleteStruct(cacheDataForTx)
		}

		cacheDataForTx.KeyName = user.PeerId + "_" + pathRequest.H
		bytes, _ := json.Marshal(requestFindPathInfo)
		cacheDataForTx.Data = bytes
		user.Db.Save(cacheDataForTx)

		sendMsgToTracker(enum.MsgType_Tracker_GetHtlcPath_351, pathRequest)
		return make(map[string]interface{}), requestFindPathInfo.IsPrivate, nil
	} else {
		requestData.HtlcRequestFindPathInfo = requestFindPathInfo
		return getPrivateChannelForHtlc(requestData, user)
	}
}

func getPrivateChannelForHtlc(requestData *bean.HtlcRequestFindPath, user bean.User) (data interface{}, isPrivate bool, err error) {
	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}
	defer tx.Rollback()
	//get all private channel
	var nodes []dao.ChannelInfo
	err = tx.Select(
		q.Eq("PropertyId", requestData.PropertyId),
		q.Eq("IsPrivate", true),
		q.Eq("CurrState", dao.ChannelState_CanUse),
		q.Or(
			q.And(
				q.Eq("PeerIdB", requestData.RecipientUserPeerId),
				q.Eq("PeerIdA", user.PeerId)),
			q.And(
				q.Eq("PeerIdB", user.PeerId),
				q.Eq("PeerIdA", requestData.RecipientUserPeerId)),
		)).OrderBy("CreateAt").Reverse().Find(&nodes)

	retData := make(map[string]interface{})
	if nodes != nil && len(nodes) > 0 {
		for _, channel := range nodes {
			commitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channel.ChannelId, user.PeerId)
			if err == nil && commitmentTxInfo.Id > 0 {
				if commitmentTxInfo.AmountToRSMC >= requestData.Amount {
					retData["h"] = requestData.H
					retData["is_private"] = requestData.IsPrivate
					retData["property_id"] = requestData.PropertyId
					retData["amount"] = requestData.Amount
					retData["routing_packet"] = channel.ChannelId
					retData["min_cltv_expiry"] = 1
					retData["next_node_peerId"] = requestData.RecipientUserPeerId
					retData["memo"] = requestData.Description
					break
				}
			}
		}
	}
	_ = tx.Commit()
	if len(retData) == 0 {
		return nil, true, errors.New(enum.Tips_htlc_noPrivatePath)
	}
	return retData, true, nil
}

func (service *htlcForwardTxManager) GetResponseFromTrackerOfPayerRequestFindPath(channelPath string, user bean.User) (data interface{}, err error) {
	if tool.CheckIsString(&channelPath) == false {
		err = errors.New("has no channel path")
		log.Println(err)
		return nil, err
	}

	dataArr := strings.Split(channelPath, "_")
	if len(dataArr) != 3 {
		return nil, errors.New("no channel path")
	}

	h := dataArr[0]

	cacheDataForTx := &dao.CacheDataForTx{}
	err = user.Db.Select(q.Eq("KeyName", user.PeerId+"_"+h)).First(cacheDataForTx)
	if cacheDataForTx.Id == 0 {
		err = errors.New("has no channel path")
		log.Println(err)
		return nil, err
	}

	requestFindPathInfo := bean.HtlcRequestFindPathInfo{}
	_ = json.Unmarshal(cacheDataForTx.Data, &requestFindPathInfo)
	if &requestFindPathInfo == nil {
		err = errors.New("has no channel path")
		log.Println(err)
		return nil, err
	}

	splitArr := strings.Split(dataArr[1], ",")
	currChannelInfo := dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("ChannelId", splitArr[0]),
		q.Eq("CurrState", dao.ChannelState_CanUse),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).First(&currChannelInfo)
	if err != nil {
		err = errors.New("has no ChannelPath")
		log.Println(err)
		return nil, err
	}
	nextNodePeerId := currChannelInfo.PeerIdB
	if user.PeerId == currChannelInfo.PeerIdB {
		nextNodePeerId = currChannelInfo.PeerIdA
	}

	arrLength := len(strings.Split(dataArr[1], ","))
	retData := make(map[string]interface{})
	retData["h"] = h
	retData["is_private"] = false
	retData["property_id"] = requestFindPathInfo.PropertyId
	retData["amount"] = requestFindPathInfo.Amount
	retData["routing_packet"] = dataArr[1]
	retData["min_cltv_expiry"] = arrLength
	retData["next_node_peerId"] = nextNodePeerId
	retData["memo"] = requestFindPathInfo.Description

	user.Db.DeleteStruct(cacheDataForTx)

	return retData, nil
}

// step 1 alice -100040协议的alice方的逻辑 alice start a htlc as payer
func (service *htlcForwardTxManager) AliceAddHtlcAtAliceSide(msg bean.RequestMessage, user bean.User) (data interface{}, needSign bool, err error) {
	log.Println("htlc step 1 begin", time.Now())
	if tool.CheckIsString(&msg.Data) == false {
		return nil, false, errors.New(enum.Tips_common_empty + "msg data")
	}

	requestData := &bean.CreateHtlcTxForC3a{}
	err = json.Unmarshal([]byte(msg.Data), requestData)
	if err != nil {
		log.Println(err.Error())
		return nil, false, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	defer tx.Rollback()

	//region check input data 检测输入输入数据
	if requestData.Amount < config.GetOmniDustBtc() {
		return nil, false, errors.New(fmt.Sprintf(enum.Tips_common_amountMustGreater, config.GetOmniDustBtc()))
	}
	if tool.CheckIsString(&requestData.H) == false {
		return nil, false, errors.New(enum.Tips_common_empty + "h")
	}
	if tool.CheckIsString(&requestData.RoutingPacket) == false {
		return nil, false, errors.New(enum.Tips_common_empty + "routing_packet")
	}

	channelIds := strings.Split(requestData.RoutingPacket, ",")
	totalStep := len(channelIds)
	var channelInfo *dao.ChannelInfo
	var currStep = 0
	for index, channelId := range channelIds {
		temp := getChannelInfoByChannelId(tx, channelId, user.PeerId)
		if temp != nil {
			if temp.PeerIdA == msg.RecipientUserPeerId || temp.PeerIdB == msg.RecipientUserPeerId {
				channelInfo = temp
				currStep = index
				break
			}
		}
	}
	if channelInfo == nil {
		return nil, false, errors.New(enum.Tips_htlc_noChanneFromRountingPacket)
	}

	if channelInfo.CurrState == dao.ChannelState_NewTx {
		return nil, false, errors.New(enum.Tips_common_newTxMsg)
	}

	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, user.PeerId)
	duration := time.Now().Sub(fundingTransaction.CreateAt)
	if duration > time.Minute*30 {
		pass, err := checkChannelOmniAssetAmount(*channelInfo)
		if err != nil {
			return nil, false, err
		}
		if pass == false {
			err = errors.New(enum.Tips_rsmc_broadcastedChannel)
			log.Println(err)
			return nil, false, err
		}
	}

	if requestData.CltvExpiry < (totalStep - currStep) {
		requestData.CltvExpiry = totalStep - currStep
	}

	if channelInfo.CurrState < dao.ChannelState_NewTx {
		return nil, false, errors.New("do not finish funding")
	}

	if tool.CheckIsString(&requestData.LastTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "last_temp_address_private_key")
		log.Println(err)
		return nil, false, err
	}

	latestCommitmentTx, _ := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	if latestCommitmentTx.Id > 0 && latestCommitmentTx.CurrState == dao.TxInfoState_Init {
		tx.DeleteStruct(latestCommitmentTx)
	}
	latestCommitmentTx, _ = getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)

	if latestCommitmentTx.Id > 0 {
		if latestCommitmentTx.CurrState == dao.TxInfoState_CreateAndSign {
			_, err = tool.GetPubKeyFromWifAndCheck(requestData.LastTempAddressPrivateKey, latestCommitmentTx.RSMCTempAddressPubKey)
			if err != nil {
				return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongPrivateKeyForLast, requestData.LastTempAddressPrivateKey, latestCommitmentTx.RSMCTempAddressPubKey))
			}
		}
		if latestCommitmentTx.CurrState == dao.TxInfoState_Create {
			if latestCommitmentTx.TxType != dao.CommitmentTransactionType_Htlc {
				return nil, false, errors.New("error commitment tx type")
			}

			if requestData.CurrRsmcTempAddressPubKey != latestCommitmentTx.RSMCTempAddressPubKey {
				return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, requestData.CurrRsmcTempAddressPubKey, latestCommitmentTx.RSMCTempAddressPubKey))
			}

			if requestData.CurrHtlcTempAddressPubKey != latestCommitmentTx.HTLCTempAddressPubKey {
				return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, requestData.CurrHtlcTempAddressPubKey, latestCommitmentTx.HTLCTempAddressPubKey))
			}

			if latestCommitmentTx.LastCommitmentTxId > 0 {
				lastCommitmentTx := &dao.CommitmentTransaction{}
				_ = tx.One("Id", latestCommitmentTx.LastCommitmentTxId, lastCommitmentTx)
				_, err = tool.GetPubKeyFromWifAndCheck(requestData.LastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey)
				if err != nil {
					return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongPrivateKeyForLast, requestData.LastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey))
				}
			}
		}
	}
	if tool.CheckIsString(&requestData.CurrRsmcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_rsmc_temp_address_pub_key")
		log.Println(err)
		return nil, false, err
	}

	if tool.CheckIsString(&requestData.CurrHtlcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_pub_key")
		log.Println(err)
		return nil, false, err
	}

	if tool.CheckIsString(&requestData.CurrHtlcTempAddressForHt1aPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_for_ht1a_pub_key")
		log.Println(err)
		return nil, false, err
	}
	//endregion

	//这次请求的第一次发起
	htlcRequestInfo := &dao.AddHtlcRequestInfo{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("PropertyId", channelInfo.PropertyId),
		q.Eq("H", requestData.H),
		q.Eq("Amount", requestData.Amount),
		q.Eq("RoutingPacket", requestData.RoutingPacket),
		q.Eq("RecipientUserPeerId", msg.RecipientUserPeerId)).First(htlcRequestInfo)

	rawTx := dao.CommitmentTxRawTx{}
	if htlcRequestInfo.Id == 0 || latestCommitmentTx.Id == 0 || latestCommitmentTx.CurrState == dao.TxInfoState_CreateAndSign {
		htlcRequestInfo.RecipientUserPeerId = msg.RecipientUserPeerId
		htlcRequestInfo.H = requestData.H
		htlcRequestInfo.Memo = requestData.Memo
		htlcRequestInfo.PropertyId = channelInfo.PropertyId
		htlcRequestInfo.Amount = requestData.Amount
		htlcRequestInfo.ChannelId = channelInfo.ChannelId
		htlcRequestInfo.RoutingPacket = requestData.RoutingPacket
		htlcRequestInfo.CurrRsmcTempAddressPubKey = requestData.CurrRsmcTempAddressPubKey
		htlcRequestInfo.CurrHtlcTempAddressPubKey = requestData.CurrHtlcTempAddressPubKey
		htlcRequestInfo.CurrHtlcTempAddressForHt1aIndex = requestData.CurrHtlcTempAddressForHt1aIndex
		htlcRequestInfo.CurrHtlcTempAddressForHt1aPubKey = requestData.CurrHtlcTempAddressForHt1aPubKey
		htlcRequestInfo.CurrState = dao.NS_Create
		htlcRequestInfo.CreateAt = time.Now()
		htlcRequestInfo.CreateBy = user.PeerId
		_ = tx.Save(htlcRequestInfo)

		totalStep := len(channelIds)
		latestCommitmentTx, rawTx, err = htlcPayerCreateCommitmentTx_C3a(tx, channelInfo, *requestData, totalStep, currStep, latestCommitmentTx, user)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}

		//更新tracker的htlc的状态
		if channelInfo.IsPrivate == false {
			txStateRequest := trackerBean.UpdateHtlcTxStateRequest{}
			txStateRequest.Path = latestCommitmentTx.HtlcRoutingPacket
			txStateRequest.H = latestCommitmentTx.HtlcH
			txStateRequest.DirectionFlag = trackerBean.HtlcTxState_PayMoney
			txStateRequest.CurrChannelId = channelInfo.ChannelId
			sendMsgToTracker(enum.MsgType_Tracker_UpdateHtlcTxState_352, txStateRequest)
		}
	} else {
		if requestData.CurrHtlcTempAddressForHt1aPubKey != htlcRequestInfo.CurrHtlcTempAddressForHt1aPubKey {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, requestData.CurrHtlcTempAddressForHt1aPubKey, htlcRequestInfo.CurrHtlcTempAddressForHt1aPubKey))
		}
		tx.Select(q.Eq("CommitmentTxId", latestCommitmentTx.Id)).First(&rawTx)
		if rawTx.Id == 0 {
			return nil, false, errors.New("not found rawTx")
		}
	}

	c3aP2pData := &bean.CreateHtlcTxForC3aOfP2p{}
	c3aP2pData.RoutingPacket = requestData.RoutingPacket
	c3aP2pData.ChannelId = channelInfo.ChannelId
	c3aP2pData.H = requestData.H
	c3aP2pData.Amount = requestData.Amount
	c3aP2pData.Memo = requestData.Memo
	c3aP2pData.CltvExpiry = requestData.CltvExpiry
	c3aP2pData.LastTempAddressPrivateKey = requestData.LastTempAddressPrivateKey
	c3aP2pData.CurrRsmcTempAddressPubKey = requestData.CurrRsmcTempAddressPubKey
	c3aP2pData.CurrHtlcTempAddressPubKey = requestData.CurrHtlcTempAddressPubKey
	c3aP2pData.C3aRsmcPartialSignedData = rawTx.RsmcRawTxData
	c3aP2pData.C3aHtlcPartialSignedData = rawTx.HtlcRawTxData
	c3aP2pData.C3aCounterpartyPartialSignedData = rawTx.ToCounterpartyRawTxData
	c3aP2pData.CurrHtlcTempAddressForHt1aPubKey = requestData.CurrHtlcTempAddressForHt1aPubKey
	c3aP2pData.PayerCommitmentTxHash = latestCommitmentTx.CurrHash
	c3aP2pData.PayerNodeAddress = msg.SenderNodePeerId
	c3aP2pData.PayerPeerId = msg.SenderUserPeerId

	if latestCommitmentTx.CurrState == dao.TxInfoState_Init {
		txForC3a := bean.NeedAliceSignCreateHtlcTxForC3a{}
		txForC3a.H = c3aP2pData.H
		txForC3a.ChannelId = latestCommitmentTx.ChannelId
		txForC3a.C3aRsmcRawData = c3aP2pData.C3aRsmcPartialSignedData
		txForC3a.C3aHtlcRawData = c3aP2pData.C3aHtlcPartialSignedData
		txForC3a.C3aCounterpartyRawData = c3aP2pData.C3aCounterpartyPartialSignedData
		txForC3a.PayerNodeAddress = msg.SenderNodePeerId
		txForC3a.PayerPeerId = msg.SenderUserPeerId

		cacheDataForTx := &dao.CacheDataForTx{}
		cacheDataForTx.KeyName = user.PeerId + "_" + channelInfo.ChannelId
		_ = tx.Select(q.Eq("KeyName", cacheDataForTx.KeyName)).First(cacheDataForTx)
		if cacheDataForTx.Id != 0 {
			_ = tx.DeleteStruct(cacheDataForTx)
		}
		bytes, _ := json.Marshal(c3aP2pData)
		cacheDataForTx.Data = bytes
		_ = tx.Save(cacheDataForTx)

		_ = tx.Commit()

		log.Println("htlc step 1 end", time.Now())

		return txForC3a, true, nil
	}
	_ = tx.Commit()
	return c3aP2pData, false, nil
}

// step 2 alice -100100 Alice对C3a的部分签名结果
func (service *htlcForwardTxManager) OnAliceSignedC3aAtAliceSide(msg bean.RequestMessage, user bean.User) (toAlice, toBob interface{}, err error) {
	log.Println("htlc step 2 begin", time.Now())
	if tool.CheckIsString(&msg.Data) == false {
		err = errors.New(enum.Tips_common_empty + "msg.data")
		log.Println(err)
		return nil, nil, err
	}

	signedDataForC3a := bean.AliceSignedHtlcDataForC3a{}
	_ = json.Unmarshal([]byte(msg.Data), &signedDataForC3a)

	if tool.CheckIsString(&signedDataForC3a.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id")
		log.Println(err)
		return nil, nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	cacheDataForTx := &dao.CacheDataForTx{}
	tx.Select(q.Eq("KeyName", user.PeerId+"_"+signedDataForC3a.ChannelId)).First(cacheDataForTx)
	if cacheDataForTx.Id == 0 {
		return nil, nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}

	dataTo40P := &bean.CreateHtlcTxForC3aOfP2p{}
	_ = json.Unmarshal(cacheDataForTx.Data, dataTo40P)

	if len(dataTo40P.ChannelId) == 0 {
		return nil, nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}

	if tool.CheckIsString(&signedDataForC3a.C3aRsmcPartialSignedHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedDataForC3a.C3aRsmcPartialSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "c3a_rsmc_partial_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	if tool.CheckIsString(&signedDataForC3a.C3aHtlcPartialSignedHex) == false {
		err = errors.New(enum.Tips_common_empty + "c3a_htlc_partial_signed_hex")
		log.Println(err)
		return nil, nil, err
	}
	if pass, _ := rpcClient.CheckMultiSign(true, signedDataForC3a.C3aHtlcPartialSignedHex, 1); pass == false {
		err = errors.New(enum.Tips_common_wrong + "counterparty_signed_hex")
		log.Println(err)
		return nil, nil, err
	}

	if tool.CheckIsString(&signedDataForC3a.C3aCounterpartyPartialSignedHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedDataForC3a.C3aCounterpartyPartialSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "c3a_counterparty_partial_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, signedDataForC3a.ChannelId, user.PeerId)
	if err != nil {
		return nil, nil, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
	}

	if len(latestCommitmentTxInfo.RSMCTxHex) > 0 {
		pass, _ := rpcClient.CheckMultiSign(true, signedDataForC3a.C3aRsmcPartialSignedHex, 1)
		if pass == false {
			return nil, nil, errors.New("error sign c3a_rsmc_partial_signed_hex")
		}
		//封装好的签名数据，给bob的客户端签名使用
		latestCommitmentTxInfo.RSMCTxHex = signedDataForC3a.C3aRsmcPartialSignedHex
		latestCommitmentTxInfo.RSMCTxid = rpcClient.GetTxId(signedDataForC3a.C3aRsmcPartialSignedHex)
		latestCommitmentTxInfo.CurrState = dao.TxInfoState_Create
	}

	pass, _ := rpcClient.CheckMultiSign(true, signedDataForC3a.C3aHtlcPartialSignedHex, 1)
	if pass == false {
		return nil, nil, errors.New("error sign c3a_htlc_partial_signed_hex")
	}
	//封装好的签名数据，给bob的客户端签名使用
	latestCommitmentTxInfo.HtlcTxHex = signedDataForC3a.C3aHtlcPartialSignedHex
	latestCommitmentTxInfo.HTLCTxid = rpcClient.GetTxId(signedDataForC3a.C3aHtlcPartialSignedHex)

	if len(latestCommitmentTxInfo.ToCounterpartyTxHex) > 0 {
		pass, _ := rpcClient.CheckMultiSign(true, signedDataForC3a.C3aCounterpartyPartialSignedHex, 1)
		if pass == false {
			return nil, nil, errors.New("error sign c3a_counterparty_partial_signed_hex")
		}
		//封装好的签名数据，给bob的客户端签名使用
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedDataForC3a.C3aCounterpartyPartialSignedHex
		latestCommitmentTxInfo.ToCounterpartyTxid = rpcClient.GetTxId(signedDataForC3a.C3aCounterpartyPartialSignedHex)
	}

	latestCommitmentTxInfo.CurrState = dao.TxInfoState_Create
	_ = tx.Update(latestCommitmentTxInfo)
	_ = tx.Commit()

	dataTo40P.C3aRsmcPartialSignedData.Hex = signedDataForC3a.C3aRsmcPartialSignedHex
	dataTo40P.C3aHtlcPartialSignedData.Hex = signedDataForC3a.C3aHtlcPartialSignedHex
	dataTo40P.C3aCounterpartyPartialSignedData.Hex = signedDataForC3a.C3aCounterpartyPartialSignedHex

	toAliceResult := bean.AliceSignedHtlcDataForC3aResult{}
	toAliceResult.ChannelId = dataTo40P.ChannelId
	toAliceResult.CommitmentTxHash = dataTo40P.PayerCommitmentTxHash
	log.Println("htlc step 2 end", time.Now())
	return toAliceResult, dataTo40P, nil
}

// step 3 bob -40号协议 缓存来自40号协议的信息 推送110040消息，需要bob对C3a的交易进行签名
func (service *htlcForwardTxManager) BeforeBobSignAddHtlcRequestAtBobSide_40(msgData string, user bean.User) (data interface{}, err error) {
	log.Println("htlc step 3 begin", time.Now())
	requestAddHtlc := &bean.CreateHtlcTxForC3aOfP2p{}
	_ = json.Unmarshal([]byte(msgData), requestAddHtlc)
	channelId := requestAddHtlc.ChannelId

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
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found channel info")
	}

	channelInfo.CurrState = dao.ChannelState_NewTx
	_ = tx.Update(channelInfo)

	cacheDataForTx := &dao.CacheDataForTx{}
	cacheDataForTx.KeyName = requestAddHtlc.PayerCommitmentTxHash
	tx.Select(q.Eq("KeyName", cacheDataForTx.KeyName)).First(cacheDataForTx)
	if cacheDataForTx.Id != 0 {
		tx.DeleteStruct(cacheDataForTx)
	}
	cacheDataForTx.Data = []byte(msgData)
	_ = tx.Save(cacheDataForTx)

	_ = tx.Commit()

	toBobData := bean.CreateHtlcTxForC3aToBob{}
	toBobData.ChannelId = requestAddHtlc.ChannelId
	toBobData.H = requestAddHtlc.H
	toBobData.PayerCommitmentTxHash = requestAddHtlc.PayerCommitmentTxHash
	toBobData.PayerPeerId = requestAddHtlc.PayerPeerId
	toBobData.PayerNodeAddress = requestAddHtlc.PayerNodeAddress
	toBobData.C3aRsmcPartialSignedData = requestAddHtlc.C3aRsmcPartialSignedData
	toBobData.C3aCounterpartyPartialSignedData = requestAddHtlc.C3aCounterpartyPartialSignedData
	toBobData.C3aHtlcPartialSignedData = requestAddHtlc.C3aHtlcPartialSignedData
	log.Println("htlc step 3 end", time.Now())
	return toBobData, nil
}

// step 4 bob 响应-100041号协议，创建C3a的Rsmc的Rd和Br，toHtlc的Br，Ht1a，Hlock，以及C3b的toB，toRsmc，toHtlc
func (service *htlcForwardTxManager) BobSignedAddHtlcAtBobSide(jsonData string, user bean.User) (returnData interface{}, err error) {
	log.Println("htlc step 4 begin", time.Now())
	if tool.CheckIsString(&jsonData) == false {
		err := errors.New(enum.Tips_common_empty + "msg data")
		log.Println(err)
		return nil, err
	}

	requestData := bean.BobSignedC3a{}
	err = json.Unmarshal([]byte(jsonData), &requestData)
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

	if tool.CheckIsString(&requestData.PayerCommitmentTxHash) == false {
		return nil, errors.New(enum.Tips_common_empty + "payer_commitment_tx_hash")
	}

	cacheDataForTx := &dao.CacheDataForTx{}
	err = tx.Select(q.Eq("KeyName", requestData.PayerCommitmentTxHash)).First(cacheDataForTx)
	if cacheDataForTx.Id == 0 {
		return nil, errors.New(enum.Tips_common_wrong + "payer_commitment_tx_hash")
	}
	aliceMsg := string(cacheDataForTx.Data)
	if tool.CheckIsString(&aliceMsg) == false {
		return nil, errors.New(enum.Tips_common_wrong + "payer_commitment_tx_hash")
	}

	payerRequestAddHtlc := &bean.CreateHtlcTxForC3aOfP2p{}
	_ = json.Unmarshal([]byte(aliceMsg), payerRequestAddHtlc)

	if len(payerRequestAddHtlc.C3aRsmcPartialSignedData.Hex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, requestData.C3aCompleteSignedRsmcHex, 2); pass == false {
			err = errors.New(enum.Tips_common_empty + "c3a_complete_signed_rsmc_hex")
			log.Println(err)
			return nil, err
		}
		payerRequestAddHtlc.C3aRsmcPartialSignedData.Hex = requestData.C3aCompleteSignedRsmcHex
	}

	if pass, _ := rpcClient.CheckMultiSign(true, requestData.C3aCompleteSignedHtlcHex, 2); pass == false {
		err = errors.New(enum.Tips_common_empty + "c3a_complete_signed_htlc_hex")
		log.Println(err)
		return nil, err
	}
	payerRequestAddHtlc.C3aHtlcPartialSignedData.Hex = requestData.C3aCompleteSignedHtlcHex

	if len(payerRequestAddHtlc.C3aCounterpartyPartialSignedData.Hex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, requestData.C3aCompleteSignedCounterpartyHex, 2); pass == false {
			err = errors.New(enum.Tips_common_empty + "c3a_complete_signed_counterparty_hex")
			log.Println(err)
			return nil, err
		}
		payerRequestAddHtlc.C3aCounterpartyPartialSignedData.Hex = requestData.C3aCompleteSignedCounterpartyHex
	}

	channelId := payerRequestAddHtlc.ChannelId

	needBobSignData := bean.NeedBobSignHtlcTxOfC3b{}
	needBobSignData.ChannelId = channelId

	toAliceDataOf41P := bean.NeedAliceSignHtlcTxOfC3bP2p{}

	toAliceDataOf41P.PayerCommitmentTxHash = requestData.PayerCommitmentTxHash
	toAliceDataOf41P.PayeePeerId = user.PeerId
	toAliceDataOf41P.PayeeNodeAddress = user.P2PLocalPeerId

	if len(payerRequestAddHtlc.C3aRsmcPartialSignedData.Hex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, requestData.C3aCompleteSignedRsmcHex, 2); pass == false {
			return nil, errors.New(enum.Tips_common_wrong + "c3a_complete_signed_rsmc_hex")
		}
	}
	toAliceDataOf41P.C3aCompleteSignedRsmcHex = requestData.C3aCompleteSignedRsmcHex

	if pass, _ := rpcClient.CheckMultiSign(true, requestData.C3aCompleteSignedHtlcHex, 2); pass == false {
		return nil, errors.New(enum.Tips_common_wrong + "c3a_complete_signed_htlc_hex")
	}
	toAliceDataOf41P.C3aCompleteSignedHtlcHex = requestData.C3aCompleteSignedHtlcHex

	if len(payerRequestAddHtlc.C3aCounterpartyPartialSignedData.Hex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, requestData.C3aCompleteSignedCounterpartyHex, 2); pass == false {
			return nil, errors.New(enum.Tips_common_wrong + "c3a_complete_signed_counterparty_hex")
		}
	}
	toAliceDataOf41P.C3aCompleteSignedCounterpartyHex = requestData.C3aCompleteSignedCounterpartyHex

	// region check input data

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if err != nil {
		return nil, errors.New(enum.Tips_htlc_noChanneFromRountingPacket)
	}

	if channelInfo.CurrState < dao.ChannelState_NewTx {
		return nil, errors.New("do not finish funding")
	}

	toAliceDataOf41P.ChannelId = channelInfo.ChannelId
	bobChannelPubKey := channelInfo.PubKeyB
	if user.PeerId == channelInfo.PeerIdA {
		bobChannelPubKey = channelInfo.PubKeyA
	}

	latestCommitmentTxInfo, _ := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Init {
		_ = tx.DeleteStruct(latestCommitmentTxInfo)
	} else {
		latestCommitmentTxInfo, _ = getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	}

	if latestCommitmentTxInfo.Id > 0 {
		if tool.CheckIsString(&requestData.LastTempAddressPrivateKey) == false {
			err = errors.New(enum.Tips_common_empty + "last_temp_address_private_key")
			log.Println(err)
			return nil, err
		}

		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign {
			_, err = tool.GetPubKeyFromWifAndCheck(requestData.LastTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey)
			if err != nil {
				return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongPrivateKeyForLast, requestData.LastTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
			}
		}
		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Create {
			if latestCommitmentTxInfo.TxType != dao.CommitmentTransactionType_Htlc {
				return nil, errors.New("error commitment tx type")
			}

			if requestData.CurrRsmcTempAddressPubKey != latestCommitmentTxInfo.RSMCTempAddressPubKey {
				return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, requestData.CurrRsmcTempAddressPubKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
			}
			if requestData.CurrHtlcTempAddressPubKey != latestCommitmentTxInfo.HTLCTempAddressPubKey {
				return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, requestData.CurrHtlcTempAddressPubKey, latestCommitmentTxInfo.HTLCTempAddressPubKey))
			}

			if latestCommitmentTxInfo.LastCommitmentTxId > 0 {
				lastCommitmentTx := &dao.CommitmentTransaction{}
				_ = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, lastCommitmentTx)
				_, err = tool.GetPubKeyFromWifAndCheck(requestData.LastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey)
				if err != nil {
					return nil, err
				}
			}
		}
		toAliceDataOf41P.PayeeLastTempAddressPrivateKey = requestData.LastTempAddressPrivateKey
	}
	if tool.CheckIsString(&requestData.CurrRsmcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_rsmc_temp_address_pub_key")
		log.Println(err)
		return nil, err
	}
	toAliceDataOf41P.PayeeCurrRsmcTempAddressPubKey = requestData.CurrRsmcTempAddressPubKey

	if tool.CheckIsString(&requestData.CurrHtlcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_pub_key")
		log.Println(err)
		return nil, err
	}
	toAliceDataOf41P.PayeeCurrHtlcTempAddressPubKey = requestData.CurrHtlcTempAddressPubKey
	//endregion

	//region 1、验证C3a的Rsmc的签名
	var c3aRsmcTxId, c3aSignedRsmcHex string
	var c3aRsmcMultiAddress, c3aRsmcRedeemScript, c3aRsmcMultiAddressScriptPubKey string
	var c3aRsmcOutputs []bean.TransactionInputItem
	if tool.CheckIsString(&payerRequestAddHtlc.C3aRsmcPartialSignedData.Hex) {
		c3aSignedRsmcHex = requestData.C3aCompleteSignedRsmcHex
		c3aRsmcTxId = rpcClient.GetTxId(c3aSignedRsmcHex)

		// region 根据alice的临时地址+bob的通道address,获取alice2+bob的多签地址，并得到AliceSignedRsmcHex签名后的交易的input，为创建alice的RD和bob的BR做准备
		c3aRsmcMultiAddress, c3aRsmcRedeemScript, c3aRsmcMultiAddressScriptPubKey, err = createMultiSig(payerRequestAddHtlc.CurrRsmcTempAddressPubKey, bobChannelPubKey)
		if err != nil {
			return nil, err
		}

		c3aRsmcOutputs, err = getInputsForNextTxByParseTxHashVout(c3aSignedRsmcHex, c3aRsmcMultiAddress, c3aRsmcMultiAddressScriptPubKey, c3aRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if len(c3aRsmcOutputs) == 0 {
			return nil, errors.New(enum.Tips_common_wrong + "payer rsmc hex")
		}
		//endregion
	}
	//endregion

	// region 2、验证C3a的toCounterpartyTxHex
	c3aSignedToCounterpartyTxHex := ""
	if len(payerRequestAddHtlc.C3aCounterpartyPartialSignedData.Hex) > 0 {
		c3aSignedToCounterpartyTxHex = payerRequestAddHtlc.C3aCounterpartyPartialSignedData.Hex
	}
	//endregion

	// region 3、验证C3a的 htlcHex
	c3aSignedHtlcHex := requestData.C3aCompleteSignedHtlcHex
	c3aHtlcTxId := rpcClient.GetTxId(c3aSignedHtlcHex)

	// region 根据alice的htlc临时地址+bob的通道address,获取alice2+bob的多签地址，并得到AliceSignedHtlcHex签名后的交易的input，为创建bob的HBR做准备
	c3aHtlcMultiAddress, c3aHtlcRedeemScript, c3aHtlcAddrScriptPubKey, err := createMultiSig(payerRequestAddHtlc.CurrHtlcTempAddressPubKey, bobChannelPubKey)
	if err != nil {
		return nil, err
	}

	c3aHtlcOutputs, err := getInputsForNextTxByParseTxHashVout(c3aSignedHtlcHex, c3aHtlcMultiAddress, c3aHtlcAddrScriptPubKey, c3aHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//endregion
	//endregion

	//获取bob最新的承诺交易
	isFirstRequest := false
	if latestCommitmentTxInfo != nil && latestCommitmentTxInfo.Id > 0 {
		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign {
			if latestCommitmentTxInfo.TxType != dao.CommitmentTransactionType_Rsmc {
				return nil, errors.New("wrong commitment tx type " + strconv.Itoa(int(latestCommitmentTxInfo.TxType)))
			}
		}
		if latestCommitmentTxInfo.CurrState != dao.TxInfoState_CreateAndSign && latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
			return nil, errors.New("wrong commitment tx state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		}

		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign { //有上一次的承诺交易
			isFirstRequest = true
		}
	} else { // 因为没有充值，没有最初的承诺交易C1b
		isFirstRequest = true
	}

	var amountToOther = 0.0
	var amountToHtlc = 0.0
	//如果是本轮的第一次请求交易
	rawTx := dao.CommitmentTxRawTx{}
	if isFirstRequest {
		//region 4、根据对方传过来的上一个交易的临时rsmc私钥，签名最近的BR交易，保证对方确实放弃了上一个承诺交易
		err := signLastBR(tx, dao.BRType_Rmsc, *channelInfo, user.PeerId, payerRequestAddHtlc.LastTempAddressPrivateKey, latestCommitmentTxInfo.Id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		//endregion

		fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, user.PeerId)
		if fundingTransaction == nil {
			return nil, errors.New(enum.Tips_common_notFound + "fundingTransaction")
		}

		//region 5、创建C3b
		latestCommitmentTxInfo, rawTx, err = htlcPayeeCreateCommitmentTx_C3b(tx, channelInfo, requestData, *payerRequestAddHtlc, latestCommitmentTxInfo, c3aSignedToCounterpartyTxHex, user)
		//endregion
	}

	amountToOther = latestCommitmentTxInfo.AmountToCounterparty
	amountToHtlc = latestCommitmentTxInfo.AmountToHtlc

	needBobSignData.C3bCounterpartyRawData = rawTx.ToCounterpartyRawTxData
	needBobSignData.C3bRsmcRawData = rawTx.RsmcRawTxData
	needBobSignData.C3bHtlcRawData = rawTx.HtlcRawTxData

	toAliceDataOf41P.C3bCounterpartyPartialSignedData = rawTx.ToCounterpartyRawTxData
	toAliceDataOf41P.C3bRsmcPartialSignedData = rawTx.RsmcRawTxData
	toAliceDataOf41P.C3bHtlcPartialSignedData = rawTx.HtlcRawTxData
	toAliceDataOf41P.PayeeCommitmentTxHash = latestCommitmentTxInfo.CurrHash

	var myAddress = channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdA {
		myAddress = channelInfo.AddressA
	}

	if len(c3aRsmcOutputs) > 0 {

		//region 6、根据alice C3a的Rsmc输出，创建对应的BR,为下一个交易做准备，create BR2b tx  for bob
		tempC3aRsmc := &dao.CommitmentTransaction{}
		tempC3aRsmc.Id = latestCommitmentTxInfo.Id
		tempC3aRsmc.PropertyId = channelInfo.PropertyId
		tempC3aRsmc.RSMCTempAddressPubKey = payerRequestAddHtlc.CurrRsmcTempAddressPubKey
		tempC3aRsmc.RSMCMultiAddress = c3aRsmcMultiAddress
		tempC3aRsmc.RSMCMultiAddressScriptPubKey = c3aRsmcMultiAddressScriptPubKey
		tempC3aRsmc.RSMCRedeemScript = c3aRsmcRedeemScript
		tempC3aRsmc.RSMCTxHex = c3aSignedRsmcHex
		tempC3aRsmc.RSMCTxid = c3aRsmcTxId
		tempC3aRsmc.AmountToRSMC = latestCommitmentTxInfo.AmountToCounterparty

		c3aRsmcBrHexData, err := createRawBR(dao.BRType_Rmsc, channelInfo, tempC3aRsmc, c3aRsmcOutputs, myAddress, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		c3aRsmcBrHexData.PubKeyA = payerRequestAddHtlc.CurrRsmcTempAddressPubKey
		c3aRsmcBrHexData.PubKeyB = bobChannelPubKey
		needBobSignData.C3aRsmcBrRawData = c3aRsmcBrHexData
		//endregion

		//region 7、根据签名后的AliceRsmc创建alice的RD create RD tx for alice
		aliceRdOutputAddress := channelInfo.AddressA
		if user.PeerId == channelInfo.PeerIdA {
			aliceRdOutputAddress = channelInfo.AddressB
		}
		c3aRsmcRdData, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
			c3aRsmcMultiAddress,
			c3aRsmcOutputs,
			aliceRdOutputAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			amountToOther,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&c3aRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_failToCreate, "RD raw transacation"))
		}

		signHexData := bean.NeedClientSignTxData{}
		signHexData.Hex = c3aRsmcRdData["hex"].(string)
		signHexData.Inputs = c3aRsmcRdData["inputs"]
		signHexData.IsMultisig = true
		signHexData.PubKeyA = payerRequestAddHtlc.CurrRsmcTempAddressPubKey
		signHexData.PubKeyB = bobChannelPubKey
		needBobSignData.C3aRsmcRdRawData = signHexData
		toAliceDataOf41P.C3aRsmcRdPartialSignedData = signHexData
		//endregion
	}

	// region 7、根据alice C3a的Htlc输出，创建对应的BR,为下一个交易做准备，create HBR2b tx  for bob
	tempC3aHtlc := &dao.CommitmentTransaction{}
	tempC3aHtlc.Id = latestCommitmentTxInfo.Id
	tempC3aHtlc.PropertyId = channelInfo.PropertyId
	tempC3aHtlc.RSMCTempAddressPubKey = payerRequestAddHtlc.CurrHtlcTempAddressPubKey
	tempC3aHtlc.RSMCMultiAddress = c3aHtlcMultiAddress
	tempC3aHtlc.RSMCRedeemScript = c3aHtlcRedeemScript
	tempC3aHtlc.RSMCMultiAddressScriptPubKey = c3aHtlcAddrScriptPubKey
	tempC3aHtlc.RSMCTxHex = c3aSignedHtlcHex
	tempC3aHtlc.RSMCTxid = c3aHtlcTxId
	tempC3aHtlc.AmountToRSMC = latestCommitmentTxInfo.AmountToHtlc
	c3aHtlcBrHexData, err := createRawBR(dao.BRType_Htlc, channelInfo, tempC3aHtlc, c3aHtlcOutputs, myAddress, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c3aHtlcBrHexData.PubKeyA = payerRequestAddHtlc.CurrHtlcTempAddressPubKey
	c3aHtlcBrHexData.PubKeyB = bobChannelPubKey
	needBobSignData.C3aHtlcBrRawData = c3aHtlcBrHexData
	//endregion

	// region  8、h+bobChannelPubkey 锁定给bob的付款金额
	lockByHForBobTx, err := createHtlcLockByHForBobAtPayeeSide(*channelInfo, *payerRequestAddHtlc, c3aSignedHtlcHex, bobChannelPubKey, channelInfo.PropertyId, amountToHtlc)
	if err != nil {
		return nil, err
	}
	lockByHForBobTx.PubKeyA = payerRequestAddHtlc.CurrHtlcTempAddressPubKey
	lockByHForBobTx.PubKeyB = bobChannelPubKey

	needBobSignData.C3aHtlcHlockRawData = *lockByHForBobTx
	toAliceDataOf41P.C3aHtlcHlockPartialSignedData = *lockByHForBobTx
	//endregion

	// region 9、ht1a 根据signedHtlcHex（alice签名后C3a的第三个输出）作为输入生成
	ht1aTxData, err := createHT1aForAlice(*channelInfo, *payerRequestAddHtlc, c3aSignedHtlcHex, bobChannelPubKey, channelInfo.PropertyId, amountToHtlc, latestCommitmentTxInfo.HtlcCltvExpiry)
	if err != nil {
		return nil, err
	}
	ht1aTxData.PubKeyA = bobChannelPubKey
	ht1aTxData.PubKeyB = payerRequestAddHtlc.CurrHtlcTempAddressPubKey
	needBobSignData.C3aHtlcHtRawData = *ht1aTxData
	toAliceDataOf41P.C3aHtlcHtPartialSignedData = *ht1aTxData
	toAliceDataOf41P.C3aHtlcTempAddressForHtPubKey = payerRequestAddHtlc.CurrHtlcTempAddressForHt1aPubKey
	//endregion

	channelInfo.CurrState = dao.ChannelState_HtlcTx
	_ = tx.Update(channelInfo)

	if service.tempDataSendTo41PAtBobSide == nil {
		service.tempDataSendTo41PAtBobSide = make(map[string]bean.NeedAliceSignHtlcTxOfC3bP2p)
	}
	service.tempDataSendTo41PAtBobSide[user.PeerId+"_"+channelInfo.ChannelId] = toAliceDataOf41P
	_ = tx.Commit()
	log.Println("htlc step 4 end", time.Now())
	return needBobSignData, nil
}

// step 5 bob -100101 bob完成对C3b的签名，构建41号协议的消息体，推送41号协议
func (service *htlcForwardTxManager) OnBobSignedC3bAtBobSide(msg bean.RequestMessage, user bean.User) (toAlice, toBob interface{}, err error) {
	log.Println("htlc step 5 begin", time.Now())
	c3bResult := bean.BobSignedHtlcTxOfC3b{}
	err = json.Unmarshal([]byte(msg.Data), &c3bResult)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	key := user.PeerId + "_" + c3bResult.ChannelId
	toAliceDataOfP2p := service.tempDataSendTo41PAtBobSide[key]

	if len(toAliceDataOfP2p.ChannelId) == 0 {
		return nil, nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	cacheDataForTx := dao.CacheDataForTx{}
	err = tx.Select(q.Eq("KeyName", toAliceDataOfP2p.PayerCommitmentTxHash)).First(&cacheDataForTx)
	if cacheDataForTx.Id == 0 {
		return nil, nil, errors.New(enum.Tips_common_wrong + "payer_commitment_tx_hash")
	}

	aliceMsg := string(cacheDataForTx.Data)
	if tool.CheckIsString(&aliceMsg) == false {
		return nil, nil, errors.New(enum.Tips_common_empty + "payer_commitment_tx_hash")
	}

	payerRequestAddHtlc := &bean.CreateHtlcTxForC3aOfP2p{}
	_ = json.Unmarshal([]byte(aliceMsg), payerRequestAddHtlc)

	if len(payerRequestAddHtlc.C3aRsmcPartialSignedData.Hex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, toAliceDataOfP2p.C3aCompleteSignedRsmcHex, 2); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c3a_complete_signed_rsmc_hex")
		}
	}

	if pass, _ := rpcClient.CheckMultiSign(true, toAliceDataOfP2p.C3aCompleteSignedHtlcHex, 2); pass == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + "c3a_complete_signed_htlc_hex")
	}

	if len(payerRequestAddHtlc.C3aCounterpartyPartialSignedData.Hex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, toAliceDataOfP2p.C3aCompleteSignedCounterpartyHex, 2); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c3a_complete_signed_counterparty_hex")
		}
	}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", c3bResult.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if err != nil {
		return nil, nil, errors.New(enum.Tips_htlc_noChanneFromRountingPacket)
	}

	latestCommitmentTxInfo, _ := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Init {
		return nil, nil, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
	}

	bobChannelPubKey := channelInfo.PubKeyB
	var myAddress = channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdA {
		myAddress = channelInfo.AddressA
		bobChannelPubKey = channelInfo.PubKeyA
	}

	if tool.CheckIsString(&toAliceDataOfP2p.C3aRsmcRdPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(false, c3bResult.C3aRsmcRdPartialSignedHex, 1); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c3a_rsmc_rd_partial_signed_hex")
		}
		toAliceDataOfP2p.C3aRsmcRdPartialSignedData.Hex = c3bResult.C3aRsmcRdPartialSignedHex

		if pass, _ := rpcClient.CheckMultiSign(false, c3bResult.C3aRsmcBrPartialSignedHex, 1); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c3a_rsmc_br_partial_signed_hex")
		}

		c3aSignedRsmcHex := toAliceDataOfP2p.C3aCompleteSignedRsmcHex
		c3aRsmcTxId := rpcClient.GetTxId(c3aSignedRsmcHex)

		c3aRsmcMultiAddress, c3aRsmcRedeemScript, c3aRsmcMultiAddressScriptPubKey, err := createMultiSig(payerRequestAddHtlc.CurrRsmcTempAddressPubKey, bobChannelPubKey)
		if err != nil {
			return nil, nil, err
		}

		c3aRsmcOutputs, err := getInputsForNextTxByParseTxHashVout(c3aSignedRsmcHex, c3aRsmcMultiAddress, c3aRsmcMultiAddressScriptPubKey, c3aRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}
		tempC3aRsmc := &dao.CommitmentTransaction{}
		tempC3aRsmc.Id = latestCommitmentTxInfo.Id
		tempC3aRsmc.PropertyId = channelInfo.PropertyId
		tempC3aRsmc.RSMCTempAddressPubKey = payerRequestAddHtlc.CurrRsmcTempAddressPubKey
		tempC3aRsmc.RSMCMultiAddress = c3aRsmcMultiAddress
		tempC3aRsmc.RSMCMultiAddressScriptPubKey = c3aRsmcMultiAddressScriptPubKey
		tempC3aRsmc.RSMCRedeemScript = c3aRsmcRedeemScript
		tempC3aRsmc.RSMCTxHex = toAliceDataOfP2p.C3aCompleteSignedRsmcHex
		tempC3aRsmc.RSMCTxid = c3aRsmcTxId
		tempC3aRsmc.AmountToRSMC = latestCommitmentTxInfo.AmountToCounterparty
		_ = createCurrCommitmentTxPartialSignedBR(tx, dao.BRType_Rmsc, channelInfo, tempC3aRsmc, c3aRsmcOutputs, myAddress, c3bResult.C3aRsmcBrPartialSignedHex, user)
	}

	if pass, _ := rpcClient.CheckMultiSign(false, c3bResult.C3aHtlcHtPartialSignedHex, 1); pass == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + "c3a_htlc_ht_partial_signed_hex")
	}
	toAliceDataOfP2p.C3aHtlcHtPartialSignedData.Hex = c3bResult.C3aHtlcHtPartialSignedHex

	if pass, _ := rpcClient.CheckMultiSign(false, c3bResult.C3aHtlcHlockPartialSignedHex, 1); pass == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + "c3a_htlc_hlock_partial_signed_hex")
	}
	toAliceDataOfP2p.C3aHtlcHlockPartialSignedData.Hex = c3bResult.C3aHtlcHlockPartialSignedHex

	if pass, _ := rpcClient.CheckMultiSign(false, c3bResult.C3aHtlcBrPartialSignedHex, 1); pass == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + "c3a_htlc_br_partial_signed_hex")
	}

	c3aSignedHtlcHex := toAliceDataOfP2p.C3aCompleteSignedHtlcHex
	c3aHtlcMultiAddress, c3aHtlcRedeemScript, c3aHtlcAddrScriptPubKey, err := createMultiSig(payerRequestAddHtlc.CurrHtlcTempAddressPubKey, bobChannelPubKey)
	if err != nil {
		return nil, nil, err
	}

	c3aHtlcOutputs, err := getInputsForNextTxByParseTxHashVout(c3aSignedHtlcHex, c3aHtlcMultiAddress, c3aHtlcAddrScriptPubKey, c3aHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	tempC3aHtlc := &dao.CommitmentTransaction{}
	tempC3aHtlc.Id = latestCommitmentTxInfo.Id
	tempC3aHtlc.PropertyId = channelInfo.PropertyId
	tempC3aHtlc.RSMCTempAddressPubKey = payerRequestAddHtlc.CurrHtlcTempAddressPubKey
	tempC3aHtlc.RSMCMultiAddress = c3aHtlcMultiAddress
	tempC3aHtlc.RSMCRedeemScript = c3aHtlcRedeemScript
	tempC3aHtlc.RSMCMultiAddressScriptPubKey = c3aHtlcAddrScriptPubKey
	tempC3aHtlc.RSMCTxHex = c3aSignedHtlcHex
	tempC3aHtlc.RSMCTxid = rpcClient.GetTxId(c3aSignedHtlcHex)
	tempC3aHtlc.AmountToRSMC = latestCommitmentTxInfo.AmountToHtlc
	_ = createCurrCommitmentTxPartialSignedBR(tx, dao.BRType_Htlc, channelInfo, tempC3aHtlc, c3aHtlcOutputs, myAddress, c3bResult.C3aHtlcBrPartialSignedHex, user)

	if tool.CheckIsString(&toAliceDataOfP2p.C3bRsmcPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, c3bResult.C3bRsmcPartialSignedHex, 1); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c3b_rsmc_partial_signed_hex")
		}
		toAliceDataOfP2p.C3bRsmcPartialSignedData.Hex = c3bResult.C3bRsmcPartialSignedHex
	}
	if tool.CheckIsString(&toAliceDataOfP2p.C3bCounterpartyPartialSignedData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, c3bResult.C3bCounterpartyPartialSignedHex, 1); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c3b_counterparty_partial_signed_hex")
		}
		toAliceDataOfP2p.C3bCounterpartyPartialSignedData.Hex = c3bResult.C3bCounterpartyPartialSignedHex
	}

	if pass, _ := rpcClient.CheckMultiSign(true, c3bResult.C3bHtlcPartialSignedHex, 1); pass == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + "c3b_htlc_partial_signed_hex")
	}
	toAliceDataOfP2p.C3bHtlcPartialSignedData.Hex = c3bResult.C3bHtlcPartialSignedHex

	latestCommitmentTxInfo.CurrState = dao.TxInfoState_Create
	_ = tx.UpdateField(latestCommitmentTxInfo, "CurrState", dao.TxInfoState_Create)

	_ = tx.Commit()

	delete(service.tempDataSendTo41PAtBobSide, key)

	toBobData := bean.BobSignedHtlcTxOfC3bResult{}
	toBobData.ChannelId = channelInfo.ChannelId
	toBobData.CommitmentTxHash = latestCommitmentTxInfo.CurrHash
	log.Println("htlc step 5 end", time.Now())
	return toAliceDataOfP2p, toBobData, nil
}

// step 6 alice p2p 41号协议，构建需要alice签名的数据，缓存41号协议的数据， 推送（110041）信息给alice签名
func (service *htlcForwardTxManager) AfterBobSignAddHtlcAtAliceSide_41(msgData string, user bean.User) (data interface{}, needNoticeBob bool, err error) {
	log.Println("htlc step 6 begin", time.Now())
	dataFromBob := bean.NeedAliceSignHtlcTxOfC3bP2p{}
	_ = json.Unmarshal([]byte(msgData), &dataFromBob)

	channelId := dataFromBob.ChannelId
	commitmentTxHash := dataFromBob.PayerCommitmentTxHash

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	defer tx.Rollback()
	commitmentTransaction := &dao.CommitmentTransaction{}
	err = tx.Select(
		q.Eq("CurrHash", commitmentTxHash),
		q.Eq("ChannelId", channelId)).First(commitmentTransaction)
	if err != nil {
		return nil, true, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if channelInfo.Id == 0 {
		return nil, true, errors.New("not found the channel " + channelId)
	}

	htlcRequestInfo := &dao.AddHtlcRequestInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("H", commitmentTransaction.HtlcH),
		q.Eq("PropertyId", channelInfo.PropertyId)).
		OrderBy("CreateAt").
		Reverse().
		First(htlcRequestInfo)
	if err != nil {
		return nil, false, err
	}

	needAliceSignHtlcTxOfC3b := bean.NeedAliceSignHtlcTxOfC3b{}
	needAliceSignHtlcTxOfC3b.ChannelId = channelId
	needAliceSignHtlcTxOfC3b.C3aRsmcRdPartialSignedData = dataFromBob.C3aRsmcRdPartialSignedData
	needAliceSignHtlcTxOfC3b.C3aHtlcHlockPartialSignedData = dataFromBob.C3aHtlcHlockPartialSignedData
	needAliceSignHtlcTxOfC3b.C3aHtlcHtPartialSignedData = dataFromBob.C3aHtlcHtPartialSignedData
	needAliceSignHtlcTxOfC3b.C3bRsmcPartialSignedData = dataFromBob.C3bRsmcPartialSignedData
	needAliceSignHtlcTxOfC3b.C3bHtlcPartialSignedData = dataFromBob.C3bHtlcPartialSignedData
	needAliceSignHtlcTxOfC3b.C3bCounterpartyPartialSignedData = dataFromBob.C3bCounterpartyPartialSignedData

	if service.tempDataFrom41PAtAliceSide == nil {
		service.tempDataFrom41PAtAliceSide = make(map[string]bean.NeedAliceSignHtlcTxOfC3bP2p)
	}
	service.tempDataFrom41PAtAliceSide[user.PeerId+"_"+channelId] = dataFromBob
	log.Println("htlc step 6 end", time.Now())
	return needAliceSignHtlcTxOfC3b, false, nil
}

// step 7 alice 响应100102号协议，根据C3b的签名结果创建C3b的rmsc的rd，br，htlc的br，htd，hlock，以及创建C3a的htrd和htbr
func (service *htlcForwardTxManager) OnAliceSignC3bAtAliceSide(msg bean.RequestMessage, user bean.User) (interface{}, error) {
	log.Println("htlc step 7 begin", time.Now())
	aliceSignedC3b := bean.AliceSignedHtlcTxOfC3bResult{}
	_ = json.Unmarshal([]byte(msg.Data), &aliceSignedC3b)

	channelId := aliceSignedC3b.ChannelId
	dataFromBob := service.tempDataFrom41PAtAliceSide[user.PeerId+"_"+channelId]

	//为了准备给42传数据
	needBobSignData := bean.NeedBobSignHtlcSubTxOfC3bP2p{}
	needBobSignData.ChannelId = channelId
	needBobSignData.PayeeCommitmentTxHash = dataFromBob.PayeeCommitmentTxHash
	needBobSignData.C3aHtlcTempAddressForHtPubKey = dataFromBob.C3aHtlcTempAddressForHtPubKey
	needBobSignData.PayerPeerId = user.PeerId
	needBobSignData.PayerNodeAddress = user.P2PLocalPeerId

	needAliceSignData := bean.NeedAliceSignHtlcSubTxOfC3b{}
	needAliceSignData.ChannelId = channelId
	needAliceSignData.PayeePeerId = dataFromBob.PayeePeerId
	needAliceSignData.PayeeNodeAddress = dataFromBob.PayeeNodeAddress

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()
	commitmentTransaction := &dao.CommitmentTransaction{}
	err = tx.Select(
		q.Eq("CurrHash", dataFromBob.PayerCommitmentTxHash),
		q.Eq("ChannelId", channelId)).First(commitmentTransaction)
	if err != nil {
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found the channel " + channelId)
	}

	bobRsmcHexIsExist := false
	var c3bSignedRsmcHex, c3bSignedRsmcTxid, c3bSignedToCounterpartyTxHex string
	if len(dataFromBob.C3bRsmcPartialSignedData.Hex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, aliceSignedC3b.C3bRsmcCompleteSignedHex, 2); pass == false {
			return nil, errors.New(enum.Tips_common_wrong + "c3b_rsmc_complete_signed_hex")
		}
		bobRsmcHexIsExist = true
		c3bSignedRsmcHex = aliceSignedC3b.C3bRsmcCompleteSignedHex
		dataFromBob.C3bRsmcPartialSignedData.Hex = c3bSignedRsmcHex
		c3bSignedRsmcTxid = rpcClient.GetTxId(c3bSignedRsmcHex)
	}

	if len(dataFromBob.C3bCounterpartyPartialSignedData.Hex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, aliceSignedC3b.C3bCounterpartyCompleteSignedHex, 2); pass == false {
			return nil, errors.New(enum.Tips_common_wrong + "c3b_counterparty_complete_signed_hex")
		}
		c3bSignedToCounterpartyTxHex = aliceSignedC3b.C3bCounterpartyCompleteSignedHex
		dataFromBob.C3bCounterpartyPartialSignedData.Hex = c3bSignedToCounterpartyTxHex
	}

	if pass, _ := rpcClient.CheckMultiSign(true, aliceSignedC3b.C3bHtlcCompleteSignedHex, 2); pass == false {
		return nil, errors.New(enum.Tips_common_wrong + "c3b_htlc_complete_signed_hex")
	}
	c3bSignedHtlcHex := aliceSignedC3b.C3bHtlcCompleteSignedHex
	dataFromBob.C3bHtlcPartialSignedData.Hex = c3bSignedHtlcHex

	if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedC3b.C3aHtlcHtCompleteSignedHex, 2); pass == false {
		return nil, errors.New(enum.Tips_common_wrong + "c3a_htlc_ht_complete_signed_hex")
	}

	if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedC3b.C3aHtlcHlockCompleteSignedHex, 2); pass == false {
		return nil, errors.New(enum.Tips_common_wrong + "c3a_htlc_hlock_complete_signed_hex")
	}

	if tool.CheckIsString(&aliceSignedC3b.C3aRsmcRdCompleteSignedHex) {
		if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedC3b.C3aRsmcRdCompleteSignedHex, 2); pass == false {
			return nil, errors.New(enum.Tips_common_wrong + "c3a_rsmc_rd_complete_signed_hex")
		}
	}

	payerChannelPubKey := channelInfo.PubKeyA
	payerChannelAddress := channelInfo.AddressA
	payeeChannelAddress := channelInfo.AddressB
	payeeChannelPubKey := channelInfo.PubKeyB
	if user.PeerId == channelInfo.PeerIdB {
		payerChannelPubKey = channelInfo.PubKeyB
		payerChannelAddress = channelInfo.AddressB
		payeeChannelAddress = channelInfo.AddressA
		payeeChannelPubKey = channelInfo.PubKeyA
	}

	//region 1 c3b rsmcHex
	needBobSignData.C3bCompleteSignedRsmcHex = c3bSignedRsmcHex
	commitmentTransaction.FromCounterpartySideForMeTxHex = c3bSignedRsmcHex
	//endregion

	// region 2 c3b toCounterpartyTxHex
	needBobSignData.C3bCompleteSignedCounterpartyHex = c3bSignedToCounterpartyTxHex
	//endregion

	// region 3 c3b htlcHex
	needBobSignData.C3bCompleteSignedHtlcHex = c3bSignedHtlcHex
	//endregion

	var bobRsmcOutputs []bean.TransactionInputItem
	if bobRsmcHexIsExist {

		//region 4 c3b Rsmc rd
		c3bRsmcMultiAddress, c3bRsmcRedeemScript, c3bRsmcAddrScriptPubKey, err := createMultiSig(dataFromBob.PayeeCurrRsmcTempAddressPubKey, payerChannelPubKey)
		if err != nil {
			return nil, err
		}
		bobRsmcOutputs, err = getInputsForNextTxByParseTxHashVout(c3bSignedRsmcHex, c3bRsmcMultiAddress, c3bRsmcAddrScriptPubKey, c3bRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		c3bRsmcRdTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
			c3bRsmcMultiAddress,
			bobRsmcOutputs,
			payeeChannelAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			commitmentTransaction.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&c3bRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, errors.New("fail to create rd for c3b rsmc")
		}
		rdRawData := bean.NeedClientSignTxData{}
		rdRawData.Hex = c3bRsmcRdTx["hex"].(string)
		rdRawData.Inputs = c3bRsmcRdTx["inputs"]
		rdRawData.IsMultisig = true
		rdRawData.PubKeyA = dataFromBob.PayeeCurrRsmcTempAddressPubKey
		rdRawData.PubKeyB = payerChannelPubKey
		needBobSignData.C3bRsmcRdPartialData = rdRawData
		needAliceSignData.C3bRsmcRdRawData = rdRawData
		//endregion create rd tx for alice

		//region 5 c3b Rsmc br
		tempOtherSideCommitmentTx := &dao.CommitmentTransaction{}
		tempOtherSideCommitmentTx.Id = commitmentTransaction.Id
		tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
		tempOtherSideCommitmentTx.RSMCTempAddressPubKey = dataFromBob.PayeeCurrRsmcTempAddressPubKey
		tempOtherSideCommitmentTx.RSMCMultiAddress = c3bRsmcMultiAddress
		tempOtherSideCommitmentTx.RSMCRedeemScript = c3bRsmcRedeemScript
		tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = c3bRsmcAddrScriptPubKey
		tempOtherSideCommitmentTx.RSMCTxHex = c3bSignedRsmcHex
		tempOtherSideCommitmentTx.RSMCTxid = c3bSignedRsmcTxid
		tempOtherSideCommitmentTx.AmountToRSMC = commitmentTransaction.AmountToCounterparty
		rawBR, err := createRawBR(dao.BRType_Rmsc, channelInfo, tempOtherSideCommitmentTx, bobRsmcOutputs, payerChannelAddress, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		rawBR.PubKeyA = dataFromBob.PayeeCurrRsmcTempAddressPubKey
		rawBR.PubKeyB = payerChannelPubKey
		needAliceSignData.C3bRsmcBrRawData = rawBR
		//endregion create br tx for alice
	}

	//region 6 c3b htlc htd
	htlcTimeOut := commitmentTransaction.HtlcCltvExpiry
	bobHtlcMultiAddress, bobHtlcRedeemScript, bobHtlcMultiAddressScriptPubKey, err := createMultiSig(dataFromBob.PayeeCurrHtlcTempAddressPubKey, payerChannelPubKey)
	if err != nil {
		return nil, err
	}
	bobHtlcOutputs, err := getInputsForNextTxByParseTxHashVout(c3bSignedHtlcHex, bobHtlcMultiAddress, bobHtlcMultiAddressScriptPubKey, bobHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c3bHtdTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		bobHtlcMultiAddress,
		bobHtlcOutputs,
		payerChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		commitmentTransaction.AmountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		htlcTimeOut,
		&bobHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create HTD1b for C3b")
	}
	c3bHtdRawData := bean.NeedClientSignTxData{}
	c3bHtdRawData.Hex = c3bHtdTx["hex"].(string)
	c3bHtdRawData.Inputs = c3bHtdTx["inputs"]
	c3bHtdRawData.IsMultisig = true
	c3bHtdRawData.PubKeyA = dataFromBob.PayeeCurrHtlcTempAddressPubKey
	c3bHtdRawData.PubKeyB = payerChannelPubKey
	needBobSignData.C3bHtlcHtdPartialData = c3bHtdRawData
	needAliceSignData.C3bHtlcHtdRawData = c3bHtdRawData
	//endregion

	//region 7 c3b htlc Hlock
	c3bHlockTx, err := createHtlcLockByHForBobAtPayerSide(*channelInfo, dataFromBob, c3bSignedHtlcHex, commitmentTransaction.HtlcH, payeeChannelPubKey, payerChannelPubKey, channelInfo.PropertyId, commitmentTransaction.AmountToHtlc)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create HlockHex for C3b")
	}
	needBobSignData.C3bHtlcHlockPartialData = c3bHlockTx
	needAliceSignData.C3bHtlcHlockRawData = c3bHlockTx
	//endregion

	//region 8 c3b htlc br
	c3bHtlcBrTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		bobHtlcMultiAddress,
		bobHtlcOutputs,
		payerChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		commitmentTransaction.AmountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		0,
		&bobHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create HTD1b for C3b")
	}
	c3bHtlcBrRawData := bean.NeedClientSignTxData{}
	c3bHtlcBrRawData.Hex = c3bHtlcBrTx["hex"].(string)
	c3bHtlcBrRawData.Inputs = c3bHtlcBrTx["inputs"]
	c3bHtlcBrRawData.IsMultisig = true
	c3bHtlcBrRawData.PubKeyA = dataFromBob.PayeeCurrHtlcTempAddressPubKey
	c3bHtlcBrRawData.PubKeyB = payerChannelPubKey
	needAliceSignData.C3bHtlcBrRawData = c3bHtlcBrRawData
	//endregion

	//region 9  c3a htlc ht htrd
	c3aHtHex := aliceSignedC3b.C3aHtlcHtCompleteSignedHex
	c3aHtMultiAddress, c3aHtRedeemScript, c3aHtAddrScriptPubKey, err := createMultiSig(dataFromBob.C3aHtlcTempAddressForHtPubKey, payeeChannelPubKey)
	if err != nil {
		return nil, err
	}
	c3aHtOutputs, err := getInputsForNextTxByParseTxHashVout(c3aHtHex, c3aHtMultiAddress, c3aHtAddrScriptPubKey, c3aHtRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c3bHtRdTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		c3aHtMultiAddress,
		c3aHtOutputs,
		payerChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		commitmentTransaction.AmountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		1000,
		&c3aHtRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create HTD1b for C3b")
	}

	c3aHtRdRawData := bean.NeedClientSignTxData{}
	c3aHtRdRawData.Hex = c3bHtRdTx["hex"].(string)
	c3aHtRdRawData.Inputs = c3bHtRdTx["inputs"]
	c3aHtRdRawData.IsMultisig = true
	c3aHtRdRawData.PubKeyA = dataFromBob.C3aHtlcTempAddressForHtPubKey
	c3aHtRdRawData.PubKeyB = payeeChannelPubKey
	needBobSignData.C3aHtlcHtHex = c3aHtHex
	needBobSignData.C3aHtlcHtrdPartialData = c3aHtRdRawData
	needAliceSignData.C3aHtlcHtrdRawData = c3aHtRdRawData

	//endregion

	//region 10  c3a htlc ht htbr
	c3aHtBrTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		c3aHtMultiAddress,
		c3aHtOutputs,
		payeeChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		commitmentTransaction.AmountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		0,
		&c3aHtRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create HTD1b for C3b")
	}

	c3aHtBrRawData := bean.NeedClientSignTxData{}
	c3aHtBrRawData.Hex = c3aHtBrTx["hex"].(string)
	c3aHtBrRawData.Inputs = c3aHtBrTx["inputs"]
	c3aHtBrRawData.IsMultisig = true
	c3aHtBrRawData.PubKeyA = dataFromBob.C3aHtlcTempAddressForHtPubKey
	c3aHtBrRawData.PubKeyB = payeeChannelPubKey
	needBobSignData.C3aHtlcHtbrRawData = c3aHtBrRawData
	//endregion

	//region 11  c3a htlc hlock hed
	c3aHlockHex := aliceSignedC3b.C3aHtlcHlockCompleteSignedHex
	c3aHlockMultiAddress, c3aHlockRedeemScript, c3aHlockAddrScriptPubKey, err := createMultiSig(commitmentTransaction.HtlcH, payeeChannelPubKey)
	if err != nil {
		return nil, err
	}
	c3aHlocOutputs, err := getInputsForNextTxByParseTxHashVout(c3aHlockHex, c3aHlockMultiAddress, c3aHlockAddrScriptPubKey, c3aHlockRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c3aHedTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		c3aHlockMultiAddress,
		c3aHlocOutputs,
		payeeChannelAddress,
		payeeChannelAddress,
		channelInfo.PropertyId,
		commitmentTransaction.AmountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		0,
		&c3aHlockRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create Hed1b for C3a")
	}

	c3aHedRawData := bean.NeedClientSignTxData{}
	c3aHedRawData.Hex = c3aHedTx["hex"].(string)
	c3aHedRawData.Inputs = c3aHedTx["inputs"]
	c3aHedRawData.IsMultisig = true
	c3aHedRawData.PubKeyA = commitmentTransaction.HtlcH
	c3aHedRawData.PubKeyB = payeeChannelPubKey
	needBobSignData.C3aHtlcHedRawData = c3aHedRawData
	//endregion

	if service.tempDataSendTo42PAtAliceSide == nil {
		service.tempDataSendTo42PAtAliceSide = make(map[string]bean.NeedBobSignHtlcSubTxOfC3bP2p)
	}
	service.tempDataSendTo42PAtAliceSide[user.PeerId+"_"+channelId] = needBobSignData
	_ = tx.Commit()
	log.Println("htlc step 7 end", time.Now())
	return needAliceSignData, nil
}

// step 8 alice 响应 100103号协议，更新alice的承诺交易，推送42号p2p协议
func (service *htlcForwardTxManager) OnAliceSignedC3bSubTxAtAliceSide(msg bean.RequestMessage, user bean.User) (toAlice, toBob interface{}, err error) {
	log.Println("htlc step 8 begin", time.Now())
	aliceSignedC3b := bean.AliceSignedHtlcSubTxOfC3b{}
	_ = json.Unmarshal([]byte(msg.Data), &aliceSignedC3b)

	channelId := aliceSignedC3b.ChannelId
	dataFromBob_41 := service.tempDataFrom41PAtAliceSide[user.PeerId+"_"+channelId]
	needBobSignData := service.tempDataSendTo42PAtAliceSide[user.PeerId+"_"+channelId]

	bobRsmcHexIsExist := false
	if len(needBobSignData.C3bRsmcRdPartialData.Hex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedC3b.C3bRsmcRdPartialSignedHex, 1); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c3b_rsmc_rd_partial_signed_hex")
		}
		needBobSignData.C3bRsmcRdPartialData.Hex = aliceSignedC3b.C3bRsmcRdPartialSignedHex

		if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedC3b.C3bRsmcBrPartialSignedHex, 1); pass == false {
			return nil, nil, errors.New(enum.Tips_common_wrong + "c3b_rsmc_br_partial_signed_hex")
		}
		bobRsmcHexIsExist = true
	}

	if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedC3b.C3bHtlcHtdPartialSignedHex, 1); pass == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + "c3b_htlc_htd_partial_signed_hex")
	}
	needBobSignData.C3bHtlcHtdPartialData.Hex = aliceSignedC3b.C3bHtlcHtdPartialSignedHex

	if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedC3b.C3bHtlcBrPartialSignedHex, 1); pass == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + "c3b_htlc_br_partial_signed_hex")
	}

	if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedC3b.C3bHtlcHlockPartialSignedHex, 1); pass == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + "c3b_htlc_hlock_partial_signed_hex")
	}
	needBobSignData.C3bHtlcHlockPartialData.Hex = aliceSignedC3b.C3bHtlcHlockPartialSignedHex

	if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedC3b.C3aHtlcHtrdPartialSignedHex, 1); pass == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + "c3a_htlc_htrd_partial_signed_hex")
	}
	needBobSignData.C3aHtlcHtrdPartialData.Hex = aliceSignedC3b.C3aHtlcHtrdPartialSignedHex

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()
	commitmentTx := &dao.CommitmentTransaction{}
	err = tx.Select(
		q.Eq("CurrHash", dataFromBob_41.PayerCommitmentTxHash),
		q.Eq("ChannelId", channelId)).First(commitmentTx)
	if err != nil {
		return nil, nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if channelInfo.Id == 0 {
		return nil, nil, errors.New("not found the channel " + channelId)
	}

	htlcRequestInfo := &dao.AddHtlcRequestInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("H", commitmentTx.HtlcH),
		q.Eq("PropertyId", channelInfo.PropertyId)).
		OrderBy("CreateAt").
		Reverse().
		First(htlcRequestInfo)
	if err != nil {
		return nil, nil, err
	}

	htlcRequestInfo.CurrState = dao.NS_Finish
	_ = tx.Update(htlcRequestInfo)

	payerChannelPubKey := channelInfo.PubKeyA
	payerChannelAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		payerChannelPubKey = channelInfo.PubKeyB
		payerChannelAddress = channelInfo.AddressB
	}

	//region 处理付款方收到的已经签名C3a的子交易，及上一个BR的签名，RSMCBR，HBR的创建
	if commitmentTx.CurrState == dao.TxInfoState_Create {

		//region 1 根据对方传过来的上一个交易的临时rsmc私钥，签名最近的BR交易，保证对方确实放弃了上一个承诺交易
		err := signLastBR(tx, dao.BRType_Rmsc, *channelInfo, user.PeerId, dataFromBob_41.PayeeLastTempAddressPrivateKey, commitmentTx.LastCommitmentTxId)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}
		//endregion

		// region 2 保存c3b的rsmc的br到本地
		tempOtherSideCommitmentTx := &dao.CommitmentTransaction{}
		if bobRsmcHexIsExist {
			bobRsmcMultiAddress, bobRsmcRedeemScript, bobRsmcMultiAddressScriptPubKey, err := createMultiSig(dataFromBob_41.PayeeCurrRsmcTempAddressPubKey, payerChannelPubKey)
			if err != nil {
				return nil, nil, err
			}
			c3bRsmcOutputs, err := getInputsForNextTxByParseTxHashVout(dataFromBob_41.C3bRsmcPartialSignedData.Hex, bobRsmcMultiAddress, bobRsmcMultiAddressScriptPubKey, bobRsmcRedeemScript)
			if err != nil {
				log.Println(err)
				return nil, nil, err
			}
			tempOtherSideCommitmentTx.Id = commitmentTx.Id
			tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
			tempOtherSideCommitmentTx.RSMCTempAddressPubKey = dataFromBob_41.PayeeCurrRsmcTempAddressPubKey
			tempOtherSideCommitmentTx.RSMCMultiAddress = bobRsmcMultiAddress
			tempOtherSideCommitmentTx.RSMCRedeemScript = bobRsmcRedeemScript
			tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = bobRsmcMultiAddressScriptPubKey
			tempOtherSideCommitmentTx.RSMCTxHex = dataFromBob_41.C3bRsmcPartialSignedData.Hex
			tempOtherSideCommitmentTx.RSMCTxid = rpcClient.GetTxId(tempOtherSideCommitmentTx.RSMCTxHex)
			tempOtherSideCommitmentTx.AmountToRSMC = commitmentTx.AmountToCounterparty
			err = createCurrCommitmentTxPartialSignedBR(tx, dao.BRType_Rmsc, channelInfo, tempOtherSideCommitmentTx, c3bRsmcOutputs, payerChannelAddress, aliceSignedC3b.C3bRsmcBrPartialSignedHex, user)
			if err != nil {
				log.Println(err)
				return nil, nil, err
			}
		}
		//endregion

		// region 3 保存c3b的htlc的br到本地
		bobHtlcMultiAddress, bobHtlcRedeemScript, bobHtlcMultiAddressScriptPubKey, err := createMultiSig(dataFromBob_41.PayeeCurrHtlcTempAddressPubKey, payerChannelPubKey)
		if err != nil {
			return nil, nil, err
		}

		c3bHtlcOutputs, err := getInputsForNextTxByParseTxHashVout(dataFromBob_41.C3bHtlcPartialSignedData.Hex, bobHtlcMultiAddress, bobHtlcMultiAddressScriptPubKey, bobHtlcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}

		tempOtherSideCommitmentTx.Id = commitmentTx.Id
		tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
		tempOtherSideCommitmentTx.RSMCTempAddressPubKey = dataFromBob_41.PayeeCurrHtlcTempAddressPubKey
		tempOtherSideCommitmentTx.RSMCMultiAddress = bobHtlcMultiAddress
		tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = bobHtlcMultiAddressScriptPubKey
		tempOtherSideCommitmentTx.RSMCRedeemScript = bobHtlcRedeemScript
		tempOtherSideCommitmentTx.RSMCTxHex = dataFromBob_41.C3bHtlcPartialSignedData.Hex
		tempOtherSideCommitmentTx.RSMCTxid = rpcClient.GetTxId(tempOtherSideCommitmentTx.RSMCTxHex)
		tempOtherSideCommitmentTx.AmountToRSMC = commitmentTx.AmountToHtlc
		err = createCurrCommitmentTxPartialSignedBR(tx, dao.BRType_Htlc, channelInfo, tempOtherSideCommitmentTx, c3bHtlcOutputs, payerChannelAddress, aliceSignedC3b.C3bHtlcBrPartialSignedHex, user)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}
		//endregion

		//region 4 更新收到的签名交易
		_, _, err = checkHexAndUpdateC3aOn42Protocal(tx, dataFromBob_41, *htlcRequestInfo, *channelInfo, commitmentTx, user)
		if err != nil {
			return nil, nil, err
		}
		//endregion
	}
	//endregion

	commitmentTx.CurrState = dao.TxInfoState_Htlc_WaitHTRD1aSign
	commitmentTx.SignAt = time.Now()
	_ = tx.Update(commitmentTx)
	_ = tx.Commit()

	toAliceData := bean.AliceSignedHtlcSubTxOfC3bResult{}
	toAliceData.ChannelId = commitmentTx.ChannelId
	toAliceData.CommitmentTxId = commitmentTx.CurrHash
	log.Println("htlc step 8 end", time.Now())
	return toAliceData, needBobSignData, nil
}

// step 9 bob 响应 42号协议 构造需要bob签名的数据，缓存来自42号协议的数据 推送110042
func (service *htlcForwardTxManager) OnGetNeedBobSignC3bSubTxAtBobSide(msgData string, user bean.User) (interface{}, error) {
	log.Println("htlc step 9 begin", time.Now())
	c3bCacheData := bean.NeedBobSignHtlcSubTxOfC3bP2p{}
	_ = json.Unmarshal([]byte(msgData), &c3bCacheData)

	if service.tempDataFrom42PAtBobSide == nil {
		service.tempDataFrom42PAtBobSide = make(map[string]bean.NeedBobSignHtlcSubTxOfC3bP2p)
	}
	service.tempDataFrom42PAtBobSide[user.PeerId+"_"+c3bCacheData.ChannelId] = c3bCacheData

	needBobSign := bean.NeedBobSignHtlcSubTxOfC3b{}
	needBobSign.ChannelId = c3bCacheData.ChannelId
	needBobSign.C3aHtlcHedRawData = c3bCacheData.C3aHtlcHedRawData
	needBobSign.C3aHtlcHtrdPartialData = c3bCacheData.C3aHtlcHtrdPartialData
	needBobSign.C3aHtlcHtbrRawData = c3bCacheData.C3aHtlcHtbrRawData
	needBobSign.C3bRsmcRdPartialData = c3bCacheData.C3bRsmcRdPartialData
	needBobSign.C3bHtlcHlockPartialData = c3bCacheData.C3bHtlcHlockPartialData
	needBobSign.C3bHtlcHtdPartialData = c3bCacheData.C3bHtlcHtdPartialData
	log.Println("htlc step 9 end", time.Now())
	return needBobSign, nil
}

// step 10 bob 响应100104:缓存签名的结果，生成hlock的 he让bob继续签名
func (service *htlcForwardTxManager) OnBobSignedC3bSubTxAtBobSide(msg bean.RequestMessage, user bean.User) (interface{}, error) {
	log.Println("htlc step 10 begin", time.Now())
	jsonObj := bean.BobSignedHtlcSubTxOfC3b{}
	_ = json.Unmarshal([]byte(msg.Data), &jsonObj)

	if tool.CheckIsString(&jsonObj.ChannelId) == false {
		return nil, errors.New(enum.Tips_common_empty + "channel_id")
	}

	if tool.CheckIsString(&jsonObj.CurrHtlcTempAddressForHePubKey) == false {
		return nil, errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_for_he_pub_key")
	}

	c3bCacheData := service.tempDataFrom42PAtBobSide[user.PeerId+"_"+jsonObj.ChannelId]

	if pass, _ := rpcClient.CheckMultiSign(false, jsonObj.C3aHtlcHtrdCompleteSignedHex, 2); pass == false {
		return nil, errors.New("error sign c3a_htlc_htrd_complete_signed_hex")
	}
	c3bCacheData.C3aHtlcHtrdPartialData.Hex = jsonObj.C3aHtlcHtrdCompleteSignedHex

	if pass, _ := rpcClient.CheckMultiSign(false, jsonObj.C3aHtlcHedPartialSignedHex, 1); pass == false {
		return nil, errors.New("error sign c3a_htlc_hed_partial_signed_hex")
	}
	c3bCacheData.C3aHtlcHedRawData.Hex = jsonObj.C3aHtlcHedPartialSignedHex

	if pass, _ := rpcClient.CheckMultiSign(false, jsonObj.C3aHtlcHtbrPartialSignedHex, 1); pass == false {
		return nil, errors.New("error sign c3a_htlc_htbr_partial_signed_hex")
	}
	c3bCacheData.C3aHtlcHtbrRawData.Hex = jsonObj.C3aHtlcHtbrPartialSignedHex

	if tool.CheckIsString(&jsonObj.C3bRsmcRdCompleteSignedHex) {
		if pass, _ := rpcClient.CheckMultiSign(false, jsonObj.C3bRsmcRdCompleteSignedHex, 2); pass == false {
			return nil, errors.New("error sign c3b_rsmc_rd_complete_signed_hex")
		}
		c3bCacheData.C3bRsmcRdPartialData.Hex = jsonObj.C3bRsmcRdCompleteSignedHex
	}

	if pass, _ := rpcClient.CheckMultiSign(false, jsonObj.C3bHtlcHlockCompleteSignedHex, 2); pass == false {
		return nil, errors.New("error sign c3b_htlc_hlock_complete_signed_hex")
	}
	c3bCacheData.C3bHtlcHlockPartialData.Hex = jsonObj.C3bHtlcHlockCompleteSignedHex

	if pass, _ := rpcClient.CheckMultiSign(false, jsonObj.C3bHtlcHtdCompleteSignedHex, 2); pass == false {
		return nil, errors.New("error sign c3b_htlc_htd_complete_signed_hex")
	}
	c3bCacheData.C3bHtlcHtdPartialData.Hex = jsonObj.C3bHtlcHtdCompleteSignedHex

	service.tempDataFrom42PAtBobSide[user.PeerId+"_"+jsonObj.ChannelId] = c3bCacheData

	//根据Hlock，创建C3b的He H+bobChannelPubkey作为输入，aliceChannelPubkey+bob的临时地址3作为输出
	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	latestCommitmentTx, err := getLatestCommitmentTxUseDbTx(tx, jsonObj.ChannelId, user.PeerId)
	if err != nil {
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", latestCommitmentTx.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found the channel " + jsonObj.ChannelId)
	}

	c3bHeRawData, err := saveHtlcHeTxForPayee(tx, *channelInfo, latestCommitmentTx, jsonObj, jsonObj.C3bHtlcHlockCompleteSignedHex, user)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	needBobSignData := bean.NeedBobSignHtlcHeTxOfC3b{}
	needBobSignData.ChannelId = jsonObj.ChannelId
	needBobSignData.C3bHtlcHlockHeRawData = *c3bHeRawData
	needBobSignData.PayerPeerId = c3bCacheData.PayerPeerId
	needBobSignData.PayerNodeAddress = c3bCacheData.PayerNodeAddress
	log.Println("htlc step 10 end", time.Now())
	return needBobSignData, nil
}

// step 11 bob 响应100105:收款方完成Hlock的he的部分签名，更新C3b的信息，最后推送43给alice和推送正向H的创建htlc的结果
func (service *htlcForwardTxManager) OnBobSignHtRdAtBobSide_42(msgData string, user bean.User) (toAlice, toBob interface{}, err error) {
	log.Println("htlc step 11 begin", time.Now())
	jsonObj := bean.BobSignedHtlcHeTxOfC3b{}
	_ = json.Unmarshal([]byte(msgData), &jsonObj)

	if tool.CheckIsString(&jsonObj.ChannelId) == false {
		return nil, nil, errors.New(enum.Tips_common_empty + "channel_id")
	}

	dataFrom42P := service.tempDataFrom42PAtBobSide[user.PeerId+"_"+jsonObj.ChannelId]
	if len(dataFrom42P.ChannelId) == 0 {
		return nil, nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}

	if pass, _ := rpcClient.CheckMultiSign(false, jsonObj.C3bHtlcHlockHePartialSignedHex, 1); pass == false {
		return nil, nil, errors.New("error sign c3b_htlc_hlock_he_partial_signed_hex")
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	c3aRetData := bean.C3aSignedHerdTxOfC3bP2p{}
	c3aRetData.ChannelId = dataFrom42P.ChannelId

	latestCommitmentTx := &dao.CommitmentTransaction{}
	err = tx.Select(q.Eq("CurrHash", dataFrom42P.PayeeCommitmentTxHash)).First(latestCommitmentTx)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(q.Eq("ChannelId", latestCommitmentTx.ChannelId)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	bobChannelPubKey := channelInfo.PubKeyB
	bobChannelAddress := channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdA {
		bobChannelPubKey = channelInfo.PubKeyA
		bobChannelAddress = channelInfo.AddressA
	}

	//region 1 返回给alice的htrd签名数据
	c3aRetData.C3aHtlcHtrdCompleteSignedHex = dataFrom42P.C3aHtlcHtrdPartialData.Hex
	//endregion
	c3aRetData.C3aHtlcHedPartialSignedHex = dataFrom42P.C3aHtlcHedRawData.Hex

	if latestCommitmentTx.CurrState == dao.TxInfoState_Create {
		//创建ht的BR
		c3aHtMultiAddress, c3aHtRedeemScript, c3aHtMultiAddressScriptPubKey, err := createMultiSig(dataFrom42P.C3aHtlcTempAddressForHtPubKey, bobChannelPubKey)
		if err != nil {
			return nil, nil, err
		}

		c3aHtOutputs, err := getInputsForNextTxByParseTxHashVout(dataFrom42P.C3aHtlcHtHex, c3aHtMultiAddress, c3aHtMultiAddressScriptPubKey, c3aHtRedeemScript)
		if err != nil {
			return nil, nil, err
		}

		tempHt1a := &dao.CommitmentTransaction{}
		tempHt1a.Id = latestCommitmentTx.Id
		tempHt1a.PropertyId = channelInfo.PropertyId
		tempHt1a.RSMCTempAddressPubKey = dataFrom42P.C3aHtlcTempAddressForHtPubKey
		tempHt1a.RSMCMultiAddress = c3aHtMultiAddress
		tempHt1a.RSMCRedeemScript = c3aHtRedeemScript
		tempHt1a.RSMCMultiAddressScriptPubKey = c3aHtMultiAddressScriptPubKey
		tempHt1a.RSMCTxHex = dataFrom42P.C3aHtlcHtHex
		tempHt1a.RSMCTxid = c3aHtOutputs[0].Txid
		tempHt1a.AmountToRSMC = latestCommitmentTx.AmountToHtlc
		err = createCurrCommitmentTxPartialSignedBR(tx, dao.BRType_Ht1a, channelInfo, tempHt1a, c3aHtOutputs, bobChannelAddress, dataFrom42P.C3aHtlcHtbrRawData.Hex, user)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}

		latestCommitmentTx, _, err = checkHexAndUpdateC3bOn42Protocal(tx, dataFrom42P, *channelInfo, latestCommitmentTx, user)
		if err != nil {
			log.Println(err.Error())
			return nil, nil, err
		}

		//更新He交易
		_ = updateHtlcHeTxForPayee(tx, *channelInfo, latestCommitmentTx, jsonObj.C3bHtlcHlockHePartialSignedHex)
	}

	channelInfo.CurrState = dao.ChannelState_HtlcTx
	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	key := user.PeerId + "_" + channelInfo.ChannelId
	delete(service.tempDataFrom42PAtBobSide, key)
	log.Println("htlc step 11 end", time.Now())
	return c3aRetData, latestCommitmentTx, nil
}

// step 12 响应43号协议: 保存htrd和hed 推送110043给Alice
func (service *htlcForwardTxManager) OnGetHtrdTxDataFromBobAtAliceSide_43(msgData string, user bean.User) (data interface{}, err error) {
	log.Println("htlc step 12 begin", time.Now())
	c3aHtrdData := bean.C3aSignedHerdTxOfC3bP2p{}
	_ = json.Unmarshal([]byte(msgData), &c3aHtrdData)

	if tool.CheckIsString(&c3aHtrdData.ChannelId) == false {
		return nil, errors.New(enum.Tips_common_empty + "channel_id")
	}

	if pass, _ := rpcClient.CheckMultiSign(false, c3aHtrdData.C3aHtlcHedPartialSignedHex, 1); pass == false {
		return nil, errors.New("error sign c3a_htlc_hed_partial_data")
	}

	if pass, _ := rpcClient.CheckMultiSign(false, c3aHtrdData.C3aHtlcHtrdCompleteSignedHex, 2); pass == false {
		return nil, errors.New("error sign c3a_htlc_htrd_complete_signed_hex")
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	latestCommitmentTx, err := getLatestCommitmentTxUseDbTx(tx, c3aHtrdData.ChannelId, user.PeerId)
	if err != nil {
		return nil, err
	}

	channelInfo := dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", latestCommitmentTx.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(&channelInfo)
	if &channelInfo == nil {
		return nil, errors.New("not found the channel " + c3aHtrdData.ChannelId)
	}

	if latestCommitmentTx.CurrState == dao.TxInfoState_Htlc_WaitHTRD1aSign {
		_, err := saveHtRD1a(tx, c3aHtrdData.C3aHtlcHtrdCompleteSignedHex, *latestCommitmentTx, user)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
		_ = createHed1a(tx, c3aHtrdData.C3aHtlcHedPartialSignedHex, channelInfo, *latestCommitmentTx, user)
	}

	latestCommitmentTx.CurrState = dao.TxInfoState_Htlc_GetH
	_ = tx.Update(latestCommitmentTx)

	channelInfo.CurrState = dao.ChannelState_HtlcTx
	_ = tx.Update(&channelInfo)

	//同步通道信息到tracker
	sendChannelStateToTracker(channelInfo, *latestCommitmentTx)

	key := user.PeerId + "_" + channelInfo.ChannelId
	cacheDataForTx := &dao.CacheDataForTx{}
	tx.Select(q.Eq("KeyName", key)).First(cacheDataForTx)
	if cacheDataForTx.Id != 0 {
		_ = tx.DeleteStruct(cacheDataForTx)
	}
	delete(service.tempDataFrom41PAtAliceSide, key)
	delete(service.tempDataSendTo42PAtAliceSide, key)
	_ = tx.Commit()
	log.Println("htlc step 12 end", time.Now())
	return latestCommitmentTx, nil
}

// 创建付款方C3a
func htlcPayerCreateCommitmentTx_C3a(tx storm.Node, channelInfo *dao.ChannelInfo, requestData bean.CreateHtlcTxForC3a, totalStep int, currStep int, latestCommitmentTx *dao.CommitmentTransaction, user bean.User) (*dao.CommitmentTransaction, dao.CommitmentTxRawTx, error) {
	rawTx := dao.CommitmentTxRawTx{}
	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, user.PeerId)
	if fundingTransaction == nil {
		return nil, rawTx, errors.New("not found fundingTransaction")
	}
	// htlc的资产分配方案
	var outputBean = commitmentTxOutputBean{}
	amountAndFee, _ := decimal.NewFromFloat(requestData.Amount).Mul(decimal.NewFromFloat(1 + config.GetHtlcFee()*float64(totalStep-(currStep+1)))).Round(8).Float64()
	outputBean.RsmcTempPubKey = requestData.CurrRsmcTempAddressPubKey
	outputBean.HtlcTempPubKey = requestData.CurrHtlcTempAddressPubKey

	aliceIsPayer := true
	if user.PeerId == channelInfo.PeerIdB {
		aliceIsPayer = false
	}
	outputBean.AmountToHtlc = amountAndFee
	if aliceIsPayer { //Alice pay money to bob Alice是付款方
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(amountAndFee)).Round(8).Float64()
		outputBean.AmountToCounterparty = fundingTransaction.AmountB
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
		outputBean.OppositeSideChannelAddress = channelInfo.AddressB
	} else { //	bob pay money to alice
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Sub(decimal.NewFromFloat(amountAndFee)).Round(8).Float64()
		outputBean.AmountToCounterparty = fundingTransaction.AmountA
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
		outputBean.OppositeSideChannelAddress = channelInfo.AddressA
	}
	if latestCommitmentTx.Id > 0 {
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(latestCommitmentTx.AmountToRSMC).Sub(decimal.NewFromFloat(amountAndFee)).Round(8).Float64()
		outputBean.AmountToCounterparty = latestCommitmentTx.AmountToCounterparty
	}

	newCommitmentTxInfo, err := createCommitmentTx(user.PeerId, channelInfo, fundingTransaction, outputBean, &user)
	if err != nil {
		log.Println(err)
		return nil, rawTx, err
	}
	newCommitmentTxInfo.TxType = dao.CommitmentTransactionType_Htlc
	newCommitmentTxInfo.RSMCTempAddressIndex = requestData.CurrRsmcTempAddressIndex
	newCommitmentTxInfo.HTLCTempAddressIndex = requestData.CurrHtlcTempAddressIndex

	listUnspent, err := GetAddressListUnspent(tx, *channelInfo)
	if err != nil {
		return nil, rawTx, err
	}

	allUsedTxidTemp := ""
	// rsmc
	if newCommitmentTxInfo.AmountToRSMC > 0 {
		rsmcTxData, usedTxid, err := rpcClient.OmniCreateRawTransactionUseSingleInput(
			int(newCommitmentTxInfo.TxType),
			listUnspent,
			channelInfo.ChannelAddress,
			newCommitmentTxInfo.RSMCMultiAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, rawTx, err
		}
		allUsedTxidTemp += usedTxid
		newCommitmentTxInfo.RsmcInputTxid = usedTxid
		newCommitmentTxInfo.RSMCTxHex = rsmcTxData["hex"].(string)

		signHexData := bean.NeedClientSignTxData{}
		signHexData.Hex = newCommitmentTxInfo.RSMCTxHex
		signHexData.Inputs = rsmcTxData["inputs"]
		signHexData.IsMultisig = true
		signHexData.PubKeyA = channelInfo.PubKeyA
		signHexData.PubKeyB = channelInfo.PubKeyB
		rawTx.RsmcRawTxData = signHexData
	}

	//htlc
	if newCommitmentTxInfo.AmountToHtlc > 0 {
		htlcTxData, usedTxid, err := rpcClient.OmniCreateRawTransactionUseSingleInput(
			int(newCommitmentTxInfo.TxType),
			listUnspent,
			channelInfo.ChannelAddress,
			newCommitmentTxInfo.HTLCMultiAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToHtlc,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, allUsedTxidTemp)
		if err != nil {
			log.Println(err)
			return nil, rawTx, err
		}
		allUsedTxidTemp += "," + usedTxid
		newCommitmentTxInfo.HtlcRoutingPacket = requestData.RoutingPacket

		newCommitmentTxInfo.HtlcCltvExpiry = requestData.CltvExpiry
		newCommitmentTxInfo.BeginBlockHeight = conn2tracker.GetBlockCount()

		newCommitmentTxInfo.HtlcTxHex = htlcTxData["hex"].(string)
		newCommitmentTxInfo.HtlcH = requestData.H
		if aliceIsPayer {
			newCommitmentTxInfo.HtlcSender = channelInfo.PeerIdA
		} else {
			newCommitmentTxInfo.HtlcSender = channelInfo.PeerIdB
		}

		signHexData := bean.NeedClientSignTxData{}
		signHexData.Hex = newCommitmentTxInfo.HtlcTxHex
		signHexData.Inputs = htlcTxData["inputs"]
		signHexData.IsMultisig = true
		signHexData.PubKeyA = channelInfo.PubKeyA
		signHexData.PubKeyB = channelInfo.PubKeyB
		rawTx.HtlcRawTxData = signHexData

	}

	//create to Bob tx
	if newCommitmentTxInfo.AmountToCounterparty > 0 {
		toBobTxData, err := rpcClient.OmniCreateRawTransactionUseRestInput(
			int(newCommitmentTxInfo.TxType),
			listUnspent,
			channelInfo.ChannelAddress,
			allUsedTxidTemp,
			outputBean.OppositeSideChannelAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			&channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, rawTx, err
		}
		newCommitmentTxInfo.ToCounterpartyTxHex = toBobTxData["hex"].(string)

		signHexData := bean.NeedClientSignTxData{}
		signHexData.Hex = newCommitmentTxInfo.ToCounterpartyTxHex
		signHexData.Inputs = toBobTxData["inputs"]
		signHexData.IsMultisig = true
		signHexData.PubKeyA = channelInfo.PubKeyA
		signHexData.PubKeyB = channelInfo.PubKeyB
		rawTx.ToCounterpartyRawTxData = signHexData

	}

	newCommitmentTxInfo.CurrState = dao.TxInfoState_Init
	newCommitmentTxInfo.LastHash = ""
	newCommitmentTxInfo.CurrHash = ""
	if latestCommitmentTx.Id > 0 {
		newCommitmentTxInfo.LastCommitmentTxId = latestCommitmentTx.Id
		newCommitmentTxInfo.LastHash = latestCommitmentTx.CurrHash
	}
	err = tx.Save(newCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, rawTx, err
	}

	rawTx.CommitmentTxId = newCommitmentTxInfo.Id
	tx.Save(&rawTx)

	bytes, err := json.Marshal(newCommitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	newCommitmentTxInfo.CurrHash = msgHash
	err = tx.Update(newCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, rawTx, err
	}
	return newCommitmentTxInfo, rawTx, nil
}

// 创建收款方C3b
func htlcPayeeCreateCommitmentTx_C3b(tx storm.Node, channelInfo *dao.ChannelInfo, reqData bean.BobSignedC3a, payerData bean.CreateHtlcTxForC3aOfP2p, latestCommitmentTx *dao.CommitmentTransaction, signedToOtherHex string, user bean.User) (*dao.CommitmentTransaction, dao.CommitmentTxRawTx, error) {

	channelIds := strings.Split(payerData.RoutingPacket, ",")
	var totalStep = len(channelIds)
	var currStep = 0
	for index, channelId := range channelIds {
		if channelId == channelInfo.ChannelId {
			currStep = index
			break
		}
	}
	rawTx := dao.CommitmentTxRawTx{}
	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, user.PeerId)
	if fundingTransaction == nil {
		return nil, rawTx, errors.New("not found fundingTransaction")
	}

	// htlc的资产分配方案
	var outputBean = commitmentTxOutputBean{}
	decimal.DivisionPrecision = 8
	amountAndFee, _ := decimal.NewFromFloat(payerData.Amount).Mul(decimal.NewFromFloat((1 + config.GetHtlcFee()*float64(totalStep-(currStep+1))))).Round(8).Float64()
	outputBean.RsmcTempPubKey = reqData.CurrRsmcTempAddressPubKey
	outputBean.HtlcTempPubKey = reqData.CurrHtlcTempAddressPubKey

	bobIsPayee := true
	if user.PeerId == channelInfo.PeerIdA {
		bobIsPayee = false
	}
	outputBean.AmountToHtlc = amountAndFee
	if bobIsPayee { //Alice pay money to bob
		outputBean.AmountToRsmc = fundingTransaction.AmountB
		outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(amountAndFee)).Round(8).Float64()
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
		outputBean.OppositeSideChannelAddress = channelInfo.AddressA
	} else { //	bob pay money to alice
		outputBean.AmountToRsmc = fundingTransaction.AmountA
		outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Sub(decimal.NewFromFloat(amountAndFee)).Round(8).Float64()
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
		outputBean.OppositeSideChannelAddress = channelInfo.AddressB
	}
	if latestCommitmentTx.Id > 0 {
		outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(latestCommitmentTx.AmountToCounterparty).Sub(decimal.NewFromFloat(amountAndFee)).Round(8).Float64()
		outputBean.AmountToRsmc = latestCommitmentTx.AmountToRSMC
	}

	newCommitmentTxInfo, err := createCommitmentTx(user.PeerId, channelInfo, fundingTransaction, outputBean, &user)
	if err != nil {
		log.Println(err)
		return nil, rawTx, err
	}
	newCommitmentTxInfo.FromCounterpartySideForMeTxHex = signedToOtherHex
	newCommitmentTxInfo.TxType = dao.CommitmentTransactionType_Htlc
	newCommitmentTxInfo.RSMCTempAddressIndex = reqData.CurrRsmcTempAddressIndex
	newCommitmentTxInfo.HTLCTempAddressIndex = reqData.CurrHtlcTempAddressIndex

	listUnspent, err := GetAddressListUnspent(tx, *channelInfo)
	if err != nil {
		return nil, rawTx, err
	}

	allUsedTxidTemp := ""
	// rsmc
	if newCommitmentTxInfo.AmountToRSMC > 0 {
		rsmcTxData, usedTxid, err := rpcClient.OmniCreateRawTransactionUseSingleInput(
			int(newCommitmentTxInfo.TxType),
			listUnspent,
			channelInfo.ChannelAddress,
			newCommitmentTxInfo.RSMCMultiAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, rawTx, err
		}
		allUsedTxidTemp += usedTxid
		newCommitmentTxInfo.RsmcInputTxid = usedTxid
		newCommitmentTxInfo.RSMCTxHex = rsmcTxData["hex"].(string)

		signHexData := bean.NeedClientSignTxData{}
		signHexData.Hex = newCommitmentTxInfo.RSMCTxHex
		signHexData.Inputs = rsmcTxData["inputs"]
		signHexData.IsMultisig = true
		signHexData.PubKeyA = channelInfo.PubKeyA
		signHexData.PubKeyB = channelInfo.PubKeyB
		rawTx.RsmcRawTxData = signHexData
	}

	// htlc
	if newCommitmentTxInfo.AmountToHtlc > 0 {
		htlcTxData, usedTxid, err := rpcClient.OmniCreateRawTransactionUseSingleInput(
			int(newCommitmentTxInfo.TxType),
			listUnspent,
			channelInfo.ChannelAddress,
			newCommitmentTxInfo.HTLCMultiAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToHtlc,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, allUsedTxidTemp)
		if err != nil {
			log.Println(err)
			return nil, rawTx, err
		}
		allUsedTxidTemp += "," + usedTxid
		newCommitmentTxInfo.HtlcRoutingPacket = payerData.RoutingPacket
		newCommitmentTxInfo.HtlcCltvExpiry = payerData.CltvExpiry
		newCommitmentTxInfo.BeginBlockHeight = conn2tracker.GetBlockCount()
		newCommitmentTxInfo.HtlcTxHex = htlcTxData["hex"].(string)

		signHexData := bean.NeedClientSignTxData{}
		signHexData.Hex = newCommitmentTxInfo.HtlcTxHex
		signHexData.Inputs = htlcTxData["inputs"]
		signHexData.IsMultisig = true
		signHexData.PubKeyA = channelInfo.PubKeyA
		signHexData.PubKeyB = channelInfo.PubKeyB
		rawTx.HtlcRawTxData = signHexData

		newCommitmentTxInfo.HtlcH = payerData.H
		if bobIsPayee {
			newCommitmentTxInfo.HtlcSender = channelInfo.PeerIdA
		} else {
			newCommitmentTxInfo.HtlcSender = channelInfo.PeerIdB
		}
	}

	//create for other side tx
	if newCommitmentTxInfo.AmountToCounterparty > 0 {
		toBobTxData, err := rpcClient.OmniCreateRawTransactionUseRestInput(
			int(newCommitmentTxInfo.TxType),
			listUnspent,
			channelInfo.ChannelAddress,
			allUsedTxidTemp,
			outputBean.OppositeSideChannelAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			&channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, rawTx, err
		}
		newCommitmentTxInfo.ToCounterpartyTxHex = toBobTxData["hex"].(string)

		signHexData := bean.NeedClientSignTxData{}
		signHexData.Hex = newCommitmentTxInfo.ToCounterpartyTxHex
		signHexData.Inputs = toBobTxData["inputs"]
		signHexData.IsMultisig = true
		signHexData.PubKeyA = channelInfo.PubKeyA
		signHexData.PubKeyB = channelInfo.PubKeyB
		rawTx.ToCounterpartyRawTxData = signHexData
	}

	newCommitmentTxInfo.CurrState = dao.TxInfoState_Init
	newCommitmentTxInfo.LastHash = ""
	newCommitmentTxInfo.CurrHash = ""
	if latestCommitmentTx.Id > 0 {
		newCommitmentTxInfo.LastCommitmentTxId = latestCommitmentTx.Id
		newCommitmentTxInfo.LastHash = latestCommitmentTx.CurrHash
	}
	err = tx.Save(newCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, rawTx, err
	}
	rawTx.CommitmentTxId = newCommitmentTxInfo.Id
	tx.Save(&rawTx)

	bytes, err := json.Marshal(newCommitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	newCommitmentTxInfo.CurrHash = msgHash
	err = tx.Update(newCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, rawTx, err
	}
	return newCommitmentTxInfo, rawTx, nil
}

// 付款方更新C3a的信息
func checkHexAndUpdateC3aOn42Protocal(tx storm.Node, jsonObj bean.NeedAliceSignHtlcTxOfC3bP2p, htlcRequestInfo dao.AddHtlcRequestInfo, channelInfo dao.ChannelInfo, commitmentTransaction *dao.CommitmentTransaction, user bean.User) (retData *string, needNoticePayee bool, err error) {

	payeePubKey := channelInfo.PubKeyB
	payerAddress := channelInfo.AddressA
	otherSideAddress := channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdB {
		payerAddress = channelInfo.AddressB
		otherSideAddress = channelInfo.AddressA
		payeePubKey = channelInfo.PubKeyA
	}

	//region 1、检测 signedToOtherHex
	if len(commitmentTransaction.ToCounterpartyTxHex) > 0 {
		signedToCounterpartyHex := jsonObj.C3aCompleteSignedCounterpartyHex
		if tool.CheckIsString(&signedToCounterpartyHex) == false {
			err = errors.New("signedToOtherHex is empty at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		_, err = omnicore.VerifyOmniTxHex(signedToCounterpartyHex,
			commitmentTransaction.PropertyId,
			commitmentTransaction.AmountToCounterparty,
			otherSideAddress,
			true)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}
		commitmentTransaction.ToCounterpartyTxHex = signedToCounterpartyHex
		commitmentTransaction.ToCounterpartyTxid = rpcClient.GetTxId(signedToCounterpartyHex)
	}
	//endregion

	//region 2、检测 signedRsmcHex
	signedRsmcHex := jsonObj.C3aCompleteSignedRsmcHex
	if tool.CheckIsString(&signedRsmcHex) {
		_, err = omnicore.VerifyOmniTxHex(signedRsmcHex,
			commitmentTransaction.PropertyId,
			commitmentTransaction.AmountToRSMC,
			commitmentTransaction.RSMCMultiAddress,
			true)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}

		commitmentTransaction.RSMCTxHex = signedRsmcHex
		commitmentTransaction.RSMCTxid = rpcClient.GetTxId(signedRsmcHex)
	}
	//endregion

	//region 3、检测 signedHtlcHex
	signedHtlcHex := jsonObj.C3aCompleteSignedHtlcHex
	if tool.CheckIsString(&signedHtlcHex) == false {
		err = errors.New("signedHtlcHex is empty at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	_, err = omnicore.VerifyOmniTxHex(signedHtlcHex,
		commitmentTransaction.PropertyId,
		commitmentTransaction.AmountToHtlc,
		commitmentTransaction.HTLCMultiAddress,
		true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	commitmentTransaction.HtlcTxHex = signedHtlcHex
	commitmentTransaction.HTLCTxid = rpcClient.GetTxId(signedHtlcHex)
	//endregion

	//region 4、rsmc Rd的保存
	payerRsmcRdHex := jsonObj.C3aRsmcRdPartialSignedData.Hex
	if tool.CheckIsString(&payerRsmcRdHex) {
		_, err = omnicore.VerifyOmniTxHex(payerRsmcRdHex,
			commitmentTransaction.PropertyId,
			commitmentTransaction.AmountToRSMC,
			payerAddress,
			true)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}

		err = saveRdTx(tx, &channelInfo, signedRsmcHex, payerRsmcRdHex, commitmentTransaction, payerAddress, &user)
		if err != nil {
			return nil, false, err
		}
	}
	//endregion

	//region 5、对ht1a进行二次签名，并保存
	payerHt1aHex := jsonObj.C3aHtlcHtPartialSignedData.Hex
	multiAddress, _, _, err := createMultiSig(htlcRequestInfo.CurrHtlcTempAddressForHt1aPubKey, payeePubKey)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	_, err = omnicore.VerifyOmniTxHex(payerHt1aHex,
		commitmentTransaction.PropertyId,
		commitmentTransaction.AmountToHtlc,
		multiAddress,
		true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	htlcTimeOut := commitmentTransaction.HtlcCltvExpiry
	ht1a, err := saveHT1aForAlice(tx, channelInfo, commitmentTransaction, payerHt1aHex, htlcRequestInfo, payeePubKey, htlcTimeOut, user)
	if err != nil {
		err = errors.New("fail to sign  payerHt1aHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}

	//endregion

	//region 6、为bob存储lockByHForBobHex
	lockByHForBobHex := jsonObj.C3aHtlcHlockPartialSignedData.Hex
	if tool.CheckIsString(&lockByHForBobHex) == false {
		err = errors.New("lockByHForBobHex is empty at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	_, err = saveHtlcLockByHTxAtPayerSide(tx, channelInfo, commitmentTransaction, lockByHForBobHex, user)
	if err != nil {
		err = errors.New("fail to lockByHForBobHex at 41 protocol")
		log.Println(err)
		return nil, false, err
	}
	//endregion
	return &ht1a.RSMCTxHex, true, nil
}

// 收款方更新C3b的信息
func checkHexAndUpdateC3bOn42Protocal(tx storm.Node, jsonObj bean.NeedBobSignHtlcSubTxOfC3bP2p, channelInfo dao.ChannelInfo, latestCommitmentTx *dao.CommitmentTransaction, user bean.User) (data *dao.CommitmentTransaction, needNoticePayee bool, err error) {
	bobChannelAddress := channelInfo.AddressB
	aliceChannelAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdA {
		aliceChannelAddress = channelInfo.AddressB
		bobChannelAddress = channelInfo.AddressA
	}
	//region 1、检测 signedToOtherHex
	signedToCounterpartyTxHex := jsonObj.C3bCompleteSignedCounterpartyHex
	if tool.CheckIsString(&signedToCounterpartyTxHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedToCounterpartyTxHex, 2); pass == false {
			err = errors.New("signedToCounterpartyTxHex is empty at 42 protocol")
			log.Println(err)
			return nil, true, err
		}
		_, err = omnicore.VerifyOmniTxHex(signedToCounterpartyTxHex,
			latestCommitmentTx.PropertyId,
			latestCommitmentTx.AmountToCounterparty,
			aliceChannelAddress,
			true)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}
		latestCommitmentTx.ToCounterpartyTxHex = signedToCounterpartyTxHex
		latestCommitmentTx.ToCounterpartyTxid = rpcClient.GetTxId(signedToCounterpartyTxHex)
	}
	//endregion

	//region 2、检测 signedRsmcHex
	signedRsmcHex := ""
	if len(latestCommitmentTx.RSMCTxHex) > 0 {
		signedRsmcHex = jsonObj.C3bCompleteSignedRsmcHex
		if pass, _ := rpcClient.CheckMultiSign(true, signedRsmcHex, 2); pass == false {
			err = errors.New("signedRsmcHex is empty at 42 protocol")
			log.Println(err)
			return nil, true, err
		}
		_, err = omnicore.VerifyOmniTxHex(signedRsmcHex,
			latestCommitmentTx.PropertyId,
			latestCommitmentTx.AmountToRSMC,
			latestCommitmentTx.RSMCMultiAddress,
			true)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}

		latestCommitmentTx.RSMCTxHex = signedRsmcHex
		latestCommitmentTx.RSMCTxid = rpcClient.GetTxId(signedRsmcHex)
	}
	//endregion

	//region 3、检测 signedHtlcHex
	signedHtlcHex := jsonObj.C3bCompleteSignedHtlcHex
	if pass, _ := rpcClient.CheckMultiSign(true, signedHtlcHex, 2); pass == false {
		err = errors.New("signedHtlcHex is empty at 42 protocol")
		log.Println(err)
		return nil, true, err
	}

	_, err = omnicore.VerifyOmniTxHex(signedHtlcHex,
		latestCommitmentTx.PropertyId,
		latestCommitmentTx.AmountToHtlc,
		latestCommitmentTx.HTLCMultiAddress,
		true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	latestCommitmentTx.HtlcTxHex = signedHtlcHex
	latestCommitmentTx.HTLCTxid = rpcClient.GetTxId(signedRsmcHex)
	//endregion

	//region 4、rsmc Rd
	if len(latestCommitmentTx.RSMCTxHex) > 0 {
		payeeRsmcRdHex := jsonObj.C3bRsmcRdPartialData.Hex
		if pass, _ := rpcClient.CheckMultiSign(false, payeeRsmcRdHex, 2); pass == false {
			err = errors.New("signedRsmcHex is empty at 42 protocol")
			log.Println(err)
			return nil, true, err
		}

		_, err = omnicore.VerifyOmniTxHex(payeeRsmcRdHex,
			latestCommitmentTx.PropertyId,
			latestCommitmentTx.AmountToRSMC,
			bobChannelAddress,
			true)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}

		err = saveRdTx(tx, &channelInfo, signedRsmcHex, payeeRsmcRdHex, latestCommitmentTx, bobChannelAddress, &user)
		if err != nil {
			return nil, false, err
		}
	}
	//endregion

	// region  5 保存Hlock h+bobChannelPubkey锁定给bob的付款金额 有了H对应的R，就能解锁
	lockByHForBobHex := jsonObj.C3bHtlcHlockPartialData.Hex
	if tool.CheckIsString(&lockByHForBobHex) == false {
		err = errors.New("payeeHlockHex is empty at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	_, err = saveHtlcLockByHForBobAtPayeeSide(tx, channelInfo, latestCommitmentTx, lockByHForBobHex, user)
	if err != nil {
		err = errors.New("fail to signHtlcLockByHTxAtPayerSide at 41 protocol")
		log.Println(err)
		return nil, false, err
	}
	//endregion

	// region  6、签名HTD1b 超时退给alice的钱
	payeeHTD1bHex := jsonObj.C3bHtlcHtdPartialData.Hex
	if tool.CheckIsString(&payeeHTD1bHex) == false {
		err = errors.New("payeeHTD1bHex is empty at 41 protocol")
		log.Println(err)
		return nil, true, err
	}

	_, err = omnicore.VerifyOmniTxHex(payeeHTD1bHex,
		latestCommitmentTx.PropertyId,
		latestCommitmentTx.AmountToHtlc,
		aliceChannelAddress,
		true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	err = saveHTD1bTx(tx, signedHtlcHex, payeeHTD1bHex, *latestCommitmentTx, aliceChannelAddress, &user)
	if err != nil {
		return nil, false, err
	}
	//endregion
	latestCommitmentTx.CurrState = dao.TxInfoState_Htlc_GetH
	bytes, err := json.Marshal(latestCommitmentTx)
	msgHash := tool.SignMsgWithSha256(bytes)
	latestCommitmentTx.CurrHash = msgHash
	_ = tx.Update(latestCommitmentTx)

	return latestCommitmentTx, false, nil
}
