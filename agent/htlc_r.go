package agent

import (
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	conn2tracker "github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/service"
	"github.com/shopspring/decimal"
	"strings"
)

func ROwnerGetHtlcRFromLocal(dataTx interface{}, user *bean.User) (r string) {
	tx := dataTx.(*dao.CommitmentTransaction)
	hAndRImage := &dao.HtlcHAndRImage{}
	_ = user.Db.Select(q.Eq("H", tx.HtlcH)).First(hAndRImage)
	if hAndRImage.Id > 0 {
		return hAndRImage.R
	}
	return ""
}

func InterUserGetNextNode(dataTx interface{}, user *bean.User) (channelId string, amount float64, msg *bean.RequestMessage) {
	currNodeTx := dataTx.(*dao.CommitmentTransaction)
	currChannelId := currNodeTx.ChannelId
	htlcPathArr := strings.Split(currNodeTx.HtlcRoutingPacket, ",")
	totalStep := len(htlcPathArr)
	currStep := 0
	for ; currStep < len(htlcPathArr); currStep++ {
		if htlcPathArr[currStep] == currChannelId {
			break
		}
	}
	currStep++
	if currStep < len(htlcPathArr) {
		channelId = htlcPathArr[currStep]
		latestCommitmentTx := getLatestCommitmentTx(channelId, *user)
		msg = &bean.RequestMessage{}
		msg.RecipientUserPeerId = latestCommitmentTx.PeerIdA
		if user.PeerId == latestCommitmentTx.PeerIdA {
			msg.RecipientUserPeerId = latestCommitmentTx.PeerIdB
		}
		msg.RecipientNodePeerId = conn2tracker.GetUserP2pNodeId(msg.RecipientUserPeerId)
		amount, _ = decimal.NewFromFloat(currNodeTx.HtlcAmountToPayee).Mul(decimal.NewFromFloat(1 + config.HtlcFeeRate*float64(totalStep-currStep-1))).Round(8).Float64()
		if len(msg.RecipientNodePeerId) == 0 {
			return "", 0, nil
		}
		msg.SenderNodePeerId = user.P2PLocalPeerId
		msg.SenderUserPeerId = user.PeerId
		return channelId, amount, msg
	}
	return "", 0, nil
}

func InterUserGetHtlcRFromLocalForPreNode(dataTx interface{}, user *bean.User) (r, channelId string, msg *bean.RequestMessage) {
	currNodeTx := dataTx.(*dao.CommitmentTransaction)
	currChannelId := currNodeTx.ChannelId
	htlcPathArr := strings.Split(currNodeTx.HtlcRoutingPacket, ",")
	index := len(htlcPathArr) - 1
	for ; index >= 0; index-- {
		if htlcPathArr[index] == currChannelId {
			break
		}
	}
	if index > 0 {
		newNodeTx := &dao.CommitmentTransaction{}
		_ = user.Db.Select(q.Eq("HtlcH", currNodeTx.HtlcH), q.Eq("ChannelId", htlcPathArr[index-1])).First(newNodeTx)
		if newNodeTx.Id > 0 {
			msg = &bean.RequestMessage{}
			msg.RecipientUserPeerId = newNodeTx.PeerIdA
			if user.PeerId == newNodeTx.PeerIdA {
				msg.RecipientUserPeerId = newNodeTx.PeerIdB
			}
			msg.RecipientNodePeerId = conn2tracker.GetUserP2pNodeId(msg.RecipientUserPeerId)
			if len(msg.RecipientNodePeerId) == 0 {
				return "", "", nil
			}
			return currNodeTx.HtlcR, newNodeTx.ChannelId, msg
		}
	}
	return "", "", nil
}

func HtlcBobSignedHeRdAtBobSide(dataTx interface{}, user *bean.User) (signedData *bean.BobSignHerdForC3b, err error) {
	needSignData := dataTx.(bean.HtlcBobSendRResult)
	signedData = &bean.BobSignHerdForC3b{ChannelId: needSignData.ChannelId}

	currCommitmentTx := getCurrCommitmentTx(needSignData.ChannelId, *user)
	htTx := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	_ = user.Db.Select(q.Eq("ChannelId", needSignData.ChannelId), q.Eq("CommitmentTxId", currCommitmentTx.Id)).First(htTx)
	if htTx.Id > 0 {
		wallet, _ := service.HDWalletService.GetAddressByIndex(user, uint32(htTx.RSMCTempAddressIndex))
		_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcHerdRawData.Hex, convertBean(needSignData.C3bHtlcHerdRawData.Inputs), wallet.Wif)
		if err != nil {
			return signedData, err
		}
		signedData.C3bHtlcHerdPartialSignedHex = hex
	}
	return signedData, nil
}

func HtlcAliceSignedHeRdAtAliceSide(dataTx interface{}, user *bean.User) (signedData *bean.AliceSignHerdTxOfC3e, err error) {
	needSignData := dataTx.(bean.NeedAliceSignHerdTxOfC3bP2p)

	signedData = &bean.AliceSignHerdTxOfC3e{ChannelId: needSignData.ChannelId}
	channelWalletInfo, err := getChannelWalletInfo(needSignData.ChannelId, user)
	if err != nil {
		return nil, err
	}

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcHerdPartialSignedData.Hex, convertBean(needSignData.C3bHtlcHerdPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bHtlcHerdCompleteSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcHebrRawData.Hex, convertBean(needSignData.C3bHtlcHebrRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bHtlcHebrPartialSignedHex = hex

	return signedData, nil
}
