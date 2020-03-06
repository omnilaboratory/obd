package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/config"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"strings"
	"time"
)

type channelManager struct{}

var ChannelService = channelManager{}

// AliceOpenChannel init ChannelInfo
func (c *channelManager) AliceOpenChannel(msg bean.RequestMessage, peerIdA string) (data *bean.OpenChannelInfo, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New("empty inputData")
	}

	data = &bean.OpenChannelInfo{}
	err = json.Unmarshal([]byte(msg.Data), &data)
	if err != nil {
		return nil, err
	}

	data.FundingAddress, err = getAddressFromPubKey(data.FundingPubKey)
	if err != nil {
		return nil, err
	}

	data.ChainHash = config.Init_node_chain_hash
	data.TemporaryChannelId = bean.ChannelIdService.NextTemporaryChanID()

	channelInfo := &dao.ChannelInfo{}
	channelInfo.OpenChannelInfo = *data
	channelInfo.PeerIdA = peerIdA
	channelInfo.PeerIdB = msg.RecipientPeerId
	channelInfo.PubKeyA = data.FundingPubKey
	channelInfo.AddressA = data.FundingAddress
	channelInfo.CurrState = dao.ChannelState_Create
	channelInfo.CreateAt = time.Now()
	channelInfo.CreateBy = peerIdA

	err = db.Save(channelInfo)
	return data, err
}

func (c *channelManager) BobAcceptChannel(jsonData string, peerIdB string) (channelInfo *dao.ChannelInfo, err error) {
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
	err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("PeerIdB", peerIdB), q.Eq("CurrState", dao.ChannelState_Create)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, errors.New("can not find the channel " + reqData.TemporaryChannelId + " on Create state")
	}

	if channelInfo.PeerIdB != peerIdB {
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
	}
	if reqData.Approval {
		channelInfo.CurrState = dao.ChannelState_CanUse
	} else {
		channelInfo.CurrState = dao.ChannelState_OpenChannelDefuse
	}
	channelInfo.AcceptAt = time.Now()
	err = db.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return channelInfo, err
}

// GetChannelByTemporaryChanId
func (c *channelManager) GetChannelByTemporaryChanId(jsonData string) (node *dao.ChannelInfo, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("wrong TemporaryChannelId")
	}
	return c.GetChannelByTemporaryChanIdArray(jsonData)
}

// GetChannelByTemporaryChanIdArray
func (c *channelManager) GetChannelByTemporaryChanIdArray(chanId string) (node *dao.ChannelInfo, err error) {
	node = &dao.ChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", chanId)).First(node)
	return node, err
}

// DelChannelByTemporaryChanId
func (c *channelManager) DelChannelByTemporaryChanId(jsonData string) (node *dao.ChannelInfo, err error) {

	array := gjson.Parse(jsonData).Array()
	if len(array) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	var tempChanId chainhash.Hash
	for index, value := range array {
		tempChanId[index] = byte(value.Num)
	}
	node = &dao.ChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", tempChanId)).First(node)
	if err == nil {
		err = db.DeleteStruct(node)
	}
	return node, err
}

// AllItem
func (c *channelManager) AllItem(peerId string) (data []dao.ChannelInfo, err error) {
	var infos []dao.ChannelInfo
	err = db.Select(q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).OrderBy("CreateAt").Reverse().Find(&infos)
	return infos, err
}
func (c *channelManager) GetChannelInfoByChannelId(jsonData string, peerId string) (info *dao.ChannelInfo, err error) {
	array := gjson.Parse(jsonData).Array()
	if len(array) != 32 {
		return nil, errors.New("wrong ChannelId")
	}

	var channelId bean.ChannelID
	for index, value := range array {
		channelId[index] = byte(value.Num)
	}
	info = &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", channelId), q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).First(info)
	return info, err
}
func (c *channelManager) GetChannelInfoById(jsonData string, peerId string) (info *dao.ChannelInfo, err error) {
	id, err := strconv.Atoi(jsonData)
	if err != nil {
		return nil, err
	}
	info = &dao.ChannelInfo{}
	err = db.Select(q.Eq("Id", id), q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).First(info)
	return info, err
}

// TotalCount
func (c *channelManager) TotalCount(peerId string) (count int, err error) {
	return db.Select(q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).Count(&dao.ChannelInfo{})
}

//ForceCloseChannel
func (c *channelManager) ForceCloseChannel(jsonData string, user *bean.User) (interface{}, error) {
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
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_CanUse)).First(channelInfo)
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

	if tool.CheckIsString(&lastCommitmentTx.RSMCTxHash) {
		commitmentTxid, err := rpcClient.SendRawTransaction(lastCommitmentTx.RSMCTxHash)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(commitmentTxid)
	}
	if tool.CheckIsString(&lastCommitmentTx.ToOtherTxHash) {
		commitmentTxidToBob, err := rpcClient.SendRawTransaction(lastCommitmentTx.ToOtherTxHash)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(commitmentTxidToBob)
	}

	lastRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(lastRevocableDeliveryTx)
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

func (c *channelManager) SendBreachRemedyTransaction(jsonData string, user *bean.User) (transaction *dao.BreachRemedyTransaction, err error) {
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
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_Close)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	lastBRTx := &dao.BreachRemedyTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CurrState", dao.TxInfoState_CreateAndSign), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(lastBRTx)
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

func (c *channelManager) RequestCloseChannel(jsonData string, user *bean.User) (interface{}, *string, error) {
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
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_CanUse)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, nil, err
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

	if lastCommitmentTx.CurrState != dao.TxInfoState_Htlc_GetR && lastCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, nil, errors.New("latest commitment tx state is wrong")
	}

	dbData := &dao.CloseChannel{}
	dbData.ChannelId = reqData.ChannelId
	dbData.Owner = user.PeerId
	dbData.CommitmentTxId = lastCommitmentTx.Id
	dbData.CurrState = 0
	dbData.CreateAt = time.Now()
	dataBytes, _ := json.Marshal(dbData)
	dbData.RequestHex = tool.SignMsgWithSha256(dataBytes)
	err = db.Save(dbData)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	toData := make(map[string]interface{})
	toData["channel_id"] = reqData.ChannelId
	toData["request_close_channel_hash"] = dbData.RequestHex
	return toData, &targetUser, nil
}

func (c *channelManager) CloseChannelSign(jsonData string, user *bean.User) (interface{}, *string, error) {

	if tool.CheckIsString(&jsonData) == false {
		return nil, nil, errors.New("empty inputData")
	}
	reqData := &bean.CloseChannelSign{}
	err := json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	if tool.CheckIsString(&reqData.ChannelId) == false {
		err = errors.New("empty channel_id")
		log.Println(err)
		return nil, nil, err
	}

	if tool.CheckIsString(&reqData.RequestCloseChannelHash) == false {
		err = errors.New("empty request_close_channel_hash")
		log.Println(err)
		return nil, nil, err
	}

	closeChannelStarterData := &dao.CloseChannel{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", 0), q.Eq("RequestHex", reqData.RequestCloseChannelHash)).First(closeChannelStarterData)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_CanUse)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	targetUser := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		targetUser = channelInfo.PeerIdB
	}

	if reqData.Approval == false {
		log.Println("disagree close channel")
		return nil, &targetUser, errors.New("disagree close channel")
	}

	lastCommitmentTx, err := getLatestCommitmentTx(channelInfo.ChannelId, targetUser)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	if lastCommitmentTx.CurrState != dao.TxInfoState_Htlc_GetR && lastCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, nil, errors.New("latest commitment tx state is wrong")
	}

	if tool.CheckIsString(&lastCommitmentTx.RSMCTxHash) {
		commitmentTxid, err := rpcClient.SendRawTransaction(lastCommitmentTx.RSMCTxHash)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}
		log.Println(commitmentTxid)
	}
	if tool.CheckIsString(&lastCommitmentTx.ToOtherTxHash) {
		commitmentTxidToBob, err := rpcClient.SendRawTransaction(lastCommitmentTx.ToOtherTxHash)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}
		log.Println(commitmentTxidToBob)
	}

	lastRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", targetUser)).OrderBy("CreateAt").Reverse().First(lastRevocableDeliveryTx)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	revocableDeliveryTxid, err := rpcClient.SendRawTransaction(lastRevocableDeliveryTx.TxHash)
	if err != nil {
		log.Println(err)
		msg := err.Error()
		if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
			return nil, nil, err
		}
	}
	log.Println(revocableDeliveryTxid)

	tx, err := db.Begin(true)
	defer tx.Rollback()

	lastCommitmentTx.CurrState = dao.TxInfoState_SendHex
	lastCommitmentTx.SendAt = time.Now()
	err = tx.Update(lastCommitmentTx)
	if err != nil {
		return nil, nil, err
	}

	lastRevocableDeliveryTx.CurrState = dao.TxInfoState_SendHex
	lastRevocableDeliveryTx.SendAt = time.Now()
	err = tx.Update(lastRevocableDeliveryTx)
	if err != nil {
		return nil, nil, err
	}

	err = addRDTxToWaitDB(tx, lastRevocableDeliveryTx)
	if err != nil {
		return nil, nil, err
	}

	channelInfo.CurrState = dao.ChannelState_Close
	channelInfo.CloseAt = time.Now()
	err = tx.Update(channelInfo)
	if err != nil {
		return nil, nil, err
	}

	closeChannelStarterData.CurrState = 1
	_ = tx.Update(closeChannelStarterData)

	err = tx.Commit()
	if err != nil {
		return nil, nil, err
	}

	return channelInfo, &targetUser, nil
}

func addRDTxToWaitDB(tx storm.Node, lastRevocableDeliveryTx *dao.RevocableDeliveryTransaction) (err error) {
	node := &dao.RDTxWaitingSend{}
	count, err := tx.Select(q.Eq("TransactionHex", lastRevocableDeliveryTx.TxHash)).Count(node)
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
