package lightclient

import (
	"encoding/json"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	trackerBean "github.com/omnilaboratory/obd/tracker/bean"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/url"
	"reflect"
	"strings"
	"time"
)

var conn *websocket.Conn
var ticker3m *time.Ticker
var isReset = true

func ConnectToTracker() (err error) {

	u := url.URL{Scheme: "ws", Host: config.TrackerHost, Path: "/ws"}
	log.Printf("begin to connect to tracker: %s", u.String())

	conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("fail to dial tracker:", err)
		return err
	}

	chainNodeType, trackerP2pAddress, err := conn2tracker.GetChainNodeType()
	if err != nil {
		return err
	}
	config.ChainNodeType = chainNodeType
	config.BootstrapPeers, _ = config.StringsToAddrs(strings.Split(trackerP2pAddress, ","))

	err = StartP2PNode()
	if err != nil {
		return err
	}

	if service.TrackerChan == nil {
		service.TrackerChan = make(chan []byte)
	}

	if isReset {
		go readDataFromWs()
	}

	if ticker3m == nil {
		startSchedule()
	}
	return nil
}

func readDataFromWs() {
	isReset = false
	ticker := time.NewTicker(time.Minute * 2)
	defer ticker.Stop()

	defer func(ticker *time.Ticker) {
		if r := recover(); r != nil {
			log.Println("tracker goroutine recover")
			ticker.Stop()
			conn = nil
			isReset = true
		}
	}(ticker)
	defer conn.Close()

	// read message
	go func() {
		for {
			if conn == nil {
				isReset = true
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				isReset = true
				log.Println("socket to tracker get err:", err)
				conn = nil
				return
			}

			//log.Println("get data from tracker", string(message))
			replyMessage := bean.ReplyMessage{}
			err = json.Unmarshal(message, &replyMessage)
			if err == nil {
				switch replyMessage.Type {
				case enum.MsgType_Tracker_GetHtlcPath_351:
					v := reflect.ValueOf(replyMessage.Result)
					requestMessage := bean.RequestMessage{}
					requestMessage.Type = replyMessage.Type
					requestMessage.Data = ""
					if v.Kind() == reflect.Map {
						dataMap := replyMessage.Result.(map[string]interface{})
						requestMessage.RecipientUserPeerId = dataMap["senderPeerId"].(string)
						requestMessage.Data = dataMap["h"].(string) + "_" + dataMap["path"].(string) + "_" + tool.FloatToString(dataMap["amount"].(float64), 8)
						//requestMessage.Data = dataMap["h"].(string) + "_" + dataMap["path"].(string)
					}
					htlcTrackerDealModule(requestMessage)
				case enum.MsgType_Tracker_Connect_301:
					config.ChainNodeType = replyMessage.Result.(string)
					go SynData()
				}
			}
		}
	}()

	// heartbeat and check whether
	for {
		select {
		case t := <-ticker.C:
			if conn != nil {
				info := make(map[string]interface{})
				info["type"] = enum.MsgType_Tracker_HeartBeat_302
				info["data"] = t.String()
				bytes, err := json.Marshal(info)
				err = conn.WriteMessage(websocket.TextMessage, bytes)
				if err != nil {
					log.Println("HeartBeat:", err)
					isReset = true
					log.Println("socket to tracker get err:", err)
					conn = nil
					return
				}
			} else {
				return
			}
		}
	}
}

func SynData() {
	log.Println("synData to tracker")
	updateP2pAddressLogin()
	sycUserInfos()
	sycChannelInfos()
}

func updateP2pAddressLogin() {
	info := make(map[string]interface{})
	info["type"] = enum.MsgType_Tracker_NodeLogin_303
	nodeLoginInfo := &bean.ObdNodeLoginRequest{}
	nodeLoginInfo.NodeId = tool.GetObdNodeId()
	nodeLoginInfo.P2PAddress = localServerDest
	info["data"] = nodeLoginInfo
	bytes, err := json.Marshal(info)
	if err != nil {
		log.Println(err)
	} else {
		sendMsgToTracker(bytes)
	}
}

func sycUserInfos() {

	nodes := make([]bean.ObdNodeUserLoginRequest, 0)
	for userId, _ := range GlobalWsClientManager.OnlineClientMap {
		user := bean.ObdNodeUserLoginRequest{}
		user.UserId = userId
		nodes = append(nodes, user)
	}
	if len(nodes) > 0 {
		log.Println("syn UserInfo data to tracker", nodes)
		info := make(map[string]interface{})
		info["type"] = enum.MsgType_Tracker_UpdateUserInfo_353
		info["data"] = nodes
		bytes, err := json.Marshal(&info)
		if err == nil {
			sendMsgToTracker(bytes)
		}
	}
}

//同步通道信息
func sycChannelInfos() {
	nodes := getChannelInfos()
	if len(nodes) > 0 {
		info := make(map[string]interface{})
		info["type"] = enum.MsgType_Tracker_UpdateChannelInfo_350
		info["data"] = nodes
		bytes, err := json.Marshal(info)
		if err == nil {
			sendMsgToTracker(bytes)
		}
	}
}

func getChannelInfos() []bean.ChannelInfoRequest {
	_dir := config.DataDirectory + "/" + config.ChainNodeType
	files, _ := ioutil.ReadDir(_dir)

	dbNames := make([]string, 0)
	userPeerIds := make([]string, 0)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "user_") && strings.HasSuffix(f.Name(), ".db") {
			peerId := strings.TrimPrefix(f.Name(), "user_")
			peerId = strings.TrimSuffix(peerId, ".db")
			value, exists := GlobalWsClientManager.OnlineClientMap[peerId]
			if exists && value != nil {
				userPeerIds = append(userPeerIds, peerId)
			} else {
				dbNames = append(dbNames, f.Name())
			}
		}
	}

	nodes := make([]bean.ChannelInfoRequest, 0)

	for _, peerId := range userPeerIds {
		client, _ := GlobalWsClientManager.OnlineClientMap[peerId]
		if client != nil {
			nodes = checkChannel(peerId, client.User.Db, nodes)
		}
	}

	for _, dbName := range dbNames {
		db, err := storm.Open(_dir + "/" + dbName)
		if err == nil {
			userId := strings.TrimLeft(dbName, "user_")
			userId = strings.TrimRight(userId, ".db")
			nodes = checkChannel(userId, db, nodes)
			_ = db.Close()
		}
	}
	return nodes
}

func lockChannel(userId, channelId string) (err error) {
	var userDb *storm.DB
	value, exists := GlobalWsClientManager.OnlineClientMap[userId]
	if exists && value.User.Db != nil {
		userDb = value.User.Db
	} else {
		_dir := "dbdata" + config.ChainNodeType
		userDb, err = storm.Open(_dir + "/" + "user_" + userId + ".db")
	}
	var channelInfo dao.ChannelInfo
	if userDb != nil {
		err = userDb.Select(
			q.Eq("IsPrivate", false),
			q.Eq("CurrState", bean.ChannelState_CanUse),
			q.Eq("ChannelId", channelId)).First(&channelInfo)
		if err == nil {
			return userDb.UpdateField(&channelInfo, "CurrState", bean.ChannelState_LockByTracker)
		}
	}
	return nil
}

func unlockChannel(userId, channelId string) (err error) {
	var userDb *storm.DB
	value, exists := GlobalWsClientManager.OnlineClientMap[userId]
	dbIsOnOpen := false
	if exists && value.User.Db != nil {
		userDb = value.User.Db
		dbIsOnOpen = true
	} else {
		_dir := "dbdata" + config.ChainNodeType
		userDb, err = storm.Open(_dir + "/" + "user_" + userId + ".db")
	}
	var channelInfo dao.ChannelInfo
	if userDb != nil {
		err = userDb.Select(
			q.Or(q.Eq("CurrState", bean.ChannelState_LockByTracker),
				q.Eq("CurrState", bean.ChannelState_CanUse)),
			q.Eq("ChannelId", channelId)).First(&channelInfo)
		if err == nil {
			if channelInfo.CurrState != bean.ChannelState_CanUse {
				err = userDb.UpdateField(&channelInfo, "CurrState", bean.ChannelState_CanUse)
			}
		}
		if dbIsOnOpen == false {
			_ = userDb.Close()
		}
	}
	return err
}

func checkChannel(userId string, db storm.Node, nodes []bean.ChannelInfoRequest) []bean.ChannelInfoRequest {
	var channelInfos []dao.ChannelInfo
	err := db.Select(
		q.Eq("IsPrivate", false),
		q.Or(
			q.Eq("CurrState", bean.ChannelState_CanUse),
			q.Eq("CurrState", bean.ChannelState_Close),
			q.Eq("CurrState", bean.ChannelState_HtlcTx))).Find(&channelInfos)
	if err == nil {
		for _, channelInfo := range channelInfos {
			commitmentTransaction := dao.CommitmentTransaction{}
			err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId)).OrderBy("CreateAt").Reverse().First(&commitmentTransaction)
			request := bean.ChannelInfoRequest{}
			request.ChannelId = channelInfo.ChannelId
			request.PropertyId = channelInfo.PropertyId
			request.PeerIdA = channelInfo.PeerIdA
			request.PeerIdB = channelInfo.PeerIdB
			request.CurrState = channelInfo.CurrState
			if commitmentTransaction.Id > 0 {
				request.AmountA = commitmentTransaction.AmountToRSMC
				request.AmountB = commitmentTransaction.AmountToCounterparty
				request.IsAlice = false
				if commitmentTransaction.Owner == channelInfo.PeerIdA {
					request.IsAlice = true
					request.AmountA = commitmentTransaction.AmountToRSMC
					request.AmountB = commitmentTransaction.AmountToCounterparty
				} else {
					request.AmountA = commitmentTransaction.AmountToCounterparty
					request.AmountB = commitmentTransaction.AmountToRSMC
				}
			} else {
				request.AmountA = channelInfo.Amount
				request.AmountB = 0
				request.IsAlice = false
				if channelInfo.FunderPeerId == userId {
					request.IsAlice = true
				}
			}
			nodes = append(nodes, request)
		}
	}
	return nodes
}

func sendMsgToTracker(msg []byte) {
	//log.Println("send to tracker", string(msg))
	if conn == nil {
		isReset = true
		err := ConnectToTracker()
		if err != nil {
			log.Println(err)
			return
		}
	}
	err := conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func startSchedule() {
	go func() {
		for {
			select {
			case msg := <-service.TrackerChan:
				sendMsgToTracker(msg)
				msgType := gjson.Parse(string(msg)).Get("type").Int()
				if msgType == 350 {
					message := &trackerBean.RequestMessage{}
					err := json.Unmarshal(msg, message)
					if err == nil {
						marshal, err := json.Marshal(message.Data)
						if err == nil {
							sendChannelInfoToIndirectTracker(string(marshal))
						}
					}
				}
			}
		}
	}()

	go func() {
		ticker3m = time.NewTicker(1 * time.Minute)
		defer ticker3m.Stop()
		for {
			select {
			case t := <-ticker3m.C:
				if conn == nil {
					log.Println("reconnect tracker ", t)
					_ = ConnectToTracker()
				}
			}
		}
	}()
}
