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
)

type htlcTxManager struct {
	operationFlag sync.Mutex
}

var HtlcTxService htlcTxManager

// query bob,and ask bob
func (service *htlcTxManager) RequestOpenHtlc(msgData string, user bean.User) (data interface{}, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, errors.New("empty json data")
	}

	reqData := &bean.HtlcRequestCreate{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	htlcCreateRandHInfo := &dao.HtlcCreateRandHInfo{}
	err = db.Select(q.Eq("CreateBy", user.PeerId), q.Eq("CurrState", dao.NS_Finish), q.Eq("H", reqData.H)).First(htlcCreateRandHInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	channelAliceInfos := getAllChannels(htlcCreateRandHInfo.SenderPeerId)
	if len(channelAliceInfos) == 0 {
		return nil, errors.New("sender's channel not found")
	}
	//if has the channel direct
	for _, item := range channelAliceInfos {
		if item.PeerIdA == htlcCreateRandHInfo.SenderPeerId || item.PeerIdB == htlcCreateRandHInfo.RecipientPeerId {
			return nil, errors.New("has direct channel")
		}
		if item.PeerIdB == htlcCreateRandHInfo.SenderPeerId || item.PeerIdA == htlcCreateRandHInfo.RecipientPeerId {
			return nil, errors.New("has direct channel")
		}
	}

	channelCarlInfos := getAllChannels(htlcCreateRandHInfo.RecipientPeerId)
	if len(channelCarlInfos) == 0 {
		return nil, errors.New("recipient's channel not found")
	}

	bob, aliceChannel, carlChannel := getTwoChannelOfSingleHop(*htlcCreateRandHInfo, channelAliceInfos, channelCarlInfos)
	if tool.CheckIsString(&bob) == false {
		return nil, errors.New("no inter channel can use")
	}

	log.Println(bob)
	log.Println(aliceChannel)
	log.Println(carlChannel)

	return data, nil
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
	return data, nil
}
