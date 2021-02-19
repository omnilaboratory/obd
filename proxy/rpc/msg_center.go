package rpc

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"log"
	"strings"
)

var (
	loginChan          = make(chan bean.ReplyMessage)
	logoutChan         = make(chan bean.ReplyMessage)
	changePasswordChan = make(chan bean.ReplyMessage)
	openChannelChan    = make(chan bean.ReplyMessage)
	fundChannelChan    = make(chan bean.ReplyMessage)
	rsmcChan           = make(chan bean.ReplyMessage)
	addInvoiceChan     = make(chan bean.ReplyMessage)
	payInvoiceChan     = make(chan bean.ReplyMessage)
)

func readDataFromObd() {
	for {
		log.Println("等待信息")
		_, message, err := connObd.ReadMessage()
		log.Println("收到信息")
		if err != nil {
			log.Println(err)
			connObd = nil
			return
		}
		log.Println(string(message))
		replyMessage := bean.ReplyMessage{}
		_ = json.Unmarshal(message, &replyMessage)

		//log.Println(replyMessage)
		if currUserInfo != nil {
			if strings.Contains(replyMessage.To, currUserInfo.UserPeerId) == false {
				continue
			}
		}

		switch replyMessage.Type {
		case enum.MsgType_UserLogin_2001:
			if strings.Contains(replyMessage.From, replyMessage.To) {
				loginChan <- replyMessage
			}
		case enum.MsgType_User_UpdateAdminToken_2008:
			changePasswordChan <- replyMessage
		case enum.MsgType_UserLogout_2002:
			logoutChan <- replyMessage

		case enum.MsgType_SendChannelOpen_32:
			openChannelChan <- replyMessage
		case enum.MsgType_RecvChannelAccept_33:
			openChannelChan <- replyMessage

		case enum.MsgType_FundingCreate_SendBtcFundingCreated_340:
			fundChannelChan <- replyMessage
		case enum.MsgType_FundingSign_RecvBtcSign_350:
			fundChannelChan <- replyMessage
		case enum.MsgType_ClientSign_AssetFunding_AliceSignRD_1134:
			fundChannelChan <- replyMessage

		case enum.MsgType_ClientSign_CommitmentTx_AliceSignC2a_360:
			rsmcChan <- replyMessage

		case enum.MsgType_HTLC_Invoice_402:
			addInvoiceChan <- replyMessage

		case enum.MsgType_HTLC_FindPath_401:
			if replyMessage.Status == false {
				payInvoiceChan <- replyMessage
			}
			break
		case enum.MsgType_HTLC_FinishTransferH_43:
			payInvoiceChan <- replyMessage
		default:
			continue
		}
	}
}

func sendMsgToObd(info interface{}, RecipientNodePeerId, RecipientUserPeerId string, msgType enum.MsgType) {
	var infoBytes []byte
	if info != nil {
		infoBytes, _ = json.Marshal(info)
	}
	requestMessage := bean.RequestMessage{Data: string(infoBytes), Type: msgType}
	requestMessage.RecipientNodePeerId = RecipientNodePeerId
	requestMessage.RecipientUserPeerId = RecipientUserPeerId
	msg, _ := json.Marshal(requestMessage)
	err := connObd.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
}
