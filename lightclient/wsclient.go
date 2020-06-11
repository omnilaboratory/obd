package lightclient

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"log"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

func (client *Client) Write() {
	defer func() {
		e := client.Socket.Close()
		if e != nil {
			log.Println(e)
		} else {
			log.Println("socket closed after writing...")
		}
	}()

	for {
		select {
		case data, ok := <-client.SendChannel:
			if !ok {
				_ = client.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Println("send data", string(data))
			_ = client.Socket.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func (client *Client) Read() {
	defer func() {
		_ = service.UserService.UserLogout(client.User)
		if client.User != nil {
			delete(globalWsClientManager.OnlineUserMap, client.User.PeerId)
			delete(service.OnlineUserMap, client.User.PeerId)
			client.User = nil
		}
		globalWsClientManager.Disconnected <- client
		_ = client.Socket.Close()
		log.Println("socket closed after reading...")
	}()

	for {
		_, dataReq, err := client.Socket.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var msg bean.RequestMessage
		log.Println("request data: ", string(dataReq))
		parse := gjson.Parse(string(dataReq))

		if parse.Exists() == false {
			log.Println("wrong json input")
			client.sendToMyself(enum.MsgType_Error_0, false, string(dataReq))
			continue
		}

		msg.Type = enum.MsgType(parse.Get("type").Int())
		if enum.CheckExist(msg.Type) == false {
			data := "not exist the msg type"
			log.Println(data)
			client.sendToMyself(msg.Type, false, data)
			continue
		}

		msg.Data = parse.Get("data").String()
		msg.RawData = string(dataReq)
		msg.SenderUserPeerId = parse.Get("sender_user_peer_id").String()
		msg.RecipientUserPeerId = parse.Get("recipient_user_peer_id").String()
		msg.RecipientNodePeerId = parse.Get("recipient_node_peer_id").String()
		msg.PubKey = parse.Get("pub_key").String()
		msg.Signature = parse.Get("signature").String()

		// check the Recipient is online
		if tool.CheckIsString(&msg.RecipientUserPeerId) {
			_, err := client.FindUser(&msg.RecipientUserPeerId)
			if err != nil {
				if tool.CheckIsString(&msg.RecipientNodePeerId) == false {
					client.sendToMyself(msg.Type, false, "can not find target user")
					continue
				}
			}
		}

		// check the data whether is right signature
		if tool.CheckIsString(&msg.PubKey) && tool.CheckIsString(&msg.Signature) {
			rpcClient := rpc.NewClient()
			result, err := rpcClient.VerifyMessage(msg.PubKey, msg.Signature, msg.Data)
			if err != nil {
				client.sendToMyself(msg.Type, false, err.Error())
				continue
			}
			if gjson.Parse(result).Bool() == false {
				client.sendToMyself(msg.Type, false, "error signature")
				continue
			}
		}

		var sendType = enum.SendTargetType_SendToNone
		status := false
		var dataOut []byte
		var needLogin = true

		if msg.Type <= enum.MsgType_UserLogin_2001 && msg.Type >= enum.MsgType_Common_End_2999 {
			needLogin = false
		}

		if msg.Type <= enum.MsgType_UserLogin_2001 && msg.Type > enum.MsgType_User_End_2099 {
			if msg.Type == enum.MsgType_GetMnemonic_2004 {
				sendType, dataOut, status = client.hdWalletModule(msg)
			} else {
				sendType, dataOut, status = client.userModule(msg)
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
				client.sendToMyself(msg.Type, false, "please login")
				continue
			} else { // already login

				if msg.Type == enum.MsgType_SendChannelOpen_32 || msg.Type == enum.MsgType_SendChannelAccept_33 ||
					msg.Type == enum.MsgType_FundingCreate_SendBtcFundingCreated_340 || msg.Type == enum.MsgType_FundingSign_SendBtcSign_350 ||
					msg.Type == enum.MsgType_FundingCreate_SendAssetFundingCreated_34 || msg.Type == enum.MsgType_FundingSign_SendAssetFundingSigned_35 ||
					msg.Type == enum.MsgType_CommitmentTx_SendCommitmentTransactionCreated_351 ||
					msg.Type == enum.MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352 ||
					msg.Type == enum.MsgType_HTLC_SendAddHTLC_40 || msg.Type == enum.MsgType_HTLC_SendAddHTLCSigned_41 ||
					msg.Type == enum.MsgType_HTLC_SendVerifyR_45 || msg.Type == enum.MsgType_HTLC_SendSignVerifyR_46 ||
					msg.Type == enum.MsgType_HTLC_SendRequestCloseCurrTx_49 || msg.Type == enum.MsgType_HTLC_SendCloseSigned_50 ||
					msg.Type == enum.MsgType_Atomic_SendSwap_80 || msg.Type == enum.MsgType_Atomic_SendSwapAccept_81 ||
					msg.Type == enum.MsgType_SendCloseChannelRequest_38 || msg.Type == enum.MsgType_SendCloseChannelSign_39 {
					if tool.CheckIsString(&msg.RecipientUserPeerId) == false {
						client.sendToMyself(msg.Type, false, "error recipient_user_peer_id")
						continue
					}
					if tool.CheckIsString(&msg.RecipientNodePeerId) == false {
						client.sendToMyself(msg.Type, false, "error recipient_node_peer_id")
						continue
					}
					if p2pChannelMap[msg.RecipientNodePeerId] == nil {
						client.sendToMyself(msg.Type, false, "not connect recipient_node_peer_id")
						continue
					}
					if msg.RecipientNodePeerId == P2PLocalPeerId {
						if _, err = FindUserOnLine(&msg.RecipientUserPeerId); err != nil {
							client.sendToMyself(msg.Type, false, err.Error())
							continue
						}
					} else {
						if flag := service.HttpGetUserStateFromTracker(msg.RecipientUserPeerId); flag == 0 {
							client.sendToMyself(msg.Type, false, errors.New("recipient_user_peer_id not online").Error())
							continue
						}
					}
				}

				msg.SenderUserPeerId = client.User.PeerId
				for {
					//-3000 -3001
					if msg.Type <= enum.MsgType_Mnemonic_CreateAddress_3000 &&
						msg.Type >= enum.MsgType_Mnemonic_GetAddressByIndex_3001 {
						sendType, dataOut, status = client.hdWalletModule(msg)
						break
					}

					//-32 -33  及查询
					if msg.Type == enum.MsgType_SendChannelOpen_32 ||
						msg.Type == enum.MsgType_SendChannelAccept_33 ||
						(msg.Type <= enum.MsgType_ChannelOpen_AllItem_3150 &&
							msg.Type >= enum.MsgType_GetChannelInfoByDbId_3155) {
						sendType, dataOut, status = client.channelModule(msg)
						break
					}

					//-34 -340 及查询
					if msg.Type == enum.MsgType_FundingCreate_SendAssetFundingCreated_34 ||
						msg.Type == enum.MsgType_FundingCreate_SendBtcFundingCreated_340 ||
						(msg.Type <= enum.MsgType_FundingCreate_Asset_AllItem_3100 &&
							msg.Type >= enum.MsgType_FundingCreate_Btc_ItemRDByTempChannelIdAndTxId_3110) {
						sendType, dataOut, status = client.fundingTransactionModule(msg)
						break
					}

					//-35 -350
					if msg.Type == enum.MsgType_FundingSign_SendAssetFundingSigned_35 ||
						msg.Type == enum.MsgType_FundingSign_SendBtcSign_350 {
						sendType, dataOut, status = client.fundingSignModule(msg)
						break
					}

					//-351 及查询
					if msg.Type == enum.MsgType_CommitmentTx_SendCommitmentTransactionCreated_351 ||
						(msg.Type <= enum.MsgType_CommitmentTx_ItemsByChanId_3200 &&
							msg.Type >= enum.MsgType_CommitmentTx_AllBRByChanId_3208) {
						sendType, dataOut, status = client.commitmentTxModule(msg)
						break
					}

					//-352
					if msg.Type == enum.MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352 {
						sendType, dataOut, status = client.commitmentTxSignModule(msg)
						break
					}

					//-38 -39
					if msg.Type == enum.MsgType_SendCloseChannelRequest_38 ||
						msg.Type == enum.MsgType_SendCloseChannelSign_39 {
						sendType, dataOut, status = client.channelModule(msg)
						break
					}

					//-40 -41
					if msg.Type == enum.MsgType_HTLC_FindPath_401 ||
						msg.Type == enum.MsgType_HTLC_Invoice_402 ||
						msg.Type == enum.MsgType_HTLC_SendAddHTLC_40 ||
						msg.Type == enum.MsgType_HTLC_SendAddHTLCSigned_41 {
						sendType, dataOut, status = client.htlcHModule(msg)
						break
					}

					//-42 -43 -44 -45 -46 -47
					if msg.Type == enum.MsgType_HTLC_SendVerifyR_45 ||
						msg.Type == enum.MsgType_HTLC_SendSignVerifyR_46 {
						sendType, dataOut, status = client.htlcTxModule(msg)
						break
					}

					// -48 -49
					if msg.Type == enum.MsgType_HTLC_SendRequestCloseCurrTx_49 ||
						msg.Type == enum.MsgType_HTLC_SendCloseSigned_50 {
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

		if len(dataOut) == 0 {
			dataOut = dataReq
		}

		//broadcast except me
		if sendType == enum.SendTargetType_SendToExceptMe {
			for itemClient := range globalWsClientManager.ClientsMap {
				if itemClient != client {
					jsonMessage := getReplyObj(string(dataOut), msg.Type, status, client, itemClient)
					itemClient.SendChannel <- jsonMessage
				}
			}
		}
		//broadcast to all
		if sendType == enum.SendTargetType_SendToAll {
			jsonMessage := getReplyObj(string(dataOut), msg.Type, status, client, nil)
			globalWsClientManager.Broadcast <- jsonMessage
		}
	}
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
	node := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &node)
	if err == nil {
		parse := gjson.Parse(data)
		jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromId, To: toClientId, Result: parse.Value()})
	} else {
		if strings.Contains(err.Error(), " array into Go value of type map") {
			parse := gjson.Parse(data)
			jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromId, To: toClientId, Result: parse.Value()})
		} else {
			jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromId, To: toClientId, Result: data})
		}
	}
	return jsonMessage
}

func getP2PReplyObj(data string, msgType enum.MsgType, status bool, fromId, toClientId string) []byte {

	parse := gjson.Parse(data)
	result := parse.Value()
	if parse.Exists() == false {
		result = data
	}
	jsonMessage, _ := json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromId, To: toClientId, Result: result})
	return jsonMessage
}

func (client *Client) sendToMyself(msgType enum.MsgType, status bool, data string) {
	jsonMessage := getReplyObj(data, msgType, status, client, client)
	client.SendChannel <- jsonMessage
}

func (client *Client) sendToSomeone(msgType enum.MsgType, status bool, recipientPeerId string, data string) error {
	if tool.CheckIsString(&recipientPeerId) {
		if _, err := client.FindUser(&recipientPeerId); err == nil {
			itemClient := globalWsClientManager.OnlineUserMap[recipientPeerId]
			if itemClient != nil && itemClient.User != nil {
				jsonMessage := getReplyObj(data, msgType, status, client, itemClient)
				itemClient.SendChannel <- jsonMessage
				return nil
			}
		}
	}
	return errors.New("recipient not exist or online")
}

//发送消息给对方，分为同节点和不同节点的两种情况
func (client *Client) sendDataToP2PUser(msg bean.RequestMessage, status bool, data string) error {
	msg.SenderUserPeerId = client.User.PeerId
	msg.SenderNodePeerId = client.User.P2PLocalPeerId
	if tool.CheckIsString(&msg.RecipientUserPeerId) && tool.CheckIsString(&msg.RecipientNodePeerId) {
		//如果是同一个obd节点
		if msg.RecipientNodePeerId == P2PLocalPeerId {
			if _, err := FindUserOnLine(&msg.RecipientUserPeerId); err == nil {
				itemClient := globalWsClientManager.OnlineUserMap[msg.RecipientUserPeerId]
				if itemClient != nil && itemClient.User != nil {
					//因为数据库，分库，需要对特定的消息进行处理
					if status {
						//收到请求后，首先对消息进行处理
						retData, err := routerOfP2PNode(msg.Type, data, itemClient)
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
					fromId := msg.SenderUserPeerId + "@" + p2pChannelMap[msg.SenderNodePeerId].Address
					toId := msg.RecipientUserPeerId + "@" + p2pChannelMap[msg.RecipientNodePeerId].Address
					jsonMessage := getP2PReplyObj(data, msg.Type, status, fromId, toId)
					itemClient.SendChannel <- jsonMessage
					return nil
				}
			}
		} else { //不通的p2p的节点 需要转发到对方的节点
			msgToOther := bean.RequestMessage{}
			msgToOther.Type = msg.Type
			msgToOther.SenderNodePeerId = P2PLocalPeerId
			msgToOther.SenderUserPeerId = msg.SenderUserPeerId
			msgToOther.RecipientUserPeerId = msg.RecipientUserPeerId
			msgToOther.RecipientNodePeerId = msg.RecipientNodePeerId
			msgToOther.Data = data
			bytes, err := json.Marshal(msgToOther)
			if err == nil {
				return SendP2PMsg(msg.RecipientNodePeerId, string(bytes))
			}
		}
	}
	return errors.New("recipient not exist or online")
}

//当p2p收到消息后
func getDataFromP2PSomeone(msg bean.RequestMessage) error {
	if tool.CheckIsString(&msg.RecipientUserPeerId) && tool.CheckIsString(&msg.RecipientNodePeerId) {
		if msg.RecipientNodePeerId == P2PLocalPeerId {
			if _, err := FindUserOnLine(&msg.RecipientUserPeerId); err == nil {
				itemClient := globalWsClientManager.OnlineUserMap[msg.RecipientUserPeerId]
				if itemClient != nil && itemClient.User != nil {
					//收到数据后，需要对其进行加工
					retData, err := routerOfP2PNode(msg.Type, msg.Data, itemClient)
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

					fromId := msg.SenderUserPeerId + "@" + p2pChannelMap[msg.SenderNodePeerId].Address
					toId := msg.RecipientUserPeerId + "@" + p2pChannelMap[msg.RecipientNodePeerId].Address
					jsonMessage := getP2PReplyObj(msg.Data, msg.Type, true, fromId, toId)
					itemClient.SendChannel <- jsonMessage
					return nil
				}
			}
		}
	}
	return errors.New("recipient not exist or online")
}

func p2pMiddleNodeTransferData(msg *bean.RequestMessage, itemClient Client, data string, retData string) string {
	//需要节点之间本身的通信 bob的节点响应352后，发送353到alice节点，353处理完成后，需要对353的结果消息进行分发
	//当前的353消息本身是从bob发给Alice的

	if msg.Type == enum.MsgType_ChannelOpen_32 {
		msg.Type = enum.MsgType_RecvChannelOpen_32
	}

	if msg.Type == enum.MsgType_ChannelAccept_33 {
		msg.Type = enum.MsgType_RecvChannelAccept_33
	}

	if msg.Type == enum.MsgType_FundingCreate_AssetFundingCreated_34 {
		msg.Type = enum.MsgType_FundingCreate_RecvAssetFundingCreated_34
	}

	if msg.Type == enum.MsgType_FundingSign_AssetFundingSigned_35 {
		msg.Type = enum.MsgType_FundingSign_RecvAssetFundingSigned_35
	}

	if msg.Type == enum.MsgType_FundingCreate_BtcFundingCreated_340 {
		msg.Type = enum.MsgType_FundingCreate_RecvBtcFundingCreated_340
	}

	if msg.Type == enum.MsgType_FundingSign_BtcSign_350 {
		msg.Type = enum.MsgType_FundingSign_RecvBtcSign_350
	}

	if msg.Type == enum.MsgType_CommitmentTx_CommitmentTransactionCreated_351 {
		msg.Type = enum.MsgType_CommitmentTx_RecvCommitmentTransactionCreated_351
	}

	if msg.Type == enum.MsgType_CloseChannelRequest_38 {
		msg.Type = enum.MsgType_RecvCloseChannelRequest_38
	}

	if msg.Type == enum.MsgType_CloseChannelSign_39 {
		msg.Type = enum.MsgType_RecvCloseChannelSign_39
	}

	if msg.Type == enum.MsgType_HTLC_AddHTLC_40 {
		msg.Type = enum.MsgType_HTLC_RecvAddHTLC_40
	}

	if msg.Type == enum.MsgType_CommitmentTxSigned_ToAliceSign_353 {
		//	发给bob的信息
		if gjson.Parse(retData).Get("bobData").Exists() {
			newMsg := bean.RequestMessage{}
			newMsg.Type = enum.MsgType_CommitmentTxSigned_SecondToBobSign_354
			newMsg.SenderUserPeerId = itemClient.User.PeerId
			newMsg.SenderNodePeerId = P2PLocalPeerId
			newMsg.RecipientUserPeerId = msg.SenderUserPeerId
			newMsg.RecipientNodePeerId = msg.SenderNodePeerId
			payeeData := gjson.Parse(retData).Get("bobData").String()
			//转发给bob，
			_ = itemClient.sendDataToP2PUser(newMsg, true, payeeData)
		}

		//发给alice
		msg.Type = enum.MsgType_CommitmentTxSigned_RecvRevokeAndAcknowledgeCommitmentTransaction_352
		payerData := gjson.Parse(retData).Get("aliceData").String()
		data = payerData
	}

	//当354处理完成，就改成352的返回 353和354对用户是透明的
	if msg.Type == enum.MsgType_CommitmentTxSigned_SecondToBobSign_354 {
		msg.Type = enum.MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352
		msg.SenderUserPeerId = msg.RecipientUserPeerId
		msg.SenderNodePeerId = msg.RecipientNodePeerId
	}

	if msg.Type == enum.MsgType_HTLC_PayerSignC3b_42 {
		newMsg := bean.RequestMessage{}
		newMsg.Type = enum.MsgType_HTLC_PayeeCreateHTRD1a_43
		newMsg.SenderUserPeerId = itemClient.User.PeerId
		newMsg.SenderNodePeerId = P2PLocalPeerId
		newMsg.RecipientUserPeerId = msg.SenderUserPeerId
		newMsg.RecipientNodePeerId = msg.SenderNodePeerId
		newMsg.Data = data
		//转发给bob
		_ = itemClient.sendDataToP2PUser(newMsg, true, newMsg.Data)
		return ""
	}

	//当43处理完成，就改成41的返回 42和43对用户是透明的
	if msg.Type == enum.MsgType_HTLC_PayeeCreateHTRD1a_43 {
		newMsg := bean.RequestMessage{}
		newMsg.Type = enum.MsgType_HTLC_PayerSignHTRD1a_44
		newMsg.SenderUserPeerId = itemClient.User.PeerId
		newMsg.SenderNodePeerId = P2PLocalPeerId
		newMsg.RecipientUserPeerId = msg.SenderUserPeerId
		newMsg.RecipientNodePeerId = msg.SenderNodePeerId
		newMsg.Data = data
		//转发给payer alice，
		_ = itemClient.sendDataToP2PUser(newMsg, true, newMsg.Data)
		return ""
	}

	//当44处理完成，就改成41的返回 42和43对用户是透明的
	if msg.Type == enum.MsgType_HTLC_PayerSignHTRD1a_44 {
		msg.Type = enum.MsgType_HTLC_RecvAddHTLCSigned_41
	}

	if msg.Type == enum.MsgType_HTLC_VerifyR_45 {
		msg.Type = enum.MsgType_HTLC_RecvVerifyR_45
	}

	//当47处理完成，发送48号协议给收款方
	if msg.Type == enum.MsgType_HTLC_SendHerdHex_47 {

		newMsg := bean.RequestMessage{}
		newMsg.Type = enum.MsgType_HTLC_SignHedHex_48
		newMsg.SenderUserPeerId = itemClient.User.PeerId
		newMsg.SenderNodePeerId = P2PLocalPeerId
		newMsg.RecipientUserPeerId = msg.SenderUserPeerId
		newMsg.RecipientNodePeerId = msg.SenderNodePeerId
		payerData := gjson.Parse(retData).Get("payerData").String()
		//转发给payer alice，
		_ = itemClient.sendDataToP2PUser(newMsg, true, payerData)

		//发给bob的
		msg.Type = enum.MsgType_HTLC_RecvSignVerifyR_46
		payeeData := gjson.Parse(retData).Get("payeeData").String()
		data = payeeData
	}

	//当48理完成，就改成46的返回 47和48对用户是透明的
	if msg.Type == enum.MsgType_HTLC_SignHedHex_48 {
		msg.SenderUserPeerId = msg.RecipientUserPeerId
		msg.SenderNodePeerId = msg.RecipientNodePeerId
		msg.Type = enum.MsgType_HTLC_SendSignVerifyR_46
	}

	if msg.Type == enum.MsgType_HTLC_RequestCloseCurrTx_49 {
		msg.Type = enum.MsgType_HTLC_RecvRequestCloseCurrTx_49
	}

	if msg.Type == enum.MsgType_HTLC_CloseHtlcRequestSignBR_51 {
		//	发给bob的信息
		newMsg := bean.RequestMessage{}
		newMsg.Type = enum.MsgType_HTLC_CloseHtlcUpdateCnb_52
		newMsg.SenderUserPeerId = itemClient.User.PeerId
		newMsg.SenderNodePeerId = P2PLocalPeerId
		newMsg.RecipientUserPeerId = msg.SenderUserPeerId
		newMsg.RecipientNodePeerId = msg.SenderNodePeerId
		payeeData := gjson.Parse(retData).Get("bobData").String()
		//转发给bob，
		_ = itemClient.sendDataToP2PUser(newMsg, true, payeeData)
		//发给alice
		msg.Type = enum.MsgType_HTLC_RecvCloseSigned_50
		payerData := gjson.Parse(retData).Get("aliceData").String()
		data = payerData
	}

	if msg.Type == enum.MsgType_HTLC_CloseHtlcUpdateCnb_52 {
		msg.SenderUserPeerId = msg.RecipientUserPeerId
		msg.SenderNodePeerId = msg.RecipientNodePeerId
		msg.Type = enum.MsgType_HTLC_SendCloseSigned_50
	}

	if msg.Type == enum.MsgType_Atomic_Swap_N80 {
		msg.Type = enum.MsgType_Atomic_RecvSwap_N80
	}

	if msg.Type == enum.MsgType_Atomic_SwapAccept_N81 {
		msg.Type = enum.MsgType_Atomic_RecvSwapAccept_N81
	}

	return data
}

func (client *Client) FindUser(peerId *string) (*Client, error) {
	if tool.CheckIsString(peerId) {
		itemClient := globalWsClientManager.OnlineUserMap[*peerId]
		if itemClient != nil && itemClient.User != nil {
			return itemClient, nil
		}
	}
	return nil, errors.New("user not exist or online")
}
func FindUserOnLine(peerId *string) (*Client, error) {
	if tool.CheckIsString(peerId) {
		itemClient := globalWsClientManager.OnlineUserMap[*peerId]
		if itemClient != nil && itemClient.User != nil {
			return itemClient, nil
		}
	}
	return nil, errors.New(*peerId + " not exist or online")
}
