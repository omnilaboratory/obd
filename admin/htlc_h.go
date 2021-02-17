package admin

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/service"
	"log"
	"strings"
	"time"
)

func HtlcCreateInvoice(msg *bean.RequestMessage, user *bean.User) (err error) {
	wallet, _ := service.HDWalletService.CreateNewAddress(user)
	requestData := &bean.HtlcRequestInvoice{}
	err = json.Unmarshal([]byte(msg.Data), requestData)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	requestData.H = wallet.PubKey
	marshal, _ := json.Marshal(requestData)
	msg.Data = string(marshal)
	invoiceImage := dao.HtlcHAndRImage{H: wallet.PubKey, Index: wallet.Index, R: wallet.Wif, CreateAt: time.Now()}
	_ = user.Db.Save(&invoiceImage)
	return nil
}

func HtlcFindPathByInvoice(invoice string) {

}

func HtlcBeforeAliceAddHtlcAtAliceSide(msg *bean.RequestMessage, user *bean.User) (err error) {
	needSignData := &bean.CreateHtlcTxForC3a{}
	err = json.Unmarshal([]byte(msg.Data), needSignData)
	if err != nil {
		return err
	}

	channelIds := strings.Split(needSignData.RoutingPacket, ",")
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

	commitmentTransaction := getLatestCommitmentTx(channelInfo.ChannelId, *user)
	if commitmentTransaction.Id > 0 {
		latestTempAddressInfo, _ := service.HDWalletService.GetAddressByIndex(user, uint32(commitmentTransaction.RSMCTempAddressIndex))
		needSignData.LastTempAddressPrivateKey = latestTempAddressInfo.Wif
	}

	address, _ := service.HDWalletService.CreateNewAddress(user)
	needSignData.CurrRsmcTempAddressIndex = address.Index
	needSignData.CurrRsmcTempAddressPubKey = address.PubKey

	address, _ = service.HDWalletService.CreateNewAddress(user)
	needSignData.CurrHtlcTempAddressIndex = address.Index
	needSignData.CurrHtlcTempAddressPubKey = address.PubKey

	address, _ = service.HDWalletService.CreateNewAddress(user)
	needSignData.CurrHtlcTempAddressForHt1aIndex = address.Index
	needSignData.CurrHtlcTempAddressForHt1aPubKey = address.PubKey
	marshal, _ := json.Marshal(needSignData)
	msg.Data = string(marshal)

	return nil
}

func HtlcAliceSignC3aAtAliceSide(data interface{}, user *bean.User) (signedData *bean.AliceSignedHtlcDataForC3a, err error) {
	needSignData := data.(bean.NeedAliceSignCreateHtlcTxForC3a)

	channelWalletInfo, err := getChannelWalletInfo(needSignData.ChannelId, user)
	if err != nil {
		return nil, err
	}

	signedData = &bean.AliceSignedHtlcDataForC3a{}
	signedData.ChannelId = needSignData.ChannelId

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aRsmcRawData.Hex, convertBean(needSignData.C3aRsmcRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3aRsmcPartialSignedHex = hex

	if len(needSignData.C3aCounterpartyRawData.Hex) > 0 {
		_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aCounterpartyRawData.Hex, convertBean(needSignData.C3aCounterpartyRawData.Inputs), channelWalletInfo.Wif)
		if err != nil {
			return nil, err
		}
		signedData.C3aCounterpartyPartialSignedHex = hex
	}

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcRawData.Hex, convertBean(needSignData.C3aHtlcRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3aHtlcPartialSignedHex = hex

	return signedData, nil
}

func HtlcBobSignAddHtlcRequestAtBobSide_40(needSignData bean.CreateHtlcTxForC3aToBob, user *bean.User) (signedData bean.BobSignedC3a, err error) {
	signedData = bean.BobSignedC3a{PayerCommitmentTxHash: needSignData.PayerCommitmentTxHash}

	address, _ := service.HDWalletService.CreateNewAddress(user)
	signedData.CurrRsmcTempAddressIndex = address.Index
	signedData.CurrRsmcTempAddressPubKey = address.PubKey

	address, _ = service.HDWalletService.CreateNewAddress(user)
	signedData.CurrHtlcTempAddressIndex = address.Index
	signedData.CurrHtlcTempAddressPubKey = address.PubKey

	channelWalletInfo, err := getChannelWalletInfo(needSignData.ChannelId, user)
	if err != nil {
		return signedData, err
	}

	commitmentTransaction := getLatestCommitmentTx(needSignData.ChannelId, *user)
	if commitmentTransaction.Id > 0 {
		latestTempAddressInfo, err := service.HDWalletService.GetAddressByIndex(user, uint32(commitmentTransaction.RSMCTempAddressIndex))
		if err != nil {
			return signedData, err
		}
		signedData.LastTempAddressPrivateKey = latestTempAddressInfo.Wif
	}

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aRsmcPartialSignedData.Hex, convertBean(needSignData.C3aRsmcPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aCompleteSignedRsmcHex = hex

	if len(needSignData.C3aCounterpartyPartialSignedData.Hex) > 0 {
		_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aCounterpartyPartialSignedData.Hex, convertBean(needSignData.C3aCounterpartyPartialSignedData.Inputs), channelWalletInfo.Wif)
		if err != nil {
			return signedData, err
		}
		signedData.C3aCompleteSignedCounterpartyHex = hex
	}

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcPartialSignedData.Hex, convertBean(needSignData.C3aHtlcPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aCompleteSignedHtlcHex = hex

	return signedData, nil
}

func HtlcBobSignC3b(needSignData *bean.NeedBobSignHtlcTxOfC3b, user *bean.User) (signedData *bean.BobSignedHtlcTxOfC3b, err error) {

	signedData = &bean.BobSignedHtlcTxOfC3b{ChannelId: needSignData.ChannelId}

	channelWalletInfo, _ := getChannelWalletInfo(needSignData.ChannelId, user)
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aRsmcRdRawData.Hex, convertBean(needSignData.C3aRsmcRdRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aRsmcRdPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aRsmcBrRawData.Hex, convertBean(needSignData.C3aRsmcBrRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aRsmcBrPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcHtRawData.Hex, convertBean(needSignData.C3aHtlcHtRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aHtlcHtPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcHlockRawData.Hex, convertBean(needSignData.C3aHtlcHlockRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aHtlcHlockPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcBrRawData.Hex, convertBean(needSignData.C3aHtlcBrRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aHtlcBrPartialSignedHex = hex

	if len(needSignData.C3bRsmcRawData.Hex) > 0 {
		_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bRsmcRawData.Hex, convertBean(needSignData.C3bRsmcRawData.Inputs), channelWalletInfo.Wif)
		if err != nil {
			return signedData, err
		}
		signedData.C3bRsmcPartialSignedHex = hex
	}

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bCounterpartyRawData.Hex, convertBean(needSignData.C3bCounterpartyRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bCounterpartyPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcRawData.Hex, convertBean(needSignData.C3bHtlcRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bHtlcPartialSignedHex = hex
	return signedData, nil
}

func HtlcAliceSignC3b(dataHex interface{}, user *bean.User) (signedData *bean.AliceSignedHtlcTxOfC3bResult, err error) {
	needSignData := dataHex.(bean.NeedAliceSignHtlcTxOfC3b)
	signedData = &bean.AliceSignedHtlcTxOfC3bResult{ChannelId: needSignData.ChannelId}
	currCommitmentTx := getCurrCommitmentTx(needSignData.ChannelId, *user)
	wallet, _ := service.HDWalletService.GetAddressByIndex(user, uint32(currCommitmentTx.RSMCTempAddressIndex))

	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aRsmcRdPartialSignedData.Hex, convertBean(needSignData.C3aRsmcRdPartialSignedData.Inputs), wallet.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aRsmcRdCompleteSignedHex = hex

	wallet, _ = service.HDWalletService.GetAddressByIndex(user, uint32(currCommitmentTx.HTLCTempAddressIndex))
	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcHtPartialSignedData.Hex, convertBean(needSignData.C3aHtlcHtPartialSignedData.Inputs), wallet.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aHtlcHtCompleteSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcHlockPartialSignedData.Hex, convertBean(needSignData.C3aHtlcHlockPartialSignedData.Inputs), wallet.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aHtlcHlockCompleteSignedHex = hex

	channelWalletInfo, _ := getChannelWalletInfo(needSignData.ChannelId, user)
	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bRsmcPartialSignedData.Hex, convertBean(needSignData.C3bRsmcPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bRsmcCompleteSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcPartialSignedData.Hex, convertBean(needSignData.C3bHtlcPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bHtlcCompleteSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bCounterpartyPartialSignedData.Hex, convertBean(needSignData.C3bCounterpartyPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bCounterpartyCompleteSignedHex = hex

	return signedData, nil
}

func HtlcAliceSignedC3bSubTxAtAliceSide(dataHex interface{}, user *bean.User) (signedData *bean.AliceSignedHtlcSubTxOfC3b, err error) {
	needSignData := dataHex.(bean.NeedAliceSignHtlcSubTxOfC3b)

	signedData = &bean.AliceSignedHtlcSubTxOfC3b{ChannelId: needSignData.ChannelId}

	currCommitmentTx := getCurrCommitmentTx(needSignData.ChannelId, *user)

	htlcRequestInfo := &dao.AddHtlcRequestInfo{}
	_ = user.Db.Select(
		q.Eq("ChannelId", needSignData.ChannelId),
		q.Eq("H", currCommitmentTx.HtlcH),
		q.Eq("RoutingPacket", currCommitmentTx.HtlcRoutingPacket)).First(htlcRequestInfo)

	if htlcRequestInfo.Id > 0 {
		wallet, _ := service.HDWalletService.GetAddressByIndex(user, uint32(htlcRequestInfo.CurrHtlcTempAddressForHt1aIndex))
		_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcHtrdRawData.Hex, convertBean(needSignData.C3aHtlcHtrdRawData.Inputs), wallet.Wif)
		if err != nil {
			return nil, err
		}
		signedData.C3aHtlcHtrdPartialSignedHex = hex
	}

	channelWalletInfo, _ := getChannelWalletInfo(needSignData.ChannelId, user)
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bRsmcRdRawData.Hex, convertBean(needSignData.C3bRsmcRdRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3bRsmcRdPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bRsmcBrRawData.Hex, convertBean(needSignData.C3bRsmcBrRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3bRsmcBrPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcHtdRawData.Hex, convertBean(needSignData.C3bHtlcHtdRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3bHtlcHtdPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcHlockRawData.Hex, convertBean(needSignData.C3bHtlcHlockRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3bHtlcHlockPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcBrRawData.Hex, convertBean(needSignData.C3bHtlcBrRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3bHtlcBrPartialSignedHex = hex

	return signedData, nil
}

func HtlcBobSignedC3bSubTxAtBobSide(dataHex interface{}, user *bean.User) (signedData *bean.BobSignedHtlcSubTxOfC3b, err error) {
	needSignData := dataHex.(bean.NeedBobSignHtlcSubTxOfC3b)

	signedData = &bean.BobSignedHtlcSubTxOfC3b{ChannelId: needSignData.ChannelId}
	newAddress, _ := service.HDWalletService.CreateNewAddress(user)
	signedData.CurrHtlcTempAddressForHeIndex = newAddress.Index
	signedData.CurrHtlcTempAddressForHePubKey = newAddress.PubKey

	currCommitmentTx := getCurrCommitmentTx(needSignData.ChannelId, *user)
	wallet, _ := service.HDWalletService.GetAddressByIndex(user, uint32(currCommitmentTx.RSMCTempAddressIndex))
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bRsmcRdPartialData.Hex, convertBean(needSignData.C3bRsmcRdPartialData.Inputs), wallet.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bRsmcRdCompleteSignedHex = hex

	wallet, _ = service.HDWalletService.GetAddressByIndex(user, uint32(currCommitmentTx.HTLCTempAddressIndex))
	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcHtdPartialData.Hex, convertBean(needSignData.C3bHtlcHtdPartialData.Inputs), wallet.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bHtlcHtdCompleteSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcHlockPartialData.Hex, convertBean(needSignData.C3bHtlcHlockPartialData.Inputs), wallet.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bHtlcHlockCompleteSignedHex = hex

	channelWalletInfo, err := getChannelWalletInfo(needSignData.ChannelId, user)
	if err != nil {
		return nil, err
	}
	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcHtrdPartialData.Hex, convertBean(needSignData.C3aHtlcHtrdPartialData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aHtlcHtrdCompleteSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcHtbrRawData.Hex, convertBean(needSignData.C3aHtlcHtbrRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aHtlcHtbrPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcHedRawData.Hex, convertBean(needSignData.C3aHtlcHedRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aHtlcHedPartialSignedHex = hex

	return signedData, nil
}

func HtlcBobSignHtRdAtBobSide(dataHex interface{}, user *bean.User) (signedData *bean.BobSignedHtlcHeTxOfC3b, err error) {
	needSignData := dataHex.(bean.NeedBobSignHtlcHeTxOfC3b)

	signedData = &bean.BobSignedHtlcHeTxOfC3b{ChannelId: needSignData.ChannelId}
	channelWalletInfo, err := getChannelWalletInfo(needSignData.ChannelId, user)
	if err != nil {
		return nil, err
	}
	_, hex, err := omnicore.OmniSignRawTransactionForUnsend(needSignData.C3bHtlcHlockHeRawData.Hex, convertBean(needSignData.C3bHtlcHlockHeRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3bHtlcHlockHePartialSignedHex = hex

	return signedData, nil
}
