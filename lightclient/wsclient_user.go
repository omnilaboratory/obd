package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
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
				PeerId:   gjson.Get(msg.Data, "peer_id").String(),
				Password: gjson.Get(msg.Data, "password").String(),
			}
			err := service.UserService.UserLogin(&user)
			if err == nil {
				if client.User != nil {
					GlobalWsClientManager.OnlineUserMap[user.PeerId] = client
					service.OnlineUserMap[user.PeerId] = true
				}
				data = client.User.PeerId + " login"
				status = true
				sendType = enum.SendTargetType_SendToAll
			} else {
				client.sendToMyself(msg.Type, true, err.Error())
				sendType = enum.SendTargetType_SendToSomeone
			}
		}
	case enum.MsgType_UserLogout_2:
		if client.User != nil {
			data = client.User.PeerId + " logout"
			sendType = enum.SendTargetType_SendToAll
			if client.User != nil {
				delete(GlobalWsClientManager.OnlineUserMap, client.User.PeerId)
				delete(service.OnlineUserMap, client.User.PeerId)
			}
			client.User = nil
			status = true
		} else {
			client.sendToMyself(msg.Type, true, "please login")
			sendType = enum.SendTargetType_SendToSomeone
		}

	// Added by Kevin 2019-11-25
	case enum.MsgType_UserSignUp:
		
	}

	return sendType, []byte(data), status
}
