package rpc

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"log"
)

var (
	loginChan            chan bean.ReplyMessage
	logoutChan           chan bean.ReplyMessage
	updateLoginTokenChan chan bean.ReplyMessage
)

func readDataFromObd() {
	for {
		_, message, err := connObd.ReadMessage()
		if err != nil {
			connObd = nil
			return
		}

		replyMessage := bean.ReplyMessage{}
		err = json.Unmarshal(message, &replyMessage)
		if err == nil {
			log.Println(replyMessage)
			switch replyMessage.Type {
			case enum.MsgType_UserLogin_2001:
				loginChan <- replyMessage
			case enum.MsgType_User_UpdateAdminToken_2008:
				updateLoginTokenChan <- replyMessage
			case enum.MsgType_UserLogout_2002:
				logoutChan <- replyMessage
			}
		}
	}
}

func sendMsgToObd(info interface{}, msgType enum.MsgType) {
	var infoBytes []byte
	if info != nil {
		infoBytes, _ = json.Marshal(info)
	}
	requestMessage := bean.RequestMessage{Data: string(infoBytes), Type: msgType}
	msg, _ := json.Marshal(requestMessage)
	err := connObd.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
}
