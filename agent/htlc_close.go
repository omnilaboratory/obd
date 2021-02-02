package agent

import (
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/service"
)

func HtlcRequestCloseHtlc(c3xData interface{}, user *bean.User) (reqData *bean.HtlcCloseRequestCurrTx, err error) {
	c3x := c3xData.(*dao.CommitmentTransaction)

	reqData = &bean.HtlcCloseRequestCurrTx{}
	reqData.ChannelId = c3x.ChannelId

	walletInfo, _ := service.HDWalletService.GetAddressByIndex(user, uint32(c3x.RSMCTempAddressIndex))
	reqData.LastRsmcTempAddressPrivateKey = walletInfo.Wif

	walletInfo, _ = service.HDWalletService.GetAddressByIndex(user, uint32(c3x.HTLCTempAddressIndex))
	reqData.LastHtlcTempAddressPrivateKey = walletInfo.Wif

	htTx := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	user.Db.Select(q.Eq("ChannelId", c3x.ChannelId), q.Eq("CommitmentTxId", c3x.Id)).First(htTx)
	if htTx.Id > 0 {
		walletInfo, _ = service.HDWalletService.GetAddressByIndex(user, uint32(htTx.RSMCTempAddressIndex))
		reqData.LastHtlcTempAddressForHtnxPrivateKey = walletInfo.Wif
	}

	walletInfo, _ = service.HDWalletService.CreateNewAddress(user)
	reqData.CurrTempAddressIndex = walletInfo.Index
	reqData.CurrTempAddressPubKey = walletInfo.PubKey
	return reqData, nil
}

func HtlcCloseAliceSignedCxa(hexData interface{}, user *bean.User) (signedData *bean.AliceSignedRsmcDataForC4a, err error) {
	needSignData := hexData.(bean.NeedAliceSignRsmcDataForC4a)
	signedData = &bean.AliceSignedRsmcDataForC4a{ChannelId: needSignData.ChannelId}

	channelWalletInfo, _ := getChannelWalletInfo(needSignData.ChannelId, user)
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C4aRsmcRawData.Hex, convertBean(needSignData.C4aRsmcRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.RsmcPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C4aCounterpartyRawData.Hex, convertBean(needSignData.C4aCounterpartyRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.CounterpartyPartialSignedHex = hex

	return signedData, nil
}

func HtlcCloseBobSignCloseHtlcRequest(hexData interface{}, user *bean.User) (signedData *bean.HtlcBobSignCloseCurrTx, err error) {
	needSignData := hexData.(*bean.AliceRequestCloseHtlcCurrTxOfP2pToBobClient)
	signedData = &bean.HtlcBobSignCloseCurrTx{MsgHash: needSignData.MsgHash}

	latestCommitmentTx := getCurrCommitmentTx(needSignData.ChannelId, *user)
	htTx := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	user.Db.Select(q.Eq("ChannelId", needSignData.ChannelId), q.Eq("CommitmentTxId", latestCommitmentTx.Id)).First(htTx)
	if htTx.Id == 0 {
		return nil, errors.New("not found the htTx")
	}
	wallet, _ := service.HDWalletService.GetAddressByIndex(user, uint32(latestCommitmentTx.RSMCTempAddressIndex))
	signedData.LastRsmcTempAddressPrivateKey = wallet.Wif
	wallet, _ = service.HDWalletService.GetAddressByIndex(user, uint32(latestCommitmentTx.HTLCTempAddressIndex))
	signedData.LastHtlcTempAddressPrivateKey = wallet.Wif
	wallet, _ = service.HDWalletService.GetAddressByIndex(user, uint32(htTx.RSMCTempAddressIndex))
	signedData.LastHtlcTempAddressForHtnxPrivateKey = wallet.Wif

	wallet, _ = service.HDWalletService.CreateNewAddress(user)
	signedData.CurrTempAddressIndex = wallet.Index
	signedData.CurrTempAddressPubKey = wallet.PubKey

	channelWalletInfo, _ := getChannelWalletInfo(needSignData.ChannelId, user)
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C4aRsmcPartialSignedData.Hex, convertBean(needSignData.C4aRsmcPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4aRsmcCompleteSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C4aCounterpartyPartialSignedData.Hex, convertBean(needSignData.C4aCounterpartyPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4aCounterpartyCompleteSignedHex = hex

	return signedData, nil
}

func HtlcCloseBobSignedCxb(hexData interface{}, user *bean.User) (signedData *bean.BobSignedRsmcDataForC4b, err error) {
	needSignData := hexData.(bean.NeedBobSignRawDataForC4b)
	signedData = &bean.BobSignedRsmcDataForC4b{ChannelId: needSignData.ChannelId}
	signedData.C4aBrId = int64(needSignData.C4aBrRawData.BrId)

	channelWalletInfo, _ := getChannelWalletInfo(needSignData.ChannelId, user)
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C4bRsmcRawData.Hex, convertBean(needSignData.C4bRsmcRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4bRsmcPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C4bCounterpartyRawData.Hex, convertBean(needSignData.C4bCounterpartyRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4bCounterpartyPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C4aRdRawData.Hex, convertBean(needSignData.C4aRdRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4aRdPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C4aBrRawData.Hex, convertBean(needSignData.C4aBrRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4aBrPartialSignedHex = hex

	return signedData, nil
}

func HtlcCloseAliceSignedCxb(hexData interface{}, user *bean.User) (signedData *bean.AliceSignedRsmcTxForC4b, err error) {
	needSignData := hexData.(bean.NeedAliceSignRsmcTxForC4b)
	signedData = &bean.AliceSignedRsmcTxForC4b{ChannelId: needSignData.ChannelId}

	currCommitmentTx := getCurrCommitmentTx(needSignData.ChannelId, *user)
	wallet, _ := service.HDWalletService.GetAddressByIndex(user, uint32(currCommitmentTx.RSMCTempAddressIndex))
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C4aRdPartialSignedData.Hex, convertBean(needSignData.C4aRdPartialSignedData.Inputs), wallet.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4aRdCompleteSignedHex = hex

	channelWalletInfo, _ := getChannelWalletInfo(needSignData.ChannelId, user)
	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C4bRsmcPartialSignedData.Hex, convertBean(needSignData.C4bRsmcPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4bRsmcCompleteSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C4bCounterpartyPartialSignedData.Hex, convertBean(needSignData.C4bCounterpartyPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4bCounterpartyCompleteSignedHex = hex

	return signedData, nil
}

func HtlcCloseAliceSignedCxbBubTx(hexData interface{}, user *bean.User) (signedData *bean.AliceSignedRdTxForC4b, err error) {
	needSignData := hexData.(bean.NeedAliceSignRdTxForC4b)
	signedData = &bean.AliceSignedRdTxForC4b{ChannelId: needSignData.ChannelId}

	channelWalletInfo, _ := getChannelWalletInfo(needSignData.ChannelId, user)
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C4bRdRawData.Hex, convertBean(needSignData.C4bRdRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4bRdPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C4bBrRawData.Hex, convertBean(needSignData.C4bBrRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4bBrPartialSignedHex = hex
	signedData.C4bBrId = int64(needSignData.C4bBrRawData.BrId)

	return signedData, nil
}
func HtlcCloseBobSignedCxbSubTx(hexData interface{}, user *bean.User) (signedData *bean.BobSignedRdTxForC4b, err error) {
	needSignData := hexData.(bean.NeedBobSignRdTxForC4b)
	signedData = &bean.BobSignedRdTxForC4b{ChannelId: needSignData.ChannelId}

	currCommitmentTx := getCurrCommitmentTx(needSignData.ChannelId, *user)
	wallet, _ := service.HDWalletService.GetAddressByIndex(user, uint32(currCommitmentTx.RSMCTempAddressIndex))

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C4bRdPartialSignedData.Hex, convertBean(needSignData.C4bRdPartialSignedData.Inputs), wallet.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C4bRdCompleteSignedHex = hex

	return signedData, nil
}
