package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"log"
	"sync"
)

//close htlc or close channel
type htlcCloseTxManager struct {
	operationFlag sync.Mutex
}

// htlc 关闭htlc交易
var HtlcCloseTxService htlcCloseTxManager

// -47
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

	return commitmentTxInfo, nil
}

func (service *htlcCloseTxManager) SignCloseHtlc(msgData string, user bean.User) error {

	return nil
}

func (service *htlcCloseTxManager) RequestCloseChannel(msgData string, user bean.User) error {

	return nil
}

func (service *htlcCloseTxManager) SignCloseChannel(msgData string, user bean.User) error {

	return nil
}
