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

	walletInfo, err := getChannelWalletInfoByAddress(sendInfo.ToAddress, user)
	if err != nil {
		return nil, err
	}

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

func AliceSignFundAsset(msg bean.RequestMessage, hexData map[string]interface{}, user *bean.User) (signedData map[string]interface{}, err error) {
	if user == nil || user.IsAdmin == false {
		return hexData, nil
	}

	sendInfo := &bean.FundingAsset{}
	_ = json.Unmarshal([]byte(msg.Data), sendInfo)
	walletInfo, err := getChannelWalletInfoByAddress(sendInfo.ToAddress, user)
	if err != nil {
		return nil, err
	}
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

func AliceFirstSignFundBtcRedeemTx(hexData interface{}, user *bean.User) (signedData *bean.NeedClientSignHexData, err error) {
	signData := hexData.(bean.NeedClientSignHexData)
	channelInfo := dao.ChannelInfo{}
	user.Db.Select(q.Eq("TemporaryChannelId", signData.TemporaryChannelId), q.Eq("CurrState", bean.ChannelState_WaitFundAsset)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New("not found the channel")
	}

	walletInfo, err := service.HDWalletService.GetAddressByIndex(user, uint32(channelInfo.FunderAddressIndex))
	if err != nil {
		return nil, err
	}

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

	walletInfo, err := getChannelWalletInfoByTemplateId(fundingBtcOfP2p.TemporaryChannelId, user)
	if err != nil {
		return nil, err
	}

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(fundingBtcOfP2p.SignData.Hex, convertBean(fundingBtcOfP2p.SignData.Inputs), walletInfo.Wif)
	if err != nil {
		return nil, err
	}

	resultData.SignedMinerRedeemTransactionHex = hex
	resultData.FundingTxid = fundingBtcOfP2p.FundingTxid
	resultData.TemporaryChannelId = fundingBtcOfP2p.TemporaryChannelId
	resultData.Approval = true
	return resultData, nil
}

func AliceCreateTempWalletForC1a(msg *bean.RequestMessage, user *bean.User) (err error) {
	reqData := &bean.SendRequestAssetFunding{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return err
	}
	reqData.TempAddressIndex = address.Index
	reqData.TempAddressPubKey = address.PubKey
	marshal, _ := json.Marshal(reqData)
	msg.Data = string(marshal)
	return nil
}

func AliceSignC1a(hexData interface{}, user *bean.User) (signedData *bean.AliceSignC1aOfAssetFunding, err error) {
	signData := hexData.(bean.NeedClientSignHexData)

	walletInfo, err := getChannelWalletInfoByTemplateId(signData.TemporaryChannelId, user)
	if err != nil {
		return nil, err
	}
	_, signedHex, err := omnicore.OmniSignRawTransactionForUnsend(signData.Hex, convertBean(signData.Inputs), walletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData = &bean.AliceSignC1aOfAssetFunding{}
	signedData.SignedC1aHex = signedHex
	return signedData, nil
}

func BobSignC1a(data string, user *bean.User) (resultData *bean.SignAssetFunding, err error) {
	resultData = &bean.SignAssetFunding{}

	p2pData := bean.FundingAssetOfP2p{}
	_ = json.Unmarshal([]byte(data), &p2pData)

	resultData.TemporaryChannelId = p2pData.TemporaryChannelId

	channelAddressInfo, err := getChannelWalletInfoByTemplateId(p2pData.TemporaryChannelId, user)
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

	channelAddressInfo, err := getChannelWalletInfoByTemplateId(resultData.TemporaryChannelId, user)
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

func AliceSignRdOfC1a(hexData map[string]interface{}, user *bean.User) (signedData *bean.AliceSignRDOfAssetFunding, err error) {

	channelId := hexData["channel_id"].(string)
	commitmentTxInfo := &dao.CommitmentTransaction{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(commitmentTxInfo)
	walletInfo, _ := service.HDWalletService.GetAddressByIndex(user, uint32(commitmentTxInfo.RSMCTempAddressIndex))

	rdHex := hexData["hex"].(string)
	rdInputs := hexData["inputs"]
	_, signedHex, err := omnicore.OmniSignRawTransactionForUnsend(rdHex, convertBean(rdInputs), walletInfo.Wif)
	if err != nil {
		return nil, err
	}
	hexData["Hex"] = signedHex
	signedData = &bean.AliceSignRDOfAssetFunding{}
	signedData.ChannelId = channelId
	signedData.RdSignedHex = signedHex
	return signedData, nil
}
