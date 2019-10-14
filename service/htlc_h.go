package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
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

	data = &bean.HtlcHRespond{}
	data.PropertyId = htlcHRequest.PropertyId
	data.Amount = htlcHRequest.Amount
	data.RequestHash = ""
	return data, nil
}
func (service *htlcHMessageManager) DealHtlcResponse(jsonData string, creator *bean.User) (data *dao.HtlcCreateRandHInfo, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty json data")
	}
	htlcHRespond := &bean.HtlcHRespond{}
	err = json.Unmarshal([]byte(jsonData), htlcHRespond)
	if err != nil {
		return nil, err
	}

	return data, nil
}
