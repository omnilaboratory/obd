package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"github.com/tidwall/gjson"
)

func (c *Client) userModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	var data string

	switch msg.Type {
	case enum.MsgType_UserLogin:
		if c.User != nil {
			c.sendToMyself(msg.Type, true, "already login")
			sendType = enum.SendTargetType_SendToSomeone
		} else {
			user := bean.User{
				PeerId: gjson.Get(msg.Data, "peer_id").String(),
			}
			if len(user.PeerId) > 0 {
				c.User = &user
				service.UserService.UserLogin(&user)
				data = c.User.PeerId + " login"
				sendType = enum.SendTargetType_SendToAll
				status = true
			} else {
				c.sendToMyself(msg.Type, true, "error peer_id")
				sendType = enum.SendTargetType_SendToSomeone
			}
		}
	case enum.MsgType_UserLogout:
		if c.User != nil {
			data = c.User.PeerId + " logout"
			c.User = nil
			sendType = enum.SendTargetType_SendToAll
			status = true
		} else {
			c.sendToMyself(msg.Type, true, "please login")
			sendType = enum.SendTargetType_SendToSomeone
		}
	}
	return sendType, []byte(data), status
}
