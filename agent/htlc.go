package agent

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/service"
	"strings"
)

func BeforeAliceAddHtlcAtAliceSide(msg *bean.RequestMessage, user *bean.User) (err error) {
	requestData := &bean.CreateHtlcTxForC3a{}
	err = json.Unmarshal([]byte(msg.Data), requestData)
	if err != nil {
		return err
	}

	channelIds := strings.Split(requestData.RoutingPacket, ",")
	var channelInfo *dao.ChannelInfo
	for _, channelId := range channelIds {
		temp := getLockChannelForHtlc(channelId, user)
		if temp != nil {
			if temp.PeerIdA == msg.RecipientUserPeerId || temp.PeerIdB == msg.RecipientUserPeerId {
				channelInfo = temp
				break
			}
		}
	}
	if channelInfo == nil {
		return errors.New("not found the channel")
	}

	commitmentTransaction := &dao.CommitmentTransaction{}
	user.Db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(commitmentTransaction)
	if commitmentTransaction.Id > 0 {
		latestTempAddressInfo, err := service.HDWalletService.GetAddressByIndex(user, uint32(commitmentTransaction.RSMCTempAddressIndex))
		if err != nil {
			return err
		}
		requestData.LastTempAddressPrivateKey = latestTempAddressInfo.Wif
	}

	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return err
	}
	requestData.CurrRsmcTempAddressIndex = address.Index
	requestData.CurrRsmcTempAddressPubKey = address.PubKey

	address, err = service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return err
	}
	requestData.CurrHtlcTempAddressIndex = address.Index
	requestData.CurrHtlcTempAddressPubKey = address.PubKey

	address, err = service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return err
	}
	requestData.CurrHtlcTempAddressForHt1aIndex = address.Index
	requestData.CurrHtlcTempAddressForHt1aPubKey = address.PubKey
	marshal, _ := json.Marshal(requestData)
	msg.Data = string(marshal)

	return nil
}

func AliceSignC3aAtAliceSide(data interface{}, user *bean.User) (signedData *bean.AliceSignedHtlcDataForC3a, err error) {
	txForC3a := data.(bean.NeedAliceSignCreateHtlcTxForC3a)
	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", txForC3a.ChannelId)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found the channel")
	}

	addressIndex := channelInfo.FunderAddressIndex
	if channelInfo.PeerIdB == user.PeerId {
		addressIndex = channelInfo.FundeeAddressIndex
	}

	channelAddressInfo, err := service.HDWalletService.GetAddressByIndex(user, uint32(addressIndex))
	if err != nil {
		return nil, err
	}

	signedData = &bean.AliceSignedHtlcDataForC3a{}
	signedData.ChannelId = channelInfo.ChannelId

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(txForC3a.C3aRsmcRawData.Hex, convertBean(txForC3a.C3aRsmcRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3aRsmcPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(txForC3a.C3aCounterpartyRawData.Hex, convertBean(txForC3a.C3aCounterpartyRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3aCounterpartyPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(txForC3a.C3aHtlcRawData.Hex, convertBean(txForC3a.C3aHtlcRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3aHtlcPartialSignedHex = hex

	return signedData, nil
}

func BobSignAddHtlcRequestAtBobSide_40(data bean.CreateHtlcTxForC3aToBob, user *bean.User) (resultData bean.BobSignedC3a, err error) {
	resultData = bean.BobSignedC3a{PayerCommitmentTxHash: data.PayerCommitmentTxHash}

	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return resultData, err
	}
	resultData.CurrRsmcTempAddressIndex = address.Index
	resultData.CurrRsmcTempAddressPubKey = address.PubKey

	address, err = service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return resultData, err
	}
	resultData.CurrHtlcTempAddressIndex = address.Index
	resultData.CurrHtlcTempAddressPubKey = address.PubKey

	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", data.ChannelId)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return resultData, errors.New("not found the channel")
	}

	addressIndex := channelInfo.FunderAddressIndex
	if channelInfo.PeerIdB == user.PeerId {
		addressIndex = channelInfo.FundeeAddressIndex
	}
	commitmentTransaction := &dao.CommitmentTransaction{}
	user.Db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(commitmentTransaction)
	if commitmentTransaction.Id > 0 {
		latestTempAddressInfo, err := service.HDWalletService.GetAddressByIndex(user, uint32(commitmentTransaction.RSMCTempAddressIndex))
		if err != nil {
			return resultData, err
		}
		resultData.LastTempAddressPrivateKey = latestTempAddressInfo.Wif
	}

	channelAddressInfo, err := service.HDWalletService.GetAddressByIndex(user, uint32(addressIndex))
	if err != nil {
		return resultData, err
	}

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(data.C3aRsmcPartialSignedData.Hex, convertBean(data.C3aRsmcPartialSignedData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return resultData, err
	}
	resultData.C3aCompleteSignedRsmcHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(data.C3aCounterpartyPartialSignedData.Hex, convertBean(data.C3aCounterpartyPartialSignedData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return resultData, err
	}
	resultData.C3aCompleteSignedCounterpartyHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(data.C3aHtlcPartialSignedData.Hex, convertBean(data.C3aHtlcPartialSignedData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return resultData, err
	}
	resultData.C3aCompleteSignedHtlcHex = hex

	return resultData, nil
}

func BobSignC3b() {

}

func convertBean(inputs interface{}) (result []bean.TransactionInputItem) {
	result = make([]bean.TransactionInputItem, 0)
	for _, item := range inputs.([]map[string]interface{}) {
		temp := bean.TransactionInputItem{
			ScriptPubKey: item["scriptPubKey"].(string),
			RedeemScript: item["redeemScript"].(string),
		}
		result = append(result, temp)
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
