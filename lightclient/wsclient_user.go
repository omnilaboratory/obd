package lightclient

import (
	"encoding/json"
	"errors"
	"github.com/tidwall/gjson"
	"obd/bean"
	"obd/bean/enum"
	"obd/service"
	"obd/tool"
)

func loginRetData(client Client) string {
	retData := make(map[string]interface{})
	retData["userPeerId"] = client.User.PeerId
	retData["nodePeerId"] = client.User.P2PLocalPeerId
	retData["nodeAddress"] = client.User.P2PLocalAddress
	bytes, _ := json.Marshal(retData)
	return string(bytes)
}

func (client *Client) userModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	var data string
	switch msg.Type {
	case enum.MsgType_UserLogin_1:

		mnemonic := gjson.Get(msg.Data, "mnemonic").String()
		if client.User != nil {
			if client.User.Mnemonic != mnemonic {
				_ = service.UserService.UserLogout(client.User)
				delete(GlobalWsClientManager.OnlineUserMap, client.User.PeerId)
				delete(service.OnlineUserMap, client.User.PeerId)
				client.User = nil
			}
		}

		if client.User != nil {
			data = loginRetData(*client)
			client.sendToMyself(msg.Type, true, data)
			sendType = enum.SendTargetType_SendToSomeone
		} else {
			user := bean.User{
				Mnemonic:        mnemonic,
				P2PLocalAddress: localServerDest,
				P2PLocalPeerId:  P2PLocalPeerId,
			}
			var err error = nil
			peerId := tool.SignMsgWithSha256([]byte(user.Mnemonic))
			if GlobalWsClientManager.OnlineUserMap[peerId] != nil {
				err = errors.New("user has logined at other node")
			} else {
				err = service.UserService.UserLogin(&user)
			}
			if err == nil {
				client.User = &user
				GlobalWsClientManager.OnlineUserMap[user.PeerId] = client
				service.OnlineUserMap[user.PeerId] = true
				data = loginRetData(*client)
				status = true
				client.sendToMyself(msg.Type, status, data)
				sendType = enum.SendTargetType_SendToExceptMe
			} else {
				client.sendToMyself(msg.Type, status, err.Error())
				sendType = enum.SendTargetType_SendToSomeone
			}
		}
	case enum.MsgType_UserLogout_2:
		if client.User != nil {
			data = client.User.PeerId + " logout"
			status = true
			client.sendToMyself(msg.Type, status, "logout success")
			if client.User != nil {
				delete(GlobalWsClientManager.OnlineUserMap, client.User.PeerId)
				delete(service.OnlineUserMap, client.User.PeerId)
			}
			sendType = enum.SendTargetType_SendToExceptMe
			client.User = nil
		} else {
			client.sendToMyself(msg.Type, status, "please login")
			sendType = enum.SendTargetType_SendToSomeone
		}
	case enum.MsgType_p2p_ConnectServer_3:
		localP2PAddress, err := ConnP2PServer(msg.Data)
		if err != nil {
			data = err.Error()
		} else {
			status = true
			data = localP2PAddress
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_p2p_SendDataToServer_4:
		SendP2PMsg(msg.RecipientNodePeerId, msg.RawData)

	// Added by Kevin 2019-11-25
	// Process GetMnemonic
	case enum.MsgType_GetMnemonic_101:
		if client.User != nil { // The user already login.
			client.sendToMyself(msg.Type, true, "already login")
			sendType = enum.SendTargetType_SendToSomeone
		} else {
			// get Mnemonic
			mnemonic, err := service.HDWalletService.Bip39GenMnemonic(256)
			if err == nil { //get  successful.
				data = mnemonic
				status = true
			} else {
				data = err.Error()
			}
			client.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		}
	}
	return sendType, []byte(data), status
}
