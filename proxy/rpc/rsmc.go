package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/proxy/pb"
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
