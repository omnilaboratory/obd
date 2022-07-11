package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/lightclient"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/service"
	"log"
	"strings"
)

var connObd *websocket.Conn
var currUserInfo *pb.LoginResponse

func checkLogin() (user *bean.User, err error) {
	if obcClient.User == nil {
		return nil, errors.New("please login")
	}
	return obcClient.User, nil
}

func checkTargetUserIsOnline(requestMessage bean.RequestMessage) (err error) {
	_, err = lightclient.FindUserOnLine(requestMessage)
	if err != nil {
		return err
	}
	if lightclient.P2pChannelMap[requestMessage.RecipientNodePeerId] == nil {
		err = lightclient.ScanAndConnNode(requestMessage.RecipientNodePeerId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (server *RpcServer) GenSeed(ctx context.Context, in *pb.GenSeedRequest) (resp *pb.GenSeedResponse, err error) {
	log.Println("GenSeed")

	requestMessage := bean.RequestMessage{
		Type:             enum.MsgType_GetMnemonic_2004,
		SenderNodePeerId: obcClient.User.P2PLocalPeerId,
		SenderUserPeerId: obcClient.User.PeerId,
	}
	_, dataBytes, status := obcClient.HdWalletModule(requestMessage)
	data := string(dataBytes)
	if status == false {
		return nil, errors.New(data)
	}
	resp = &pb.GenSeedResponse{
		CipherSeedMnemonic: data,
	}
	return resp, nil
}

func (server *RpcServer) Login(ctx context.Context, in *pb.LoginRequest) (resp *pb.LoginResponse, err error) {

	log.Println("Login")

	if obcClient.User != nil {
		if obcClient.User.Mnemonic != in.Mnemonic {
			return nil, errors.New("user '" + in.Mnemonic + "' is online")
		}
	}

	if len(in.LoginToken) < 6 {
		return nil, errors.New("wrong login_token")
	}
	info := loginInfo{Mnemonic: in.Mnemonic, LoginToken: in.LoginToken}

	infoBytes, _ := json.Marshal(info)
	requestMessage := bean.RequestMessage{
		Type: enum.MsgType_UserLogin_2001,
		Data: string(infoBytes),
	}
	_, dataBytes, status := obcClient.UserModule(requestMessage)
	data := string(dataBytes)
	if status == false {
		return nil, errors.New(data)
	}

	dataMap := make(map[string]interface{})
	_ = json.Unmarshal(dataBytes, &dataMap)
	resp = &pb.LoginResponse{
		UserPeerId:    dataMap["userPeerId"].(string),
		NodePeerId:    dataMap["nodePeerId"].(string),
		NodeAddress:   dataMap["nodeAddress"].(string),
		HtlcFeeRate:   dataMap["htlcFeeRate"].(float64),
		HtlcMaxFee:    dataMap["htlcMaxFee"].(float64),
		ChainNodeType: dataMap["chainNodeType"].(string),
	}
	currUserInfo = resp
	return resp, nil
}

func (server *RpcServer) GetInfo(ctx context.Context, in *pb.GetInfoRequest) (resp *pb.GetInfoResponse, err error) {
	requestMessage := bean.RequestMessage{
		Type: enum.MsgType_User_GetInfo_2009,
	}
	_, dataBytes, status := obcClient.UserModule(requestMessage)
	if status == false {
		return nil, errors.New(string(dataBytes))
	}

	dataMap := make(map[string]interface{})
	_ = json.Unmarshal(dataBytes, &dataMap)
	resp = &pb.GetInfoResponse{
		UserPeerId:    dataMap["userPeerId"].(string),
		NodePeerId:    dataMap["nodePeerId"].(string),
		NodeAddress:   dataMap["nodeAddress"].(string),
		HtlcFeeRate:   dataMap["htlcFeeRate"].(float64),
		HtlcMaxFee:    dataMap["htlcMaxFee"].(float64),
		ChainNodeType: dataMap["chainNodeType"].(string),
	}
	resp.IsAdmin = true
	return resp, nil
}
func (server *RpcServer) NextAddr(ctx context.Context, in *pb.AddrRequest) (resp *pb.AddrResponse, err error) {
	log.Println("NextAddr")

	user, err := checkLogin()
	if err != nil {
		return nil, err
	}

	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return nil, err
	}
	resp = &pb.AddrResponse{
		Index:  int64(address.Index),
		Addr:   address.Address,
		PubKey: address.PubKey,
		Wif:    address.Wif,
	}
	return resp, nil
}

func (server *RpcServer) GetAddressInfo(ctx context.Context, in *pb.GetAddressInfoRequest) (resp *pb.AddrResponse, err error) {
	log.Println("GetAddressInfo")

	if in.Addr == "" {
		return nil, errors.New("empty addr")
	}

	user, err := checkLogin()
	if err != nil {
		return nil, err
	}

	for i := 0; i < obcClient.User.CurrAddrIndex; i++ {
		wallet, _ := service.HDWalletService.GetAddressByIndex(user, uint32(i))
		if wallet.Address == in.GetAddr() {
			resp = &pb.AddrResponse{
				Index:  int64(wallet.Index),
				Addr:   wallet.Address,
				PubKey: wallet.PubKey,
				Wif:    wallet.Wif,
			}
			break
		}
	}
	if resp == nil {
		return nil, errors.New("not found")
	}
	return resp, nil
}
func (server *RpcServer) NewAddress(ctx context.Context, in *pb.NewAddressRequest) (resp *pb.NewAddressResponse, err error) {
	log.Println("NextAddr")

	user, err := checkLogin()
	if err != nil {
		return nil, err
	}

	address, err := service.HDWalletService.CreateNewAddress(user)
	if err != nil {
		return nil, err
	}
	resp = &pb.NewAddressResponse{
		Index:  int64(address.Index),
		Addr:   address.Address,
		PubKey: address.PubKey,
		Wif:    address.Wif,
	}
	return resp, nil
}

func (server *RpcServer) EstimateFee(ctx context.Context, in *pb.EstimateFeeRequest) (resp *pb.EstimateFeeResponse, err error) {
	log.Println("EstimateFee")
	minerFee := omnicore.GetMinerFee(in.ConfTarget)
	resp = &pb.EstimateFeeResponse{
		SatPerKw: int64(100000000 * minerFee),
	}
	return resp, nil
}

func (server *RpcServer) Logout(ctx context.Context, in *pb.LogoutRequest) (resp *pb.LogoutResponse, err error) {
	log.Println("Logout")

	_, err = checkLogin()
	if err != nil {
		return nil, err
	}

	requestMessage := bean.RequestMessage{
		Type:             enum.MsgType_UserLogout_2002,
		SenderNodePeerId: obcClient.User.P2PLocalPeerId,
		SenderUserPeerId: obcClient.User.PeerId,
	}
	_, dataBytes, status := obcClient.UserModule(requestMessage)
	data := string(dataBytes)
	if status == false {
		return nil, errors.New(data)
	}
	return &pb.LogoutResponse{}, nil
}

func (server *RpcServer) ChangePassword(ctx context.Context, in *pb.ChangePasswordRequest) (resp *pb.ChangePasswordResponse, err error) {
	log.Println("ChangePassword")

	_, err = checkLogin()
	if err != nil {
		return nil, err
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
	infoBytes, _ := json.Marshal(token)
	requestMessage := bean.RequestMessage{
		Type:             enum.MsgType_User_UpdateAdminToken_2008,
		SenderNodePeerId: obcClient.User.P2PLocalPeerId,
		SenderUserPeerId: obcClient.User.PeerId,
		Data:             string(infoBytes)}
	_, dataBytes, status := obcClient.UserModule(requestMessage)
	data := string(dataBytes)
	if status == false {
		return nil, errors.New(data)
	}

	resp = &pb.ChangePasswordResponse{
		Result: data,
	}
	return resp, nil
}
func (server *RpcServer) ListPeers(ctx context.Context, in *pb.ListPeersRequest) (resp *pb.ListPeersResponse, err error) {
	log.Println("ListPeers")

	_, err = checkLogin()
	if err != nil {
		return nil, err
	}

	resp = &pb.ListPeersResponse{}
	for key, value := range lightclient.P2pChannelMap {
		if key != lightclient.P2PLocalNodeId {
			peer := &pb.Peer{Address: value.Address, PubKey: key}
			resp.Peers = append(resp.Peers, peer)
		}
	}
	return resp, nil
}
