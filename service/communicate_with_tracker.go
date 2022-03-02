package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tracker/tkrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"time"
)

//var TrackerChan chan []byte

// sendChannelStateToTracker
func sendChannelStateToTracker(channelInfo dao.ChannelInfo, commitmentTx dao.CommitmentTransaction) {
	if channelInfo.IsPrivate {
		return
	}
	infoRequest := &tkrpc.ChannelInfo{}
	//infoRequest := bean.ChannelInfoRequest{}
	infoRequest.ChannelId = channelInfo.ChannelId
	infoRequest.PropertyId = channelInfo.PropertyId
	infoRequest.CurrState =tkrpc.ChannelState(int( channelInfo.CurrState))
	infoRequest.PeerIda = channelInfo.PeerIdA
	infoRequest.PeerIdb = channelInfo.PeerIdB

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
	_,err:=ITclient.UpdateChannelInfo(todo,infoRequest)
	if err != nil {
		log.Println("UpdateChannelInfo err",err)
	}
	//nodes := make([]bean.ChannelInfoRequest, 0)
	//nodes = append(nodes, infoRequest)
	//go sendMsgToTracker(enum.MsgType_Tracker_UpdateChannelInfo_350, nodes)
}

func noticeTrackerUserLogin(user dao.User) {
	_,err:=ITclient.UpdateUserInfo(todo,&tkrpc.UpdateUserInfoReq{UserId:user.PeerId,NodeId:user.P2PLocalPeerId,P2PAddress:user.P2PLocalAddress,IsOnline: 1})
	if err != nil {
		log.Println("UpdateUserInfo err",err)
	}
	//loginRequest := bean.ObdNodeUserLoginRequest{UserId: user.PeerId}
	//sendMsgToTracker(enum.MsgType_Tracker_UserLogin_304, loginRequest)
}

func noticeTrackerUserLogout(user dao.User) {
	//loginRequest := bean.ObdNodeUserLoginRequest{UserId: user.PeerId}
	//sendMsgToTracker(enum.MsgType_Tracker_UserLogout_305, loginRequest)
	_,err:=ITclient.UpdateUserInfo(todo,&tkrpc.UpdateUserInfoReq{UserId:user.PeerId,NodeId:user.P2PLocalPeerId,P2PAddress:user.P2PLocalAddress,IsOnline: 2})
	if err != nil {
		log.Println("UpdateUserInfo err",err)
	}
}

var ITclient tkrpc.InfoTrackerClient
func init(){
	tc,_:=context.WithTimeout(context.Background(),5*time.Second)
	creds := credentials.NewTLS(&tls.Config{
		//TLS Config values here
	})
	conn, err := grpc.DialContext(tc,config.TrackerHostGrpc, grpc.WithBlock(),grpc.WithTransportCredentials(creds))  //grpc.WithInsecure(),
	if err == nil {
		ITclient = tkrpc.NewInfoTrackerClient(conn)
		log.Println("connected tracker grpc",config.TrackerHostGrpc)
	}
	if err != nil {
		panic(fmt.Errorf("Tracker grpc Client err %s",err.Error()))
	}
}
