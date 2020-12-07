package lightclient

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
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
	case enum.MsgType_UserLogin_2001:
		mnemonic := gjson.Get(msg.Data, "mnemonic").String()
		if client.User != nil {
			if client.User.Mnemonic != mnemonic {
				_ = service.UserService.UserLogout(client.User)
				delete(globalWsClientManager.OnlineClientMap, client.User.PeerId)
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
				P2PLocalPeerId:  p2PLocalPeerId,
			}
			var err error = nil
			peerId := tool.GetUserPeerId(user.Mnemonic)
			if globalWsClientManager.OnlineClientMap[peerId] != nil {
				err = errors.New("user has logined at other node")
			} else {
				err = service.UserService.UserLogin(&user)
			}
			if err == nil {
				client.User = &user
				globalWsClientManager.OnlineClientMap[user.PeerId] = client
				service.OnlineUserMap[user.PeerId] = &user
				data = loginRetData(*client)
				status = true
				client.sendToMyself(msg.Type, status, data)
				sendType = enum.SendTargetType_SendToExceptMe
			} else {
				client.sendToMyself(msg.Type, status, err.Error())
				sendType = enum.SendTargetType_SendToSomeone
			}
		}
	case enum.MsgType_UserLogout_2002:
		if client.User != nil {
			data = client.User.PeerId + " logout"
			status = true
			client.sendToMyself(msg.Type, status, "logout success")
			if client.User != nil {
				delete(globalWsClientManager.OnlineClientMap, client.User.PeerId)
				delete(service.OnlineUserMap, client.User.PeerId)
			}
			sendType = enum.SendTargetType_SendToExceptMe
			client.User = nil
		} else {
			client.sendToMyself(msg.Type, status, "please login")
			sendType = enum.SendTargetType_SendToSomeone
		}
	case enum.MsgType_p2p_ConnectPeer_2003:
		remoteNodeAddress := gjson.Get(msg.Data, "remote_node_address")
		if remoteNodeAddress.Exists() == false {
			data = errors.New("remote_node_address not exist").Error()
		} else {
			localP2PAddress, err := connP2PServer(remoteNodeAddress.Str)
			if err != nil {
				data = err.Error()
			} else {
				status = true
				data = localP2PAddress
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_GetObdNodeInfo_2005:
		bytes, err := json.Marshal(bean.CurrObdNodeInfo)
		if err != nil {
			data = err.Error()
		} else {
			status = true
			data = string(bytes)
		}
		client.sendToMyself(msg.Type, true, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_GetMiniBtcFundAmount_2006:
		fee := service.GetBtcMinerFundMiniAmount()
		data = tool.FloatToString(fee, 8)
		client.sendToMyself(msg.Type, true, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_HeartBeat_2007:
		client.sendToMyself(msg.Type, true, data)
		sendType = enum.SendTargetType_SendToSomeone
	// Process GetMnemonic
	case enum.MsgType_GetMnemonic_2004:
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
