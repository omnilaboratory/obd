package lightclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/tool"
	"log"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type Client struct {
	Id            string
	IsGRpcRequest bool
	User          *bean.User
	Socket        *websocket.Conn
	SendChannel   chan []byte
	GrpcChan      chan []byte
}

func (client *Client) Write() {
	defer func() {
		_ = client.Socket.Close()
	}()

	for {
		select {
		case data, ok := <-client.SendChannel:
			if !ok {
				_ = client.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := client.Socket.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("fail to send data to client ", string(data))
				log.Println(err)
			}
		}
	}
}

func (client *Client) Read() {
	defer func() {
		GlobalWsClientManager.Disconnected <- client
	}()

	for {
		_, dataReq, err := client.Socket.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var msg bean.RequestMessage
		//log.Println("input data: ", string(dataReq))

		temp := make(map[string]interface{})
		err = json.Unmarshal(dataReq, &temp)
		if err != nil {
			log.Println(err)
			client.SendToMyself(enum.MsgType_Error_0, false, "error json format")
			continue
		}
		temp = nil

		jsonParse := gjson.Parse(string(dataReq))
		if jsonParse.Value() == nil || jsonParse.Exists() == false || jsonParse.IsObject() == false {
			log.Println("wrong json input")
			client.SendToMyself(enum.MsgType_Error_0, false, "wrong json input")
			continue
		}

		msg.Type = enum.MsgType(jsonParse.Get("type").Int())
		if enum.CheckExist(msg.Type) == false {
			data := "not exist the msg type"
			log.Println(data)
			client.SendToMyself(msg.Type, false, data)
			continue
		}

		msg.SenderNodePeerId = P2PLocalNodeId
		msg.SenderUserPeerId = client.Id
		if client.User != nil {
			msg.SenderUserPeerId = client.User.PeerId
		}

		msg.RecipientNodePeerId = jsonParse.Get("recipient_node_peer_id").String()
		msg.RecipientUserPeerId = jsonParse.Get("recipient_user_peer_id").String()

		// check the Recipient is online
		if tool.CheckIsString(&msg.RecipientUserPeerId) {
			_, err = FindUserOnLine(msg)
			if err != nil {
				client.SendToMyself(msg.Type, false, fmt.Sprintf(enum.Tips_user_notExistOrOnline, msg.RecipientUserPeerId))
				continue
			}
		}
		msg.Data = jsonParse.Get("data").String()

		var sendType = enum.SendTargetType_SendToNone
		status := false
		var dataOut []byte
		var needLogin = true

		if msg.Type <= enum.MsgType_UserLogin_2001 && msg.Type >= enum.MsgType_Common_End_2999 {
			needLogin = false
		}

		if msg.Type <= enum.MsgType_UserLogin_2001 && msg.Type > enum.MsgType_User_End_2099 {
			if msg.Type == enum.MsgType_GetMnemonic_2004 {
				sendType, dataOut, status = client.HdWalletModule(msg)
			} else {
				sendType, dataOut, status = client.UserModule(msg)
			}
			needLogin = false
		}

		if msg.Type <= enum.MsgType_Core_GetNewAddress_2101 && msg.Type > enum.MsgType_Core_Omni_End_2199 {
			sendType, dataOut, status = client.omniCoreModule(msg)
			needLogin = false
		}

		if needLogin {
			//not login
			if client.User == nil {
				client.SendToMyself(msg.Type, false, "please login")
				continue
			} else { // already login

				if msg.Type == enum.MsgType_SendChannelOpen_32 || msg.Type == enum.MsgType_SendChannelAccept_33 ||
					msg.Type == enum.MsgType_FundingCreate_SendBtcFundingCreated_340 || msg.Type == enum.MsgType_FundingSign_SendBtcSign_350 ||
					msg.Type == enum.MsgType_FundingCreate_SendAssetFundingCreated_34 || msg.Type == enum.MsgType_FundingSign_SendAssetFundingSigned_35 ||
					msg.Type == enum.MsgType_Funding_134 ||
					msg.Type == enum.MsgType_ClientSign_Duplex_BtcFundingMinerRDTx_341 ||
					msg.Type == enum.MsgType_CommitmentTx_SendCommitmentTransactionCreated_351 ||
					msg.Type == enum.MsgType_ClientSign_CommitmentTx_AliceSignC2a_360 ||
					msg.Type == enum.MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352 ||
					msg.Type == enum.MsgType_ClientSign_CommitmentTx_BobSignC2b_361 ||
					msg.Type == enum.MsgType_ClientSign_CommitmentTx_AliceSignC2b_Rd_363 ||
					msg.Type == enum.MsgType_HTLC_SendAddHTLC_40 ||
					msg.Type == enum.MsgType_HTLC_ClientSign_Alice_C3a_100 ||
					msg.Type == enum.MsgType_HTLC_ClientSign_Bob_C3b_101 ||
					msg.Type == enum.MsgType_HTLC_ClientSign_Alice_C3bSub_103 ||
					msg.Type == enum.MsgType_HTLC_ClientSign_Alice_He_105 ||
					msg.Type == enum.MsgType_HTLC_ClientSign_Bob_HeSub_106 ||
					msg.Type == enum.MsgType_HTLC_ClientSign_Alice_HeSub_46 ||
					msg.Type == enum.MsgType_HTLC_SendVerifyR_45 ||
					msg.Type == enum.MsgType_HTLC_Close_SendRequestCloseCurrTx_49 ||
					msg.Type == enum.MsgType_HTLC_Close_ClientSign_Bob_C4b_111 ||
					msg.Type == enum.MsgType_HTLC_Close_ClientSign_Alice_C4bSub_113 ||
					msg.Type == enum.MsgType_HTLC_Close_SendCloseSigned_50 ||
					msg.Type == enum.MsgType_Atomic_SendSwap_80 || msg.Type == enum.MsgType_Atomic_SendSwapAccept_81 {
					if tool.CheckIsString(&msg.RecipientUserPeerId) == false {
						client.SendToMyself(msg.Type, false, enum.Tips_common_empty+" recipient_user_peer_id")
						continue
					}
					if tool.CheckIsString(&msg.RecipientNodePeerId) == false {
						client.SendToMyself(msg.Type, false, enum.Tips_common_empty+"error recipient_node_peer_id")
						continue
					}
					if P2pChannelMap[msg.RecipientNodePeerId] == nil {
						err = ScanAndConnNode(msg.RecipientNodePeerId)
						if err != nil {
							client.SendToMyself(msg.Type, false, fmt.Sprintf(enum.Tips_common_errorObdPeerId, msg.RecipientNodePeerId))
							continue
						}
					}
				}

				for {
					//-3000 -3001
					if msg.Type <= enum.MsgType_Mnemonic_CreateAddress_3000 &&
						msg.Type >= enum.MsgType_Mnemonic_GetAddressByIndex_3001 {
						sendType, dataOut, status = client.HdWalletModule(msg)
						break
					}

					//-32 -33  -38 and query for channel
					if msg.Type == enum.MsgType_SendChannelOpen_32 ||
						msg.Type == enum.MsgType_SendChannelAccept_33 ||
						msg.Type == enum.MsgType_SendCloseChannelRequest_38 ||
						(msg.Type <= enum.MsgType_ChannelOpen_AllItem_3150 &&
							msg.Type >= enum.MsgType_CheckChannelAddessExist_3156) {
						sendType, dataOut, status = client.ChannelModule(msg)
						break
					}

					//-34 -340 and query
					if msg.Type == enum.MsgType_FundingCreate_SendAssetFundingCreated_34 ||
						msg.Type == enum.MsgType_FundingCreate_SendBtcFundingCreated_340 ||
						msg.Type == enum.MsgType_ClientSign_AssetFunding_AliceSignC1a_1034 ||
						msg.Type == enum.MsgType_ClientSign_AssetFunding_AliceSignRD_1134 ||
						msg.Type == enum.MsgType_ClientSign_Duplex_BtcFundingMinerRDTx_341 ||
						msg.Type == enum.MsgType_Funding_134 ||
						(msg.Type <= enum.MsgType_FundingCreate_Asset_AllItem_3100 &&
							msg.Type >= enum.MsgType_FundingCreate_Btc_ItemByChannelId_3111) {
						sendType, dataOut, status = client.FundingTransactionModule(msg)
						break
					}

					//-35 -350
					if msg.Type == enum.MsgType_FundingSign_SendAssetFundingSigned_35 ||
						msg.Type == enum.MsgType_ClientSign_AssetFunding_RdAndBr_1035 ||
						msg.Type == enum.MsgType_FundingSign_SendBtcSign_350 {
						sendType, dataOut, status = client.fundingSignModule(msg)
						break
					}

					//-351 及查询
					if msg.Type == enum.MsgType_CommitmentTx_SendCommitmentTransactionCreated_351 ||
						msg.Type == enum.MsgType_ClientSign_CommitmentTx_AliceSignC2a_360 ||
						msg.Type == enum.MsgType_ClientSign_CommitmentTx_AliceSignC2b_362 ||
						msg.Type == enum.MsgType_ClientSign_CommitmentTx_AliceSignC2b_Rd_363 ||
						(msg.Type <= enum.MsgType_CommitmentTx_ItemsByChanId_3200 &&
							msg.Type >= enum.MsgType_CommitmentTx_DelItemByChanId_3209) {
						sendType, dataOut, status = client.CommitmentTxModule(msg)
						break
					}

					//-htlc query
					if msg.Type <= enum.MsgType_Htlc_GetLatestHT1aOrHE1b_3250 &&
						msg.Type >= enum.MsgType_Htlc_GetHT1aOrHE1bBySomeCommitmentId_3251 {
						sendType, dataOut, status = client.htlcQueryModule(msg)
						break
					}

					//-352
					if msg.Type == enum.MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352 ||
						msg.Type == enum.MsgType_ClientSign_CommitmentTx_BobSignC2b_361 ||
						msg.Type == enum.MsgType_ClientSign_CommitmentTx_BobSignC2b_Rd_364 {
						sendType, dataOut, status = client.commitmentTxSignModule(msg)
						break
					}

					//-40 -41
					if msg.Type == enum.MsgType_HTLC_FindPath_401 ||
						msg.Type == enum.MsgType_HTLC_Invoice_402 ||
						msg.Type == enum.MsgType_HTLC_ParseInvoice_403 ||
						msg.Type == enum.MsgType_HTLC_SendAddHTLC_40 ||
						msg.Type == enum.MsgType_HTLC_ClientSign_Alice_C3a_100 ||
						msg.Type == enum.MsgType_HTLC_ClientSign_Bob_C3b_101 ||
						msg.Type == enum.MsgType_HTLC_ClientSign_Alice_C3b_102 ||
						msg.Type == enum.MsgType_HTLC_ClientSign_Alice_C3bSub_103 ||
						msg.Type == enum.MsgType_HTLC_ClientSign_Bob_C3bSub_104 ||
						msg.Type == enum.MsgType_HTLC_ClientSign_Alice_He_105 ||
						msg.Type == enum.MsgType_HTLC_SendAddHTLCSigned_41 {
						sendType, dataOut, status = client.HtlcHModule(msg)
						break
					}

					//-42 -43 -44 -45 -46 -47
					if msg.Type == enum.MsgType_HTLC_SendVerifyR_45 ||
						msg.Type == enum.MsgType_HTLC_ClientSign_Bob_HeSub_106 ||
						msg.Type == enum.MsgType_HTLC_ClientSign_Alice_HeSub_46 {
						sendType, dataOut, status = client.htlcTxModule(msg)
						break
					}

					// -49 -50
					if msg.Type == enum.MsgType_HTLC_Close_SendRequestCloseCurrTx_49 ||
						msg.Type == enum.MsgType_HTLC_Close_ClientSign_Alice_C4a_110 ||
						msg.Type == enum.MsgType_HTLC_Close_ClientSign_Bob_C4b_111 ||
						msg.Type == enum.MsgType_HTLC_Close_ClientSign_Alice_C4b_112 ||
						msg.Type == enum.MsgType_HTLC_Close_ClientSign_Alice_C4bSub_113 ||
						msg.Type == enum.MsgType_HTLC_Close_ClientSign_Bob_C4bSubResult_114 ||
						msg.Type == enum.MsgType_HTLC_Close_SendCloseSigned_50 {
						sendType, dataOut, status = client.htlcCloseModule(msg)
						break
					}
					// -80 -81
					if msg.Type == enum.MsgType_Atomic_SendSwap_80 ||
						msg.Type == enum.MsgType_Atomic_SendSwapAccept_81 {
						sendType, dataOut, status = client.atomicSwapModule(msg)
						break
					}
					break
				}
			}
		}

		if status == false && len(dataOut) == 0 && sendType == enum.SendTargetType_SendToNone {
			data := "the msg type has no module"
			log.Println(data)
			client.SendToMyself(msg.Type, false, data)
			continue
		}

		if len(dataOut) == 0 {
			dataOut = dataReq
		}

		//broadcast except me
		if sendType == enum.SendTargetType_SendToExceptMe {
			for itemClient := range GlobalWsClientManager.ClientsMap {
				if itemClient != client {
					jsonMessage := getReplyObj(string(dataOut), msg.Type, status, client, itemClient)
					itemClient.SendChannel <- jsonMessage
				}
			}
		}
		//broadcast to all
		if sendType == enum.SendTargetType_SendToAll {
			jsonMessage := getReplyObj(string(dataOut), msg.Type, status, client, nil)
			GlobalWsClientManager.Broadcast <- jsonMessage
		}
	}
}

func (client *Client) SendToMyself(msgType enum.MsgType, status bool, data string) {
	if client.SendChannel != nil {
		jsonMessage := getReplyObj(data, msgType, status, client, client)
		client.SendChannel <- jsonMessage
	}
}

// send p2p msg, check whether they are at the same obd node
func (client *Client) sendDataToP2PUser(msg bean.RequestMessage, status bool, data string) error {
	msg.SenderUserPeerId = client.User.PeerId
	msg.SenderNodePeerId = client.User.P2PLocalPeerId
	if tool.CheckIsString(&msg.RecipientUserPeerId) && tool.CheckIsString(&msg.RecipientNodePeerId) {
		//if they at the same obd node
		if msg.RecipientNodePeerId == P2PLocalNodeId {
			if _, err := FindUserOnLine(msg); err == nil {
				itemClient := GlobalWsClientManager.OnlineClientMap[msg.RecipientUserPeerId]
				if itemClient != nil && itemClient.User != nil {
					if status {
						retData, isGoOn, err := routerOfP2PNode(msg, data, itemClient)
						if isGoOn == false {
							return nil
						}
						if err != nil {
							return err
						} else {
							if tool.CheckIsString(&retData) {
								data = retData
							}
						}
						data = p2pMiddleNodeTransferData(&msg, *itemClient, data, retData)
						if len(data) == 0 {
							return nil
						}
					}
					fromId := msg.SenderUserPeerId + "@" + P2pChannelMap[msg.SenderNodePeerId].Address
					toId := msg.RecipientUserPeerId + "@" + P2pChannelMap[msg.RecipientNodePeerId].Address
					jsonMessage := getP2PReplyObj(data, msg.Type, status, fromId, toId)
					if itemClient.SendChannel != nil {
						itemClient.SendChannel <- jsonMessage
					}
					if itemClient.IsGRpcRequest && itemClient.GrpcChan != nil {
						if msg.Type == enum.MsgType_RecvChannelAccept_33 ||
							msg.Type == enum.MsgType_HTLC_FinishTransferH_43 {
							go func() {
								itemClient.GrpcChan <- jsonMessage
							}()
						}
					}
					return nil
				}
			}
		} else { //at the different obd node,p2p transfer msg to other node
			msgToOther := bean.RequestMessage{}
			msgToOther.Type = msg.Type
			msgToOther.SenderNodePeerId = P2PLocalNodeId
			msgToOther.SenderUserPeerId = msg.SenderUserPeerId
			msgToOther.RecipientUserPeerId = msg.RecipientUserPeerId
			msgToOther.RecipientNodePeerId = msg.RecipientNodePeerId
			msgToOther.Data = data
			bytes, err := json.Marshal(msgToOther)
			if err == nil {
				return sendP2PMsg(msg.RecipientNodePeerId, string(bytes))
			}
		}
	}
	return errors.New(fmt.Sprintf(enum.Tips_user_notExistOrOnline, msg.RecipientUserPeerId))
}

//当p2p收到消息后
func getDataFromP2PSomeone(msg bean.RequestMessage) error {
	if tool.CheckIsString(&msg.RecipientUserPeerId) && tool.CheckIsString(&msg.RecipientNodePeerId) {
		if msg.RecipientNodePeerId == P2PLocalNodeId {
			if _, err := FindUserOnLine(msg); err == nil {
				itemClient := GlobalWsClientManager.OnlineClientMap[msg.RecipientUserPeerId]
				if itemClient != nil && itemClient.User != nil {
					//收到数据后，需要对其进行加工
					retData, isGoOn, err := routerOfP2PNode(msg, msg.Data, itemClient)
					if isGoOn == false {
						return nil
					}
					if err != nil {
						return err
					} else {
						if tool.CheckIsString(&retData) {
							msg.Data = retData
						}
					}

					msg.Data = p2pMiddleNodeTransferData(&msg, *itemClient, msg.Data, retData)
					if len(msg.Data) == 0 {
						return nil
					}

					fromId := msg.SenderUserPeerId + "@" + P2pChannelMap[msg.SenderNodePeerId].Address
					toId := msg.RecipientUserPeerId + "@" + P2pChannelMap[msg.RecipientNodePeerId].Address
					jsonMessage := getP2PReplyObj(msg.Data, msg.Type, true, fromId, toId)
					if itemClient.SendChannel != nil {
						itemClient.SendChannel <- jsonMessage
					}
					if itemClient.IsGRpcRequest && itemClient.GrpcChan != nil {
						if msg.Type == enum.MsgType_RecvChannelAccept_33 ||
							msg.Type == enum.MsgType_HTLC_FinishTransferH_43 {
							go func() {
								itemClient.GrpcChan <- jsonMessage
							}()
						}
					}
					return nil
				}
			}
		}
	}
	return errors.New(fmt.Sprintf(enum.Tips_user_notExistOrOnline, msg.RecipientUserPeerId))
}

func getReplyObj(data string, msgType enum.MsgType, status bool, fromClient, toClient *Client) []byte {
	var jsonMessage []byte

	fromId := fromClient.Id
	if fromClient.User != nil {
		fromId = fromClient.User.PeerId
	}

	toClientId := "all"
	if toClient != nil {
		toClientId = toClient.Id
		if toClient.User != nil {
			toClientId = toClient.User.PeerId
		}
	}

	if strings.Contains(fromId, "@/") == false {
		fromId = fromId + "@" + localServerDest
	}

	parse := gjson.Parse(data)
	result := parse.Value()
	if strings.HasPrefix(data, "{") == false && strings.HasPrefix(data, "[") == false {
		result = data
	}

	jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromId, To: toClientId, Result: result})

	return jsonMessage
}

func getP2PReplyObj(data string, msgType enum.MsgType, status bool, fromId, toClientId string) []byte {
	parse := gjson.Parse(data)
	result := parse.Value()
	if strings.HasPrefix(data, "{") == false && strings.HasPrefix(data, "[") == false {
		result = data
	}

	jsonMessage, _ := json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromId, To: toClientId, Result: result})
	return jsonMessage
}
