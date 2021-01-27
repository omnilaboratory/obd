package agent

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
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

	return nil
}
