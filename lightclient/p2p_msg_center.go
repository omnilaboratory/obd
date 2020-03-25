package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"errors"
)

func routerOfSameNode(msgType enum.MsgType, data string, user *bean.User) (err error) {
	err = errors.New("fail to deal msg in the inter node")
	status := false
	switch msgType {
	case enum.MsgType_ChannelOpen_N32:
		err := service.ChannelService.BeforeBobOpenChannel(data, user)
		if err == nil {
			status = true
		}
	case enum.MsgType_ChannelAccept_N33:
		_, err := service.ChannelService.AfterBobAcceptChannel(data, user)
		if err == nil {
			status = true
		}
	case enum.MsgType_FundingCreate_BtcCreate_N3400:
		_, err := service.FundingTransactionService.BeforeBobSignBtcFunding(data, user)
		if err == nil {
			status = true
		}
	case enum.MsgType_FundingSign_BtcSign_N3500:
		_, err := service.FundingTransactionService.AfterBobSignBtcFunding(data, user)
		if err == nil {
			status = true
		}
	default:
		status = true
	}

	if status {
		err = nil
	}
	return err
}
