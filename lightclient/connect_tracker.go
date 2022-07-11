package lightclient

import (
	"context"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/tkrpc"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

func ConnectToTracker() (err error) {

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
	 SynData()
	//HeartBeat will set obd-node online info.
	go HeartBeat()

	return nil
}

var ITclient=service.ITclient
func SynData() {
	log.Println("synData to tracker")

	//sycUserInfos()
	ureq:=&tkrpc.UpdateUserInfosReq{ }
	for userId, _ := range GlobalWsClientManager.OnlineClientMap {
		req:=&tkrpc.UpdateUserInfoReq{UserId: userId}
		ureq.UpdateUserInfoReqs=append(ureq.UpdateUserInfoReqs,req)
	}
	_,err:= ITclient.UpdateUserInfos(todo,ureq)
	if err != nil {
		panic(err)
	}
	log.Println("synced userinfos")

	//sycChannelInfos()
	creq:=&tkrpc.UpdateChannelInfosReq{}
	channels := getChannelInfos()
	for _, channel := range channels {
		cInfo:= &tkrpc.ChannelInfo{}
		cInfo.ChannelId  =channel.ChannelId
		cInfo.PropertyId =channel.PropertyId
		cInfo.CurrState  =tkrpc.ChannelState(int(channel.CurrState))
		cInfo.PeerIda    =channel.PeerIdA
		cInfo.PeerIdb    =channel.PeerIdB
		cInfo.AmountA    =channel.AmountA
		cInfo.AmountB    =channel.AmountB
		cInfo.IsAlice    =channel.IsAlice
		cInfo.NodeId     =tool.GetObdNodeId()
		creq.ChannelInfos=append(creq.ChannelInfos,cInfo)
	}
	_,err= ITclient.UpdateChannelInfos(todo,creq)
	if err != nil {
		panic(err)
	}
	log.Println("synced ChannelInfos")

}

var todo=context.TODO()

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

func HeartBeat() {
	ninfo := &tkrpc.UpdateNodeInfoReq{NodeId: tool.GetObdNodeId(), P2PAddress: localServerDest, IsOnline: 1}
	for {
		cliStream, err := ITclient.HeartBeat(context.TODO(), nil)
		if err == nil {
			for {
				err = cliStream.Send(ninfo)
				if err != nil {
					break
				}
				time.Sleep(1 * time.Minute)
			}
		}
		if err != nil {
			log.Println("heartBeat error", err)
		}
		time.Sleep(2 * time.Second)
	}
}
