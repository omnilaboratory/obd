package lightclient

import (
	"encoding/json"
	"obd/bean"
	"obd/bean/enum"
	"obd/service"
)

//htlc h module
func (client *Client) htlcHDealModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""

	switch msg.Type {
	case enum.MsgType_HTLC_Invoice_N4003:
		htlcHRequest := &bean.HtlcRequestFindPath{}
		err := json.Unmarshal([]byte(msg.Data), htlcHRequest)
		if err != nil {
			data = err.Error()
		} else {
			if _, err := client.FindUser(&htlcHRequest.RecipientUserPeerId); err != nil {
				data = err.Error()
			} else {
				respond, err := service.HtlcHMessageService.AddHTLC(msg.Data, client.User)
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
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_FindPath_N4001:
		respond, err := service.HtlcForwardTxService.PayerRequestFindPath(msg.Data, *client.User)
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
	case enum.MsgType_HTLC_AddHTLC_N40:
		respond, err := service.HtlcForwardTxService.PayerAddHtlc_40(msg.Data, *client.User)
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
		if status {
			_ = client.sendDataToP2PUser(msg, true, data)
		}
		client.sendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_AddHTLCSigned_N41:
		returnData, err := service.HtlcForwardTxService.PayeeSignGetAddHtlc_41(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			approval := returnData["approval"]
			bytes, err := json.Marshal(returnData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
			if status {
				msg.Type = enum.MsgType_HTLC_PayerSignC3b_N42
				_ = client.sendDataToP2PUser(msg, status, data)
			}
			if approval == false {
				client.sendToMyself(msg.Type, status, data)
			}
		}
		if err != nil {
			client.sendToMyself(msg.Type, status, data)
		}
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
	case enum.MsgType_HTLC_GetRFromLCommitTx_N4103:
		respond, err := service.HtlcQueryService.GetRFromCommitmentTx(msg.Data, *client.User)
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
	case enum.MsgType_HTLC_GetPathInfoByH_N4104:
		respond, err := service.HtlcQueryService.GetPathInfoByH(msg.Data, *client.User)
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
	case enum.MsgType_HTLC_GetRInfoByHOfOwner_N4105:
		respond, err := service.HtlcQueryService.GetRByHOfReceiver(msg.Data, *client.User)
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
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_VerifyR_N47:
		respond, senderPeerId, err := service.HtlcBackwardTxService.VerifyRAndCreateTxs(msg.Data, *client.User)
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
	case enum.MsgType_HTLC_CloseSigned_N49:
		outData, targetUser, err := service.HtlcCloseTxService.CloseHTLCSigned(msg.Data, *client.User)
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
func (client *Client) atomicSwapModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_Atomic_Swap_N80:
		outData, targetUser, err := service.AtomicSwapService.AtomicSwap(msg.Data, *client.User)
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
		break
	case enum.MsgType_Atomic_Swap_Accept_N81:
		outData, targetUser, err := service.AtomicSwapService.AtomicSwapAccepted(msg.Data, *client.User)
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
		break
	}
	return sendType, []byte(data), status
}
