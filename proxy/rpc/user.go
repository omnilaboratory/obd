package rpc

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/proxy/pb"
	"log"
	"net/url"
	"strings"
)

var connObd *websocket.Conn

type UserRpc struct {
}

func (user *UserRpc) Login(ctx context.Context, in *pb.LoginRequest) (resp *pb.LoginResponse, err error) {
	loginChan = make(chan bean.ReplyMessage)
	defer close(loginChan)

	if connObd == nil {
		u := url.URL{Scheme: "ws", Host: "127.0.0.1:60020", Path: "/ws" + config.ChainNodeType}
		log.Printf("begin to connect to tracker: %s", u.String())

		connObd, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Println("fail to dial obd:", err)
			return nil, err
		}
		go readDataFromObd()
	}
	if len(in.LoginToken) < 6 {
		return nil, errors.New("wrong login_token")
	}

	info := loginInfo{Mnemonic: in.Mnemonic, LoginToken: in.LoginToken}
	sendMsgToObd(info, enum.MsgType_UserLogin_2001)
	data := <-loginChan

	node := data.Result.(map[string]interface{})
	resp = &pb.LoginResponse{
		UserPeerId:    node["userPeerId"].(string),
		NodePeerId:    node["nodePeerId"].(string),
		NodeAddress:   node["nodeAddress"].(string),
		HtlcFeeRate:   node["htlcFeeRate"].(float64),
		HtlcMaxFee:    node["htlcMaxFee"].(float64),
		ChainNodeType: node["chainNodeType"].(string),
	}
	return resp, nil
}

func (user *UserRpc) Logout(ctx context.Context, in *pb.LogoutRequest) (resp *pb.LogoutResponse, err error) {
	log.Println("log out")
	if connObd == nil {
		return nil, errors.New("please login first")
	}

	logoutChan = make(chan bean.ReplyMessage)
	defer close(logoutChan)

	sendMsgToObd(nil, enum.MsgType_UserLogout_2002)

	data := <-logoutChan
	if data.Status == true {
		_ = connObd.Close()
		connObd = nil
	}
	return &pb.LogoutResponse{}, nil
}

func (user *UserRpc) UpdateLoginToken(ctx context.Context, in *pb.UpdateLoginTokenRequest) (resp *pb.UpdateLoginTokenResponse, err error) {
	if connObd == nil {
		return nil, errors.New("please login first")
	}

	if len(in.OldLoginToken) < 6 {
		return nil, errors.New("wrong oldLoginToken")
	}

	in.NewLoginToken = strings.TrimLeft(in.NewLoginToken, " ")
	in.NewLoginToken = strings.TrimRight(in.NewLoginToken, " ")
	if len(in.NewLoginToken) < 6 {
		return nil, errors.New("wrong newLoginToken")
	}

	token := updateLoginToken{OldLoginToken: in.OldLoginToken, NewLoginToken: in.NewLoginToken}

	updateLoginTokenChan = make(chan bean.ReplyMessage)
	defer close(updateLoginTokenChan)

	sendMsgToObd(token, enum.MsgType_User_UpdateAdminToken_2008)

	data := <-updateLoginTokenChan
	if data.Status == false {
		return nil, errors.New(data.Result.(string))
	}
	resp = &pb.UpdateLoginTokenResponse{
		Result: data.Result.(string),
	}
	return resp, nil
}
