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

	tx, err := db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()
	createRandHInfo := &dao.HtlcCreateRandHInfo{}
	createRandHInfo.SenderPeerId = creator.PeerId
	createRandHInfo.RecipientPeerId = htlcHRequest.RecipientPeerId
	createRandHInfo.PropertyId = htlcHRequest.PropertyId
	createRandHInfo.Amount = htlcHRequest.Amount
	createRandHInfo.CurrState = dao.NS_Create
	createRandHInfo.CreateAt = time.Now()
	createRandHInfo.CreateBy = creator.PeerId
	err = tx.Save(createRandHInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	bytes, err := json.Marshal(createRandHInfo)
	msgHash := tool.SignMsg(bytes)
	createRandHInfo.RequestHash = msgHash
	err = tx.Update(createRandHInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	data = &bean.HtlcHRespond{}
	data.PropertyId = htlcHRequest.PropertyId
	data.Amount = htlcHRequest.Amount
	data.RequestHash = msgHash

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

	createRandHInfo := dao.HtlcCreateRandHInfo{}
	err = db.Select(q.Eq("RequestHash", htlcHRespond.RequestHash), q.Eq("CurrState", dao.NS_Create)).First(&createRandHInfo)
	if err != nil {
		return nil, err
	}

	return data, nil
}
