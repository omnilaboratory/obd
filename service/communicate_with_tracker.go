package service

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	trackerBean "github.com/omnilaboratory/obd/tracker/bean"
	"github.com/tidwall/gjson"
	"strings"
)

var TrackerChan chan []byte

// sendChannelStateToTracker
func sendChannelStateToTracker(channelInfo dao.ChannelInfo, commitmentTx dao.CommitmentTransaction) {
	if channelInfo.IsPrivate {
		return
	}
	infoRequest := bean.ChannelInfoRequest{}
	infoRequest.ChannelId = channelInfo.ChannelId
	infoRequest.PropertyId = channelInfo.PropertyId
	infoRequest.CurrState = channelInfo.CurrState
	infoRequest.PeerIdA = channelInfo.PeerIdA
	infoRequest.PeerIdB = channelInfo.PeerIdB

	infoRequest.IsAlice = false
	if commitmentTx.Id > 0 {
		if commitmentTx.Owner == channelInfo.PeerIdA {
			infoRequest.IsAlice = true
			infoRequest.AmountA = commitmentTx.AmountToRSMC
			infoRequest.AmountB = commitmentTx.AmountToCounterparty
		} else {
			infoRequest.AmountB = commitmentTx.AmountToRSMC
			infoRequest.AmountA = commitmentTx.AmountToCounterparty
		}
	}
	nodes := make([]bean.ChannelInfoRequest, 0)
	nodes = append(nodes, infoRequest)
	go sendMsgToTracker(enum.MsgType_Tracker_UpdateChannelInfo_350, nodes)
}

func noticeTrackerUserLogin(user dao.User) {
	loginRequest := bean.ObdNodeUserLoginRequest{}
	loginRequest.UserId = user.PeerId
	sendMsgToTracker(enum.MsgType_Tracker_UserLogin_304, loginRequest)
}

func noticeTrackerUserLogout(user dao.User) {
	loginRequest := bean.ObdNodeUserLoginRequest{}
	loginRequest.UserId = user.PeerId
	sendMsgToTracker(enum.MsgType_Tracker_UserLogout_305, loginRequest)
}

func sendMsgToTracker(msgType enum.MsgType, data interface{}) {

	message := trackerBean.RequestMessage{}
	message.Type = msgType

	dataBytes, _ := json.Marshal(data)
	dataStr := string(dataBytes)

	parse := gjson.Parse(dataStr)
	result := parse.Value()
	if strings.HasPrefix(dataStr, "{") == false && strings.HasPrefix(dataStr, "[") == false {
		result = dataStr
	}

	message.Data = result
	//log.Println(message.Data)
	bytes, _ := json.Marshal(message)
	if TrackerChan != nil {
		TrackerChan <- bytes
	}
}
