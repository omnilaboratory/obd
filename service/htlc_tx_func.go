package service

import (
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"errors"
	"github.com/asdine/storm/q"
)

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
	if err := FindUserIsOnline(bobPeerId); err != nil {
		return nil, err
	}

	commitmentTxInfo, err := getLatestCommitmentTx(aliceChannel.ChannelId, alicePeerId)
	if err != nil {
		return nil, err
	}
	if commitmentTxInfo.AmountToRSMC < (htlcCreateRandHInfo.Amount + tool.GetHtlcFee()) {
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
	_ = db.Select(q.Eq("PeerIdB", peerId)).Find(&channelBInfos)

	channelInfos = []dao.ChannelInfo{}
	if channelAInfos != nil && len(channelAInfos) > 0 {
		channelInfos = append(channelInfos, channelAInfos...)
	}
	if channelBInfos != nil && len(channelBInfos) > 0 {
		channelInfos = append(channelInfos, channelBInfos...)
	}
	return channelInfos
}
