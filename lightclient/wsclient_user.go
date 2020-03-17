package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"LightningOnOmni/tool"
	"errors"
	"github.com/tidwall/gjson"
)

func (client *Client) userModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	var data string
	switch msg.Type {
	case enum.MsgType_UserLogin_1:
		if client.User != nil {
			client.sendToMyself(msg.Type, true, "already login")
			sendType = enum.SendTargetType_SendToSomeone
		} else {
			user := bean.User{
				Mnemonic: gjson.Get(msg.Data, "mnemonic").String(),
			}
			var err error = nil
			peerId := tool.SignMsgWithSha256([]byte(user.Mnemonic))
			if GlobalWsClientManager.OnlineUserMap[peerId] != nil {
				err = errors.New("user has login")
			} else {
				err = service.UserService.UserLogin(&user)
			}
			if err == nil {
				client.User = &user
				GlobalWsClientManager.OnlineUserMap[user.PeerId] = client
				service.OnlineUserMap[user.PeerId] = true
				data = user.PeerId + " login"
				status = true
				client.sendToMyself(msg.Type, status, "login success")
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
		ConnP2PServer(msg.Data)
	case enum.MsgType_p2p_SendDataToServer_4:
		ClientSendMsg(msg.Data)
	case enum.MsgType_p2p_SendDataToClient_5:
		ServerSendMsg(msg.Data)
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
