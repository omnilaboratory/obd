package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"LightningOnOmni/tool"
	"encoding/json"
	"strconv"
)

func (c *Client) channelModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	//get openChannelReq from funder then send to fundee
	case enum.MsgType_ChannelOpen:
		if tool.CheckIsString(&msg.RecipientPeerId) == false {
			data = "no target fundee"
		} else {
			if msg.RecipientPeerId == c.User.PeerId {
				data = "can not open channel to yourself"
			} else {
				node, err := service.ChannelService.OpenChannel(msg, c.User.PeerId)
				if err != nil {
					data = err.Error()
				} else {
					bytes, err := json.Marshal(node)
					if err != nil {
						data = err.Error()
					} else {
						data = string(bytes)
						status = true
					}
				}
			}
		}
		c.sendToSomeone(msg.Type, status, msg.RecipientPeerId, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_ChannelOpen_ItemByTempId:
		node, err := service.ChannelService.GetChannelByTemporaryChanId(msg.Data)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_ChannelOpen_DelItemByTempId:
		node, err := service.ChannelService.DelChannelByTemporaryChanId(msg.Data)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_ChannelOpen_AllItem:
		nodes, err := service.ChannelService.AllItem(c.User.PeerId)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(nodes)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_ChannelOpen_Count:
		node, err := service.ChannelService.TotalCount()
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(node)
			status = true
		}
		c.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	//get acceptChannelReq from fundee then send to funder
	case enum.MsgType_ChannelAccept:
		if c.User == nil {
			c.sendToMyself(msg.Type, true, "please login first")
			sendType = enum.SendTargetType_SendToSomeone
		} else {
			node, err := service.ChannelService.AcceptChannel(msg.Data, c.User.PeerId)
			if err != nil {
				data = err.Error()
			} else {
				bytes, err := json.Marshal(node)
				if err != nil {
					data = err.Error()
				} else {
					data = string(bytes)
					status = true
				}
			}
			if node != nil {
				c.sendToSomeone(msg.Type, status, node.FunderPeerId, data)
			}
			c.sendToMyself(msg.Type, status, data)
			sendType = enum.SendTargetType_SendToSomeone
		}
	}
	return sendType, []byte(data), status
}
