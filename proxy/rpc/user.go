package rpc

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/proxy/pb"
	"log"
	"strings"
)

var connObd *websocket.Conn
var currUserInfo *pb.LoginResponse

type UserRpc struct {
}

func (user *UserRpc) Login(ctx context.Context, in *pb.LoginRequest) (resp *pb.LoginResponse, err error) {

	log.Println("Login")
	if len(in.LoginToken) < 6 {
		return nil, errors.New("wrong login_token")
	}

	info := loginInfo{Mnemonic: in.Mnemonic, LoginToken: in.LoginToken}
	sendMsgToObd(info, "", "", enum.MsgType_UserLogin_2001)

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
	currUserInfo = resp
	return resp, nil
}

func (user *UserRpc) Logout(ctx context.Context, in *pb.LogoutRequest) (resp *pb.LogoutResponse, err error) {
	log.Println("Logout")
	if connObd == nil {
		return nil, errors.New("please login first")
	}

	if logoutChan == nil {
		logoutChan = make(chan bean.ReplyMessage)
	}

	sendMsgToObd(nil, "", "", enum.MsgType_UserLogout_2002)

	data := <-logoutChan
	if data.Status == true {
		_ = connObd.Close()
		connObd = nil
	}
	return &pb.LogoutResponse{}, nil
}

func (user *UserRpc) ChangePassword(ctx context.Context, in *pb.ChangePasswordRequest) (resp *pb.ChangePasswordResponse, err error) {
	log.Println("ChangePassword")
	if connObd == nil {
		return nil, errors.New("please login first")
	}

	if len(in.CurrentPassword) < 6 {
		return nil, errors.New("wrong current_password")
	}

	in.NewPassword = strings.TrimLeft(in.NewPassword, " ")
	in.NewPassword = strings.TrimRight(in.NewPassword, " ")
	if len(in.NewPassword) < 6 {
		return nil, errors.New("wrong new_password")
	}

	token := updateLoginToken{CurrentPassword: in.CurrentPassword, NewPassword: in.NewPassword}

	if changePasswordChan == nil {
		changePasswordChan = make(chan bean.ReplyMessage)
	}

	sendMsgToObd(token, "", "", enum.MsgType_User_UpdateAdminToken_2008)

	data := <-changePasswordChan
	if data.Status == false {
		return nil, errors.New(data.Result.(string))
	}
	resp = &pb.ChangePasswordResponse{
		Result: data.Result.(string),
	}
	return resp, nil
}
