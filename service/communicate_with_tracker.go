package service

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
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
	if commitmentTx.Owner == channelInfo.PeerIdA {
		infoRequest.IsAlice = true
		infoRequest.AmountA = commitmentTx.AmountToRSMC
		infoRequest.AmountB = commitmentTx.AmountToCounterparty
	} else {
		infoRequest.AmountB = commitmentTx.AmountToRSMC
		infoRequest.AmountA = commitmentTx.AmountToCounterparty
	}
	nodes := make([]bean.ChannelInfoRequest, 0)
	nodes = append(nodes, infoRequest)
	sendMsgToTracker(enum.MsgType_Tracker_UpdateChannelInfo_350, nodes)
}

func httpGetHtlcStateFromTracker(path string, h string) (flag int) {
	url := "http://" + config.TrackerHost + "/api/v1/getHtlcTxState?path=" + path + "&h=" + h
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(body))
		return int(gjson.Get(string(body), "data").Get("flag").Int())
	}
	return 0
}

func httpGetChannelStateFromTracker(channelId string) (flag int) {
	url := "http://" + config.TrackerHost + "/api/v1/getChannelState?channelId=" + channelId
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(body))
		return int(gjson.Get(string(body), "data").Get("state").Int())
	}
	return 0
}

func HttpGetUserStateFromTracker(userId string) (flag int) {
	url := "http://" + config.TrackerHost + "/api/v1/getUserState?userId=" + userId
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(body))
		return int(gjson.Get(string(body), "data").Get("state").Int())
	}
	return 0
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

	message := bean.RequestMessage{}
	message.Type = msgType

	dataBytes, _ := json.Marshal(data)
	dataStr := string(dataBytes)

	parse := gjson.Parse(dataStr)
	result := parse.Value()
	if strings.HasPrefix(dataStr, "{") == false && strings.HasPrefix(dataStr, "[") == false {
		result = dataStr
	}

	message.Data = result
	log.Println(message.Data)
	bytes, _ := json.Marshal(message)
	if TrackerChan != nil {
		TrackerChan <- bytes
	}
}
