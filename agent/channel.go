package agent

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/service"
)

func BeforeBobAcceptOpenChannel(msg *bean.RequestMessage, user *bean.User) (resultData *bean.SendSignOpenChannel, err error) {

	aliceOpenChannelInfo := bean.RequestOpenChannel{}
	err = json.Unmarshal([]byte(msg.Data), &aliceOpenChannelInfo)
	if err != nil {
		return nil, err
	}

	resultData = &bean.SendSignOpenChannel{}
	resultData.TemporaryChannelId = aliceOpenChannelInfo.TemporaryChannelId

	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return nil, err
	}
	resultData.FundeeAddressIndex = address.Index
	resultData.FundingPubKey = address.PubKey
	resultData.Approval = true

	return resultData, nil
}
