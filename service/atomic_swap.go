package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"log"
	"sync"
	"time"
)

//close htlc or close channel
type atomicSwapManager struct {
	operationFlag sync.Mutex
}

var AtomicSwapService atomicSwapManager

//type MsgType_Atomic_Swap_N80  可以理解为付款凭证
func (this *atomicSwapManager) AtomicSwap(msgData string, user bean.User) (outData interface{}, targetUser string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}
	reqData := &bean.AtomicSwapRequest{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.RecipientPeerId) == false {
		return nil, "", errors.New("error recipient_peer_id")
	}

	if reqData.TimeLocker < 10 {
		return nil, "", errors.New("error time_locker")
	}

	err = FindUserIsOnline(reqData.RecipientPeerId)
	if err != nil {
		return nil, "", err
	}

	if reqData.RecipientPeerId == user.PeerId {
		return nil, "", errors.New("you should not send msg to yourself")
	}

	if tool.CheckIsString(&reqData.ChannelIdFrom) == false {
		return nil, "", errors.New("error channel_id_from")
	}

	if reqData.PropertySent < 0 {
		return nil, "", errors.New("error property_sent")
	}

	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelIdFrom),
		q.Eq("CurrState", dao.ChannelState_CanUse),
		q.Eq("PropertyId", reqData.PropertySent),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(&dao.ChannelInfo{})
	if err != nil {
		return nil, "", errors.New("not found this channel_id_from")
	}
	if tool.CheckIsString(&reqData.ChannelIdTo) == false {
		return nil, "", errors.New("error channel_id_to")
	}
	if reqData.PropertyReceived < 0 {
		return nil, "", errors.New("error property_received")
	}
	if reqData.ExchangeRate < 0 {
		return nil, "", errors.New("error exchange_rate")
	}
	if tool.CheckIsString(&reqData.TransactionId) == false {
		return nil, "", errors.New("error transaction_id")
	}

	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelIdTo),
		q.Eq("CurrState", dao.ChannelState_CanUse),
		q.Eq("PropertyId", reqData.PropertyReceived),
		q.Or(
			q.Eq("PeerIdA", reqData.RecipientPeerId),
			q.Eq("PeerIdB", reqData.RecipientPeerId))).
		First(&dao.ChannelInfo{})
	if err != nil {
		return nil, "", errors.New("not found this channel_id_to")
	}

	err = db.Select(
		q.Eq("CurrHash", reqData.TransactionId),
		q.Eq("TxType", dao.CommitmentTransactionType_Htlc),
		q.Eq("CurrState", dao.TxInfoState_Htlc_GetH),
		q.Eq("ChannelId", reqData.ChannelIdFrom),
		q.Eq("Owner", user.PeerId)).
		First(&dao.CommitmentTransaction{})
	if err != nil {
		return nil, "", errors.New("error transaction_id")
	}

	swapInfo := &dao.AtomicSwapInfo{}

	//check if have send the same info
	err = db.Select(
		q.Eq("TransactionId", reqData.TransactionId),
		q.Eq("ChannelIdFrom", reqData.ChannelIdFrom),
		q.Eq("ChannelIdTo", reqData.ChannelIdTo),
		q.Eq("PropertySent", reqData.PropertySent),
		q.Eq("PropertyReceived", reqData.PropertyReceived),
		q.Eq("CreateBy", user.PeerId),
	).First(swapInfo)

	swapInfo.LatestEditAt = time.Now()
	swapInfo.AtomicSwapRequest = *reqData
	if err == nil {
		err = db.Update(swapInfo)
	} else {
		swapInfo.CreateBy = user.PeerId
		swapInfo.CreateAt = time.Now()
		err = db.Save(swapInfo)
	}
	if err != nil {
		return nil, "", errors.New("fail to save db,try again")
	}
	return reqData, reqData.RecipientPeerId, nil
}

//MsgType_Atomic_Swap_Accept_N81 可以理解为发货凭证
func (this *atomicSwapManager) AtomicSwapAccepted(msgData string, user bean.User) (outData interface{}, targetUser string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}
	reqData := &bean.AtomicSwapAccepted{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.RecipientPeerId) == false {
		return nil, "", errors.New("error recipient_peer_id")
	}

	if reqData.TimeLocker < 10 {
		return nil, "", errors.New("error time_locker")
	}

	err = FindUserIsOnline(reqData.RecipientPeerId)
	if err != nil {
		return nil, "", err
	}

	if reqData.RecipientPeerId == user.PeerId {
		return nil, "", errors.New("you should not send msg to yourself")
	}

	if tool.CheckIsString(&reqData.ChannelIdFrom) == false {
		return nil, "", errors.New("error channel_id_from")
	}

	if reqData.PropertySent < 0 {
		return nil, "", errors.New("error property_sent")
	}

	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelIdFrom),
		q.Eq("CurrState", dao.ChannelState_CanUse),
		q.Eq("PropertyId", reqData.PropertySent),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(&dao.ChannelInfo{})
	if err != nil {
		return nil, "", errors.New("not found this channel_id_from")
	}

	if tool.CheckIsString(&reqData.ChannelIdTo) == false {
		return nil, "", errors.New("error channel_id_to")
	}

	if reqData.PropertyReceived < 0 {
		return nil, "", errors.New("error property_received")
	}
	if reqData.ExchangeRate < 0 {
		return nil, "", errors.New("error exchange_rate")
	}
	if tool.CheckIsString(&reqData.TransactionId) == false {
		return nil, "", errors.New("error transaction_id")
	}
	if tool.CheckIsString(&reqData.TargetTransactionId) == false {
		return nil, "", errors.New("error target_transaction_id")
	}

	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelIdTo),
		q.Eq("CurrState", dao.ChannelState_CanUse),
		q.Eq("PropertyId", reqData.PropertyReceived),
		q.Or(
			q.Eq("PeerIdA", reqData.RecipientPeerId),
			q.Eq("PeerIdB", reqData.RecipientPeerId))).
		First(&dao.ChannelInfo{})
	if err != nil {
		return nil, "", errors.New("not found this channel_id_to")
	}

	err = db.Select(
		q.Eq("CurrHash", reqData.TransactionId),
		q.Eq("TxType", dao.CommitmentTransactionType_Htlc),
		q.Eq("CurrState", dao.TxInfoState_Htlc_GetH),
		q.Eq("ChannelId", reqData.ChannelIdFrom),
		q.Eq("Owner", user.PeerId)).
		First(&dao.CommitmentTransaction{})
	if err != nil {
		return nil, "", errors.New("error transaction_id")
	}

	targetTx := &dao.AtomicSwapInfo{}
	err = db.Select(
		q.Eq("TransactionId", reqData.TargetTransactionId),
		q.Eq("RecipientPeerId", user.PeerId),
	).First(targetTx)
	if err != nil {
		return nil, "", errors.New("not found the target swap transaction")
	}

	swapInfo := &dao.AtomicSwapAcceptedInfo{}

	err = db.Select(
		q.Eq("TransactionId", reqData.TransactionId),
		q.Eq("ChannelIdFrom", reqData.ChannelIdFrom),
		q.Eq("ChannelIdTo", reqData.ChannelIdTo),
		q.Eq("PropertySent", reqData.PropertySent),
		q.Eq("PropertyReceived", reqData.PropertyReceived),
		q.Eq("TargetTransactionId", reqData.TargetTransactionId),
		q.Eq("CreateBy", user.PeerId),
	).First(swapInfo)

	swapInfo.AtomicSwapAccepted = *reqData
	swapInfo.LatestEditAt = time.Now()
	if err == nil {
		err = db.Update(swapInfo)
	} else {
		swapInfo.CreateBy = user.PeerId
		swapInfo.CreateAt = time.Now()
		err = db.Save(swapInfo)
	}
	if err != nil {
		return nil, "", errors.New("fail to save db,try again")
	}

	return reqData, reqData.RecipientPeerId, nil
}
