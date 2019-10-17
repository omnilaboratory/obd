package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"log"
	"sync"
	"time"
)

type htlcTxManager struct {
	operationFlag sync.Mutex
}

var HtlcTxService htlcTxManager

// query bob,and ask bob
func (service *htlcTxManager) AliceFindPathOfSingleHopAndSendToBob(msgData string, user bean.User) (data map[string]interface{}, bob string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcRequestCreate{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	htlcCreateRandHInfo := &dao.HtlcCreateRandHInfo{}
	err = db.Select(q.Eq("CreateBy", user.PeerId), q.Eq("CurrState", dao.NS_Finish), q.Eq("H", reqData.H)).First(htlcCreateRandHInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	channelAliceInfos := getAllChannels(htlcCreateRandHInfo.SenderPeerId)
	if len(channelAliceInfos) == 0 {
		return nil, "", errors.New("sender's channel not found")
	}
	//if has the channel direct
	for _, item := range channelAliceInfos {
		if item.PeerIdA == htlcCreateRandHInfo.SenderPeerId && item.PeerIdB == htlcCreateRandHInfo.RecipientPeerId {
			return nil, "", errors.New("has direct channel")
		}
		if item.PeerIdB == htlcCreateRandHInfo.SenderPeerId && item.PeerIdA == htlcCreateRandHInfo.RecipientPeerId {
			return nil, "", errors.New("has direct channel")
		}
	}

	channelCarlInfos := getAllChannels(htlcCreateRandHInfo.RecipientPeerId)
	if len(channelCarlInfos) == 0 {
		return nil, "", errors.New("recipient's channel not found")
	}

	bob, aliceChannel, carlChannel := getTwoChannelOfSingleHop(*htlcCreateRandHInfo, channelAliceInfos, channelCarlInfos)
	if tool.CheckIsString(&bob) == false {
		return nil, "", errors.New("no inter channel can use")
	}

	log.Println(aliceChannel)
	log.Println(carlChannel)
	log.Println(bob)

	// operate db
	htlcSingleHopTxBaseInfo := &dao.HtlcSingleHopTxBaseInfo{}
	htlcSingleHopTxBaseInfo.FirstChannelId = aliceChannel.Id
	htlcSingleHopTxBaseInfo.SecondChannelId = carlChannel.Id
	htlcSingleHopTxBaseInfo.InterNodePeerId = bob
	htlcSingleHopTxBaseInfo.HtlcCreateRandHInfoRequestHash = htlcCreateRandHInfo.RequestHash
	htlcSingleHopTxBaseInfo.CurrState = dao.NS_Create
	htlcSingleHopTxBaseInfo.CreateBy = user.PeerId
	htlcSingleHopTxBaseInfo.CreateAt = time.Now()
	err = db.Save(htlcSingleHopTxBaseInfo)
	if err != nil {
		return nil, "", err
	}

	data = make(map[string]interface{})
	data["request_hash"] = htlcCreateRandHInfo.RequestHash
	data["h"] = htlcCreateRandHInfo.H
	return data, bob, nil
}

func (service *htlcTxManager) BobConfirmPath(msgData string, user bean.User) (data map[string]interface{}, senderPeerId string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	htlcSignRequestCreate := &bean.HtlcSignRequestCreate{}
	err = json.Unmarshal([]byte(msgData), htlcSignRequestCreate)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if htlcSignRequestCreate.Approval {
		if tool.CheckIsString(&htlcSignRequestCreate.ChannelAddressPrivateKey) == false {
			return nil, "", errors.New("channel_address_private_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.LastTempAddressPrivateKey) == false {
			return nil, "", errors.New("last_temp_address_private_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.CurrRsmcTempAddressPubKey) == false {
			return nil, "", errors.New("curr_rsmc_temp_address_pub_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.CurrRsmcTempAddressPrivateKey) == false {
			return nil, "", errors.New("curr_rsmc_temp_address_private_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.CurrHtlcTempAddressForHt1aPubKey) == false {
			return nil, "", errors.New("curr_htlc_temp_address_for_ht1a_pub_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.CurrHtlcTempAddressForHt1aPrivateKey) == false {
			return nil, "", errors.New("curr_htlc_temp_address_for_ht1a_private_key is empty")
		}
	}

	tx, err := db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	defer tx.Rollback()

	htlcCreateRandHInfo := &dao.HtlcCreateRandHInfo{}
	err = tx.Select(q.Eq("RequestHash", htlcSignRequestCreate.RequestHash)).First(htlcCreateRandHInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	htlcSingleHopTxBaseInfo := &dao.HtlcSingleHopTxBaseInfo{}
	err = tx.Select(q.Eq("HtlcCreateRandHInfoRequestHash", htlcSignRequestCreate.RequestHash)).First(htlcSingleHopTxBaseInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	htlcSingleHopTxBaseInfo.CurrState = dao.NS_Refuse
	if htlcSignRequestCreate.Approval {
		htlcSingleHopTxBaseInfo.CurrState = dao.NS_Finish

		htlcSingleHopTxBaseInfo.BobCurrRsmcTempPubKey = htlcSignRequestCreate.CurrRsmcTempAddressPubKey
		htlcSingleHopTxBaseInfo.BobCurrHtlcTempPubKey = htlcSignRequestCreate.CurrHtlcTempAddressPubKey

		//锁定两个通道
		aliceChannel := &dao.ChannelInfo{}
		err := tx.One("Id", htlcSingleHopTxBaseInfo.FirstChannelId, aliceChannel)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}
		carlChannel := &dao.ChannelInfo{}
		err = tx.One("Id", htlcSingleHopTxBaseInfo.SecondChannelId, carlChannel)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}

		aliceChannel.CurrState = dao.ChannelState_HtlcBegin
		err = tx.Update(aliceChannel)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}

		carlChannel.CurrState = dao.ChannelState_HtlcBegin
		err = tx.Update(carlChannel)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}

		if aliceChannel.PeerIdB == user.PeerId {
			tempAddrPrivateKeyMap[aliceChannel.PubKeyB] = htlcSignRequestCreate.ChannelAddressPrivateKey
		} else {
			tempAddrPrivateKeyMap[aliceChannel.PubKeyA] = htlcSignRequestCreate.ChannelAddressPrivateKey
		}
		bobLatestCommitmentTx, err := getLatestCommitmentTx(aliceChannel.ChannelId, user.PeerId)
		if err == nil {
			tempAddrPrivateKeyMap[bobLatestCommitmentTx.RSMCTempAddressPubKey] = htlcSignRequestCreate.LastTempAddressPrivateKey
		}
		tempAddrPrivateKeyMap[htlcSingleHopTxBaseInfo.BobCurrRsmcTempPubKey] = htlcSignRequestCreate.CurrRsmcTempAddressPrivateKey
		tempAddrPrivateKeyMap[htlcSingleHopTxBaseInfo.BobCurrHtlcTempPubKey] = htlcSignRequestCreate.CurrHtlcTempAddressPrivateKey
	}
	htlcSingleHopTxBaseInfo.SignBy = user.PeerId
	htlcSingleHopTxBaseInfo.SignAt = time.Now()
	err = tx.Update(htlcSingleHopTxBaseInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	data = make(map[string]interface{})
	data["approval"] = htlcSignRequestCreate.Approval
	data["request_hash"] = htlcSignRequestCreate.RequestHash
	return data, htlcCreateRandHInfo.SenderPeerId, nil
}

// -44
func (service *htlcTxManager) AliceOpenHtlcChannel(msgData string, user bean.User) (outData map[string]interface{}, targetUser string, err error) {
	if tool.CheckIsString(&msgData) == false {
		err = errors.New("empty json data")
		log.Println(err)
		return nil, "", err
	}
	htlcRequestOpen := &bean.HtlcRequestOpen{}
	err = json.Unmarshal([]byte(msgData), htlcRequestOpen)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&htlcRequestOpen.RequestHash) == false {
		err = errors.New("empty request_hash")
		log.Println(err)
		return nil, "", err
	}

	singleHopTxBaseInfo := dao.HtlcSingleHopTxBaseInfo{}
	err = db.Select(q.Eq("", htlcRequestOpen.RequestHash)).First(&singleHopTxBaseInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&htlcRequestOpen.ChannelAddressPrivateKey) == false {
		err = errors.New("channel_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.LastTempAddressPrivateKey) == false {
		err = errors.New("last_temp_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.CurrRsmcTempAddressPubKey) == false {
		err = errors.New("curr_rsmc_temp_address_pub_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.CurrRsmcTempAddressPrivateKey) == false {
		err = errors.New("curr_rsmc_temp_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.CurrHtlcTempAddressForHt1aPubKey) == false {
		err = errors.New("curr_htlc_temp_address_for_ht1a_pub_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.CurrHtlcTempAddressForHt1aPrivateKey) == false {
		err = errors.New("curr_htlc_temp_address_for_ht1a_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	//1、上一个交易必然是RSMC交易，所以需要结算上一个交易，为其创建BR交易
	//2、然后创建HTLC的commitment交易（Cna和Cnb），它有一个输入（三个btc的input），三个输出（rsmc，bob，htlc）
	//3、关于htlc的输出，也是把资金放到一个临时多签地址里面，这个资金在Alice(交易发起方)一方会创建一个锁定一天的交易（HT1a）
	//4、HT1a的构造，Cna的第三个输出作为输入，
	// 	其输出就是产生htlc里面的rsmc（为何要用这种呢？这个本身是alice自己的余额，所以提现是需要限制的，限制就是rsmc）
	// 	和CommitmentTx一样，要产生rsmc，就是要创建一个临时多签地址，所以又需要一组私钥
	// 	所以alice要创建BR，Cna，HT1a，HED1a,HTRD1a

	//launch database transaction, if anything goes wrong, roll back.
	tx, err := db.Begin(true)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback()

	channelInfo := dao.ChannelInfo{}
	err = tx.One("Id", singleHopTxBaseInfo.FirstChannelId, &channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if user.PeerId == channelInfo.PeerIdA {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = htlcRequestOpen.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = htlcRequestOpen.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	}

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = tx.Select(q.Eq("ChannelId", singleHopTxBaseInfo.FirstChannelId), q.Eq("CurrState", dao.FundingTransactionState_Accept)).OrderBy("CreateAt").Reverse().First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	//PeerIdA(概念中的Alice) 对上一次承诺交易的废弃
	err = htlcAliceAbortLastCommitmentTx(tx, channelInfo, user, *fundingTransaction, *htlcRequestOpen)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	//PeerIdB(概念中的Bob) 对上一次承诺交易的废弃
	err = htlcBobAbortLastCommitmentTx(tx, channelInfo, user, *fundingTransaction, *htlcRequestOpen)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	err = htlcCreateAliceTxes(tx, user)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	return nil, "", nil
}
