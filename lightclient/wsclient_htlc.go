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
	case enum.MsgType_HTLC_AddHTLC_N40, enum.MsgType_HTLC_Invoice_N4003:
		htlcHRequest := &bean.HtlcHRequest{}
		err := json.Unmarshal([]byte(msg.Data), htlcHRequest)
		if err != nil {
			data = err.Error()
		} else {
			if _, err := client.FindUser(&htlcHRequest.RecipientPeerId); err != nil {
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
	case enum.MsgType_HTLC_AddHTLCSigned_N41:
		respond, senderUser, err := service.HtlcHMessageService.AddHTLCSigned(msg.Data, client.User)
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
	case enum.MsgType_HTLC_FindPathAndSendH_N42:
		respond, bob, err := service.HtlcForwardTxService.AliceFindPathAndSendToBob(msg.Data, *client.User)
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
		client.sendToMyself(msg.Type, status, data)
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
		client.sendToMyself(msg.Type, status, data)
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
		client.sendToMyself(msg.Type, status, data)
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
	case enum.MsgType_HTLC_CloseChannelSigned_N51:
		outData, targetUser, err := service.HtlcCloseTxService.CloseHtlcChannelSigned(msg.Data, *client.User)
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
