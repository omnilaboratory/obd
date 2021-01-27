package lightclient

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/agent"
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
		client.SendToMyself(enum.MsgType_HTLC_FindPath_401, status, data)
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
			respond, err := service.HtlcForwardTxService.CreateHtlcInvoice(msg)
			if err != nil {
				data = err.Error()
			} else {
				status = true
				data = respond.(string)
			}
		}
		client.SendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_FindPath_401:
		tempClientMap[client.User.PeerId] = client
		respond, isPrivate, err := service.HtlcForwardTxService.PayerRequestFindPath(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
			client.SendToMyself(msg.Type, status, data)
		} else {
			if isPrivate {
				bytes, err := json.Marshal(respond)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
				client.SendToMyself(msg.Type, status, data)
			} else {
				status = true
			}
		}
		sendType = enum.SendTargetType_SendToSomeone

	case enum.MsgType_HTLC_SendAddHTLC_40:

		if client.User.IsAdmin {
			agent.BeforeAliceAddHtlcAtAliceSide(&msg, client.User)
		}

		respond, needSign, err := service.HtlcForwardTxService.AliceAddHtlcAtAliceSide(msg, *client.User)
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
			if status {
				if needSign == false {
					msg.Type = enum.MsgType_HTLC_AddHTLC_40
					err = client.sendDataToP2PUser(msg, true, data)
					if err != nil {
						status = false
						data = err.Error()
					}
				}
			}
		}

		if status && needSign {
			if client.User.IsAdmin {
				//签名完成
				signedData, err := agent.AliceSignC3aAtAliceSide(respond, client.User)
				if err == nil {
					signedDataBytes, _ := json.Marshal(signedData)
					msg.Data = string(signedDataBytes)
					log.Println(msg.Data)

					_, toBob, err := service.HtlcForwardTxService.OnAliceSignedC3aAtAliceSide(msg, *client.User)
					if err != nil {
						data = err.Error()
					} else {
						bytes, err := json.Marshal(toBob)
						if err != nil {
							data = err.Error()
						} else {
							data = string(bytes)
							status = true
						}

						if status {
							msg.Type = enum.MsgType_HTLC_AddHTLC_40
							err = client.sendDataToP2PUser(msg, true, data)
							if err != nil {
								status = false
								data = err.Error()
							}
						}
					}
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_SendAddHTLC_40
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_HTLC_ClientSign_Alice_C3a_100:
		toAlice, toBob, err := service.HtlcForwardTxService.OnAliceSignedC3aAtAliceSide(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toBob)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}

			if status {
				msg.Type = enum.MsgType_HTLC_AddHTLC_40
				err = client.sendDataToP2PUser(msg, true, data)
				if err != nil {
					status = false
					data = err.Error()
				}
			}

			if status {
				bytes, err := json.Marshal(toAlice)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_ClientSign_Alice_C3a_100
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_HTLC_SendAddHTLCSigned_41:
		returnData, err := service.HtlcForwardTxService.BobSignedAddHtlcAtBobSide(msg.Data, *client.User)
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
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_ClientSign_Bob_C3b_101:
		toAlice, toBob, err := service.HtlcForwardTxService.OnBobSignedC3bAtBobSide(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toAlice)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
			if status {
				msg.Type = enum.MsgType_HTLC_NeedPayerSignC3b_41
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					status = false
					data = err.Error()
				}
			}
			if status {
				bytes, err := json.Marshal(toBob)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_ClientSign_Bob_C3b_101
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_ClientSign_Alice_C3b_102:
		returnData, err := service.HtlcForwardTxService.OnAliceSignC3bAtAliceSide(msg, *client.User)
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
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_ClientSign_Alice_C3bSub_103:
		toAlice, toBob, err := service.HtlcForwardTxService.OnAliceSignedC3bSubTxAtAliceSide(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toBob)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}

			if status {
				msg.Type = enum.MsgType_HTLC_PayeeCreateHTRD1a_42
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					status = false
					data = err.Error()
				}
			}
			if status {
				bytes, err := json.Marshal(toAlice)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_ClientSign_Alice_C3bSub_103
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_ClientSign_Bob_C3bSub_104:
		returnData, err := service.HtlcForwardTxService.OnBobSignedC3bSubTxAtBobSide(msg, *client.User)
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
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_ClientSign_Alice_He_105:
		toAlice, toBob, err := service.HtlcForwardTxService.OnBobSignHtRdAtBobSide_42(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toAlice)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
			if status {
				msg.Type = enum.MsgType_HTLC_PayerSignHTRD1a_43
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					status = false
					data = err.Error()
				}
			}
			if status {
				bytes, err := json.Marshal(toBob)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_ClientSign_Alice_He_105
		client.SendToMyself(msg.Type, status, data)
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
		client.SendToMyself(msg.Type, status, data)
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
		client.SendToMyself(msg.Type, status, data)
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
	case enum.MsgType_HTLC_SendVerifyR_45:
		respond, err := service.HtlcBackwardTxService.SendRToPreviousNodeAtBobSide(msg, *client.User)
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
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_ClientSign_Bob_HeSub_106:
		respond, err := service.HtlcBackwardTxService.OnBobSignedHeRdAtBobSide(msg, *client.User)
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
		msg.Type = enum.MsgType_HTLC_ClientSign_Bob_HeSub_106
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_ClientSign_Alice_HeSub_46:
		toAlice, toBob, err := service.HtlcBackwardTxService.OnAliceSignedHeRdAtAliceSide(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toBob)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				msg.Type = enum.MsgType_HTLC_SendHerdHex_46
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
			if status {
				bytes, err := json.Marshal(toAlice)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}

		}
		msg.Type = enum.MsgType_HTLC_ClientSign_Alice_HeSub_46
		client.SendToMyself(msg.Type, status, data)
	}
	return sendType, []byte(data), status
}

//htlc tx
func (client *Client) htlcCloseModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_HTLC_Close_SendRequestCloseCurrTx_49:
		outData, needSign, err := service.HtlcCloseTxService.RequestCloseHtlc(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				if needSign == false {
					msg.Type = enum.MsgType_HTLC_Close_RequestCloseCurrTx_49
					err = client.sendDataToP2PUser(msg, status, data)
					if err != nil {
						data = err.Error()
						status = false
					}
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_Close_SendRequestCloseCurrTx_49
		client.SendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HTLC_Close_ClientSign_Alice_C4a_110:
		toAlice, toBob, err := service.HtlcCloseTxService.OnAliceSignedCxa(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toBob)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
			if status {
				msg.Type = enum.MsgType_HTLC_Close_RequestCloseCurrTx_49
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}

			if status {
				bytes, err = json.Marshal(toAlice)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_Close_ClientSign_Alice_C4a_110
		client.SendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone

	case enum.MsgType_HTLC_Close_SendCloseSigned_50:
		outData, err := service.HtlcCloseTxService.OnBobSignCloseHtlcRequest(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(outData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_Close_ClientSign_Bob_C4b_111:
		toAlice, toBob, err := service.HtlcCloseTxService.OnBobSignedCxb(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toAlice)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
			if status {
				msg.Type = enum.MsgType_HTLC_CloseHtlcRequestSignBR_50
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
			if status {
				bytes, err := json.Marshal(toBob)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_Close_ClientSign_Bob_C4b_111
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_Close_ClientSign_Alice_C4b_112:
		toAlice, err := service.HtlcCloseTxService.OnAliceSignedCxb(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toAlice)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_Close_ClientSign_Alice_C4bSub_113:
		toAlice, toBob, err := service.HtlcCloseTxService.OnAliceSignedCxbBubTx(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toBob)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}

			if status {
				msg.Type = enum.MsgType_HTLC_CloseHtlcUpdateCnb_51
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
			if status {
				bytes, err := json.Marshal(toAlice)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		msg.Type = enum.MsgType_HTLC_Close_ClientSign_Alice_C4bSub_113
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_HTLC_Close_ClientSign_Bob_C4bSubResult_114:
		toBob, err := service.HtlcCloseTxService.OnBobSignedCxbSubTx(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(toBob)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
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
		client.SendToMyself(msg.Type, status, data)
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
		client.SendToMyself(msg.Type, status, data)
		break
	}
	return sendType, []byte(data), status
}
