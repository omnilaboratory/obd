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
	invoiceImage := dao.HtlcInvoiceImage{H: wallet.PubKey, Index: wallet.Index, R: wallet.Wif, CreateBy: user.PeerId, CreateAt: time.Now()}
	_ = user.Db.Save(&invoiceImage)
	return nil
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
		latestTempAddressInfo, err := service.HDWalletService.GetAddressByIndex(user, uint32(commitmentTransaction.RSMCTempAddressIndex))
		if err != nil {
			return err
		}
		needSignData.LastTempAddressPrivateKey = latestTempAddressInfo.Wif
	}

	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return err
	}
	needSignData.CurrRsmcTempAddressIndex = address.Index
	needSignData.CurrRsmcTempAddressPubKey = address.PubKey

	address, err = service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return err
	}
	needSignData.CurrHtlcTempAddressIndex = address.Index
	needSignData.CurrHtlcTempAddressPubKey = address.PubKey

	address, err = service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return err
	}
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

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aCounterpartyRawData.Hex, convertBean(needSignData.C3aCounterpartyRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3aCounterpartyPartialSignedHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcRawData.Hex, convertBean(needSignData.C3aHtlcRawData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return nil, err
	}
	signedData.C3aHtlcPartialSignedHex = hex

	return signedData, nil
}

func HtlcBobSignAddHtlcRequestAtBobSide_40(needSignData bean.CreateHtlcTxForC3aToBob, user *bean.User) (signedData bean.BobSignedC3a, err error) {
	signedData = bean.BobSignedC3a{PayerCommitmentTxHash: needSignData.PayerCommitmentTxHash}

	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return signedData, err
	}
	signedData.CurrRsmcTempAddressIndex = address.Index
	signedData.CurrRsmcTempAddressPubKey = address.PubKey

	address, err = service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return signedData, err
	}
	signedData.CurrHtlcTempAddressIndex = address.Index
	signedData.CurrHtlcTempAddressPubKey = address.PubKey

	channelWalletInfo, err := getChannelWalletInfo(needSignData.ChannelId, user)
	if err != nil {
		return signedData, err
	}

	commitmentTransaction := &dao.CommitmentTransaction{}
	user.Db.Select(q.Eq("ChannelId", needSignData.ChannelId), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(commitmentTransaction)
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

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aCounterpartyPartialSignedData.Hex, convertBean(needSignData.C3aCounterpartyPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aCompleteSignedCounterpartyHex = hex

	_, hex, err = omnicore.OmniSignRawTransactionForUnsend(needSignData.C3aHtlcPartialSignedData.Hex, convertBean(needSignData.C3aHtlcPartialSignedData.Inputs), channelWalletInfo.Wif)
	if err != nil {
		return signedData, err
	}
	signedData.C3aCompleteSignedHtlcHex = hex

	return signedData, nil
}

func HtlcBobSignC3b() {

}
