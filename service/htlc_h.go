package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"sync"
)

type htlcHMessageManager struct {
	operationFlag sync.Mutex
}

var HtlcHMessageService htlcHMessageManager

func (service *htlcHMessageManager) DealHtlcRequest(jsonData string, creator *bean.User) (data *bean.HtlcHRespond, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty json data")
	}
	htlcHRequest := &bean.HtlcHRequest{}
	err = json.Unmarshal([]byte(jsonData), htlcHRequest)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
