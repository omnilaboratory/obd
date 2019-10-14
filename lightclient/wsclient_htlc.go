package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"encoding/json"
)

func (client *Client) htlcHDealModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""

	switch msg.Type {
	case enum.MsgType_HTLC_RequestH:
		htlcHRequest := &bean.HtlcHRequest{}
		err := json.Unmarshal([]byte(msg.Data), htlcHRequest)
		if err != nil {
			data = err.Error()
		} else {
			if _, err := client.FindUser(&htlcHRequest.RecipientPeerId); err != nil {
				data = err.Error()
			} else {
				respond, err := service.HtlcHMessageService.DealHtlcRequest(msg.Data, client.User)
				bytes, err := json.Marshal(respond)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
					_ = client.sendToSomeone(msg.Type, status, htlcHRequest.RecipientPeerId, data)
				}
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_RespondH:
		respond, err := service.HtlcHMessageService.DealHtlcResponse(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, respond.SenderPeerId, data)
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone
	}

	return sendType, []byte(data), status
}
