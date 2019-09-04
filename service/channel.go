package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/config"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"time"
)

type channelManager struct{}

var ChannelService = channelManager{}

// OpenChannel init data
func (c *channelManager) OpenChannel(msg bean.RequestMessage, peerIdA string) (data *bean.OpenChannelInfo, err error) {

	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New("empty inputData")
	}

	data = &bean.OpenChannelInfo{}
	json.Unmarshal([]byte(msg.Data), &data)

	if len(data.FundingPubKey) != 34 {
		return nil, errors.New("wrong fundingPubKey")
	}

	//TODO need to validate pubKey
	isMine, err := rpcClient.ValidateAddress(data.FundingPubKey)
	if err != nil {
		return nil, err
	}
	if isMine == false {
		return nil, errors.New("invalid fundingPubKey")
	}

	data.ChainHash = config.Init_node_chain_hash
	tempId := bean.ChannelIdService.NextTemporaryChanID()
	data.TemporaryChannelId = tempId

	node := &dao.ChannelInfo{}
	node.OpenChannelInfo = *data
	node.PeerIdA = peerIdA
	node.PeerIdB = msg.RecipientPeerId
	node.PubKeyA = data.FundingPubKey
	node.CurrState = dao.ChannelState_Create
	node.CreateAt = time.Now()

	db.Save(node)
	return data, err
}

func (c *channelManager) AcceptChannel(jsonData string, peerIdB string) (node *dao.ChannelInfo, err error) {
	data := &bean.AcceptChannelInfo{}
	err = json.Unmarshal([]byte(jsonData), &data)

	if err != nil {
		return nil, err
	}

	if len(data.TemporaryChannelId) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	if data.Attitude {
		if len(data.FundingPubKey) != 34 {
			return nil, errors.New("wrong PubKeyB")
		}
		isMine, err := rpcClient.ValidateAddress(data.FundingPubKey)
		if err != nil {
			return nil, err
		}
		if isMine == false {
			return nil, errors.New("invalid fundingPubKey")
		}
	}

	node = &dao.ChannelInfo{}
	db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(node)

	if &node.Id == nil {
		return nil, errors.New("invalid TemporaryChannelId")
	}
	if node.CurrState != dao.ChannelState_Create {
		return nil, errors.New("invalid ChannelState " + strconv.Itoa(int(node.CurrState)))
	}

	if data.Attitude {
		node.PubKeyB = data.FundingPubKey

		multiSig, err := rpcClient.CreateMultiSig(2, []string{node.PubKeyA, node.PubKeyB})
		if err != nil {
			return nil, err
		}
		node.ChannelPubKey = gjson.Get(multiSig, "address").String()
		node.RedeemScript = gjson.Get(multiSig, "redeemScript").String()
	}
	node.AcceptAt = time.Now()
	if data.Attitude {
		node.CurrState = dao.ChannelState_Accept
	} else {
		node.CurrState = dao.ChannelState_OpenChannelDefuse
	}
	err = db.Update(node)

	return node, err
}

// GetChannelByTemporaryChanId
func (c *channelManager) GetChannelByTemporaryChanId(jsonData string) (node *dao.ChannelInfo, err error) {
	array := gjson.Parse(jsonData).Array()
	if len(array) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	var tempChanId chainhash.Hash
	for index, value := range array {
		tempChanId[index] = byte(value.Num)
	}
	return c.GetChannelByTemporaryChanIdArray(tempChanId)
}

// GetChannelByTemporaryChanIdArray
func (c *channelManager) GetChannelByTemporaryChanIdArray(chanId chainhash.Hash) (node *dao.ChannelInfo, err error) {
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

// TotalCount
func (c *channelManager) TotalCount(peerId string) (count int, err error) {
	return db.Select(q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).Count(&dao.ChannelInfo{})
}

//CloseChannel
func (c *channelManager) CloseChannel(jsonData string, user *bean.User) (interface{}, error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty inputData")
	}

	reqData := &bean.CloseChannel{}
	err := json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		return nil, errors.New("wrong channelId")
	}

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		return nil, errors.New("empty ChannelAddressPrivateKey")
	}
	if tool.CheckIsString(&reqData.LastTempPrivateKey) == false {
		return nil, errors.New("empty LastTempPrivateKey")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	creatorSide := 0
	if user.PeerId == channelInfo.PeerIdB {
		creatorSide = 1
	}

	lastCommitmentTx := &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CreatorSide", creatorSide)).OrderBy("CreateAt").Reverse().First(lastCommitmentTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	commitmentTxid, chex, err := rpcClient.BtcSignAndSendRawTransaction(lastCommitmentTx.TxHexFirstSign, reqData.ChannelAddressPrivateKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(commitmentTxid)

	lastRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CreatorSide", creatorSide)).OrderBy("CreateAt").Reverse().First(lastRevocableDeliveryTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	revocableDeliveryTxid, rdhex, err := rpcClient.BtcSignAndSendRawTransaction(lastCommitmentTx.TxHexFirstSign, reqData.LastTempPrivateKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(revocableDeliveryTxid)

	tx, err := db.Begin(true)
	defer tx.Rollback()

	lastCommitmentTx.CurrState = dao.TxInfoState_MyselfSign
	lastCommitmentTx.TxHexEndSign = chex
	lastCommitmentTx.EndSignAt = time.Now()
	tx.Update(lastCommitmentTx)

	lastRevocableDeliveryTx.CurrState = dao.TxInfoState_MyselfSign
	lastRevocableDeliveryTx.TxHexEndSign = rdhex
	lastRevocableDeliveryTx.EndSignAt = time.Now()
	tx.Update(lastRevocableDeliveryTx)

	channelInfo.CurrState = dao.ChannelState_Close
	channelInfo.CloseAt = time.Now()
	tx.Update(channelInfo)

	tx.Commit()

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

	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		return nil, errors.New("wrong channelId")
	}

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		return nil, errors.New("empty ChannelAddressPrivateKey")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	creatorSide := 1
	if user.PeerId == channelInfo.PeerIdB {
		creatorSide = 0
	}

	lastBRTx := &dao.BreachRemedyTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Eq("CreatorSide", creatorSide)).OrderBy("CreateAt").Reverse().First(lastBRTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	brtxid, brhex, err := rpcClient.BtcSignAndSendRawTransaction(lastBRTx.TxHexFirstSign, reqData.ChannelAddressPrivateKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(brtxid)

	lastBRTx.TxHexEndSign = brhex
	lastBRTx.EndSignAt = time.Now()
	lastBRTx.CurrState = dao.TxInfoState_MyselfSign
	err = db.Update(lastBRTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return lastBRTx, nil
}
