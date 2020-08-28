package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
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
func (this *atomicSwapManager) AtomicSwap(msg bean.RequestMessage, user bean.User) (outData interface{}, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New("empty json data")
	}
	reqData := &bean.AtomicSwapRequest{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.RecipientUserPeerId) == false {
		return nil, errors.New("error recipient_user_peer_id")
	}

	if msg.RecipientUserPeerId != reqData.RecipientUserPeerId {
		return nil, errors.New("wrong recipient_user_peer_id")
	}

	if reqData.TimeLocker < 10 {
		return nil, errors.New("error time_locker")
	}
	if reqData.RecipientUserPeerId == user.PeerId {
		return nil, errors.New("you should not send msg to yourself")
	}

	err = findUserIsOnline(msg.RecipientNodePeerId, reqData.RecipientUserPeerId)
	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&reqData.ChannelIdFrom) == false {
		return nil, errors.New("error channel_id_from")
	}

	_, err = rpcClient.OmniGetProperty(reqData.PropertySent)
	if err != nil {
		log.Println(err)
		return nil, errors.New("error property_sent")
	}

	err = user.Db.Select(
		q.Eq("ChannelId", reqData.ChannelIdFrom),
		q.Eq("CurrState", dao.ChannelState_HtlcTx),
		q.Eq("PropertyId", reqData.PropertySent),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(&dao.ChannelInfo{})
	if err != nil {
		return nil, errors.New("not found this channel_id_from")
	}

	if tool.CheckIsString(&reqData.ChannelIdTo) == false {
		return nil, errors.New("error channel_id_to")
	}

	_, err = rpcClient.OmniGetProperty(reqData.PropertyReceived)
	if err != nil {
		log.Println(err)
		return nil, errors.New("error property_received")
	}

	if reqData.ExchangeRate < 0 {
		return nil, errors.New("error exchange_rate")
	}
	if tool.CheckIsString(&reqData.TransactionId) == false {
		return nil, errors.New("error transaction_id")
	}

	flag := httpGetChannelStateFromTracker(reqData.ChannelIdTo)
	if flag == 0 {
		return nil, errors.New("not found this channel_id_to")
	}

	err = user.Db.Select(
		q.Eq("CurrHash", reqData.TransactionId),
		q.Eq("TxType", dao.CommitmentTransactionType_Htlc),
		q.Eq("CurrState", dao.TxInfoState_Htlc_GetH),
		q.Eq("ChannelId", reqData.ChannelIdFrom),
		q.Eq("Owner", user.PeerId)).
		First(&dao.CommitmentTransaction{})
	if err != nil {
		return nil, errors.New("error transaction_id")
	}

	swapInfo := &dao.AtomicSwapInfo{}
	//check if have send the same info
	err = user.Db.Select(
		q.Eq("TransactionId", reqData.TransactionId),
		q.Eq("ChannelIdFrom", reqData.ChannelIdFrom),
		q.Eq("ChannelIdTo", reqData.ChannelIdTo),
		q.Eq("PropertySent", reqData.PropertySent),
		q.Eq("PropertyReceived", reqData.PropertyReceived),
		q.Eq("CreateBy", user.PeerId),
	).First(swapInfo)

	swapInfo.LatestEditAt = time.Now()
	swapInfo.AtomicSwapRequest = *reqData
	if swapInfo.Id > 0 {
		err = user.Db.Update(swapInfo)
	} else {
		swapInfo.CreateBy = user.PeerId
		swapInfo.CreateAt = time.Now()
		err = user.Db.Save(swapInfo)
	}
	if err != nil {
		return nil, errors.New("fail to save db,try again")
	}
	return reqData, nil
}

//80的信息到达接受者的obd节点的处理
func (this *atomicSwapManager) BeforeSignAtomicSwapAtBobSide(data string, user *bean.User) (retData interface{}, err error) {
	reqData := &bean.AtomicSwapRequest{}
	err = json.Unmarshal([]byte(data), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	swapInfo := &dao.AtomicSwapInfo{}
	//check if have send the same info
	err = user.Db.Select(
		q.Eq("TransactionId", reqData.TransactionId),
		q.Eq("ChannelIdFrom", reqData.ChannelIdFrom),
		q.Eq("ChannelIdTo", reqData.ChannelIdTo),
		q.Eq("PropertySent", reqData.PropertySent),
		q.Eq("PropertyReceived", reqData.PropertyReceived),
		q.Eq("CreateBy", user.PeerId),
	).First(swapInfo)

	swapInfo.LatestEditAt = time.Now()
	swapInfo.AtomicSwapRequest = *reqData
	if swapInfo.Id > 0 {
		err = user.Db.Update(swapInfo)
	} else {
		swapInfo.CreateBy = user.PeerId
		swapInfo.CreateAt = time.Now()
		err = user.Db.Save(swapInfo)
	}
	if err != nil {
		log.Println(err)
	}

	return reqData, nil
}

//MsgType_Atomic_Swap_Accept_N81 可以理解为发货凭证
func (this *atomicSwapManager) AtomicSwapAccepted(msg bean.RequestMessage, user bean.User) (outData interface{}, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New("empty json data")
	}
	reqData := &bean.AtomicSwapAccepted{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.RecipientUserPeerId) == false {
		return nil, errors.New("error recipient_user_peer_id")
	}

	if reqData.RecipientUserPeerId == user.PeerId {
		return nil, errors.New("you should not send msg to yourself")
	}

	if msg.RecipientUserPeerId != reqData.RecipientUserPeerId {
		return nil, errors.New("wrong recipient_user_peer_id")
	}

	err = findUserIsOnline(msg.RecipientNodePeerId, reqData.RecipientUserPeerId)
	if err != nil {
		return nil, err
	}

	if reqData.TimeLocker < 10 {
		return nil, errors.New("error time_locker")
	}

	if tool.CheckIsString(&reqData.ChannelIdFrom) == false {
		return nil, errors.New("error channel_id_from")
	}

	_, err = rpcClient.OmniGetProperty(reqData.PropertySent)
	if err != nil {
		log.Println(err)
		return nil, errors.New("error property_sent")
	}

	err = user.Db.Select(
		q.Eq("ChannelId", reqData.ChannelIdFrom),
		q.Eq("CurrState", dao.ChannelState_HtlcTx),
		q.Eq("PropertyId", reqData.PropertySent),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(&dao.ChannelInfo{})
	if err != nil {
		return nil, errors.New("not found this channel of channel_id_from")
	}

	if tool.CheckIsString(&reqData.ChannelIdTo) == false {
		return nil, errors.New("error channel_id_to")
	}

	_, err = rpcClient.OmniGetProperty(reqData.PropertyReceived)
	if err != nil {
		log.Println(err)
		return nil, errors.New("error property_received")
	}

	if reqData.ExchangeRate < 0 {
		return nil, errors.New("error exchange_rate")
	}
	if tool.CheckIsString(&reqData.TransactionId) == false {
		return nil, errors.New("error transaction_id")
	}

	err = user.Db.Select(
		q.Eq("CurrHash", reqData.TransactionId),
		q.Eq("TxType", dao.CommitmentTransactionType_Htlc),
		q.Eq("CurrState", dao.TxInfoState_Htlc_GetH),
		q.Eq("ChannelId", reqData.ChannelIdFrom),
		q.Eq("Owner", user.PeerId)).
		First(&dao.CommitmentTransaction{})
	if err != nil {
		return nil, errors.New("error transaction_id")
	}

	if tool.CheckIsString(&reqData.TargetTransactionId) == false {
		return nil, errors.New("error target_transaction_id")
	}

	flag := httpGetChannelStateFromTracker(reqData.ChannelIdTo)
	if flag == 0 {
		return nil, errors.New("not found this channel_id_to")
	}

	targetTx := &dao.AtomicSwapInfo{}
	err = user.Db.Select(
		q.Eq("TransactionId", reqData.TargetTransactionId),
		q.Eq("RecipientUserPeerId", user.PeerId),
	).First(targetTx)
	if err != nil {
		return nil, errors.New("not found the target swap transaction")
	}

	swapInfo := &dao.AtomicSwapAcceptedInfo{}
	err = user.Db.Select(
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
	if swapInfo.Id > 0 {
		err = user.Db.Update(swapInfo)
	} else {
		swapInfo.CreateBy = user.PeerId
		swapInfo.CreateAt = time.Now()
		err = user.Db.Save(swapInfo)
	}

	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to save db,try again")
	}

	return reqData, nil
}

//81的信息到达接受者的obd节点的处理
func (this *atomicSwapManager) BeforeSignAtomicSwapAcceptedAtAliceSide(data string, user *bean.User) (retData interface{}, err error) {
	reqData := &bean.AtomicSwapAccepted{}
	err = json.Unmarshal([]byte(data), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	swapInfo := &dao.AtomicSwapAcceptedInfo{}
	err = user.Db.Select(
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
	if swapInfo.Id > 0 {
		err = user.Db.Update(swapInfo)
	} else {
		swapInfo.CreateBy = user.PeerId
		swapInfo.CreateAt = time.Now()
		err = user.Db.Save(swapInfo)
	}
	if err != nil {
		log.Println(err)
	}

	return reqData, nil
}
