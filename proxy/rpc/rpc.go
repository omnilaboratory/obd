package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/service"
	"log"
)

type RpcServer struct{}

// for testing
func (r *RpcServer) Hello(ctx context.Context,
	in *pb.HelloRequest) (*pb.HelloResponse, error) {

	log.Println("hello " + in.GetSayhi())
	resp, err := Hello(in.Sayhi)
	if err != nil {
		return nil, err
	}

	return &pb.HelloResponse{Resp: resp}, nil
}

// for testing
func Hello(sayhi string) (string, error) {
	returnMsg := "You sent: [" + sayhi + "]. We're testing proxy mode of OBD."
	return returnMsg, nil
}

func (s *RpcServer) ConnectPeer(ctx context.Context, in *pb.ConnectPeerRequest) (resp *pb.ConnectPeerResponse, err error) {
	log.Println("ConnectPeer")

	if len(in.Addr) == 0 {
		return nil, errors.New("wrong addr")
	}
	input := make(map[string]interface{})
	input["remote_node_address"] = in.Addr
	marshal, _ := json.Marshal(input)
	requestMessage := bean.RequestMessage{
		Type: enum.MsgType_p2p_ConnectPeer_2003,
		Data: string(marshal)}
	_, bytes, status := obcClient.UserModule(requestMessage)
	data := string(bytes)
	if status == false {
		return nil, errors.New(data)
	}
	resp = &pb.ConnectPeerResponse{}
	return resp, nil
}
func (s *RpcServer) DisconnectPeer(ctx context.Context, in *pb.DisconnectPeerRequest) (resp *pb.DisconnectPeerResponse, err error) {
	log.Println("ConnectPeer")

	if len(in.Addr) == 0 {
		return nil, errors.New("wrong addr")
	}
	input := make(map[string]interface{})
	input["remote_node_address"] = in.Addr
	marshal, _ := json.Marshal(input)
	requestMessage := bean.RequestMessage{
		Type: enum.MsgType_p2p_DisconnectPeer_2010,
		Data: string(marshal)}
	_, bytes, status := obcClient.UserModule(requestMessage)
	data := string(bytes)
	if status == false {
		return nil, errors.New(data)
	}
	resp = &pb.DisconnectPeerResponse{}
	return resp, nil
}

func (s *RpcServer) PendingChannels(ctx context.Context, in *pb.PendingChannelsRequest) (resp *pb.ListChannelsResponse, err error) {
	log.Println("PendingChannels")

	user, err := checkLogin()
	if err != nil {
		return nil, err
	}

	inputData := make(map[string]interface{})
	inputData["is_pending"] = true
	inputData["page_size"] = in.PageSize
	inputData["page_index"] = in.PageIndex
	marshal, _ := json.Marshal(inputData)
	respData, err := service.ChannelService.AllItem(string(marshal), *user)
	if err != nil {
		return nil, err
	}
	resp = &pb.ListChannelsResponse{}
	list := respData.Data.([]service.ChannelVO)
	for _, item := range list {
		if len(item.ChannelId) == 0 {
			continue
		}
		node := &pb.Channel{}
		node.ChanId = item.ChannelId
		node.Private = item.IsPrivate
		node.Active = true
		if item.CurrState == bean.ChannelState_Close {
			node.Active = false
		}
		node.PropertyId = item.PropertyId
		node.Capacity = int64(item.AssetAmount * 100000000)
		node.Initiator = false
		if user.PeerId == item.PeerIdA {
			node.Initiator = true
		}
		node.LocalBalance = int64(item.BalanceA * 100000000)
		node.RemoteBalance = int64(item.BalanceB * 100000000)
		node.NumUpdates = item.NumUpdates
		resp.Channels = append(resp.Channels, node)
	}
	return resp, nil
}
func (s *RpcServer) ListChannels(ctx context.Context, in *pb.ListChannelsRequest) (resp *pb.ListChannelsResponse, err error) {
	log.Println("ListChannels")
	user, err := checkLogin()
	if err != nil {
		return nil, err
	}
	marshal, _ := json.Marshal(in)
	respData, err := service.ChannelService.AllItem(string(marshal), *user)
	if err != nil {
		return nil, err
	}
	resp = &pb.ListChannelsResponse{}
	list := respData.Data.([]service.ChannelVO)
	for _, item := range list {
		if len(item.ChannelId) == 0 {
			continue
		}
		node := &pb.Channel{}
		node.ChanId = item.ChannelId
		node.Private = item.IsPrivate
		node.Active = true
		if item.CurrState == bean.ChannelState_Close {
			node.Active = false
		}
		node.PropertyId = item.PropertyId
		node.Capacity = int64(item.AssetAmount * 100000000)
		node.Initiator = false
		if user.PeerId == item.PeerIdA {
			node.Initiator = true
		}
		node.LocalBalance = int64(item.BalanceA * 100000000)
		node.RemoteBalance = int64(item.BalanceB * 100000000)
		node.NumUpdates = item.NumUpdates
		resp.Channels = append(resp.Channels, node)
	}
	return resp, err
}

func (s *RpcServer) LatestTransaction(ctx context.Context, in *pb.LatestTransactionRequest) (*pb.Transaction, error) {
	log.Println("LatestRsmcTx")

	if len(in.ChannelId) == 0 {
		return nil, errors.New("wrong channelId")
	}

	user, err := checkLogin()
	if err != nil {
		return nil, err
	}
	marshal, _ := json.Marshal(in)
	respData, err := service.CommitmentTxService.GetLatestCommitmentTxByChannelId(string(marshal), user)
	if err != nil {
		return nil, err
	}
	resp := &pb.Transaction{
		TxHash:    respData.CurrHash,
		ChannelId: respData.ChannelId,
		AmountA:   respData.AmountToRSMC,
		AmountB:   respData.AmountToCounterparty,
		PeerA:     respData.PeerIdA,
		PeerB:     respData.PeerIdB,
		CurrState: int32(respData.CurrState),
		TxType:    int32(respData.TxType),
		H:         respData.HtlcH,
		R:         respData.HtlcR,
	}
	return resp, nil
}

func (s *RpcServer) GetTransactions(ctx context.Context, in *pb.GetTransactionsRequest) (*pb.TransactionDetails, error) {
	log.Println("LatestRsmcTx")

	if len(in.ChannelId) == 0 {
		return nil, errors.New("wrong channelId")
	}

	user, err := checkLogin()
	if err != nil {
		return nil, err
	}
	marshal, _ := json.Marshal(in)
	transactions, count, err := service.CommitmentTxService.GetItemsByChannelId(string(marshal), user)
	log.Println(count)
	log.Println(transactions)
	if err != nil {
		return nil, err
	}
	resp := &pb.TransactionDetails{TotalCount: int32(*count), PageIndex: in.PageIndex, PageSize: in.PageSize}
	for _, item := range transactions {
		node := &pb.Transaction{
			TxHash:     item.CurrHash,
			ChannelId:  item.ChannelId,
			AmountA:    item.AmountToRSMC,
			AmountB:    item.AmountToCounterparty,
			PeerA:      item.PeerIdA,
			PeerB:      item.PeerIdB,
			CurrState:  int32(item.CurrState),
			TxType:     int32(item.TxType),
			H:          item.HtlcH,
			R:          item.HtlcR,
			AmountHtlc: item.AmountToHtlc,
		}
		resp.Transactions = append(resp.Transactions, node)
	}
	return resp, nil
}

func (s *RpcServer) ChannelBalance(ctx context.Context, in *pb.ChannelBalanceRequest) (*pb.ChannelBalanceResponse, error) {
	log.Println("ChannelBalance")

	user, err := checkLogin()
	if err != nil {
		return nil, err
	}
	data, err := service.ChannelService.ChannelBalance(user)
	if err != nil {
		return nil, err
	}

	resp := &pb.ChannelBalanceResponse{
		LocalBalance:             data["local_balance"],
		RemoteBalance:            data["remote_balance"],
		UnsettledLocalBalance:    data["unsettled_local_balance"],
		UnsettledRemoteBalance:   data["unsettled_remote_balance"],
		PendingOpenLocalBalance:  data["pending_open_local_balance"],
		PendingOpenRemoteBalance: data["pending_open_remote_balance"],
	}
	return resp, nil
}
