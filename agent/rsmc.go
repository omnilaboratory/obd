package agent

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/service"
)

func RsmcAliceCreateTx(msg *bean.RequestMessage, user *bean.User) (err error) {
	reqData := &bean.RequestCreateCommitmentTx{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		return err
	}
	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", reqData.ChannelId)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return errors.New("not found the channel")
	}

	latestCommitmentTx := getLatestCommitmentTx(channelInfo.ChannelId, *user)
	if latestCommitmentTx.Id > 0 {
		lastWalletInfo, err := service.HDWalletService.GetAddressByIndex(user, uint32(latestCommitmentTx.RSMCTempAddressIndex))
		if err != nil {
			return err
		}
		reqData.LastTempAddressPrivateKey = lastWalletInfo.Wif
	}
	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return err
	}
	reqData.CurrTempAddressIndex = address.Index
	reqData.CurrTempAddressPubKey = address.PubKey
	marshal, _ := json.Marshal(reqData)
	msg.Data = string(marshal)
	return nil
}

func RsmcAliceFirstSignC2a(signData interface{}, user *bean.User) (signedDataForC2a *bean.AliceSignedRsmcDataForC2a, err error) {
	needSignData := signData.(bean.NeedAliceSignRsmcDataForC2a)
	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", needSignData.ChannelId)).First(&channelInfo)
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

	signedDataForC2a = &bean.AliceSignedRsmcDataForC2a{}
	signedDataForC2a.ChannelId = channelInfo.ChannelId
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.RsmcRawData.Hex, convertBean(needSignData.RsmcRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedDataForC2a.RsmcSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.CounterpartyRawData.Hex, convertBean(needSignData.CounterpartyRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedDataForC2a.CounterpartySignedHex = hex
	return signedDataForC2a, nil
}

func RsmcBobSecondSignC2a(needSignData *bean.PayerRequestCommitmentTxToBobClient, user *bean.User) (signedData *bean.PayeeSendSignCommitmentTx, err error) {

	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", needSignData.ChannelId)).First(&channelInfo)
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
	signedData = &bean.PayeeSendSignCommitmentTx{}
	signedData.ChannelId = channelInfo.ChannelId
	signedData.MsgHash = needSignData.MsgHash
	signedData.Approval = true

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.RsmcRawData.Hex, convertBean(needSignData.RsmcRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C2aRsmcSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.CounterpartyRawData.Hex, convertBean(needSignData.CounterpartyRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C2aCounterpartySignedHex = hex

	latestCommitmentTx := getLatestCommitmentTx(channelInfo.ChannelId, *user)
	if latestCommitmentTx.Id > 0 {
		latestRsmcAddressInfo, _ := service.HDWalletService.GetAddressByIndex(user, uint32(latestCommitmentTx.RSMCTempAddressIndex))
		signedData.LastTempAddressPrivateKey = latestRsmcAddressInfo.Wif
	}

	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return nil, err
	}
	signedData.CurrTempAddressIndex = address.Index
	signedData.CurrTempAddressPubKey = address.PubKey

	return signedData, nil
}

func RsmcBobFirstSignC2b(c2bData interface{}, user *bean.User) (signedDataForC2b *bean.BobSignedRsmcDataForC2b, err error) {

	needSignData := c2bData.(bean.NeedBobSignRawDataForC2b)

	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", needSignData.ChannelId)).First(&channelInfo)
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

	signedDataForC2b = &bean.BobSignedRsmcDataForC2b{}
	signedDataForC2b.ChannelId = needSignData.ChannelId

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C2aRdRawData.Hex, convertBean(needSignData.C2aRdRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedDataForC2b.C2aRdSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C2aBrRawData.Hex, convertBean(needSignData.C2aBrRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedDataForC2b.C2aBrSignedHex = hex
	signedDataForC2b.C2aBrId = int64(needSignData.C2aBrRawData.BrId)

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C2bRsmcRawData.Hex, convertBean(needSignData.C2bRsmcRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedDataForC2b.C2bRsmcSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C2bCounterpartyRawData.Hex, convertBean(needSignData.C2bCounterpartyRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedDataForC2b.C2bCounterpartySignedHex = hex

	return signedDataForC2b, nil
}

func RsmcAliceSignC2b(data interface{}, user *bean.User) (signedData *bean.AliceSignedRsmcTxForC2b, err error) {

	needSignData := data.(bean.NeedAliceSignRsmcTxForC2b)

	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", needSignData.ChannelId)).First(&channelInfo)
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

	signedData = &bean.AliceSignedRsmcTxForC2b{}
	signedData.ChannelId = needSignData.ChannelId

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C2bRsmcPartialData.Hex, convertBean(needSignData.C2bRsmcPartialData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C2bRsmcSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C2bCounterpartyPartialData.Hex, convertBean(needSignData.C2bCounterpartyPartialData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C2bCounterpartySignedHex = hex

	commitmentTransaction := getCurrCommitmentTx(channelInfo.ChannelId, *user)
	tempAddressInfo, _ := service.HDWalletService.GetAddressByIndex(user, uint32(commitmentTransaction.RSMCTempAddressIndex))
	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C2aRdPartialData.Hex, convertBean(needSignData.C2aRdPartialData.Inputs), tempAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C2aRdSignedHex = hex

	return signedData, nil
}

func RsmcAliceSignRdOfC2b(data interface{}, user *bean.User) (signedData *bean.AliceSignedRdTxForC2b, err error) {

	needSignData := data.(bean.NeedAliceSignRdTxForC2b)

	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", needSignData.ChannelId)).First(&channelInfo)
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

	signedData = &bean.AliceSignedRdTxForC2b{}
	signedData.ChannelId = needSignData.ChannelId

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C2bRdRawData.Hex, convertBean(needSignData.C2bRdRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C2bRdSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C2bBrRawData.Hex, convertBean(needSignData.C2bBrRawData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C2bBrSignedHex = hex
	signedData.C2bBrId = int64(needSignData.C2bBrRawData.BrId)

	return signedData, nil
}

func RsmcBobSignRdOfC2b(needSignData bean.NeedBobSignRdTxForC2b, user *bean.User) (signedData *bean.BobSignedRdTxForC2b, err error) {
	signedData = &bean.BobSignedRdTxForC2b{}
	signedData.ChannelId = needSignData.ChannelId
	commitmentTransaction := getCurrCommitmentTx(signedData.ChannelId, *user)
	wallet, _ := service.HDWalletService.GetAddressByIndex(user, uint32(commitmentTransaction.RSMCTempAddressIndex))

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C2bRdPartialData.Hex, convertBean(needSignData.C2bRdPartialData.Inputs), wallet.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C2bRdSignedHex = hex

	return signedData, nil
}
