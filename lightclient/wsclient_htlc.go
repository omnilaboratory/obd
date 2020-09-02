package lightclient

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/service"
	"log"
)

var tempClientMap = make(map[string]*Client)

func htlcTrackerDealModule(msg bean.RequestMessage) {
	status := false
	data := ""
	client := tempClientMap[msg.RecipientUserPeerId]
	if client == nil {
		log.Println("not found client")
		return
	}
	switch msg.Type {
	case enum.MsgType_Tracker_GetHtlcPath_351:
		respond, err := service.HtlcForwardTxService.GetResponseFromTrackerOfPayerRequestFindPath(msg.Data, *client.User)
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
		client.sendToMyself(enum.MsgType_HTLC_FindPath_401, status, data)
	}
}

//htlc h module
func (client *Client) htlcHModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""

	switch msg.Type {
	case enum.MsgType_HTLC_Invoice_402:
		htlcHRequest := &bean.HtlcRequestInvoice{}
		err := json.Unmarshal([]byte(msg.Data), htlcHRequest)
		if err != nil {
			data = err.Error()
		} else {
			respond, err := service.HtlcForwardTxService.CreateHtlcInvoice(msg, *client.User)
			if err != nil {
				data = err.Error()
			} else {
				status = true
				data = respond.(string)
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_FindPath_401:
		respond, isPrivate, err := service.HtlcForwardTxService.PayerRequestFindPath(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
			client.sendToMyself(msg.Type, status, data)
		} else {
			if isPrivate {
				bytes, err := json.Marshal(respond)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
				client.sendToMyself(msg.Type, status, data)
			} else {
				status = true
				tempClientMap[client.User.PeerId] = client
			}
		}
		sendType = enum.SendTargetType_SendToSomeone

	case enum.MsgType_HTLC_SendAddHTLC_40:
		respond, err := service.HtlcForwardTxService.UpdateAddHtlc_40(msg, *client.User)
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
			msg.Type = enum.MsgType_HTLC_AddHTLC_40
			err = client.sendDataToP2PUser(msg, true, data)
			if err != nil {
				status = false
				data = err.Error()
			}
		}
		msg.Type = enum.MsgType_HTLC_SendAddHTLC_40
		client.sendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_SendAddHTLCSigned_41:
		returnData, err := service.HtlcForwardTxService.PayeeSignGetAddHtlc_41(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(returnData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
			if status {
				msg.Type = enum.MsgType_HTLC_PayerSignC3b_42
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					status = false
					data = err.Error()
				}
			}
		}
		if status == false {
			msg.Type = enum.MsgType_HTLC_SendAddHTLCSigned_41
			client.sendToMyself(msg.Type, status, data)
		}
	}
	return sendType, []byte(data), status
}

//htlc tx
func (client *Client) htlcQueryModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToSomeone
	data := ""
	switch msg.Type {
	case enum.MsgType_Htlc_GetLatestHT1aOrHE1b_3250:
		respond, err := service.HtlcQueryTxManager.GetLatestHT1aOrHE1b(msg.Data, *client.User)
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
	case enum.MsgType_Htlc_GetHT1aOrHE1bBySomeCommitmentId_3251:
		respond, err := service.HtlcQueryTxManager.GetHT1aOrHE1bBySomeCommitmentId(msg.Data, *client.User)
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
	default:
		sendType = enum.SendTargetType_SendToNone
	}
	return sendType, []byte(data), status
}

//htlc tx
func (client *Client) htlcTxModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToSomeone
	data := ""
	switch msg.Type {
	// Coding by Kevin 2019-10-28
	case enum.MsgType_HTLC_SendVerifyR_45:
		respond, err := service.HtlcBackwardTxService.SendRToPreviousNode_Step1(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				msg.Type = enum.MsgType_HTLC_VerifyR_45
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_SendVerifyR_45
		client.sendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_SendSignVerifyR_46:
		respond, err := service.HtlcBackwardTxService.VerifyRAndCreateTxs_Step3(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(respond)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				msg.Type = enum.MsgType_HTLC_SendHerdHex_47
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		if status == false {
			client.sendToMyself(msg.Type, status, data)
		}
	}
	return sendType, []byte(data), status
}

//htlc tx
func (client *Client) htlcCloseModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_HTLC_SendRequestCloseCurrTx_49:
		outData, err := service.HtlcCloseTxService.RequestCloseHtlc(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				msg.Type = enum.MsgType_HTLC_RequestCloseCurrTx_49
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_SendRequestCloseCurrTx_49
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_SendCloseSigned_50:
		outData, err := service.HtlcCloseTxService.CloseHTLCSigned(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				msg.Type = enum.MsgType_HTLC_CloseHtlcRequestSignBR_51
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		if err != nil {
			client.sendToMyself(msg.Type, status, data)
		}
	}
	return sendType, []byte(data), status
}
func (client *Client) atomicSwapModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_Atomic_SendSwap_80:
		outData, err := service.AtomicSwapService.AtomicSwap(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				msg.Type = enum.MsgType_Atomic_Swap_80
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		msg.Type = enum.MsgType_Atomic_SendSwap_80
		client.sendToMyself(msg.Type, status, data)
		break
	case enum.MsgType_Atomic_SendSwapAccept_81:
		outData, err := service.AtomicSwapService.AtomicSwapAccepted(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				msg.Type = enum.MsgType_Atomic_SwapAccept_81
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		msg.Type = enum.MsgType_Atomic_SendSwapAccept_81
		client.sendToMyself(msg.Type, status, data)
		break
	}
	return sendType, []byte(data), status
}
