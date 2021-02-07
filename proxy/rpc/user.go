package rpc

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/proxy/pb"
	"log"
	"net/url"
)

var connObd *websocket.Conn
var loginChan chan bean.ReplyMessage

type loginInfo struct {
	Mnemonic string `json:"mnemonic"`
	IsAdmin  bool   `json:"is_admin"`
}

type UserRpc struct {
}

func (user *UserRpc) Login(ctx context.Context, in *pb.LoginRequest) (resp *pb.LoginResponse, err error) {
	loginChan = make(chan bean.ReplyMessage)
	defer close(loginChan)

	if connObd == nil {
		u := url.URL{Scheme: "ws", Host: "127.0.0.1:60020", Path: "/ws" + config.ChainNodeType}
		log.Printf("begin to connect to tracker: %s", u.String())

		connObd, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Println("fail to dial tracker:", err)
			return nil, err
		}
		go readDataFromObd()
	}

	info := loginInfo{Mnemonic: in.Mnemonic, IsAdmin: true}
	marshal, _ := json.Marshal(info)
	requestMessage := bean.RequestMessage{Data: string(marshal), Type: enum.MsgType_UserLogin_2001}
	marshal, _ = json.Marshal(requestMessage)
	sendMsgToObd(marshal)

	data := <-loginChan

	node := data.Result.(map[string]interface{})
	resp = &pb.LoginResponse{
		UserPeerId:    node["userPeerId"].(string),
		NodePeerId:    node["nodePeerId"].(string),
		NodeAddress:   node["nodeAddress"].(string),
		HtlcFeeRate:   node["htlcFeeRate"].(float64),
		HtlcMaxFee:    node["htlcMaxFee"].(float64),
		ChainNodeType: node["chainNodeType"].(string),
	}
	return resp, nil
}

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
			}
		}
	}
}

func sendMsgToObd(msg []byte) {
	err := connObd.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
}
