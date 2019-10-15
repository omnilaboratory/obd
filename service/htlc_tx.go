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
func (service *htlcTxManager) RequestOpenHtlc(msgData string, user bean.User) (data map[string]interface{}, bob string, err error) {
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
		if item.PeerIdA == htlcCreateRandHInfo.SenderPeerId || item.PeerIdB == htlcCreateRandHInfo.RecipientPeerId {
			return nil, "", errors.New("has direct channel")
		}
		if item.PeerIdB == htlcCreateRandHInfo.SenderPeerId || item.PeerIdA == htlcCreateRandHInfo.RecipientPeerId {
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

	if FindUserIsOnline(bob) != nil {
		return nil, "", errors.New("inter node: " + bob + " not online")
	}

	log.Println(aliceChannel)
	aliceCommitmentTxInfo, err := getLatestCommitmentTx(aliceChannel.ChannelId, htlcCreateRandHInfo.SenderPeerId)
	if err != nil {
		return nil, "", err
	}
	amountAliceNeed := aliceCommitmentTxInfo.AmountToRSMC + tool.GetHtlcFee()
	if aliceCommitmentTxInfo.AmountToRSMC < amountAliceNeed {
		return nil, "", errors.New("sender node not enough money")
	}

	log.Println(bob)
	log.Println(carlChannel)
	bobCommitmentTxInfo, err := getLatestCommitmentTx(carlChannel.ChannelId, bob)
	if err != nil {
		return nil, "", err
	}
	if bobCommitmentTxInfo.AmountToRSMC < htlcCreateRandHInfo.Amount {
		return nil, "", errors.New("inter node not enough money")
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
			for _, carlChannel := range channelCarlInfos {
				if bobPeerId == carlChannel.PeerIdA || bobPeerId == carlChannel.PeerIdB {
					return bobPeerId, &aliceChannel, &carlChannel
				}
			}
		} else {
			bobPeerId := aliceChannel.PeerIdA
			for _, carlChannel := range channelCarlInfos {
				if bobPeerId == carlChannel.PeerIdA || bobPeerId == carlChannel.PeerIdB {
					return bobPeerId, &aliceChannel, &carlChannel
				}
			}
		}
	}
	return "", nil, nil
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

func (service *htlcTxManager) SignOpenHtlc(msgData string, user bean.User) (data interface{}, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, errors.New("empty json data")
	}

	htlcSignRequestCreate := &bean.HtlcSignRequestCreate{}
	err = json.Unmarshal([]byte(msgData), htlcSignRequestCreate)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	tx, err := db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	htlcSingleHopTxBaseInfo := &dao.HtlcSingleHopTxBaseInfo{}
	err = tx.Select(q.Eq("HtlcCreateRandHInfoRequestHash", htlcSignRequestCreate.RequestHash)).First(htlcSingleHopTxBaseInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	htlcSingleHopTxBaseInfo.CurrState = dao.NS_Refuse
	if htlcSignRequestCreate.Approval {
		htlcSingleHopTxBaseInfo.CurrState = dao.NS_Finish
		//锁定两个通道
		aliceChannel := &dao.ChannelInfo{}
		err := tx.One("Id", htlcSingleHopTxBaseInfo.FirstChannelId, aliceChannel)
		if err != nil {
			return nil, err
		}
		carlChannel := &dao.ChannelInfo{}
		err = tx.One("Id", htlcSingleHopTxBaseInfo.SecondChannelId, carlChannel)
		if err != nil {
			return nil, err
		}
	}
	htlcSingleHopTxBaseInfo.SignBy = user.PeerId
	htlcSingleHopTxBaseInfo.SignAt = time.Now()

	err = tx.Update(htlcSingleHopTxBaseInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return data, nil
}
