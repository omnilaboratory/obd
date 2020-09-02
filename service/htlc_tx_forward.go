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
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/tool"
	trackerBean "github.com/omnilaboratory/obd/tracker/bean"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type htlcForwardTxManager struct {
	operationFlag sync.Mutex
	//缓存来自alice的请求开通htlc的交易的数据
	addHtlcTempDataAt40P map[string]string
	htlcInvoiceTempData  map[string]bean.HtlcRequestFindPathInfo
}

// htlc pay money  付款
var HtlcForwardTxService htlcForwardTxManager

func (service *htlcForwardTxManager) CreateHtlcInvoice(msg bean.RequestMessage, user bean.User) (data interface{}, err error) {

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
	_, err = rpcClient.OmniGetProperty(requestData.PropertyId)
	if err != nil {
		return nil, err
	} else {
		propertyId := ""
		tool.ConvertNumToString(int(requestData.PropertyId), &propertyId)
		code, err := tool.GetMsgLengthFromInt(len(propertyId))
		if err != nil {
			return nil, err
		}
		addr += "p" + code + propertyId
	}

	code, err := tool.GetMsgLengthFromInt(len(msg.SenderNodePeerId))
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

	if requestFindPathInfo.PropertyId < 0 {
		return nil, requestFindPathInfo.IsPrivate, errors.New(enum.Tips_common_wrong + "property_id")
	}

	_, err = rpcClient.OmniGetProperty(requestFindPathInfo.PropertyId)
	if err != nil {
		return nil, requestFindPathInfo.IsPrivate, err
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
		pathRequest.PayeePeerId = requestFindPathInfo.RecipientUserPeerId
		sendMsgToTracker(enum.MsgType_Tracker_GetHtlcPath_351, pathRequest)
		service.htlcInvoiceTempData[user.PeerId+"_"+pathRequest.H] = requestFindPathInfo
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
	requestFindPathInfo := service.htlcInvoiceTempData[user.PeerId+"_"+h]
	if &requestFindPathInfo == nil {
		return nil, errors.New("has no channel path")
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

	delete(service.htlcInvoiceTempData, user.PeerId+"_"+h)
	return retData, nil
}

// 40协议的alice方的逻辑 alice start htlc as payer
func (service *htlcForwardTxManager) UpdateAddHtlc_40(msg bean.RequestMessage, user bean.User) (data *bean.AliceRequestAddHtlc, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New(enum.Tips_common_empty + "msg data")
	}

	requestData := &bean.AddHtlcRequest{}
	err = json.Unmarshal([]byte(msg.Data), requestData)
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

	//region check input data 检测输入输入数据
	if requestData.Amount < config.GetOmniDustBtc() {
		return nil, errors.New(fmt.Sprintf(enum.Tips_common_amountMustGreater, config.GetOmniDustBtc()))
	}
	if tool.CheckIsString(&requestData.H) == false {
		return nil, errors.New(enum.Tips_common_empty + "h")
	}
	if tool.CheckIsString(&requestData.RoutingPacket) == false {
		return nil, errors.New(enum.Tips_common_empty + "routing_packet")
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
		return nil, errors.New(enum.Tips_htlc_noChanneFromRountingPacket)
	}

	if channelInfo.CurrState == dao.ChannelState_NewTx {
		return nil, errors.New(enum.Tips_common_newTxMsg)
	}

	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, user.PeerId)
	duration := time.Now().Sub(fundingTransaction.CreateAt)
	if duration > time.Minute*30 {
		if checkChannelOmniAssetAmount(*channelInfo) == false {
			err = errors.New(enum.Tips_rsmc_broadcastedChannel)
			log.Println(err)
			return nil, err
		}
	}

	if requestData.CltvExpiry < (totalStep - currStep) {
		requestData.CltvExpiry = totalStep - currStep
	}

	err = checkBtcFundFinish(channelInfo.ChannelAddress, false)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&requestData.ChannelAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "channel_address_private_key")
		log.Println(err)
		return nil, err
	}

	myChannelPubKey := channelInfo.PubKeyA
	if user.PeerId == channelInfo.PeerIdB {
		myChannelPubKey = channelInfo.PubKeyB
	}
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.ChannelAddressPrivateKey, myChannelPubKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongChannelPrivateKey, requestData.ChannelAddressPrivateKey, myChannelPubKey))
	}

	if tool.CheckIsString(&requestData.LastTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "last_temp_address_private_key")
		log.Println(err)
		return nil, err
	}

	latestCommitmentTx, _ := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	if latestCommitmentTx.Id > 0 {
		if latestCommitmentTx.CurrState == dao.TxInfoState_CreateAndSign {
			_, err = tool.GetPubKeyFromWifAndCheck(requestData.LastTempAddressPrivateKey, latestCommitmentTx.RSMCTempAddressPubKey)
			if err != nil {
				return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongPrivateKeyForLast, requestData.LastTempAddressPrivateKey, latestCommitmentTx.RSMCTempAddressPubKey))
			}
		}
		if latestCommitmentTx.CurrState == dao.TxInfoState_Create {
			if latestCommitmentTx.TxType != dao.CommitmentTransactionType_Htlc {
				return nil, errors.New("error commitment tx type")
			}

			if requestData.CurrRsmcTempAddressPubKey != latestCommitmentTx.RSMCTempAddressPubKey {
				return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, requestData.CurrRsmcTempAddressPubKey, latestCommitmentTx.RSMCTempAddressPubKey))
			}

			if requestData.CurrHtlcTempAddressPubKey != latestCommitmentTx.HTLCTempAddressPubKey {
				return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, requestData.CurrHtlcTempAddressPubKey, latestCommitmentTx.HTLCTempAddressPubKey))
			}

			if latestCommitmentTx.LastCommitmentTxId > 0 {
				lastCommitmentTx := &dao.CommitmentTransaction{}
				_ = tx.One("Id", latestCommitmentTx.LastCommitmentTxId, lastCommitmentTx)
				_, err = tool.GetPubKeyFromWifAndCheck(requestData.LastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey)
				if err != nil {
					return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongPrivateKeyForLast, requestData.LastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey))
				}
			}
		}
	}
	if tool.CheckIsString(&requestData.CurrRsmcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_rsmc_temp_address_pub_key")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&requestData.CurrRsmcTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_rsmc_temp_address_private_key")
		log.Println(err)
		return nil, err
	}
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrRsmcTempAddressPrivateKey, requestData.CurrRsmcTempAddressPubKey)
	if err != nil {
		return nil, errors.New(enum.Tips_common_wrong + "curr_rsmc_temp_address_private_key")
	}

	if tool.CheckIsString(&requestData.CurrHtlcTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_private_key")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&requestData.CurrHtlcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_pub_key")
		log.Println(err)
		return nil, err
	}
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrHtlcTempAddressPrivateKey, requestData.CurrHtlcTempAddressPubKey)
	if err != nil {
		return nil, errors.New(enum.Tips_common_wrong + "curr_htlc_temp_address_private_key")
	}

	if tool.CheckIsString(&requestData.CurrHtlcTempAddressForHt1aPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_for_ht1a_pub_key")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&requestData.CurrHtlcTempAddressForHt1aPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_for_ht1a_private_key")
		log.Println(err)
		return nil, err
	}
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrHtlcTempAddressForHt1aPrivateKey, requestData.CurrHtlcTempAddressForHt1aPubKey)
	if err != nil {
		return nil, errors.New(enum.Tips_common_wrong + "curr_htlc_temp_address_for_ht1a_private_key")
	}

	tempAddrPrivateKeyMap[myChannelPubKey] = requestData.ChannelAddressPrivateKey
	tempAddrPrivateKeyMap[requestData.CurrRsmcTempAddressPubKey] = requestData.CurrRsmcTempAddressPrivateKey
	tempAddrPrivateKeyMap[requestData.CurrHtlcTempAddressPubKey] = requestData.CurrHtlcTempAddressPrivateKey
	tempAddrPrivateKeyMap[requestData.CurrHtlcTempAddressForHt1aPubKey] = requestData.CurrHtlcTempAddressForHt1aPrivateKey
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
		latestCommitmentTx, err = htlcPayerCreateCommitmentTx_C3a(tx, channelInfo, *requestData, totalStep, currStep, latestCommitmentTx, user)
		if err != nil {
			log.Println(err)
			return nil, err
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
			return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, requestData.CurrHtlcTempAddressForHt1aPubKey, htlcRequestInfo.CurrHtlcTempAddressForHt1aPubKey))
		}
	}
	_ = tx.Commit()

	retData := &bean.AliceRequestAddHtlc{}
	retData.RoutingPacket = requestData.RoutingPacket
	retData.ChannelId = channelInfo.ChannelId
	retData.H = requestData.H
	retData.Amount = requestData.Amount
	retData.Memo = requestData.Memo
	retData.CltvExpiry = requestData.CltvExpiry
	retData.LastTempAddressPrivateKey = requestData.LastTempAddressPrivateKey
	retData.CurrRsmcTempAddressPubKey = requestData.CurrRsmcTempAddressPubKey
	retData.CurrHtlcTempAddressPubKey = requestData.CurrHtlcTempAddressPubKey
	retData.CurrHtlcTempAddressForHt1aPubKey = requestData.CurrHtlcTempAddressForHt1aPubKey
	retData.PayerCommitmentTxHash = latestCommitmentTx.CurrHash
	retData.RsmcTxHex = latestCommitmentTx.RSMCTxHex
	retData.HtlcTxHex = latestCommitmentTx.HtlcTxHex
	retData.ToCounterpartyTxHex = latestCommitmentTx.ToCounterpartyTxHex
	retData.PayerNodeAddress = msg.SenderNodePeerId
	retData.PayerPeerId = msg.SenderUserPeerId
	return retData, nil
}

// 40号协议 收款方的obd节点对来自付款方obd节点的消息的处理并通过ws转发给收款方
func (service *htlcForwardTxManager) BeforeBobSignPayerAddHtlcRequestAtBobSide_40(msgData string, user bean.User) (data *bean.AliceRequestAddHtlc, err error) {
	requestAddHtlc := &bean.AliceRequestAddHtlc{}
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
	if channelInfo == nil {
		return nil, errors.New("not found channel info")
	}

	channelInfo.CurrState = dao.ChannelState_NewTx
	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	if service.addHtlcTempDataAt40P == nil {
		service.addHtlcTempDataAt40P = make(map[string]string)
	}
	service.addHtlcTempDataAt40P[requestAddHtlc.PayerCommitmentTxHash] = msgData

	return requestAddHtlc, nil
}

// -100041号协议，bob方签收
func (service *htlcForwardTxManager) PayeeSignGetAddHtlc_41(jsonData string, user bean.User) (returnData *bean.AfterBobSignAddHtlcToAlice, err error) {
	if tool.CheckIsString(&jsonData) == false {
		err := errors.New(enum.Tips_common_empty + "msg data")
		log.Println(err)
		return nil, err
	}

	requestData := &bean.HtlcSignGetH{}
	err = json.Unmarshal([]byte(jsonData), requestData)
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

	aliceMsg := service.addHtlcTempDataAt40P[requestData.PayerCommitmentTxHash]
	if tool.CheckIsString(&aliceMsg) == false {
		return nil, errors.New(enum.Tips_common_empty + "payer_commitment_tx_hash")
	}

	payerRequestAddHtlc := &bean.AliceRequestAddHtlc{}
	_ = json.Unmarshal([]byte(aliceMsg), payerRequestAddHtlc)

	channelId := payerRequestAddHtlc.ChannelId

	returnData = &bean.AfterBobSignAddHtlcToAlice{}
	returnData.ChannelId = channelId
	returnData.PayerCommitmentTxHash = payerRequestAddHtlc.PayerCommitmentTxHash

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

	err = checkBtcFundFinish(channelInfo.ChannelAddress, false)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&requestData.ChannelAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "channel_address_private_key")
		log.Println(err)
		return nil, err
	}

	bobChannelPubKey := channelInfo.PubKeyB
	if user.PeerId == channelInfo.PeerIdA {
		bobChannelPubKey = channelInfo.PubKeyA
	}

	_, err = tool.GetPubKeyFromWifAndCheck(requestData.ChannelAddressPrivateKey, bobChannelPubKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongChannelPrivateKey, requestData.ChannelAddressPrivateKey, bobChannelPubKey))
	}

	latestCommitmentTxInfo, _ := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)

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
		returnData.PayeeLastTempAddressPrivateKey = requestData.LastTempAddressPrivateKey
	}
	if tool.CheckIsString(&requestData.CurrRsmcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_rsmc_temp_address_pub_key")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&requestData.CurrRsmcTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_rsmc_temp_address_private_key")
		log.Println(err)
		return nil, err
	}
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrRsmcTempAddressPrivateKey, requestData.CurrRsmcTempAddressPubKey)
	if err != nil {
		return nil, err
	}
	returnData.PayeeCurrRsmcTempAddressPubKey = requestData.CurrRsmcTempAddressPubKey

	if tool.CheckIsString(&requestData.CurrHtlcTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_private_key")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&requestData.CurrHtlcTempAddressPubKey) == false {
		err = errors.New(enum.Tips_common_empty + "curr_htlc_temp_address_pub_key")
		log.Println(err)
		return nil, err
	}
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrHtlcTempAddressPrivateKey, requestData.CurrHtlcTempAddressPubKey)
	if err != nil {
		return nil, err
	}
	returnData.PayeeCurrHtlcTempAddressPubKey = requestData.CurrHtlcTempAddressPubKey

	tempAddrPrivateKeyMap[bobChannelPubKey] = requestData.ChannelAddressPrivateKey
	tempAddrPrivateKeyMap[requestData.CurrRsmcTempAddressPubKey] = requestData.CurrRsmcTempAddressPrivateKey
	tempAddrPrivateKeyMap[requestData.CurrHtlcTempAddressPubKey] = requestData.CurrHtlcTempAddressPrivateKey
	//endregion

	//region 1、签名对方传过来的rsmcHex
	var aliceRsmcTxId, signedAliceRsmcHex string
	var aliceRsmcMultiAddress, aliceRsmcRedeemScript, aliceRsmcMultiAddressScriptPubKey string
	var aliceRsmcOutputs []rpc.TransactionInputItem
	if tool.CheckIsString(&payerRequestAddHtlc.RsmcTxHex) {
		aliceRsmcTxId, signedAliceRsmcHex, err = rpcClient.BtcSignRawTransaction(payerRequestAddHtlc.RsmcTxHex, requestData.ChannelAddressPrivateKey)
		if err != nil {
			return nil, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "payer rsmc hex"))
		}
		testResult, err := rpcClient.TestMemPoolAccept(signedAliceRsmcHex)
		if err != nil {
			return nil, err
		}
		if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
			return nil, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
		}
		returnData.PayerSignedRsmcHex = signedAliceRsmcHex

		// region 根据alice的临时地址+bob的通道address,获取alice2+bob的多签地址，并得到AliceSignedRsmcHex签名后的交易的input，为创建alice的RD和bob的BR做准备
		aliceRsmcMultiAddress, aliceRsmcRedeemScript, aliceRsmcMultiAddressScriptPubKey, err = createMultiSig(payerRequestAddHtlc.CurrRsmcTempAddressPubKey, bobChannelPubKey)
		if err != nil {
			return nil, err
		}

		aliceRsmcOutputs, err = getInputsForNextTxByParseTxHashVout(signedAliceRsmcHex, aliceRsmcMultiAddress, aliceRsmcMultiAddressScriptPubKey, aliceRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if len(aliceRsmcOutputs) == 0 {
			return nil, errors.New(enum.Tips_common_wrong + "payer rsmc hex")
		}
	}
	//endregion

	//endregion

	// region 2、签名对方传过来的 toCounterpartyTxHex
	signedToOtherHex := ""
	if len(payerRequestAddHtlc.ToCounterpartyTxHex) > 0 {
		_, signedToOtherHex, err = rpcClient.BtcSignRawTransaction(payerRequestAddHtlc.ToCounterpartyTxHex, requestData.ChannelAddressPrivateKey)
		if err != nil {
			return nil, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "payer to_counterparty_tx_hex "))
		}
		testResult, err := rpcClient.TestMemPoolAccept(signedToOtherHex)
		if err != nil {
			return nil, err
		}
		if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
			return nil, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
		}
	}
	returnData.PayerSignedToCounterpartyHex = signedToOtherHex
	//endregion

	// region 3、签名对方传过来的 htlcHex
	aliceHtlcTxId, aliceSignedHtlcHex, err := rpcClient.BtcSignRawTransaction(payerRequestAddHtlc.HtlcTxHex, requestData.ChannelAddressPrivateKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "payer htlcTx hex"))
	}
	testResult, err := rpcClient.TestMemPoolAccept(aliceSignedHtlcHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
		return nil, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
	}
	returnData.PayerSignedHtlcHex = aliceSignedHtlcHex
	//endregion

	// region 根据alice的htlc临时地址+bob的通道address,获取alice2+bob的多签地址，并得到AliceSignedHtlcHex签名后的交易的input，为创建bob的HBR做准备
	aliceHtlcMultiAddress, aliceHtlcRedeemScript, aliceHtlcMultiAddressScriptPubKey, err := createMultiSig(payerRequestAddHtlc.CurrHtlcTempAddressPubKey, bobChannelPubKey)
	if err != nil {
		return nil, err
	}

	aliceHtlcInputs, err := getInputsForNextTxByParseTxHashVout(aliceSignedHtlcHex, aliceHtlcMultiAddress, aliceHtlcMultiAddressScriptPubKey, aliceHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

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
	htlcTimeOut := 0
	//如果是本轮的第一次请求交易
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
		newCommitmentTxInfo, err := htlcPayeeCreateCommitmentTx_C3b(tx, channelInfo, *requestData, *payerRequestAddHtlc, latestCommitmentTxInfo, signedToOtherHex, user)
		amountToOther = newCommitmentTxInfo.AmountToCounterparty
		amountToHtlc = newCommitmentTxInfo.AmountToHtlc
		htlcTimeOut = newCommitmentTxInfo.HtlcCltvExpiry

		returnData.PayeeRsmcHex = newCommitmentTxInfo.RSMCTxHex
		returnData.PayeeToCounterpartyTxHex = newCommitmentTxInfo.ToCounterpartyTxHex
		returnData.PayeeHtlcHex = newCommitmentTxInfo.HtlcTxHex
		returnData.PayeeCommitmentTxHash = newCommitmentTxInfo.CurrHash
		//endregion

		var myAddress = channelInfo.AddressB
		if user.PeerId == channelInfo.PeerIdA {
			myAddress = channelInfo.AddressA
		}
		tempOtherSideCommitmentTx := &dao.CommitmentTransaction{}
		// region 6.0、根据alice C3a的Rsmc输出，创建对应的BR,为下一个交易做准备，create BR2b tx  for bob
		if len(aliceRsmcOutputs) > 0 {
			tempOtherSideCommitmentTx.Id = newCommitmentTxInfo.Id
			tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
			tempOtherSideCommitmentTx.RSMCTempAddressPubKey = payerRequestAddHtlc.CurrRsmcTempAddressPubKey
			tempOtherSideCommitmentTx.RSMCMultiAddress = aliceRsmcMultiAddress
			tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = aliceRsmcMultiAddressScriptPubKey
			tempOtherSideCommitmentTx.RSMCRedeemScript = aliceRsmcRedeemScript
			tempOtherSideCommitmentTx.RSMCTxHex = signedAliceRsmcHex
			tempOtherSideCommitmentTx.RSMCTxid = aliceRsmcTxId
			tempOtherSideCommitmentTx.AmountToRSMC = newCommitmentTxInfo.AmountToCounterparty
			err = createCurrCommitmentTxBR(tx, dao.BRType_Rmsc, channelInfo, tempOtherSideCommitmentTx, aliceRsmcOutputs, myAddress, requestData.ChannelAddressPrivateKey, user)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
		//endregion

		// region 6.1、根据alice C3a的Htlc输出，创建对应的BR,为下一个交易做准备，create HBR2b tx  for bob
		tempOtherSideCommitmentTx.Id = newCommitmentTxInfo.Id
		tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
		tempOtherSideCommitmentTx.RSMCTempAddressPubKey = payerRequestAddHtlc.CurrHtlcTempAddressPubKey
		tempOtherSideCommitmentTx.RSMCMultiAddress = aliceHtlcMultiAddress
		tempOtherSideCommitmentTx.RSMCRedeemScript = aliceHtlcRedeemScript
		tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = aliceHtlcMultiAddressScriptPubKey
		tempOtherSideCommitmentTx.RSMCTxHex = aliceSignedHtlcHex
		tempOtherSideCommitmentTx.RSMCTxid = aliceHtlcTxId
		tempOtherSideCommitmentTx.AmountToRSMC = newCommitmentTxInfo.AmountToHtlc
		err = createCurrCommitmentTxBR(tx, dao.BRType_Htlc, channelInfo, tempOtherSideCommitmentTx, aliceHtlcInputs, myAddress, requestData.ChannelAddressPrivateKey, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		//endregion
	} else {
		returnData.PayeeRsmcHex = latestCommitmentTxInfo.RSMCTxHex
		returnData.PayeeToCounterpartyTxHex = latestCommitmentTxInfo.ToCounterpartyTxHex
		returnData.PayeeHtlcHex = latestCommitmentTxInfo.HtlcTxHex
		returnData.PayeeCommitmentTxHash = latestCommitmentTxInfo.CurrHash
		amountToOther = latestCommitmentTxInfo.AmountToCounterparty
		amountToHtlc = latestCommitmentTxInfo.AmountToHtlc
		htlcTimeOut = latestCommitmentTxInfo.HtlcCltvExpiry
	}

	//region 9、根据签名后的AliceRsmc创建alice的RD create RD tx for alice
	var payerRsmcRdHex string
	if len(aliceRsmcOutputs) > 0 {
		aliceRdOutputAddress := channelInfo.AddressA
		if user.PeerId == channelInfo.PeerIdA {
			aliceRdOutputAddress = channelInfo.AddressB
		}
		_, payerRsmcRdHex, err = rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
			aliceRsmcMultiAddress,
			[]string{
				requestData.ChannelAddressPrivateKey,
			},
			aliceRsmcOutputs,
			aliceRdOutputAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			amountToOther,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&aliceRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, errors.New(fmt.Sprintf(enum.Tips_rsmc_failToCreate, "RD raw transacation"))
		}
	}
	returnData.PayerRsmcRdHex = payerRsmcRdHex
	//endregion create RD tx for alice

	// region  10、h+bobChannelPubkey锁定给bob的付款金额
	lockByHForBobHex, err := createHtlcLockByHForBobAtPayeeSide(*channelInfo, *payerRequestAddHtlc, aliceSignedHtlcHex, bobChannelPubKey, requestData.ChannelAddressPrivateKey, channelInfo.PropertyId, amountToHtlc)
	if err != nil {
		return nil, err
	}
	returnData.PayerLockByHForBobHex = *lockByHForBobHex
	//endregion

	// region 11、ht1a 根据signedHtlcHex（alice签名后C3a的第三个输出）作为输入生成
	payerHt1aHex, err := createHT1aForAlice(*channelInfo, *payerRequestAddHtlc, aliceSignedHtlcHex, bobChannelPubKey, requestData.ChannelAddressPrivateKey, channelInfo.PropertyId, amountToHtlc, htlcTimeOut)
	if err != nil {
		return nil, err
	}
	returnData.PayerHt1aHex = *payerHt1aHex
	//endregion

	channelInfo.CurrState = dao.ChannelState_HtlcTx
	_ = tx.Update(channelInfo)

	_ = tx.Commit()
	return returnData, nil
}

// 42号协议，alice方签收
func (service *htlcForwardTxManager) AfterBobSignAddHtlcAtAliceSide_42(msgData string, user bean.User) (data interface{}, needNoticeBob bool, err error) {

	jsonObjFromPayee := &bean.AfterBobSignAddHtlcToAlice{}
	_ = json.Unmarshal([]byte(msgData), jsonObjFromPayee)

	channelId := jsonObjFromPayee.ChannelId
	commitmentTxHash := jsonObjFromPayee.PayerCommitmentTxHash

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
	if channelInfo == nil {
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

	//returnData["commitmentTxHash"] = commitmentTxHash

	//为了准备给43传数据
	returnData := &bean.AfterAliceSignAddHtlcToBob{}
	returnData.PayeeCommitmentTxHash = jsonObjFromPayee.PayeeCommitmentTxHash

	htlcRequestInfo.CurrState = dao.NS_Finish
	_ = tx.Update(htlcRequestInfo)

	payerChannelPubKey := channelInfo.PubKeyA
	payerChannelAddress := channelInfo.AddressA
	payeeRdOutputAddress := channelInfo.AddressB
	payeeChannelPubKey := channelInfo.PubKeyB
	if user.PeerId == channelInfo.PeerIdB {
		payerChannelPubKey = channelInfo.PubKeyB
		payerChannelAddress = channelInfo.AddressB
		payeeRdOutputAddress = channelInfo.AddressA
		payeeChannelPubKey = channelInfo.PubKeyA
	}

	//region 签名收款方的C3b

	//region 1、签名对方传过来的rsmcHex
	bobRsmcHexIsExist := false
	bobSignedRsmcHex := ""
	bobSignedRsmcTxid := ""
	if len(jsonObjFromPayee.PayeeRsmcHex) > 0 {
		bobRsmcHexIsExist = true
		bobSignedRsmcTxid, bobSignedRsmcHex, err = rpcClient.BtcSignRawTransaction(jsonObjFromPayee.PayeeRsmcHex, tempAddrPrivateKeyMap[payerChannelPubKey])
		if err != nil {
			return nil, true, errors.New("fail to sign rsmc hex ")
		}
		testResult, err := rpcClient.TestMemPoolAccept(bobSignedRsmcHex)
		if err != nil {
			return nil, true, err
		}
		if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
			return nil, true, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
		}
	}
	returnData.PayeeSignedRsmcHex = bobSignedRsmcHex
	commitmentTransaction.FromCounterpartySideForMeTxHex = bobSignedRsmcHex
	//endregion

	// region 2、签名对方传过来的 toCounterpartyTxHex
	var signedToCounterpartyTxHex string
	if tool.CheckIsString(&jsonObjFromPayee.PayeeToCounterpartyTxHex) {
		_, signedToCounterpartyTxHex, err = rpcClient.BtcSignRawTransaction(jsonObjFromPayee.PayeeToCounterpartyTxHex, tempAddrPrivateKeyMap[payerChannelPubKey])
		if err != nil {
			return nil, true, errors.New("fail to sign payee_to_counterparty_tx_hex ")
		}
		testResult, err := rpcClient.TestMemPoolAccept(signedToCounterpartyTxHex)
		if err != nil {
			return nil, true, err
		}
		if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
			return nil, true, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
		}
	}
	returnData.PayeeSignedToCounterpartyHex = signedToCounterpartyTxHex
	//endregion

	// region 3、签名对方传过来的 htlcHex
	bobSignedHtlcTxid, bobSignedHtlcHex, err := rpcClient.BtcSignRawTransaction(jsonObjFromPayee.PayeeHtlcHex, tempAddrPrivateKeyMap[payerChannelPubKey])
	if err != nil {
		return nil, true, errors.New("fail to sign htlcHex hex ")
	}
	testResult, err := rpcClient.TestMemPoolAccept(bobSignedHtlcHex)
	if err != nil {
		return nil, true, err
	}
	if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
		return nil, true, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
	}
	returnData.PayeeSignedHtlcHex = bobSignedHtlcHex
	//endregion

	//region 4、根据签名后的BobRsmc创建bob的RD create RD tx for bob
	payeeRsmcRdHex := ""
	bobRsmcOutputs := []rpc.TransactionInputItem{}
	if bobRsmcHexIsExist {
		bobRsmcMultiAddress, bobRsmcRedeemScript, bobRsmcMultiAddressScriptPubKey, err := createMultiSig(jsonObjFromPayee.PayeeCurrRsmcTempAddressPubKey, payerChannelPubKey)
		if err != nil {
			return nil, true, err
		}
		bobRsmcOutputs, err = getInputsForNextTxByParseTxHashVout(bobSignedRsmcHex, bobRsmcMultiAddress, bobRsmcMultiAddressScriptPubKey, bobRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, true, err
		}

		_, payeeRsmcRdHex, err = rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
			bobRsmcMultiAddress,
			[]string{
				tempAddrPrivateKeyMap[payerChannelPubKey],
			},
			bobRsmcOutputs,
			payeeRdOutputAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			commitmentTransaction.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&bobRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, true, errors.New("fail to create rd for c3b rsmc")
		}
	}
	returnData.PayeeRsmcRdHex = payeeRsmcRdHex
	//endregion create RD tx for alice

	//region 5、payeeHTD1bHex 根据签名后的bobSignedHtlcHex 创建收款方的超时获取R后的退还alice的htlc份额的交易
	htlcTimeOut := commitmentTransaction.HtlcCltvExpiry
	bobHtlcMultiAddress, bobHtlcRedeemScript, bobHtlcMultiAddressScriptPubKey, err := createMultiSig(jsonObjFromPayee.PayeeCurrHtlcTempAddressPubKey, payerChannelPubKey)
	if err != nil {
		return nil, true, err
	}
	bobHtlcOutputs, err := getInputsForNextTxByParseTxHashVout(bobSignedHtlcHex, bobHtlcMultiAddress, bobHtlcMultiAddressScriptPubKey, bobHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}
	_, payeeHTD1bHex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		bobHtlcMultiAddress,
		[]string{
			tempAddrPrivateKeyMap[payerChannelPubKey],
		},
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
		return nil, true, errors.New("fail to create HTD1b for C3b")
	}
	returnData.PayeeHtd1bHex = payeeHTD1bHex

	//为payee 创建hlockHex
	payeeHlockHex, err := createHtlcLockByHForBobAtPayerSide(*channelInfo, *jsonObjFromPayee, bobSignedHtlcHex, commitmentTransaction.HtlcH, payeeChannelPubKey, payerChannelPubKey, tempAddrPrivateKeyMap[payerChannelPubKey], channelInfo.PropertyId, commitmentTransaction.AmountToHtlc)
	if err != nil {
		log.Println(err)
		return nil, true, errors.New("fail to create HlockHex for C3b")
	}
	returnData.PayeeHlockHex = *payeeHlockHex
	//endregion

	//endregion

	//region 处理付款方收到的已经签名C3a的子交易，及上一个BR的签名，RSMCBR，HBR的创建
	if commitmentTransaction.CurrState == dao.TxInfoState_Create {
		//region 4、根据对方传过来的上一个交易的临时rsmc私钥，签名最近的BR交易，保证对方确实放弃了上一个承诺交易
		err := signLastBR(tx, dao.BRType_Rmsc, *channelInfo, user.PeerId, jsonObjFromPayee.PayeeLastTempAddressPrivateKey, commitmentTransaction.LastCommitmentTxId)
		if err != nil {
			log.Println(err)
			return nil, true, err
		}
		//endregion

		// region 5.0、根据bob C3b的Rsmc输出，创建对应的BR,为下一个交易做准备，create BR2a tx  for alice
		tempOtherSideCommitmentTx := &dao.CommitmentTransaction{}
		if bobRsmcHexIsExist {
			bobRsmcMultiAddress, bobRsmcRedeemScript, bobRsmcMultiAddressScriptPubKey, err := createMultiSig(jsonObjFromPayee.PayeeCurrRsmcTempAddressPubKey, payerChannelPubKey)
			if err != nil {
				return nil, true, err
			}
			tempOtherSideCommitmentTx.Id = commitmentTransaction.Id
			tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
			tempOtherSideCommitmentTx.RSMCTempAddressPubKey = jsonObjFromPayee.PayeeCurrRsmcTempAddressPubKey
			tempOtherSideCommitmentTx.RSMCMultiAddress = bobRsmcMultiAddress
			tempOtherSideCommitmentTx.RSMCRedeemScript = bobRsmcRedeemScript
			tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = bobRsmcMultiAddressScriptPubKey
			tempOtherSideCommitmentTx.RSMCTxHex = bobSignedRsmcHex
			tempOtherSideCommitmentTx.RSMCTxid = bobSignedRsmcTxid
			tempOtherSideCommitmentTx.AmountToRSMC = commitmentTransaction.AmountToCounterparty
			err = createCurrCommitmentTxBR(tx, dao.BRType_Rmsc, channelInfo, tempOtherSideCommitmentTx, bobRsmcOutputs, payerChannelAddress, tempAddrPrivateKeyMap[payerChannelPubKey], user)
			if err != nil {
				log.Println(err)
				return nil, true, err
			}
		}
		//endregion

		// region 5.1、根据alice C3a的Htlc输出，创建对应的BR,为下一个交易做准备，create HBR2b tx  for bob
		tempOtherSideCommitmentTx.Id = commitmentTransaction.Id
		tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
		tempOtherSideCommitmentTx.RSMCTempAddressPubKey = jsonObjFromPayee.PayeeCurrHtlcTempAddressPubKey
		tempOtherSideCommitmentTx.RSMCMultiAddress = bobHtlcMultiAddress
		tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = bobHtlcMultiAddressScriptPubKey
		tempOtherSideCommitmentTx.RSMCRedeemScript = bobHtlcRedeemScript
		tempOtherSideCommitmentTx.RSMCTxHex = bobSignedHtlcHex
		tempOtherSideCommitmentTx.RSMCTxid = bobSignedHtlcTxid
		tempOtherSideCommitmentTx.AmountToRSMC = commitmentTransaction.AmountToHtlc
		err = createCurrCommitmentTxBR(tx, dao.BRType_Htlc, channelInfo, tempOtherSideCommitmentTx, bobHtlcOutputs, payerChannelAddress, tempAddrPrivateKeyMap[payerChannelPubKey], user)
		if err != nil {
			log.Println(err)
			return nil, true, err
		}
		//endregion

		//region 6 更新收到的签名交易
		ht1aSignedHex, needNoticePayee, err := checkHexAndUpdateC3aOn42Protocal(tx, *jsonObjFromPayee, *htlcRequestInfo, *channelInfo, commitmentTransaction, user)
		if err != nil {
			return nil, needNoticePayee, err
		}

		//为43准备，创建ht1a的Rd HTRD
		returnData.PayerHt1aSignedHex = *ht1aSignedHex
		//endregion
	} else {
		ht1a := &dao.HTLCTimeoutTxForAAndExecutionForB{}
		err = tx.Select(
			q.Eq("ChannelId", channelInfo.ChannelId),
			q.Eq("CommitmentTxId", commitmentTransaction.Id),
			q.Eq("Owner", user.PeerId)).
			First(ht1a)
		if err != nil {
			return nil, false, err
		}
		// TAG 为43准备，创建ht1a的Rd HTRD
		returnData.PayerHt1aSignedHex = ht1a.RSMCTxHex
	}
	returnData.PayerCurrHtlcTempAddressForHt1aPubKey = htlcRequestInfo.CurrHtlcTempAddressForHt1aPubKey
	returnData.PayerCommitmentTxHash = commitmentTransaction.CurrHash

	//endregion

	commitmentTransaction.CurrState = dao.TxInfoState_Htlc_WaitHTRD1aSign
	commitmentTransaction.SignAt = time.Now()
	_ = tx.Update(commitmentTransaction)

	_ = tx.Commit()
	return returnData, true, nil
}

// 43号协议 包含创建HT1a的HRD1a，以及bob自己的HBR1b
func (service *htlcForwardTxManager) AfterAliceSignAddHtlcAtBobSide_43(msgData string, user bean.User) (data map[string]interface{}, needNoticeOtherSide bool, err error) {
	jsonObj := &bean.AfterAliceSignAddHtlcToBob{}
	_ = json.Unmarshal([]byte(msgData), jsonObj)

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	aliceRetData := bean.PayeeCreateHt1aRDForPayer{}
	aliceRetData.PayerCommitmentTxHash = jsonObj.PayerCommitmentTxHash

	//bobRetData := make(map[string]interface{})
	bobCommitmentHash := jsonObj.PayeeCommitmentTxHash
	latestCommitmentTx := &dao.CommitmentTransaction{}
	err = tx.Select(q.Eq("CurrHash", bobCommitmentHash)).First(latestCommitmentTx)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}
	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(q.Eq("ChannelId", latestCommitmentTx.ChannelId)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}

	bobChannelPubKey := channelInfo.PubKeyB
	bobChannelAddress := channelInfo.AddressB
	aliceChannelAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdA {
		bobChannelPubKey = channelInfo.PubKeyA
		aliceChannelAddress = channelInfo.AddressB
		bobChannelAddress = channelInfo.AddressA
	}

	//region 根据已完成签名的Ht1a创建aliceRD hex
	ht1aSignedHex := jsonObj.PayerHt1aSignedHex
	aliceHt1aMultiAddress, aliceHt1aRedeemScript, aliceHt1aMultiAddressScriptPubKey, err := createMultiSig(jsonObj.PayerCurrHtlcTempAddressForHt1aPubKey, bobChannelPubKey)
	if err != nil {
		return nil, true, err
	}
	alicHtlaOutputs, err := getInputsForNextTxByParseTxHashVout(ht1aSignedHex, aliceHt1aMultiAddress, aliceHt1aMultiAddressScriptPubKey, aliceHt1aRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}
	bobChannelPrivateKey := tempAddrPrivateKeyMap[bobChannelPubKey]

	_, aliceHt1aRDhex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		aliceHt1aMultiAddress,
		[]string{
			bobChannelPrivateKey,
		},
		alicHtlaOutputs,
		aliceChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		latestCommitmentTx.AmountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		0,
		&aliceHt1aRedeemScript)
	if err != nil {
		err = errors.New("fail to create HTRD hex " + err.Error())
		log.Println(err)
		return nil, true, err
	}
	aliceRetData.PayerHt1aRDHex = aliceHt1aRDhex
	//endregion

	if latestCommitmentTx.CurrState == dao.TxInfoState_Create {
		//创建ht1a的BR
		tempOtherSideCommitmentTx := &dao.CommitmentTransaction{}
		tempOtherSideCommitmentTx.Id = latestCommitmentTx.Id
		tempOtherSideCommitmentTx.PropertyId = channelInfo.PropertyId
		tempOtherSideCommitmentTx.RSMCTempAddressPubKey = jsonObj.PayerCurrHtlcTempAddressForHt1aPubKey
		tempOtherSideCommitmentTx.RSMCMultiAddress = aliceHt1aMultiAddress
		tempOtherSideCommitmentTx.RSMCRedeemScript = aliceHt1aRedeemScript
		tempOtherSideCommitmentTx.RSMCMultiAddressScriptPubKey = aliceHt1aMultiAddressScriptPubKey
		tempOtherSideCommitmentTx.RSMCTxHex = ht1aSignedHex
		tempOtherSideCommitmentTx.RSMCTxid = alicHtlaOutputs[0].Txid
		tempOtherSideCommitmentTx.AmountToRSMC = latestCommitmentTx.AmountToHtlc
		err = createCurrCommitmentTxBR(tx, dao.BRType_Ht1a, channelInfo, tempOtherSideCommitmentTx, alicHtlaOutputs, bobChannelAddress, bobChannelPrivateKey, user)
		if err != nil {
			log.Println(err)
			return nil, true, err
		}

		latestCommitmentTx, needNoticeOtherSide, err = checkHexAndUpdateC3bOn43Protocal(tx, *jsonObj, *channelInfo, latestCommitmentTx, user)
		if err != nil {
			log.Println(err.Error())
			return nil, needNoticeOtherSide, err
		}

	}

	channelInfo.CurrState = dao.ChannelState_HtlcTx
	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	//同步通道信息到tracker
	sendChannelStateToTracker(*channelInfo, *latestCommitmentTx)

	//bobRetData["commitmentTx"] = latestCommitmentTx

	retData := make(map[string]interface{})
	retData["aliceData"] = aliceRetData
	retData["bobData"] = latestCommitmentTx
	return retData, true, nil
}

// 44号协议
func (service *htlcForwardTxManager) AfterBobCreateHTRDAtAliceSide_44(msgData string, user bean.User) (data interface{}, needNoticeOtherSide bool, err error) {
	jsonObj := &bean.PayeeCreateHt1aRDForPayer{}
	_ = json.Unmarshal([]byte(msgData), jsonObj)

	aliceHt1aRDHex := jsonObj.PayerHt1aRDHex
	aliceCommitmentTxHash := jsonObj.PayerCommitmentTxHash
	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	latestCommitmentTransaction := &dao.CommitmentTransaction{}
	err = tx.Select(q.Eq("CurrHash", aliceCommitmentTxHash)).First(latestCommitmentTransaction)
	if err != nil {
		log.Println(err.Error())
		return nil, true, err
	}

	if latestCommitmentTransaction.CurrState == dao.TxInfoState_Htlc_WaitHTRD1aSign {
		_, err := signHtRD1a(tx, aliceHt1aRDHex, *latestCommitmentTransaction, user)
		if err != nil {
			log.Println(err.Error())
			return nil, true, err
		}
	}
	latestCommitmentTransaction.CurrState = dao.TxInfoState_Htlc_GetH
	_ = tx.Update(latestCommitmentTransaction)

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(q.Eq("ChannelId", latestCommitmentTransaction.ChannelId)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}

	channelInfo.CurrState = dao.ChannelState_HtlcTx
	_ = tx.Update(channelInfo)

	_ = tx.Commit()
	return latestCommitmentTransaction, false, nil
}

// 创建付款方的C3a
func htlcPayerCreateCommitmentTx_C3a(tx storm.Node, channelInfo *dao.ChannelInfo, requestData bean.AddHtlcRequest, totalStep int, currStep int, latestCommitmentTx *dao.CommitmentTransaction, user bean.User) (*dao.CommitmentTransaction, error) {

	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, user.PeerId)
	if fundingTransaction == nil {
		return nil, errors.New("not found fundingTransaction")
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
		return nil, err
	}
	newCommitmentTxInfo.TxType = dao.CommitmentTransactionType_Htlc
	newCommitmentTxInfo.RSMCTempAddressIndex = requestData.CurrRsmcTempAddressIndex
	newCommitmentTxInfo.HTLCTempAddressIndex = requestData.CurrHtlcTempAddressIndex

	allUsedTxidTemp := ""
	// rsmc
	if newCommitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUseSingleInput(
			int(newCommitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			[]string{
				requestData.ChannelAddressPrivateKey,
			},
			newCommitmentTxInfo.RSMCMultiAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += usedTxid
		newCommitmentTxInfo.RsmcInputTxid = usedTxid
		newCommitmentTxInfo.RSMCTxid = txid
		newCommitmentTxInfo.RSMCTxHex = hex
	}

	//htlc
	if newCommitmentTxInfo.AmountToHtlc > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUseSingleInput(
			int(newCommitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			[]string{
				requestData.ChannelAddressPrivateKey,
			},
			newCommitmentTxInfo.HTLCMultiAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToHtlc,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, allUsedTxidTemp)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += "," + usedTxid
		newCommitmentTxInfo.HtlcRoutingPacket = requestData.RoutingPacket

		currBlockHeight, err := rpcClient.GetBlockCount()
		if err != nil {
			return nil, errors.New("fail to get blockHeight ,please try again later")
		}
		newCommitmentTxInfo.HtlcCltvExpiry = requestData.CltvExpiry
		newCommitmentTxInfo.BeginBlockHeight = currBlockHeight
		newCommitmentTxInfo.HTLCTxid = txid
		newCommitmentTxInfo.HtlcTxHex = hex
		newCommitmentTxInfo.HtlcH = requestData.H
		if aliceIsPayer {
			newCommitmentTxInfo.HtlcSender = channelInfo.PeerIdA
		} else {
			newCommitmentTxInfo.HtlcSender = channelInfo.PeerIdB
		}
	}

	//create to Bob tx
	if newCommitmentTxInfo.AmountToCounterparty > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseRestInput(
			int(newCommitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			allUsedTxidTemp,
			[]string{
				requestData.ChannelAddressPrivateKey,
			},
			outputBean.OppositeSideChannelAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			0, &channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		newCommitmentTxInfo.ToCounterpartyTxid = txid
		newCommitmentTxInfo.ToCounterpartyTxHex = hex
	}

	newCommitmentTxInfo.CurrState = dao.TxInfoState_Create
	newCommitmentTxInfo.LastHash = ""
	newCommitmentTxInfo.CurrHash = ""
	if latestCommitmentTx.Id > 0 {
		newCommitmentTxInfo.LastCommitmentTxId = latestCommitmentTx.Id
		newCommitmentTxInfo.LastHash = latestCommitmentTx.CurrHash
	}
	err = tx.Save(newCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bytes, err := json.Marshal(newCommitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	newCommitmentTxInfo.CurrHash = msgHash
	err = tx.Update(newCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return newCommitmentTxInfo, nil
}

// 收款方创建C3b
func htlcPayeeCreateCommitmentTx_C3b(tx storm.Node, channelInfo *dao.ChannelInfo,
	reqData bean.HtlcSignGetH, payerRequestAddHtlcData bean.AliceRequestAddHtlc,
	latestCommitmentTx *dao.CommitmentTransaction, signedToOtherHex string, user bean.User) (*dao.CommitmentTransaction, error) {

	channelIds := strings.Split(payerRequestAddHtlcData.RoutingPacket, ",")
	var totalStep = len(channelIds)
	var currStep = 0
	for index, channelId := range channelIds {
		if channelId == channelInfo.ChannelId {
			currStep = index
			break
		}
	}
	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, user.PeerId)
	if fundingTransaction == nil {
		return nil, errors.New("not found fundingTransaction")
	}

	// htlc的资产分配方案
	var outputBean = commitmentTxOutputBean{}
	decimal.DivisionPrecision = 8
	amountAndFee, _ := decimal.NewFromFloat(payerRequestAddHtlcData.Amount).Mul(decimal.NewFromFloat((1 + config.GetHtlcFee()*float64(totalStep-(currStep+1))))).Round(8).Float64()
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
		return nil, err
	}
	newCommitmentTxInfo.FromCounterpartySideForMeTxHex = signedToOtherHex
	newCommitmentTxInfo.TxType = dao.CommitmentTransactionType_Htlc
	newCommitmentTxInfo.RSMCTempAddressIndex = reqData.CurrRsmcTempAddressIndex
	newCommitmentTxInfo.HTLCTempAddressIndex = reqData.CurrHtlcTempAddressIndex

	allUsedTxidTemp := ""
	// rsmc
	if newCommitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUseSingleInput(
			int(newCommitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			[]string{
				reqData.ChannelAddressPrivateKey,
			},
			newCommitmentTxInfo.RSMCMultiAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += usedTxid
		newCommitmentTxInfo.RsmcInputTxid = usedTxid
		newCommitmentTxInfo.RSMCTxid = txid
		newCommitmentTxInfo.RSMCTxHex = hex
	}

	// htlc
	if newCommitmentTxInfo.AmountToHtlc > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUseSingleInput(
			int(newCommitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			[]string{
				reqData.ChannelAddressPrivateKey,
			},
			newCommitmentTxInfo.HTLCMultiAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToHtlc,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, allUsedTxidTemp)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += "," + usedTxid
		newCommitmentTxInfo.HtlcRoutingPacket = payerRequestAddHtlcData.RoutingPacket
		currBlockHeight, err := rpcClient.GetBlockCount()
		if err != nil {
			return nil, errors.New("fail to get blockHeight ,please try again later")
		}
		newCommitmentTxInfo.HtlcCltvExpiry = int(payerRequestAddHtlcData.CltvExpiry)
		newCommitmentTxInfo.BeginBlockHeight = currBlockHeight
		newCommitmentTxInfo.HTLCTxid = txid
		newCommitmentTxInfo.HtlcTxHex = hex
		newCommitmentTxInfo.HtlcH = payerRequestAddHtlcData.H
		if bobIsPayee {
			newCommitmentTxInfo.HtlcSender = channelInfo.PeerIdA
		} else {
			newCommitmentTxInfo.HtlcSender = channelInfo.PeerIdB
		}
	}

	//create for other side tx
	if newCommitmentTxInfo.AmountToCounterparty > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseRestInput(
			int(newCommitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			allUsedTxidTemp,
			[]string{
				reqData.ChannelAddressPrivateKey,
			},
			outputBean.OppositeSideChannelAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			newCommitmentTxInfo.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			0, &channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		newCommitmentTxInfo.ToCounterpartyTxid = txid
		newCommitmentTxInfo.ToCounterpartyTxHex = hex
	}

	newCommitmentTxInfo.CurrState = dao.TxInfoState_Create
	newCommitmentTxInfo.LastHash = ""
	newCommitmentTxInfo.CurrHash = ""
	if latestCommitmentTx.Id > 0 {
		newCommitmentTxInfo.LastCommitmentTxId = latestCommitmentTx.Id
		newCommitmentTxInfo.LastHash = latestCommitmentTx.CurrHash
	}
	err = tx.Save(newCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bytes, err := json.Marshal(newCommitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	newCommitmentTxInfo.CurrHash = msgHash
	err = tx.Update(newCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return newCommitmentTxInfo, nil
}

// 当付款方节点收到收款方签名后的C3a后，对传入的数据进行处理
func checkHexAndUpdateC3aOn42Protocal(tx storm.Node, jsonObj bean.AfterBobSignAddHtlcToAlice, htlcRequestInfo dao.AddHtlcRequestInfo,
	channelInfo dao.ChannelInfo, commitmentTransaction *dao.CommitmentTransaction, user bean.User) (retData *string, needNoticePayee bool, err error) {

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
		signedToCounterpartyHex := jsonObj.PayerSignedToCounterpartyHex
		if tool.CheckIsString(&signedToCounterpartyHex) == false {
			err = errors.New("signedToOtherHex is empty at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		_, err = rpcClient.TestMemPoolAccept(signedToCounterpartyHex)
		if err != nil {
			err = errors.New("wrong signedToOtherHex at 41 protocol")
			log.Println(err)
			return nil, true, err
		}

		result, err := rpcClient.OmniDecodeTransaction(signedToCounterpartyHex)
		if err != nil {
			return nil, true, err
		}
		hexJsonObj := gjson.Parse(result)
		if channelInfo.ChannelAddress != hexJsonObj.Get("sendingaddress").String() {
			err = errors.New("wrong inputAddress at signedToOtherHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
			err = errors.New("wrong propertyId at signedToOtherHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}

		if otherSideAddress != hexJsonObj.Get("referenceaddress").String() {
			err = errors.New("wrong outputAddress at signedToOtherHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if commitmentTransaction.AmountToCounterparty != hexJsonObj.Get("amount").Float() {
			err = errors.New("wrong amount at signedToOtherHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		commitmentTransaction.ToCounterpartyTxHex = signedToCounterpartyHex
		commitmentTransaction.ToCounterpartyTxid = hexJsonObj.Get("txid").String()
	}
	//endregion

	//region 2、检测 signedRsmcHex
	signedRsmcHex := jsonObj.PayerSignedRsmcHex
	if tool.CheckIsString(&signedRsmcHex) {
		_, err = rpcClient.TestMemPoolAccept(signedRsmcHex)
		if err != nil {
			err = errors.New("wrong signedRsmcHex at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		result, err := rpcClient.OmniDecodeTransaction(signedRsmcHex)
		if err != nil {
			return nil, true, err
		}
		hexJsonObj := gjson.Parse(result)

		if channelInfo.ChannelAddress != hexJsonObj.Get("sendingaddress").String() {
			err = errors.New("wrong inputAddress at signedRsmcHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
			err = errors.New("wrong propertyId at signedRsmcHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}

		if commitmentTransaction.RSMCMultiAddress != hexJsonObj.Get("referenceaddress").String() {
			err = errors.New("wrong outputAddress at signedRsmcHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if commitmentTransaction.AmountToRSMC != hexJsonObj.Get("amount").Float() {
			err = errors.New("wrong amount at signedRsmcHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		commitmentTransaction.RSMCTxHex = signedRsmcHex
		commitmentTransaction.RSMCTxid = hexJsonObj.Get("txid").String()
	}
	//endregion

	//region 3、检测 signedHtlcHex
	signedHtlcHex := jsonObj.PayerSignedHtlcHex
	if tool.CheckIsString(&signedHtlcHex) == false {
		err = errors.New("signedHtlcHex is empty at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	_, err = rpcClient.TestMemPoolAccept(signedHtlcHex)
	if err != nil {
		err = errors.New("wrong signedHtlcHex at 41 protocol")
		log.Println(err)
		return nil, true, err
	}

	result, err := rpcClient.OmniDecodeTransaction(signedHtlcHex)
	if err != nil {
		return nil, true, err
	}
	hexJsonObj := gjson.Parse(result)
	if channelInfo.ChannelAddress != hexJsonObj.Get("sendingaddress").String() {
		err = errors.New("wrong inputAddress at signedHtlcHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
		err = errors.New("wrong propertyId at signedHtlcHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}

	if commitmentTransaction.HTLCMultiAddress != hexJsonObj.Get("referenceaddress").String() {
		err = errors.New("wrong outputAddress at signedHtlcHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	if commitmentTransaction.AmountToHtlc != hexJsonObj.Get("amount").Float() {
		err = errors.New("wrong amount at signedHtlcHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	commitmentTransaction.HtlcTxHex = signedHtlcHex
	commitmentTransaction.HTLCTxid = hexJsonObj.Get("txid").String()
	//endregion

	//region 4、rsmc Rd的保存
	payerRsmcRdHex := jsonObj.PayerRsmcRdHex
	if tool.CheckIsString(&payerRsmcRdHex) {
		payerRDInputsFromRsmc, err := getInputsForNextTxByParseTxHashVout(
			signedRsmcHex,
			commitmentTransaction.RSMCMultiAddress,
			commitmentTransaction.RSMCMultiAddressScriptPubKey,
			commitmentTransaction.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, true, err
		}
		result, err = rpcClient.OmniDecodeTransactionWithPrevTxs(payerRsmcRdHex, payerRDInputsFromRsmc)
		if err != nil {
			log.Println(err)
			return nil, true, err
		}
		hexJsonObj = gjson.Parse(result)
		if commitmentTransaction.RSMCMultiAddress != hexJsonObj.Get("sendingaddress").String() {
			err = errors.New("wrong inputAddress at payerRsmcRdHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}

		if payerAddress != hexJsonObj.Get("referenceaddress").String() {
			err = errors.New("wrong outputAddress at payerRsmcRdHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
			err = errors.New("wrong propertyId at payerRsmcRdHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if commitmentTransaction.AmountToRSMC != hexJsonObj.Get("amount").Float() {
			err = errors.New("wrong amount at payerRsmcRdHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}

		err = signRdTx(tx, &channelInfo, signedRsmcHex, payerRsmcRdHex, commitmentTransaction, payerAddress, &user)
		if err != nil {
			return nil, false, err
		}
	}
	//endregion

	//region 5、对ht1a进行二次签名，并保存
	payerHt1aHex := jsonObj.PayerHt1aHex
	payerHt1aInputsFromHtlc, err := getInputsForNextTxByParseTxHashVout(
		signedHtlcHex,
		commitmentTransaction.HTLCMultiAddress,
		commitmentTransaction.HTLCMultiAddressScriptPubKey,
		commitmentTransaction.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}
	multiAddress, _, _, err := createMultiSig(htlcRequestInfo.CurrHtlcTempAddressForHt1aPubKey, payeePubKey)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	result, err = rpcClient.OmniDecodeTransactionWithPrevTxs(payerHt1aHex, payerHt1aInputsFromHtlc)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}
	hexJsonObj = gjson.Parse(result)
	if commitmentTransaction.HTLCMultiAddress != hexJsonObj.Get("sendingaddress").String() {
		err = errors.New("wrong inputAddress at payerHt1aHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}

	if multiAddress != hexJsonObj.Get("referenceaddress").String() {
		err = errors.New("wrong outputAddress at payerHt1aHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
		err = errors.New("wrong propertyId at payerHt1aHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	if commitmentTransaction.AmountToHtlc != hexJsonObj.Get("amount").Float() {
		err = errors.New("wrong amount at payerHt1aHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	htlcTimeOut := commitmentTransaction.HtlcCltvExpiry
	ht1a, err := signHT1aForAlice(tx, channelInfo, commitmentTransaction, payerHt1aHex, htlcRequestInfo, payeePubKey, htlcTimeOut, user)
	if err != nil {
		err = errors.New("fail to sign  payerHt1aHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}

	//endregion

	//region 6、为bob存储lockByHForBobHex
	lockByHForBobHex := jsonObj.PayerLockByHForBobHex
	if tool.CheckIsString(&lockByHForBobHex) == false {
		err = errors.New("lockByHForBobHex is empty at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	_, err = signHtlcLockByHTxAtPayerSide(tx, channelInfo, commitmentTransaction, lockByHForBobHex, user)
	if err != nil {
		err = errors.New("fail to lockByHForBobHex at 41 protocol")
		log.Println(err)
		return nil, false, err
	}
	//endregion
	return &ht1a.RSMCTxHex, true, nil
}

// 43号协议的收款方的逻辑
func checkHexAndUpdateC3bOn43Protocal(tx storm.Node, jsonObj bean.AfterAliceSignAddHtlcToBob, channelInfo dao.ChannelInfo, latestCommitmentTx *dao.CommitmentTransaction, user bean.User) (data *dao.CommitmentTransaction, needNoticePayee bool, err error) {
	bobChannelAddress := channelInfo.AddressB
	aliceChannelAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdA {
		aliceChannelAddress = channelInfo.AddressB
		bobChannelAddress = channelInfo.AddressA
	}
	//region 1、检测 signedToOtherHex
	signedToCounterpartyTxHex := jsonObj.PayeeSignedToCounterpartyHex
	if tool.CheckIsString(&signedToCounterpartyTxHex) {
		_, err = rpcClient.TestMemPoolAccept(signedToCounterpartyTxHex)
		if err != nil {
			err = errors.New("wrong signedToCounterpartyTxHex at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		result, err := rpcClient.OmniDecodeTransaction(signedToCounterpartyTxHex)
		if err != nil {
			return nil, true, err
		}
		hexJsonObj := gjson.Parse(result)
		if channelInfo.ChannelAddress != hexJsonObj.Get("sendingaddress").String() {
			err = errors.New("wrong inputAddress at signedToOtherHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
			err = errors.New("wrong propertyId at signedToOtherHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}

		if aliceChannelAddress != hexJsonObj.Get("referenceaddress").String() {
			err = errors.New("wrong outputAddress at signedToOtherHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if latestCommitmentTx.AmountToCounterparty != hexJsonObj.Get("amount").Float() {
			err = errors.New("wrong amount at signedToOtherHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		latestCommitmentTx.ToCounterpartyTxHex = signedToCounterpartyTxHex
		latestCommitmentTx.ToCounterpartyTxid = hexJsonObj.Get("txid").String()
	}
	//endregion

	//region 2、检测 signedRsmcHex
	signedRsmcHex := ""
	if len(latestCommitmentTx.RSMCTxHex) > 0 {
		signedRsmcHex = jsonObj.PayeeSignedRsmcHex
		if tool.CheckIsString(&signedRsmcHex) == false {
			err = errors.New("signedRsmcHex is empty at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		_, err = rpcClient.TestMemPoolAccept(signedRsmcHex)
		if err != nil {
			err = errors.New("wrong signedRsmcHex at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		result, err := rpcClient.OmniDecodeTransaction(signedRsmcHex)
		if err != nil {
			return nil, true, err
		}
		hexJsonObj := gjson.Parse(result)

		if channelInfo.ChannelAddress != hexJsonObj.Get("sendingaddress").String() {
			err = errors.New("wrong inputAddress at signedRsmcHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
			err = errors.New("wrong propertyId at signedRsmcHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}

		if latestCommitmentTx.RSMCMultiAddress != hexJsonObj.Get("referenceaddress").String() {
			err = errors.New("wrong outputAddress at signedRsmcHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if latestCommitmentTx.AmountToRSMC != hexJsonObj.Get("amount").Float() {
			err = errors.New("wrong amount at signedRsmcHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		latestCommitmentTx.RSMCTxHex = signedRsmcHex
		latestCommitmentTx.RSMCTxid = hexJsonObj.Get("txid").String()
	}
	//endregion

	//region 3、检测 signedHtlcHex
	signedHtlcHex := jsonObj.PayeeSignedHtlcHex
	if tool.CheckIsString(&signedHtlcHex) == false {
		err = errors.New("signedHtlcHex is empty at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	_, err = rpcClient.TestMemPoolAccept(signedHtlcHex)
	if err != nil {
		err = errors.New("wrong signedRsmcHex at 41 protocol")
		log.Println(err)
		return nil, true, err
	}

	result, err := rpcClient.OmniDecodeTransaction(signedHtlcHex)
	if err != nil {
		return nil, true, err
	}
	hexJsonObj := gjson.Parse(result)
	if channelInfo.ChannelAddress != hexJsonObj.Get("sendingaddress").String() {
		err = errors.New("wrong inputAddress at signedHtlcHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
		err = errors.New("wrong propertyId at signedHtlcHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}

	if latestCommitmentTx.HTLCMultiAddress != hexJsonObj.Get("referenceaddress").String() {
		err = errors.New("wrong outputAddress at signedHtlcHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	if latestCommitmentTx.AmountToHtlc != hexJsonObj.Get("amount").Float() {
		err = errors.New("wrong amount at signedHtlcHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	latestCommitmentTx.HtlcTxHex = signedHtlcHex
	latestCommitmentTx.HTLCTxid = hexJsonObj.Get("txid").String()
	//endregion

	//region 4、二次签名rsmcRd并保存
	if len(latestCommitmentTx.RSMCTxHex) > 0 {
		payeeRsmcRdHex := jsonObj.PayeeRsmcRdHex
		if tool.CheckIsString(&payeeRsmcRdHex) == false {
			err = errors.New("signedRsmcHex is empty at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		payerRDInputsFromRsmc, err := getInputsForNextTxByParseTxHashVout(
			signedRsmcHex,
			latestCommitmentTx.RSMCMultiAddress,
			latestCommitmentTx.RSMCMultiAddressScriptPubKey,
			latestCommitmentTx.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, true, err
		}
		result, err = rpcClient.OmniDecodeTransactionWithPrevTxs(payeeRsmcRdHex, payerRDInputsFromRsmc)
		if err != nil {
			log.Println(err)
			return nil, true, err
		}
		hexJsonObj = gjson.Parse(result)
		if latestCommitmentTx.RSMCMultiAddress != hexJsonObj.Get("sendingaddress").String() {
			err = errors.New("wrong inputAddress at payeeRsmcRdHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}

		if bobChannelAddress != hexJsonObj.Get("referenceaddress").String() {
			err = errors.New("wrong outputAddress at payeeRsmcRdHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
			err = errors.New("wrong propertyId at payeeRsmcRdHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}
		if latestCommitmentTx.AmountToRSMC != hexJsonObj.Get("amount").Float() {
			err = errors.New("wrong amount at payeeRsmcRdHex  at 41 protocol")
			log.Println(err)
			return nil, true, err
		}

		err = signRdTx(tx, &channelInfo, signedRsmcHex, payeeRsmcRdHex, latestCommitmentTx, bobChannelAddress, &user)
		if err != nil {
			return nil, false, err
		}
	}
	//endregion

	// region  5、h+bobChannelPubkey锁定给bob的付款金额 有了H对应的R，就能解锁
	lockByHForBobHex := jsonObj.PayeeHlockHex
	if tool.CheckIsString(&lockByHForBobHex) == false {
		err = errors.New("payeeHlockHex is empty at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	_, err = signHtlcLockByHForBobAtPayeeSide(tx, channelInfo, latestCommitmentTx, lockByHForBobHex, user)
	if err != nil {
		err = errors.New("fail to signHtlcLockByHTxAtPayerSide at 41 protocol")
		log.Println(err)
		return nil, false, err
	}
	//endregion

	// region  6、签名HTD1b 超时退给alice的钱
	payeeHTD1bHex := jsonObj.PayeeHtd1bHex
	if tool.CheckIsString(&payeeHTD1bHex) == false {
		err = errors.New("payeeHTD1bHex is empty at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	payeeHTD1bInputsFromHtlc, err := getInputsForNextTxByParseTxHashVout(
		signedHtlcHex,
		latestCommitmentTx.HTLCMultiAddress,
		latestCommitmentTx.HTLCMultiAddressScriptPubKey,
		latestCommitmentTx.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}
	result, err = rpcClient.OmniDecodeTransactionWithPrevTxs(payeeHTD1bHex, payeeHTD1bInputsFromHtlc)
	if err != nil {
		log.Println(err)
		return nil, true, err
	}
	hexJsonObj = gjson.Parse(result)
	if latestCommitmentTx.HTLCMultiAddress != hexJsonObj.Get("sendingaddress").String() {
		err = errors.New("wrong inputAddress at payeeHTD1bHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	if aliceChannelAddress != hexJsonObj.Get("referenceaddress").String() {
		err = errors.New("wrong outputAddress at payeeHTD1bHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	if channelInfo.PropertyId != hexJsonObj.Get("propertyid").Int() {
		err = errors.New("wrong propertyId at payeeHTD1bHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}
	if latestCommitmentTx.AmountToHtlc != hexJsonObj.Get("amount").Float() {
		err = errors.New("wrong amount at payeeHTD1bHex  at 41 protocol")
		log.Println(err)
		return nil, true, err
	}

	err = signHTD1bTx(tx, signedHtlcHex, payeeHTD1bHex, *latestCommitmentTx, aliceChannelAddress, &user)
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
