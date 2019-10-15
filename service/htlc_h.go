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

func (service *htlcHMessageManager) DealHtlcResponse(jsonData string, user *bean.User) (data interface{}, senderPeerId *string, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, nil, errors.New("empty json data")
	}
	htlcHRespond := &bean.HtlcHRespond{}
	err = json.Unmarshal([]byte(jsonData), htlcHRespond)
	if err != nil {
		return nil, nil, err
	}

	createRandHInfo := &dao.HtlcCreateRandHInfo{}
	err = db.Select(q.Eq("RequestHash", htlcHRespond.RequestHash), q.Eq("CurrState", dao.NS_Create)).First(createRandHInfo)
	if err != nil {
		return nil, nil, err
	}

	if htlcHRespond.Approval {
		s, _ := tool.RandBytes(32)
		temp := append([]byte(createRandHInfo.RequestHash), s...)
		log.Println(temp)
		r := tool.SignMsg(temp)
		log.Println(r)
		h := tool.SignMsg([]byte(r))
		log.Println(h)

		createRandHInfo.R = r
		createRandHInfo.H = r
		createRandHInfo.CurrState = dao.NS_Finish
	} else {
		createRandHInfo.CurrState = dao.NS_Refuse
	}
	createRandHInfo.SignAt = time.Now()
	createRandHInfo.SignBy = user.PeerId
	err = db.Update(createRandHInfo)
	if err != nil {
		return nil, &createRandHInfo.SenderPeerId, err
	}

	responseData := make(map[string]interface{})
	responseData["id"] = createRandHInfo.Id
	responseData["request_hash"] = htlcHRespond.RequestHash
	responseData["approval"] = htlcHRespond.Approval
	if htlcHRespond.Approval {
		responseData["h"] = createRandHInfo.H
	}
	return responseData, &createRandHInfo.SenderPeerId, nil
}

func (service *htlcHMessageManager) GetHtlcCreatedRandHInfoList(user *bean.User) (data interface{}, err error) {
	var createRandHInfoList []dao.HtlcCreateRandHInfo
	err = db.Select(q.Eq("CreateBy", user.PeerId)).Find(&createRandHInfoList)
	if err != nil {
		return nil, err
	}
	for _, item := range createRandHInfoList {
		item.R = ""
	}
	return createRandHInfoList, nil
}

func (service *htlcHMessageManager) GetHtlcCreatedRandHInfoItem(msgData string, user *bean.User) (data interface{}, err error) {
	var createRandHInfo dao.HtlcCreateRandHInfo
	err = db.Select(q.Eq("Id", msgData), q.Eq("CreateBy", user.PeerId)).First(&createRandHInfo)
	if err != nil {
		return nil, err
	}
	createRandHInfo.R = ""
	return createRandHInfo, nil
}

func (service *htlcHMessageManager) GetHtlcSignedRandHInfoList(user *bean.User) (data interface{}, err error) {
	var createRandHInfoList []dao.HtlcCreateRandHInfo
	err = db.Select(q.Eq("RecipientPeerId", user.PeerId), q.Eq("SignBy", user.PeerId)).Find(&createRandHInfoList)
	if err != nil {
		return nil, err
	}
	return createRandHInfoList, nil
}

func (service *htlcHMessageManager) GetHtlcSignedRandHInfoItem(msgData string, user *bean.User) (data interface{}, err error) {
	var createRandHInfo dao.HtlcCreateRandHInfo
	err = db.Select(q.Eq("Id", msgData), q.Eq("SignBy", user.PeerId)).First(&createRandHInfo)
	if err != nil {
		return nil, err
	}
	return createRandHInfo, nil
}
