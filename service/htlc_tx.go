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
func (service *htlcTxManager) FindPathOfSingleHopAndSendToBob(msgData string, user bean.User) (data map[string]interface{}, bob string, err error) {
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
	if FindUserIsOnline(bob) != nil {
		return nil, "", errors.New("inter node: " + bob + " not online")
	}

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

func getTwoChannelOfSingleHop(htlcCreateRandHInfo dao.HtlcCreateRandHInfo, channelAliceInfos []dao.ChannelInfo, channelCarlInfos []dao.ChannelInfo) (string, *dao.ChannelInfo, *dao.ChannelInfo) {
	for _, aliceChannel := range channelAliceInfos {
		if aliceChannel.PeerIdA == htlcCreateRandHInfo.SenderPeerId {
			bobPeerId := aliceChannel.PeerIdB
			carlChannel, err := getCarlChannelHasInterNodeBob(htlcCreateRandHInfo, aliceChannel, channelCarlInfos, aliceChannel.PeerIdA, bobPeerId)
			if err == nil {
				return bobPeerId, &aliceChannel, carlChannel
			}
		} else {
			bobPeerId := aliceChannel.PeerIdA
			carlChannel, err := getCarlChannelHasInterNodeBob(htlcCreateRandHInfo, aliceChannel, channelCarlInfos, aliceChannel.PeerIdB, bobPeerId)
			if err == nil {
				return bobPeerId, &aliceChannel, carlChannel
			}
		}
	}
	return "", nil, nil
}

func getCarlChannelHasInterNodeBob(htlcCreateRandHInfo dao.HtlcCreateRandHInfo, aliceChannel dao.ChannelInfo, channelCarlInfos []dao.ChannelInfo, alicePeerId, bobPeerId string) (*dao.ChannelInfo, error) {
	//alice and bob's channel, whether alice has enough money
	commitmentTxInfo, err := getLatestCommitmentTx(aliceChannel.ChannelId, alicePeerId)
	if err != nil {
		return nil, err
	}
	if commitmentTxInfo.AmountToRSMC < htlcCreateRandHInfo.Amount {
		return nil, errors.New("curr channel not have enough money")
	}
	//bob and carl's channel,whether bob has enough money
	for _, carlChannel := range channelCarlInfos {
		commitmentTxInfo, err := getLatestCommitmentTx(carlChannel.ChannelId, bobPeerId)
		if err != nil {
			continue
		}
		if commitmentTxInfo.AmountToRSMC < htlcCreateRandHInfo.Amount {
			continue
		}
		return &carlChannel, nil
	}
	return nil, errors.New("not found the channel")
}

func getAllChannels(peerId string) (channelInfos []dao.ChannelInfo) {
	var channelAInfos []dao.ChannelInfo
	_ = db.Select(q.Eq("PeerIdA", peerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&channelAInfos)
	var channelBInfos []dao.ChannelInfo
	_ = db.Select(q.Eq("PeerIdB", peerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&channelBInfos)

	channelInfos = []dao.ChannelInfo{}
	if channelAInfos != nil && len(channelAInfos) > 0 {
		channelInfos = append(channelInfos, channelAInfos...)
	}
	if channelBInfos != nil && len(channelBInfos) > 0 {
		channelInfos = append(channelInfos, channelBInfos...)
	}
	return channelInfos
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
