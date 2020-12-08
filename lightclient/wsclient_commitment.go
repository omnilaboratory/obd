package lightclient

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/service"
	"log"
	"strconv"
)

func (client *Client) commitmentTxModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_CommitmentTx_SendCommitmentTransactionCreated_351:
		retData, needSign, err := service.CommitmentTxService.CommitmentTransactionCreated(msg, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(retData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		if needSign == false {
			if status {
				msg.Type = enum.MsgType_CommitmentTx_CommitmentTransactionCreated_351
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		msg.Type = enum.MsgType_CommitmentTx_SendCommitmentTransactionCreated_351
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_ClientSign_CommitmentTx_AliceSignC2a_360:
		toAlice, retData, err := service.CommitmentTxService.OnAliceSignC2aRawTxAtAliceSide(msg, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(retData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}

			if status {
				msg.Type = enum.MsgType_CommitmentTx_CommitmentTransactionCreated_351
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

		msg.Type = enum.MsgType_ClientSign_CommitmentTx_AliceSignC2a_360
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_ItemsByChanId_3200:
		nodes, count, err := service.CommitmentTxService.GetItemsByChannelId(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			page := make(map[string]interface{})
			page["count"] = len(nodes)
			page["totalCount"] = count
			page["body"] = nodes
			bytes, err := json.Marshal(page)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_ItemById_3201:
		id, err := strconv.Atoi(msg.Data)
		if err != nil {
			log.Println(err)
			data = err.Error()
		} else {
			node, err := service.CommitmentTxService.GetItemById(id, *client.User)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_Count_3202:
		count, err := service.CommitmentTxService.TotalCount(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(count)
			status = true
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_LatestCommitmentTxByChanId_3203:
		node, err := service.CommitmentTxService.GetLatestCommitmentTxByChannelId(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_LatestRDByChanId_3204:
		node, err := service.CommitmentTxService.GetLatestRDTxByChannelId(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_LatestBRByChanId_3205:
		node, err := service.CommitmentTxService.GetLatestBRTxByChannelId(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_DelItemByChanId_3209:
		err := service.CommitmentTxService.DelItemByChannelId(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			data = "success"
			status = true
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_SendSomeCommitmentById_3206:
		node, err := service.CommitmentTxService.SendSomeCommitmentById(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_AllRDByChanId_3207:
		node, err := service.CommitmentTxService.GetAllRDByChannelId(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_CommitmentTx_AllBRByChanId_3208:
		node, err := service.CommitmentTxService.GetAllBRByChannelId(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_ClientSign_CommitmentTx_AliceSignC2b_362:
		node, err := service.CommitmentTxService.OnAliceSignedC2bTxAtAliceSide(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_ClientSign_CommitmentTx_AliceSignC2b_Rd_363:
		aliceData, bobData, _, err := service.CommitmentTxService.OnAliceSignedC2b_RDTxAtAliceSide(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bobBytes, err := json.Marshal(bobData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bobBytes)
				status = true
			}
			if status {
				msg.Type = enum.MsgType_CommitmentTxSigned_SecondToBobSign_353
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}

			aliceBytes, err := json.Marshal(aliceData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(aliceBytes)
				status = true
			}
		}
		msg.Type = enum.MsgType_ClientSign_CommitmentTx_AliceSignC2b_Rd_363
		client.SendToMyself(msg.Type, status, data)
	}
	return sendType, []byte(data), status
}
func (client *Client) commitmentTxSignModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	case enum.MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352:
		retData, needSignC2b, err := service.CommitmentTxSignedService.RevokeAndAcknowledgeCommitmentTransaction(msg, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(retData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		if status {
			if needSignC2b == false {
				msg.Type = enum.MsgType_CommitmentTxSigned_ToAliceSign_352
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					status = false
					data = err.Error()
				}
			}
		}
		msg.Type = enum.MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_ClientSign_CommitmentTx_BobSignC2b_361:
		toBobData, retData, err := service.CommitmentTxSignedService.OnBobSignC2bTransactionAtBobSide(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(retData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}

			if status {
				msg.Type = enum.MsgType_CommitmentTxSigned_ToAliceSign_352
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					status = false
					data = err.Error()
				}
			}

			if status {
				bytes, err := json.Marshal(toBobData)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
		}
		msg.Type = enum.MsgType_ClientSign_CommitmentTx_BobSignC2b_361
		client.SendToMyself(msg.Type, status, data)
	case enum.MsgType_ClientSign_CommitmentTx_BobSignC2b_Rd_364:
		retData, err := service.CommitmentTxSignedService.BobSignC2b_RdAtBobSide(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(retData)
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
