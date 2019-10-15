package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"encoding/json"
)

//htlc h module
func (client *Client) htlcHDealModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""

	switch msg.Type {
	case enum.MsgType_HTLC_RequestH_N40:
		htlcHRequest := &bean.HtlcHRequest{}
		err := json.Unmarshal([]byte(msg.Data), htlcHRequest)
		if err != nil {
			data = err.Error()
		} else {
			if _, err := client.FindUser(&htlcHRequest.RecipientPeerId); err != nil {
				data = err.Error()
			} else {
				respond, err := service.HtlcHMessageService.DealHtlcRequest(msg.Data, client.User)
				if err != nil {
					data = err.Error()
				} else {
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
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_CreatedRAndHInfoList_N4001:
		respond, err := service.HtlcHMessageService.GetHtlcCreatedRandHInfoList(client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_CreatedRAndHInfoItem_N4002:
		respond, err := service.HtlcHMessageService.GetHtlcCreatedRandHInfoItemById(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_RespondH_N41:
		respond, senderUser, err := service.HtlcHMessageService.DealHtlcResponse(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, *senderUser, data)
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_SignedRAndHInfoList_N4101:
		respond, err := service.HtlcHMessageService.GetHtlcSignedRandHInfoList(client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_SignedRAndHInfoItem_N4102:
		respond, err := service.HtlcHMessageService.GetHtlcSignedRandHInfoItem(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}

//htlc tx
func (client *Client) htlcTxModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_HTLC_CreateHtlc_N42:
		respond, err := service.HtlcTxService.RequestOpenHtlc(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_SingHtlc_N43:
		respond, err := service.HtlcTxService.SignOpenHtlc(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}
