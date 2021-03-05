package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"log"
)

func (s *RpcServer) RsmcPayment(ctx context.Context, in *pb.RsmcPaymentRequest) (*pb.RsmcPaymentResponse, error) {
	log.Println("RsmcPayment")
	_, err := checkLogin()
	if err != nil {
		return nil, err
	}
	if tool.CheckIsString(&in.RecipientInfo.RecipientNodePeerId) == false {
		return nil, errors.New("wrong recipient_node_peer_id")
	}

	if tool.CheckIsString(&in.RecipientInfo.RecipientUserPeerId) == false {
		return nil, errors.New("wrong recipient_user_peer_id")
	}
	if tool.CheckIsString(&in.ChannelId) == false {
		return nil, errors.New("wrong template_channel_id")
	}
	if in.Amount < 0 {
		return nil, errors.New("wrong amount")
	}

	request := bean.RequestCreateCommitmentTx{
		ChannelId: in.ChannelId,
		Amount:    in.Amount,
	}
	if rsmcChan == nil {
		rsmcChan = make(chan bean.ReplyMessage)
	}

	infoBytes, _ := json.Marshal(request)
	requestMessage := bean.RequestMessage{
		Type:                enum.MsgType_CommitmentTx_SendCommitmentTransactionCreated_351,
		SenderNodePeerId:    obcClient.User.P2PLocalPeerId,
		SenderUserPeerId:    obcClient.User.PeerId,
		RecipientNodePeerId: in.RecipientInfo.RecipientNodePeerId,
		RecipientUserPeerId: in.RecipientInfo.RecipientUserPeerId,
		Data:                string(infoBytes)}

	err = checkTargetUserIsOnline(requestMessage)
	if err != nil {
		return nil, err
	}

	_, dataBytes, status := obcClient.CommitmentTxModule(requestMessage)

	data := string(dataBytes)
	if status == false {
		return nil, errors.New(data)
	}
	dataMap := make(map[string]interface{})
	_ = json.Unmarshal(dataBytes, &dataMap)

	resp := &pb.RsmcPaymentResponse{
		ChannelId: dataMap["channel_id"].(string),
		AmountA:   dataMap["amount_a"].(float64),
		AmountB:   dataMap["amount_b"].(float64),
	}
	return resp, nil
}

func (s *RpcServer) LatestRsmcTx(ctx context.Context, in *pb.LatestRsmcTxRequest) (*pb.RsmcTxResponse, error) {
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
	resp := &pb.RsmcTxResponse{
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

func (s *RpcServer) TxListByChannelId(ctx context.Context, in *pb.TxListRequest) (*pb.TxListResponse, error) {
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
	resp := &pb.TxListResponse{TotalCount: int32(*count), PageIndex: in.PageIndex, PageSize: in.PageSize}
	for _, item := range transactions {
		node := &pb.RsmcTxResponse{
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
		resp.List = append(resp.List, node)
	}
	return resp, nil
}
