package service

import (
	"LightningOnOmni/bean"
	"sync"
)

type htlcTxManager struct {
	operationFlag sync.Mutex
}

var HtlcTxService htlcTxManager

// query bob,and ask bob
func (service *htlcTxManager) HtlcRequest(msgData string, user bean.User) error {

	return nil
}
