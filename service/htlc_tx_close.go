package service

import (
	"LightningOnOmni/bean"
	"sync"
)

//close htlc or close channel
type htlcCloseTxManager struct {
	operationFlag sync.Mutex
}

// htlc 关闭htlc交易
var HtlcCloseTxService htlcCloseTxManager

func (service *htlcCloseTxManager) RequestCloseHtlc(msgData string, user bean.User) error {

	return nil
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
