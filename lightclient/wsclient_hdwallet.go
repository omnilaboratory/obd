package lightclient

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/service"
	"strconv"
)

func (client *Client) HdWalletModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToSomeone
	var data string
	switch msg.Type {
	case enum.MsgType_GetMnemonic_2004:
		mnemonic, err := service.HDWalletService.GenSeed()
		if err != nil { //get  successful.
			data = err.Error()
		} else {
			data = mnemonic
			status = true
		}
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_Mnemonic_CreateAddress_3000:
		wallet, err := service.HDWalletService.CreateNewAddress(client.User)
		if err != nil { //get  successful.
			data = err.Error()
		} else {
			bytes, err := json.Marshal(wallet)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
			status = true
		}
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_Mnemonic_GetAddressByIndex_3001:
		index, err := strconv.Atoi(msg.Data)
		if err != nil || index < 0 {
			data = "error index"
		} else {
			wallet, err := service.HDWalletService.GetAddressByIndex(client.User, uint32(index))
			if err != nil { //get  successful.
				data = err.Error()
			} else {
				bytes, err := json.Marshal(wallet)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		client.SendToMyself(msg.Type, status, data)

	default:
		sendType = enum.SendTargetType_SendToNone
	}
	return sendType, []byte(data), status
}
