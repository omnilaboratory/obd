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
	json.Unmarshal([]byte(msg.Data), &data)

	if tool.CheckIsString(&data.FundingPubKey) == false {
		return nil, errors.New("wrong fundingPubKey")
	}

	//TODO need to validate pubKey
	//isMine, err := rpcClient.ValidateAddress(data.FundingPubKey)
	//if err != nil {
	//	return nil, err
	//}
	//if isMine == false {
	//	return nil, errors.New("invalid fundingPubKey")
	//}

	data.ChainHash = config.Init_node_chain_hash
	data.TemporaryChannelId = bean.ChannelIdService.NextTemporaryChanID()

	channelInfo := &dao.ChannelInfo{}
	channelInfo.OpenChannelInfo = *data
	channelInfo.PeerIdA = peerIdA
	channelInfo.PeerIdB = msg.RecipientPeerId
	channelInfo.PubKeyA = data.FundingPubKey
	channelInfo.CurrState = dao.ChannelState_Create
	channelInfo.CreateAt = time.Now()
	channelInfo.CreateBy = peerIdA

	db.Save(channelInfo)
	return data, err
}

func (c *channelManager) BobAcceptChannel(jsonData string, peerIdB string) (channelInfo *dao.ChannelInfo, err error) {
	reqData := &bean.AcceptChannelInfo{}
	err = json.Unmarshal([]byte(jsonData), &reqData)

	if err != nil {
		return nil, err
	}

	if len(reqData.TemporaryChannelId) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	if reqData.Attitude {
		if tool.CheckIsString(&reqData.FundingPubKey) == false {
			return nil, errors.New("wrong PubKeyB")
		}
		//isMine, err := rpcClient.ValidateAddress(reqData.FundingPubKey)
		//if err != nil {
		//	return nil, err
		//}
		//if isMine == false {
		//	return nil, errors.New("invalid fundingPubKey")
		//}
	}

	channelInfo = &dao.ChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("CurrState", dao.ChannelState_Create)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if reqData.Attitude {
		channelInfo.PubKeyB = reqData.FundingPubKey
		multiSig, err := rpcClient.CreateMultiSig(2, []string{channelInfo.PubKeyA, channelInfo.PubKeyB})
		if err != nil {
			log.Println(err)
			return nil, err
		}
		channelInfo.ChannelAddress = gjson.Get(multiSig, "address").String()
		channelInfo.ChannelAddressRedeemScript = gjson.Get(multiSig, "redeemScript").String()
	}
	if reqData.Attitude {
		channelInfo.CurrState = dao.ChannelState_Accept
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

	lastCommitmentTx := &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(lastCommitmentTx)
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
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(lastRevocableDeliveryTx)
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
	err = tx.Update(lastCommitmentTx)
	if err != nil {
		return nil, err
	}

	lastRevocableDeliveryTx.CurrState = dao.TxInfoState_MyselfSign
	lastRevocableDeliveryTx.TxHexEndSign = rdhex
	lastRevocableDeliveryTx.EndSignAt = time.Now()
	err = tx.Update(lastRevocableDeliveryTx)
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

	lastBRTx := &dao.BreachRemedyTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(lastBRTx)
	if err != nil {
		err = errors.New("not found the latest br")
		log.Println(err)
		return nil, err
	}

	brtxid, brhex, err := rpcClient.BtcSignAndSendRawTransaction(lastBRTx.TxHexFirstSign, reqData.ChannelAddressPrivateKey)
	if err != nil {
		err = errors.New("BtcSignAndSendRawTransaction: " + err.Error())
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
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	targetUser := channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	}

	lastCommitmentTx := &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(lastCommitmentTx)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	if user.PeerId == channelInfo.PeerIdA {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = reqData.ChannelAddressPrivateKey
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = reqData.ChannelAddressPrivateKey
	}
	tempAddrPrivateKeyMap[lastCommitmentTx.PubKey2] = reqData.LastTempPrivateKey
	reqData.ChannelAddressPrivateKey = ""
	reqData.LastTempPrivateKey = ""
	return reqData, &targetUser, nil
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

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	targetUser := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		targetUser = channelInfo.PeerIdB
	}

	if reqData.Attitude == false {
		log.Println("disagree close channel")
		return nil, &targetUser, errors.New("disagree close channel")
	}

	lastCommitmentTx := &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(lastCommitmentTx)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	var channelAddressPrivateKey = ""
	var lastTempPrivateKey = ""
	if user.PeerId == channelInfo.PeerIdB {
		channelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyA]
	} else {
		channelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyB]
	}
	lastTempPrivateKey = tempAddrPrivateKeyMap[lastCommitmentTx.PubKey2]

	if tool.CheckIsString(&channelAddressPrivateKey) == false || tool.CheckIsString(&lastTempPrivateKey) == false {
		log.Println("error private key, can't signature the transaction")
		return nil, &targetUser, errors.New("error private key, can't signature the transaction")
	}

	commitmentTxid, chex, err := rpcClient.BtcSignAndSendRawTransaction(lastCommitmentTx.TxHexFirstSign, channelAddressPrivateKey)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	log.Println(commitmentTxid)

	lastRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(lastRevocableDeliveryTx)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	revocableDeliveryTxid, rdhex, err := rpcClient.BtcSignAndSendRawTransaction(lastCommitmentTx.TxHexFirstSign, lastTempPrivateKey)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	log.Println(revocableDeliveryTxid)

	tx, err := db.Begin(true)
	defer tx.Rollback()

	lastCommitmentTx.CurrState = dao.TxInfoState_MyselfSign
	lastCommitmentTx.TxHexEndSign = chex
	lastCommitmentTx.EndSignAt = time.Now()
	err = tx.Update(lastCommitmentTx)
	if err != nil {
		return nil, nil, err
	}

	lastRevocableDeliveryTx.CurrState = dao.TxInfoState_MyselfSign
	lastRevocableDeliveryTx.TxHexEndSign = rdhex
	lastRevocableDeliveryTx.EndSignAt = time.Now()
	err = tx.Update(lastRevocableDeliveryTx)
	if err != nil {
		return nil, nil, err
	}
	//TODO add to timer

	channelInfo.CurrState = dao.ChannelState_Close
	channelInfo.CloseAt = time.Now()
	err = tx.Update(channelInfo)
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, err
	}

	return nil, &targetUser, nil
}
