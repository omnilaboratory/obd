package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"log"
	"obd/bean"
	"obd/bean/chainhash"
	"obd/config"
	"obd/dao"
	"obd/tool"
	"strconv"
	"strings"
	"time"
)

type channelManager struct{}

var ChannelService = channelManager{}

// AliceOpenChannel init ChannelInfo
func (this *channelManager) AliceOpenChannel(msg bean.RequestMessage, user *bean.User) (channelInfo *dao.ChannelInfo, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New("empty inputData")
	}

	reqData := &bean.OpenChannelInfo{}
	err = json.Unmarshal([]byte(msg.Data), &reqData)
	if err != nil {
		return nil, err
	}

	reqData.FundingAddress, err = getAddressFromPubKey(reqData.FundingPubKey)
	if err != nil {
		return nil, err
	}

	reqData.ChainHash = config.Init_node_chain_hash
	reqData.TemporaryChannelId = bean.ChannelIdService.NextTemporaryChanID()

	channelInfo = &dao.ChannelInfo{}
	channelInfo.OpenChannelInfo = *reqData
	channelInfo.PeerIdA = user.PeerId
	channelInfo.PeerIdB = msg.RecipientPeerId
	channelInfo.PubKeyA = reqData.FundingPubKey
	channelInfo.AddressA = reqData.FundingAddress
	channelInfo.CurrState = dao.ChannelState_Create
	channelInfo.CreateAt = time.Now()
	channelInfo.CreateBy = user.PeerId

	err = user.Db.Save(channelInfo)
	return channelInfo, err
}

// obd init ChannelInfo for Bob
func (this *channelManager) BeforeBobOpenChannelAtBobSide(msg string, user *bean.User) (err error) {
	if tool.CheckIsString(&msg) == false {
		return errors.New("empty inputData")
	}

	aliceChannelInfo := &dao.ChannelInfo{}
	err = json.Unmarshal([]byte(msg), &aliceChannelInfo)
	if err != nil {
		return err
	}
	channelInfo := &dao.ChannelInfo{}
	channelInfo.OpenChannelInfo = aliceChannelInfo.OpenChannelInfo
	channelInfo.PeerIdA = aliceChannelInfo.PeerIdA
	channelInfo.PeerIdB = user.PeerId
	channelInfo.PubKeyA = aliceChannelInfo.FundingPubKey
	channelInfo.AddressA = aliceChannelInfo.FundingAddress
	channelInfo.CurrState = dao.ChannelState_Create
	channelInfo.CreateAt = time.Now()
	channelInfo.CreateBy = user.PeerId
	err = user.Db.Save(channelInfo)
	return err
}

func (this *channelManager) BobAcceptChannel(jsonData string, user *bean.User) (channelInfo *dao.ChannelInfo, err error) {
	reqData := &bean.AcceptChannelInfo{}
	err = json.Unmarshal([]byte(jsonData), &reqData)

	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	if reqData.Approval {
		if tool.CheckIsString(&reqData.FundingPubKey) == false {
			return nil, errors.New("wrong FundingPubKey")
		}

		reqData.FundingAddress, err = getAddressFromPubKey(reqData.FundingPubKey)
		if err != nil {
			return nil, err
		}
	}

	channelInfo = &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("PeerIdB", user.PeerId),
		q.Eq("CurrState", dao.ChannelState_Create)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, errors.New("can not find the channel " + reqData.TemporaryChannelId + " on Create state")
	}

	if channelInfo.PeerIdB != user.PeerId {
		return nil, errors.New("you are not the peerIdB")
	}

	if reqData.Approval {
		channelInfo.PubKeyB = reqData.FundingPubKey
		channelInfo.AddressB = reqData.FundingAddress
		multiSig, err := rpcClient.CreateMultiSig(2, []string{channelInfo.PubKeyA, channelInfo.PubKeyB})
		if err != nil {
			log.Println(err)
			return nil, err
		}
		channelInfo.ChannelAddress = gjson.Get(multiSig, "address").String()
		channelInfo.ChannelAddressRedeemScript = gjson.Get(multiSig, "redeemScript").String()

		addrInfoStr, err := rpcClient.GetAddressInfo(channelInfo.ChannelAddress)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		channelInfo.ChannelAddressScriptPubKey = gjson.Parse(addrInfoStr).Get("scriptPubKey").String()
		channelInfo.CurrState = dao.ChannelState_WaitFundAsset
	} else {
		channelInfo.CurrState = dao.ChannelState_OpenChannelDefuse
	}

	channelInfo.AcceptAt = time.Now()
	err = user.Db.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return channelInfo, err
}

//当bob操作完，发送信息到Alice所在的obd，obd处理先从bob得到发给alice的信息，然后再发给Alice的轻客户端
func (this *channelManager) AfterBobAcceptChannelAtAliceSide(jsonData string, user *bean.User) (channelInfo *dao.ChannelInfo, err error) {
	bobChannelInfo := &dao.ChannelInfo{}
	err = json.Unmarshal([]byte(jsonData), &bobChannelInfo)
	if err != nil {
		return nil, err
	}

	channelInfo = &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("TemporaryChannelId", bobChannelInfo.TemporaryChannelId),
		q.Eq("PeerIdA", user.PeerId),
		q.Eq("PeerIdB", bobChannelInfo.PeerIdB),
		q.Eq("CurrState", dao.ChannelState_Create)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, errors.New("can not find the channel " + bobChannelInfo.TemporaryChannelId + " on Create state")
	}

	if bobChannelInfo.CurrState == dao.ChannelState_WaitFundAsset {
		channelInfo.PubKeyB = bobChannelInfo.PubKeyB
		channelInfo.AddressB = bobChannelInfo.AddressB
		channelInfo.ChannelAddress = bobChannelInfo.ChannelAddress
		channelInfo.ChannelAddressRedeemScript = bobChannelInfo.ChannelAddressRedeemScript
		channelInfo.ChannelAddressScriptPubKey = bobChannelInfo.ChannelAddressScriptPubKey
		channelInfo.CurrState = dao.ChannelState_WaitFundAsset
	} else {
		channelInfo.CurrState = dao.ChannelState_OpenChannelDefuse
	}
	channelInfo.AcceptAt = time.Now()
	err = user.Db.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return channelInfo, err
}

// GetChannelByTemporaryChanId
func (this *channelManager) GetChannelByTemporaryChanId(jsonData string) (node *dao.ChannelInfo, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("wrong TemporaryChannelId")
	}
	return this.GetChannelByTemporaryChannelId(jsonData)
}

// GetChannelByTemporaryChannelId
func (this *channelManager) GetChannelByTemporaryChannelId(chanId string) (node *dao.ChannelInfo, err error) {
	node = &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("TemporaryChannelId", chanId)).
		First(node)
	return node, err
}

// DelChannelByTemporaryChanId
func (this *channelManager) DelChannelByTemporaryChanId(jsonData string) (node *dao.ChannelInfo, err error) {

	array := gjson.Parse(jsonData).Array()
	if len(array) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	var tempChanId chainhash.Hash
	for index, value := range array {
		tempChanId[index] = byte(value.Num)
	}
	node = &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("TemporaryChannelId", tempChanId)).
		First(node)
	if err == nil {
		err = db.DeleteStruct(node)
	}
	return node, err
}

// AllItem
func (this *channelManager) AllItem(peerId string) (data []dao.ChannelInfo, err error) {
	var infos []dao.ChannelInfo
	err = db.Select(
		q.Or(
			q.Eq("PeerIdA", peerId),
			q.Eq("PeerIdB", peerId))).
		OrderBy("CreateAt").Reverse().
		Find(&infos)
	return infos, err
}
func (this *channelManager) GetChannelInfoByChannelId(jsonData string, peerId string) (info *dao.ChannelInfo, err error) {
	channelId := gjson.Parse(jsonData).String()
	if tool.CheckIsString(&channelId) == false {
		return nil, errors.New("wrong ChannelId")
	}

	info = &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", peerId),
			q.Eq("PeerIdB", peerId))).
		First(info)
	return info, err
}
func (this *channelManager) GetChannelInfoById(jsonData string, peerId string) (info *dao.ChannelInfo, err error) {
	id, err := strconv.Atoi(jsonData)
	if err != nil {
		return nil, err
	}
	info = &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("Id", id),
		q.Or(
			q.Eq("PeerIdA", peerId),
			q.Eq("PeerIdB", peerId))).
		First(info)
	return info, err
}

// TotalCount
func (this *channelManager) TotalCount(peerId string) (count int, err error) {
	return db.Select(
		q.Or(
			q.Eq("PeerIdA", peerId),
			q.Eq("PeerIdB", peerId))).
		Count(&dao.ChannelInfo{})
}

//ForceCloseChannel
func (this *channelManager) ForceCloseChannel(jsonData string, user *bean.User) (interface{}, error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty inputData")
	}

	reqData := &bean.CloseChannel{}
	err := json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.ChannelId) == false {
		return nil, errors.New("wrong channelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", dao.ChannelState_CanUse)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	lastCommitmentTx, err := getLatestCommitmentTx(channelInfo.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if lastCommitmentTx.CurrState != dao.TxInfoState_Htlc_GetR && lastCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, errors.New("latest commnitmennt tx state is wrong")
	}

	if tool.CheckIsString(&lastCommitmentTx.RSMCTxHex) {
		commitmentTxid, err := rpcClient.SendRawTransaction(lastCommitmentTx.RSMCTxHex)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(commitmentTxid)
	}
	if tool.CheckIsString(&lastCommitmentTx.ToOtherTxHex) {
		commitmentTxidToBob, err := rpcClient.SendRawTransaction(lastCommitmentTx.ToOtherTxHex)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(commitmentTxidToBob)
	}

	lastRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		First(lastRevocableDeliveryTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	revocableDeliveryTxid, err := rpcClient.SendRawTransaction(lastRevocableDeliveryTx.TxHash)
	if err != nil {
		log.Println(err)
		msg := err.Error()
		if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
			return nil, err
		}
	}
	log.Println(revocableDeliveryTxid)

	tx, err := db.Begin(true)
	defer tx.Rollback()

	lastCommitmentTx.CurrState = dao.TxInfoState_SendHex
	lastCommitmentTx.SendAt = time.Now()
	err = tx.Update(lastCommitmentTx)
	if err != nil {
		return nil, err
	}

	lastRevocableDeliveryTx.CurrState = dao.TxInfoState_SendHex
	lastRevocableDeliveryTx.SendAt = time.Now()
	err = tx.Update(lastRevocableDeliveryTx)
	if err != nil {
		return nil, err
	}

	err = addRDTxToWaitDB(tx, lastRevocableDeliveryTx)
	if err != nil {
		return nil, err
	}

	channelInfo.CurrState = dao.ChannelState_Close
	channelInfo.CloseAt = time.Now()
	err = tx.Update(channelInfo)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return channelInfo, nil
}

func (this *channelManager) SendBreachRemedyTransaction(jsonData string, user *bean.User) (transaction *dao.BreachRemedyTransaction, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty inputData")
	}

	reqData := &bean.SendBreachRemedyTransaction{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.ChannelId) == false {
		return nil, errors.New("wrong channelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", dao.ChannelState_Close)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	lastBRTx := &dao.BreachRemedyTransaction{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CurrState", dao.TxInfoState_CreateAndSign),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		First(lastBRTx)
	if err != nil {
		err = errors.New("not found the latest br")
		log.Println(err)
		return nil, err
	}

	brtxid, err := rpcClient.SendRawTransaction(lastBRTx.TransactionSignHex)
	if err != nil {
		err = errors.New("BtcSignAndSendRawTransaction: " + err.Error())
		log.Println(err)
		return nil, err
	}
	log.Println(brtxid)

	lastBRTx.SendAt = time.Now()
	lastBRTx.CurrState = dao.TxInfoState_SendHex
	err = db.Update(lastBRTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return lastBRTx, nil
}

func (this *channelManager) RequestCloseChannel(jsonData string, user *bean.User) (interface{}, *string, error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, nil, errors.New("empty inputData")
	}
	reqData := &bean.CloseChannel{}
	err := json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	if channelInfo.CurrState != dao.ChannelState_CanUse && channelInfo.CurrState != dao.ChannelState_HtlcTx {
		return nil, nil, errors.New("wrong channel state " + strconv.Itoa(int(channelInfo.CurrState)))
	}

	targetUser := channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	}

	lastCommitmentTx, err := getLatestCommitmentTx(channelInfo.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	if lastCommitmentTx.CurrState != dao.TxInfoState_Htlc_GetH && lastCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, nil, errors.New("latest commitment tx state is wrong")
	}

	closeChannel := &dao.CloseChannel{}
	closeChannel.ChannelId = reqData.ChannelId
	closeChannel.Owner = user.PeerId
	closeChannel.CurrState = 0
	_ = db.Select(
		q.Eq("ChannelId", closeChannel.ChannelId),
		q.Eq("Owner", closeChannel.Owner),
		q.Eq("CurrState", closeChannel.CurrState)).
		Find(closeChannel)

	if closeChannel.Id == 0 {
		closeChannel.CreateAt = time.Now()
		dataBytes, _ := json.Marshal(closeChannel)
		closeChannel.RequestHex = tool.SignMsgWithSha256(dataBytes)
		err = db.Save(closeChannel)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}
	}

	toData := make(map[string]interface{})
	toData["channel_id"] = reqData.ChannelId
	toData["request_close_channel_hash"] = closeChannel.RequestHex
	return toData, &targetUser, nil
}

func (this *channelManager) CloseChannelSign(jsonData string, user *bean.User) (interface{}, string, error) {

	if tool.CheckIsString(&jsonData) == false {
		return nil, "", errors.New("empty inputData")
	}
	reqData := &bean.CloseChannelSign{}
	err := json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.ChannelId) == false {
		err = errors.New("empty channel_id")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.RequestCloseChannelHash) == false {
		err = errors.New("empty request_close_channel_hash")
		log.Println(err)
		return nil, "", err
	}

	closeChannelStarterData := &dao.CloseChannel{}
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", 0),
		q.Eq("RequestHex", reqData.RequestCloseChannelHash)).
		First(closeChannelStarterData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	targetUser := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		targetUser = channelInfo.PeerIdB
	}

	if reqData.Approval == false {
		log.Println("disagree close channel")
		return nil, targetUser, errors.New("disagree close channel")
	}

	latestCommitmentTx, err := getLatestCommitmentTx(channelInfo.ChannelId, targetUser)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	if latestCommitmentTx.CurrState != dao.TxInfoState_Htlc_GetR && latestCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, "", errors.New("latest commitment tx state is wrong")
	}

	// 当前是处于htlc的状态，且是获取到
	if channelInfo.CurrState == dao.ChannelState_HtlcTx {
		return this.CloseHtlcChannelSigned(channelInfo, closeChannelStarterData, *user)
	}

	//region 广播承诺交易 最近的rsmc的资产分配交易 因为是omni资产，承诺交易被拆分成了两个独立的交易
	if tool.CheckIsString(&latestCommitmentTx.RSMCTxHex) {
		commitmentTxid, err := rpcClient.SendRawTransaction(latestCommitmentTx.RSMCTxHex)
		if err != nil {
			log.Println(err)
			return nil, targetUser, err
		}
		log.Println(commitmentTxid)
	}
	if tool.CheckIsString(&latestCommitmentTx.ToOtherTxHex) {
		commitmentTxidToBob, err := rpcClient.SendRawTransaction(latestCommitmentTx.ToOtherTxHex)
		if err != nil {
			log.Println(err)
			return nil, targetUser, err
		}
		log.Println(commitmentTxidToBob)
	}
	//endregion

	//region 广播RD
	lastRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("Owner", targetUser)).
		OrderBy("CreateAt").Reverse().
		First(lastRevocableDeliveryTx)
	if err != nil {
		log.Println(err)
		return nil, targetUser, err
	}

	revocableDeliveryTxid, err := rpcClient.SendRawTransaction(lastRevocableDeliveryTx.TxHash)
	if err != nil {
		log.Println(err)
		msg := err.Error()
		//如果omnicore返回的信息里面包含了non-BIP68-final (code 64)， 则说明因为需要等待1000个区块高度，广播是对的
		if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
			return nil, targetUser, err
		}
	}
	log.Println(revocableDeliveryTxid)
	//endregion

	tx, err := db.Begin(true)
	defer tx.Rollback()

	// region update state
	latestCommitmentTx.CurrState = dao.TxInfoState_SendHex
	latestCommitmentTx.SendAt = time.Now()
	err = tx.Update(latestCommitmentTx)
	if err != nil {
		return nil, "", err
	}

	lastRevocableDeliveryTx.CurrState = dao.TxInfoState_SendHex
	lastRevocableDeliveryTx.SendAt = time.Now()
	err = tx.Update(lastRevocableDeliveryTx)
	if err != nil {
		return nil, "", err
	}

	err = addRDTxToWaitDB(tx, lastRevocableDeliveryTx)
	if err != nil {
		return nil, "", err
	}

	channelInfo.CurrState = dao.ChannelState_Close
	channelInfo.CloseAt = time.Now()
	err = tx.Update(channelInfo)
	if err != nil {
		return nil, "", err
	}

	closeChannelStarterData.CurrState = 1
	_ = tx.Update(closeChannelStarterData)
	//endregion

	err = tx.Commit()
	if err != nil {
		return nil, "", err
	}
	return channelInfo, targetUser, nil
}

//  htlc  when getH close channel
func (this *channelManager) CloseHtlcChannelSigned(channelInfo *dao.ChannelInfo, closeOpStaterReqData *dao.CloseChannel, user bean.User) (outData interface{}, closeOpStarter string, err error) {
	// 提现操作的发起者
	closeOpStarter = channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		closeOpStarter = channelInfo.PeerIdA
	}

	//获取到最新的交易，并保证他的txType是CommitmentTransactionType_Htlc类型
	latestCommitmentTx, err := getHtlcLatestCommitmentTx(channelInfo.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	//region 广播主承诺交易 三笔
	if tool.CheckIsString(&latestCommitmentTx.RSMCTxHex) {
		commitmentTxid, err := rpcClient.SendRawTransaction(latestCommitmentTx.RSMCTxHex)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		log.Println(commitmentTxid)
	}

	if tool.CheckIsString(&latestCommitmentTx.ToOtherTxHex) {
		commitmentTxidToBob, err := rpcClient.SendRawTransaction(latestCommitmentTx.ToOtherTxHex)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		log.Println(commitmentTxidToBob)
	}

	latestRsmcRD := &dao.RevocableDeliveryTransaction{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", latestCommitmentTx.Id),
		q.Eq("RDType", 0),
		q.Eq("Owner", closeOpStarter)).
		OrderBy("CreateAt").Reverse().
		First(latestRsmcRD)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	latestRsmcRDTxid, err := rpcClient.SendRawTransaction(latestRsmcRD.TxHash)
	if err != nil {
		log.Println(err)
		msg := err.Error()
		if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
			return nil, "", err
		}
	}
	log.Println(latestRsmcRDTxid)

	// htlc部分
	if tool.CheckIsString(&latestCommitmentTx.HtlcTxHex) {
		commitmentTxidToHtlc, err := rpcClient.SendRawTransaction(latestCommitmentTx.HtlcTxHex)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		log.Println(commitmentTxidToHtlc)
	}
	// endregion

	// 提现人是否是这次htlc的转账发起者
	var withdrawerIsHtlcSender bool
	if latestCommitmentTx.HtlcSender == closeOpStarter {
		withdrawerIsHtlcSender = true
	} else {
		withdrawerIsHtlcSender = false
	}

	tx, err := db.Begin(true)
	defer tx.Rollback()

	// region htlc的相关交易广播
	needRsmcTx := true

	// 提现人是这次htlc的转账发起者
	if withdrawerIsHtlcSender {
		// 如果已经得到R，直接广播HED1a
		if latestCommitmentTx.CurrState == dao.TxInfoState_Htlc_GetR {
			needRsmcTx = false
			hednx := &dao.HTLCExecutionDeliveryOfR{}
			err = tx.Select(
				q.Eq("CommitmentTxId", latestCommitmentTx.Id),
				q.Eq("CurrState", dao.TxInfoState_CreateAndSign),
				q.Eq("Owner", closeOpStarter)).
				First(hednx)
			if err == nil {
				if tool.CheckIsString(&hednx.TxHash) {
					_, err := rpcClient.SendRawTransaction(hednx.TxHash)
					if err != nil {
						log.Println(err)
						return nil, "", err
					}
					hednx.CurrState = dao.TxInfoState_SendHex
					hednx.SendAt = time.Now()
					_ = tx.Update(hednx)
				}
			}
		}
	} else { // 提现人是这次htlc的转账接收者
		//如果还没有获取到R 执行HTD1b
		if latestCommitmentTx.CurrState == dao.TxInfoState_Htlc_GetH {
			needRsmcTx = false
			htdnx := &dao.HTLCTimeoutDeliveryTxB{}
			err = tx.Select(
				q.Eq("CommitmentTxId", latestCommitmentTx.Id),
				q.Eq("CurrState", dao.TxInfoState_CreateAndSign),
				q.Eq("Owner", closeOpStarter)).
				First(htdnx)
			if err == nil {
				if tool.CheckIsString(&htdnx.TxHash) {
					_, err := rpcClient.SendRawTransaction(htdnx.TxHash)
					if err != nil {
						log.Println(err)
						msg := err.Error()
						if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
							return nil, "", err
						}
					}
					_ = addHTDnxTxToWaitDB(tx, htdnx)
					htdnx.CurrState = dao.TxInfoState_SendHex
					htdnx.SendAt = time.Now()
					_ = tx.Update(htdnx)
				}
			}
		}
	}
	//如果转账方在超时后还没有得到R,或者接收方得到R后想直接提现
	if needRsmcTx {
		htnx := &dao.HTLCTimeoutTxForAAndExecutionForB{}
		err = tx.Select(
			q.Eq("ChannelId", channelInfo.ChannelId),
			q.Eq("CommitmentTxId", latestCommitmentTx.Id),
			q.Eq("Owner", closeOpStarter),
			q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
			First(htnx)
		if err == nil {
			htrd := &dao.RevocableDeliveryTransaction{}
			err = tx.Select(
				q.Eq("CommitmentTxId", htnx.Id),
				q.Eq("Owner", closeOpStarter),
				q.Eq("RDType", 1),
				q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
				First(htrd)
			if err == nil {
				if tool.CheckIsString(&htnx.RSMCTxHash) {
					//广播alice的ht1a 或者bob的he1b
					_, err := rpcClient.SendRawTransaction(htnx.RSMCTxHash)
					if err == nil { //如果已经超时 比如alic的3天超时，bob的得到R后的交易的无等待锁定
						if tool.CheckIsString(&htrd.TxHash) {
							_, err = rpcClient.SendRawTransaction(htrd.TxHash)
							if err != nil {
								log.Println(err)
								msg := err.Error()
								if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
									return nil, "", err
								}
							}
							_ = addRDTxToWaitDB(tx, htrd)
							htnx.CurrState = dao.TxInfoState_SendHex
							htnx.SendAt = time.Now()
							_ = tx.Update(htnx)
						}
					} else {
						// 如果是超时内的正常提现
						if htnx.Timeout == 0 { //如果是bob的无等待交易，那就是广播报错了
							return nil, "", err
						}

						//如果是alice的（ht1a的锁定时间内的提现交易，就需要判断时候是正常的超时广播（含有non-BIP68-final (code 64)），如果不是，就返回
						log.Println(err)
						msg := err.Error()
						if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
							return nil, "", err
						}
						_ = addHT1aTxToWaitDB(tx, htnx, htrd)
					}
				}
			}
		}
	}
	// endregion

	// region update obj state to db
	latestCommitmentTx.CurrState = dao.TxInfoState_SendHex
	latestCommitmentTx.SendAt = time.Now()
	err = tx.Update(latestCommitmentTx)
	if err != nil {
		return nil, "", err
	}

	latestRsmcRD.CurrState = dao.TxInfoState_SendHex
	latestRsmcRD.SendAt = time.Now()
	err = tx.Update(latestRsmcRD)
	if err != nil {
		return nil, "", err
	}

	err = addRDTxToWaitDB(tx, latestRsmcRD)
	if err != nil {
		return nil, "", err
	}

	channelInfo.CurrState = dao.ChannelState_Close
	channelInfo.CloseAt = time.Now()
	err = tx.Update(channelInfo)
	if err != nil {
		return nil, "", err
	}

	closeOpStaterReqData.CurrState = 1
	_ = tx.Update(closeOpStaterReqData)
	//endregion

	err = tx.Commit()
	if err != nil {
		return nil, "", err
	}
	return channelInfo, closeOpStarter, nil
}

func addRDTxToWaitDB(tx storm.Node, lastRevocableDeliveryTx *dao.RevocableDeliveryTransaction) (err error) {
	node := &dao.RDTxWaitingSend{}
	count, err := tx.Select(
		q.Eq("TransactionHex", lastRevocableDeliveryTx.TxHash)).
		Count(node)
	if err == nil {
		return err
	}
	if count > 0 {
		return errors.New("already save")
	}
	node.TransactionHex = lastRevocableDeliveryTx.TxHash
	node.Type = 0
	node.IsEnable = true
	node.CreateAt = time.Now()
	err = tx.Save(node)
	if err != nil {
		return err
	}
	return nil
}
