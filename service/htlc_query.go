package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"log"
)

type htlcQueryManager struct{}

var HtlcQueryService = htlcQueryManager{}

func (service *htlcQueryManager) GetRFromCommitmentTx(msgData string, user bean.User) (r string, err error) {

	if tool.CheckIsString(&msgData) == false {
		return r, errors.New("error input data")
	}
	reqData := bean.ChannelIdReq{}
	err = json.Unmarshal([]byte(msgData), &reqData)
	if err != nil {
		return r, err
	}
	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		return r, errors.New("error ChannelId ")
	}

	commitmentTxInfo, err := getLatestCommitmentTx(reqData.ChannelId, user.PeerId)
	if err != nil {
		return r, err
	}
	if commitmentTxInfo.TxType != dao.CommitmentTransactionType_Htlc {
		return r, errors.New("error tx type")
	}
	r = commitmentTxInfo.HtlcR
	if tool.CheckIsString(&r) == false {
		err = errors.New("empty R")
	}
	return r, err
}

func (service *htlcQueryManager) GetPathInfoByH(msgData string, user bean.User) (pathInfo *dao.HtlcSingleHopPathInfo, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, errors.New("error input data")
	}

	pathInfo = &dao.HtlcSingleHopPathInfo{}
	err = db.Select(q.Eq("H", msgData)).First(pathInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return pathInfo, nil
}
func (service *htlcQueryManager) GetRByHOfOwner(msgData string, user bean.User) (r string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return "", errors.New("error input data")
	}

	info := &dao.HtlcRAndHInfo{}
	err = db.Select(q.Eq("H", msgData), q.Eq("SignBy", user.PeerId)).First(info)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return info.R, nil
}
