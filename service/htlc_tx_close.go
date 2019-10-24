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
type htlcCloseTxManager struct {
	operationFlag sync.Mutex
}

// htlc 关闭htlc交易
var HtlcCloseTxService htlcCloseTxManager

// -48
func (service *htlcCloseTxManager) RequestCloseHtlc(msgData string, user bean.User) (outData interface{}, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, errors.New("empty json data")
	}

	reqData := &bean.HtlcRequestCloseCurrTx{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		err = errors.New("wrong channel Id")
		log.Println(err)
		return nil, err
	}

	commitmentTxInfo, err := getHtlcLatestCommitmentTx(reqData.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(commitmentTxInfo)

	channelInfo := dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", commitmentTxInfo.ChannelId), q.Eq("", dao.ChannelState_HtlcBegin)).First(&channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(channelInfo)

	htlcTimeoutTx := dao.HTLCTimeoutTxA{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", commitmentTxInfo.Id), q.Eq("Owner", user.PeerId)).First(&htlcTimeoutTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if user.PeerId == channelInfo.PeerIdA {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = reqData.ChannelAddressPrivateKey
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = reqData.ChannelAddressPrivateKey
	}
	tempAddrPrivateKeyMap[commitmentTxInfo.RSMCTempAddressPubKey] = reqData.LastRsmcTempAddressPrivateKey
	tempAddrPrivateKeyMap[commitmentTxInfo.HTLCTempAddressPubKey] = reqData.LastHtlcTempAddressPrivateKey
	tempAddrPrivateKeyMap[htlcTimeoutTx.RSMCTempAddressPubKey] = reqData.LastHtlcTempAddressForHt1aPrivateKey
	tempAddrPrivateKeyMap[reqData.CurrRsmcTempAddressPubKey] = reqData.CurrRsmcTempAddressPrivateKey

	info := dao.HtlcRequestCloseCurrTxInfo{}
	info.ChannelId = commitmentTxInfo.ChannelId
	info.CurrRsmcTempAddressPubKey = reqData.CurrRsmcTempAddressPubKey
	info.CreateAt = time.Now()
	info.CreateBy = user.PeerId
	infoBytes, _ := json.Marshal(info)
	requestHash := tool.SignMsgWithSha256(infoBytes)
	info.RequestHash = requestHash
	err = db.Save(&info)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return info, nil
}

// -49
func (service *htlcCloseTxManager) SignCloseHtlc(msgData string, user bean.User) (outData interface{}, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, errors.New("empty json data")
	}

	reqData := &bean.HtlcSignCloseCurrTx{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcRequestCloseCurrTxInfo := dao.HtlcRequestCloseCurrTxInfo{}
	err = db.Select(q.Eq("RequestHash", reqData.RequestCloseHtlcHash)).First(&htlcRequestCloseCurrTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return nil, nil
}

func (service *htlcCloseTxManager) RequestCloseChannel(msgData string, user bean.User) error {

	return nil
}

func (service *htlcCloseTxManager) SignCloseChannel(msgData string, user bean.User) error {

	return nil
}
