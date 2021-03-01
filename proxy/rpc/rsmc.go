package rpc

import (
	"context"
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

	sendMsgToObd(request, in.RecipientInfo.RecipientNodePeerId, in.RecipientInfo.RecipientUserPeerId, enum.MsgType_CommitmentTx_SendCommitmentTransactionCreated_351)

	for {
		data := <-rsmcChan
		if data.Status == false {
			return nil, errors.New(data.Result.(string))
		}
		if data.Type == enum.MsgType_ClientSign_CommitmentTx_AliceSignC2a_360 {
			dataResult := data.Result.(map[string]interface{})
			resp := &pb.RsmcPaymentResponse{
				ChannelId: dataResult["channel_id"].(string),
				AmountA:   dataResult["amount_a"].(float64),
				AmountB:   dataResult["amount_b"].(float64),
			}
			return resp, nil
		}
	}
}
