package agent

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/service"
	"log"
)

func AliceSignFundBtc(msg bean.RequestMessage, hexData map[string]interface{}, user *bean.User) (signedData map[string]interface{}, err error) {
	if user == nil || user.IsAdmin == false {
		return hexData, nil
	}

	sendInfo := &bean.FundingBtc{}
	_ = json.Unmarshal([]byte(msg.Data), sendInfo)
	channelInfo := dao.ChannelInfo{}
	user.Db.Select(q.Eq("ChannelAddress", sendInfo.ToAddress), q.Eq("CurrState", bean.ChannelState_WaitFundAsset)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found the channel")
	}
	walletInfo, _ := service.HDWalletService.GetAddressByIndex(user, uint32(channelInfo.FunderAddressIndex))
	hex := hexData["hex"].(string)
	inputs := hexData["inputs"]
	_, signedHex, err := omnicore.OmniSignRawTransactionForUnsend(hex, convertBean(inputs), walletInfo.Wif)
	if err != nil {
		return nil, err
	}

	signedData = hexData
	signedData["hex"] = signedHex

	return signedData, nil
}

func AliceFirstSignFundBtcRedeemTx(msg bean.RequestMessage, hexData interface{}, user *bean.User) (signedData *bean.NeedClientSignHexData, err error) {
	signData := hexData.(bean.NeedClientSignHexData)
	channelInfo := dao.ChannelInfo{}
	user.Db.Select(q.Eq("TemporaryChannelId", signData.TemporaryChannelId), q.Eq("CurrState", bean.ChannelState_WaitFundAsset)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found the channel")
	}

	walletInfo, _ := service.HDWalletService.GetAddressByIndex(user, uint32(channelInfo.FunderAddressIndex))
	_, signedHex, err := omnicore.OmniSignRawTransactionForUnsend(signData.Hex, convertBean(signData.Inputs), walletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signData.Hex = signedHex

	return &signData, nil
}

func BobSignFundBtcRedeemTx(data string, user *bean.User) (resultData *bean.SendSignFundingBtc, err error) {
	resultData = &bean.SendSignFundingBtc{}

	fundingBtcOfP2p := bean.FundingBtcOfP2p{}
	err = json.Unmarshal([]byte(data), &fundingBtcOfP2p)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("TemporaryChannelId", fundingBtcOfP2p.TemporaryChannelId)).First(&channelInfo)
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

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(fundingBtcOfP2p.SignData.Hex, convertBean(fundingBtcOfP2p.SignData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}

	resultData.SignedMinerRedeemTransactionHex = hex
	resultData.FundingTxid = fundingBtcOfP2p.FundingTxid
	resultData.TemporaryChannelId = fundingBtcOfP2p.TemporaryChannelId
	resultData.Approval = true
	return resultData, nil
}

func BobSignC1a(data string, user *bean.User) (resultData *bean.SignAssetFunding, err error) {
	resultData = &bean.SignAssetFunding{}

	p2pData := bean.FundingAssetOfP2p{}
	_ = json.Unmarshal([]byte(data), &p2pData)

	resultData.TemporaryChannelId = p2pData.TemporaryChannelId

	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("TemporaryChannelId", p2pData.TemporaryChannelId)).First(&channelInfo)
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

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(p2pData.SignData.Hex, convertBean(p2pData.SignData.Inputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	resultData.SignedAliceRsmcHex = hex

	return resultData, nil
}

func BobSignRdAndBrOfC1a(data string, user *bean.User) (resultData *bean.SignRdAndBrOfAssetFunding, err error) {
	resultData = &bean.SignRdAndBrOfAssetFunding{}

	reqData := &bean.NeedSignRdAndBrOfAssetFunding{}
	err = json.Unmarshal([]byte(data), reqData)

	resultData.TemporaryChannelId = reqData.TemporaryChannelId

	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("TemporaryChannelId", resultData.TemporaryChannelId)).First(&channelInfo)
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

	aliceRdSignData := reqData.AliceRdSignData
	rdHex := aliceRdSignData["hex"].(string)
	rdInputs := aliceRdSignData["inputs"]
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(rdHex, convertBean(rdInputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	resultData.RdSignedHex = hex

	aliceBrSignData := reqData.AliceBrSignData
	brHex := aliceBrSignData["hex"].(string)
	brInputs := aliceBrSignData["inputs"]
	resultData.BrId = int64(aliceBrSignData["br_id"].(float64))
	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(brHex, convertBean(brInputs), channelAddressInfo.Wif)
	if err != nil {
		return nil, err
	}
	resultData.BrSignedHex = hex
	return resultData, nil
}
