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
	case enum.MsgType_HTLC_FindPathAndSendH_N42:
		respond, bob, err := service.HtlcForwardTxService.AliceFindPathOfSingleHopAndSendToBob(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, bob, data)
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_SendH_N43:
		respond, bob, err := service.HtlcForwardTxService.SendH(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, bob, data)
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_SignGetH_N44:
		respond, senderPeerId, err := service.HtlcForwardTxService.SignGetH(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, senderPeerId, data)
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_CreateCommitmentTx_N45:
		respond, senderPeerId, err := service.HtlcForwardTxService.SenderBeginCreateHtlcCommitmentTx(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, senderPeerId, data)
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone

	// Coding by Kevin 2019-10-28
	case enum.MsgType_HTLC_SendR_N46:
		respond, senderPeerId, err :=
			service.HtlcBackwardTxService.SendRToPreviousNode(msg.Data, *client.User)

		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, senderPeerId, data)
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_SignGetR_N47:
		respond, senderPeerId, err :=
			service.HtlcBackwardTxService.CheckRAndCreateTxs(msg.Data, *client.User)

		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, senderPeerId, data)
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}

//htlc tx
func (client *Client) htlcCloseModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_HTLC_RequestCloseCurrTx_N48:
		outData, targetUser, err := service.HtlcCloseTxService.RequestCloseHtlc(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, targetUser, data)
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_SignCloseCurrTx_N49:
		outData, targetUser, err := service.HtlcCloseTxService.SignCloseHtlc(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, targetUser, data)
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_RequestCloseChannel_N50:
		outData, targetUser, err := service.HtlcCloseTxService.RequestCloseChannel(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, targetUser, data)
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_SignCloseChannel_N51:
		outData, targetUser, err := service.HtlcCloseTxService.SignCloseChannel(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				_ = client.sendToSomeone(msg.Type, status, targetUser, data)
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}
