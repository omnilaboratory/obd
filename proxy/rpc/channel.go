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

func (s *RpcServer) OpenChannel(ctx context.Context, in *pb.OpenChannelRequest) (*pb.OpenChannelResponse, error) {
	log.Println("OpenChannel")
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

	if tool.CheckIsString(&in.NodePubkeyString) == false {
		return nil, errors.New("wrong node_pubkey_string")
	}

	channelOpen := bean.SendChannelOpen{
		FundingPubKey:      in.NodePubkeyString,
		FunderAddressIndex: int(in.NodePubkeyIndex),
		IsPrivate:          in.Private,
	}

	infoBytes, _ := json.Marshal(channelOpen)
	requestMessage := bean.RequestMessage{
		Type:                enum.MsgType_SendChannelOpen_32,
		SenderNodePeerId:    obcClient.User.P2PLocalPeerId,
		SenderUserPeerId:    obcClient.User.PeerId,
		RecipientNodePeerId: in.RecipientInfo.RecipientNodePeerId,
		RecipientUserPeerId: in.RecipientInfo.RecipientUserPeerId,
		Data:                string(infoBytes)}

	err = checkTargetUserIsOnline(requestMessage)
	if err != nil {
		return nil, err
	}

	if obcClient.GrpcChan == nil {
		obcClient.GrpcChan = make(chan []byte)
	}

	obcClient.ChannelModule(requestMessage)

	message := <-obcClient.GrpcChan

	close(obcClient.GrpcChan)
	obcClient.GrpcChan = nil

	replyMessage := bean.ReplyMessage{}
	_ = json.Unmarshal(message, &replyMessage)
	if replyMessage.Status == false {
		return nil, errors.New(replyMessage.Result.(string))
	}

	dataResult := replyMessage.Result.(map[string]interface{})
	resp := &pb.OpenChannelResponse{}
	resp.TemplateChannelId = dataResult["temporary_channel_id"].(string)

	return resp, nil
}

func (s *RpcServer) FundChannel(ctx context.Context, in *pb.FundChannelRequest) (*pb.FundChannelResponse, error) {
	log.Println("FundChannel")
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
	if tool.CheckIsString(&in.TemplateChannelId) == false {
		return nil, errors.New("wrong template_channel_id")
	}
	if in.BtcAmount < 0 {
		return nil, errors.New("wrong btc_amount")
	}
	if in.PropertyId < 0 {
		return nil, errors.New("wrong property_id")
	}
	if in.AssetAmount < 0 {
		return nil, errors.New("wrong asset_amount")
	}

	requestFunding := bean.SendRequestFunding{
		TemporaryChannelId: in.TemplateChannelId,
		BtcAmount:          in.BtcAmount,
		PropertyId:         in.PropertyId,
		AssetAmount:        in.AssetAmount,
	}

	infoBytes, _ := json.Marshal(requestFunding)
	requestMessage := bean.RequestMessage{
		Type:                enum.MsgType_Funding_134,
		SenderNodePeerId:    obcClient.User.P2PLocalPeerId,
		SenderUserPeerId:    obcClient.User.PeerId,
		RecipientNodePeerId: in.RecipientInfo.RecipientNodePeerId,
		RecipientUserPeerId: in.RecipientInfo.RecipientUserPeerId,
		Data:                string(infoBytes)}

	err = checkTargetUserIsOnline(requestMessage)
	if err != nil {
		return nil, err
	}

	_, dataBytes, status := obcClient.FundingTransactionModule(requestMessage)

	data := string(dataBytes)
	if status == false {
		return nil, errors.New(data)
	}
	dataMap := make(map[string]interface{})
	_ = json.Unmarshal(dataBytes, &dataMap)

	resp := &pb.FundChannelResponse{}
	resp.ChannelId = dataMap["channel_id"].(string)
	return resp, nil
}
