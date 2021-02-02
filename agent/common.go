package agent

import (
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/service"
	"log"
)

func getLatestCommitmentTx(channelId string, user bean.User) *dao.CommitmentTransaction {
	commitmentTransaction := &dao.CommitmentTransaction{}
	user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
		OrderBy("Id").Reverse().
		First(commitmentTransaction)
	return commitmentTransaction
}

func getCurrCommitmentTx(channelId string, user bean.User) *dao.CommitmentTransaction {
	commitmentTransaction := &dao.CommitmentTransaction{}
	err := user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("CurrState", dao.TxInfoState_Create),
			q.Eq("CurrState", dao.TxInfoState_Htlc_GetH),
			q.Eq("CurrState", dao.TxInfoState_Htlc_GetR),
			q.Eq("CurrState", dao.TxInfoState_Init))).
		OrderBy("Id").Reverse().
		First(commitmentTransaction)
	if err != nil {
		log.Println(err)
		return nil
	}
	return commitmentTransaction
}

func convertBean(inputs interface{}) (result []bean.TransactionInputItem) {
	result = make([]bean.TransactionInputItem, 0)
	if inputs == nil {
		return result
	}

	if arr, ok := inputs.([]map[string]interface{}); ok {
		for _, item := range arr {
			temp := bean.TransactionInputItem{
				ScriptPubKey: item["scriptPubKey"].(string),
			}
			if item["redeemScript"] != nil {
				temp.RedeemScript = item["redeemScript"].(string)
			}
			result = append(result, temp)
		}
	} else {
		nodes := inputs.([]interface{})
		for _, item := range nodes {
			node := item.(map[string]interface{})
			temp := bean.TransactionInputItem{
				ScriptPubKey: node["scriptPubKey"].(string),
				RedeemScript: node["redeemScript"].(string),
			}
			result = append(result, temp)
		}
	}
	return result
}

func getLockChannelForHtlc(channelId string, user *bean.User) (channelInfo *dao.ChannelInfo) {
	channelInfo = &dao.ChannelInfo{}
	_ = user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId)),
		q.Or(
			q.Eq("CurrState", bean.ChannelState_CanUse),
			q.Eq("CurrState", bean.ChannelState_LockByTracker))).
		First(channelInfo)
	return channelInfo
}

func getChannelWalletInfoByTemplateId(temporaryChannelId string, user *bean.User) (walletInfo *service.Wallet, err error) {
	channelInfo := dao.ChannelInfo{}
	user.Db.Select(q.Eq("TemporaryChannelId", temporaryChannelId), q.Eq("CurrState", bean.ChannelState_WaitFundAsset)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found the channel")
	}
	addressIndex := channelInfo.FunderAddressIndex
	if channelInfo.PeerIdB == user.PeerId {
		addressIndex = channelInfo.FundeeAddressIndex
	}
	return service.HDWalletService.GetAddressByIndex(user, uint32(addressIndex))
}

func getChannelWalletInfoByAddress(address string, user *bean.User) (walletInfo *service.Wallet, err error) {
	channelInfo := dao.ChannelInfo{}
	user.Db.Select(q.Eq("ChannelAddress", address), q.Eq("CurrState", bean.ChannelState_WaitFundAsset)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found the channel")
	}
	return service.HDWalletService.GetAddressByIndex(user, uint32(channelInfo.FunderAddressIndex))
}

func getChannelWalletInfo(channelId string, user *bean.User) (walletInfo *service.Wallet, err error) {
	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", channelId)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found the channel")
	}

	addressIndex := channelInfo.FunderAddressIndex
	if channelInfo.PeerIdB == user.PeerId {
		addressIndex = channelInfo.FundeeAddressIndex
	}
	return service.HDWalletService.GetAddressByIndex(user, uint32(addressIndex))
}
