package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"time"
)

type fundingTransactionManager struct{}

var FundingTransactionService fundingTransactionManager

//funder request to fund to the multiAddr (channel)
func (service *fundingTransactionManager) CreateFundingTx(jsonData string, user *bean.User) (fundingTransaction *dao.FundingTransaction, err error) {
	data := &bean.FundingCreated{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}

	if len(data.TemporaryChannelId) == 0 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	fundingTransaction = &dao.FundingTransaction{}
	count, _ := db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).Count(fundingTransaction)
	if count == 0 {
		if tool.CheckIsString(&data.FundingTxid) == false {
			return nil, errors.New("wrong FundingTxid")
		}

		channelInfo := &dao.ChannelInfo{}
		err = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).First(channelInfo)
		if err != nil {
			return nil, err
		}
		fundingTransaction.TemporaryChannelId = data.TemporaryChannelId
		fundingTransaction.PropertyId = data.PropertyId
		fundingTransaction.PeerIdA = channelInfo.PeerIdA
		fundingTransaction.PeerIdB = channelInfo.PeerIdB

		fundingTransaction.ChannelPubKey = channelInfo.ChannelPubKey
		fundingTransaction.RedeemScript = channelInfo.RedeemScript

		if user.PeerId == channelInfo.PeerIdA {
			if data.FunderPubKey != channelInfo.PubKeyA {
				return nil, errors.New("invalid FunderPubKey")
			}
		}
		if user.PeerId == channelInfo.PeerIdB {
			if data.FunderPubKey != channelInfo.PubKeyB {
				return nil, errors.New("invalid FunderPubKey")
			}
		}

		if data.FunderPubKey == channelInfo.PubKeyA {
			fundingTransaction.FunderPubKey = channelInfo.PubKeyA
			fundingTransaction.FundeePubKey = channelInfo.PubKeyB
		} else {
			fundingTransaction.FunderPubKey = channelInfo.PubKeyB
			fundingTransaction.FundeePubKey = channelInfo.PubKeyA
		}
		fundingTransaction.AmountA = data.AmountA
		fundingTransaction.FundingTxid = data.FundingTxid
		fundingTransaction.FundingOutputIndex = data.FundingOutputIndex
		fundingTransaction.FunderPubKey2ForCommitment = data.FunderPubKey2

		hash, _ := chainhash.NewHashFromStr(fundingTransaction.FundingTxid)
		op := &bean.OutPoint{
			Hash:  *hash,
			Index: fundingTransaction.FundingOutputIndex,
		}
		fundingTransaction.ChannelId = bean.ChannelIdService.NewChanIDFromOutPoint(op)
		channelInfo.ChannelId = fundingTransaction.ChannelId

		tx, _ := db.Begin(true)
		defer tx.Rollback()

		err = tx.Update(channelInfo)
		fundingTransaction.CurrState = dao.FundingTransactionState_Create
		fundingTransaction.CreateBy = user.PeerId
		fundingTransaction.CreateAt = time.Now()
		err = tx.Save(fundingTransaction)
		tx.Commit()
	} else {
		_ = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(fundingTransaction)
		err = nil
	}
	return fundingTransaction, err
}

func createCommitmentATx(channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, user *bean.User) (*dao.CommitmentTxInfo, error) {
	commitmentTxInfo := &dao.CommitmentTxInfo{}
	commitmentTxInfo.PeerIdA = channelInfo.PeerIdA
	commitmentTxInfo.PeerIdB = channelInfo.PeerIdB
	commitmentTxInfo.ChannelId = channelInfo.ChannelId
	commitmentTxInfo.PropertyId = fundingTransaction.PropertyId
	commitmentTxInfo.CreateSide = 0

	//input
	commitmentTxInfo.InputTxid = fundingTransaction.FundingTxid
	commitmentTxInfo.InputVout = fundingTransaction.FundingOutputIndex
	commitmentTxInfo.InputAmount = fundingTransaction.AmountA + fundingTransaction.AmountB

	//output
	commitmentTxInfo.PubKeyA2 = fundingTransaction.FunderPubKey2ForCommitment
	commitmentTxInfo.PubKeyB = channelInfo.PubKeyB
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{commitmentTxInfo.PubKeyA2, commitmentTxInfo.PubKeyB})
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.MultiAddress = gjson.Get(multiAddr, "address").String()
	commitmentTxInfo.RedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	commitmentTxInfo.AmountM = fundingTransaction.AmountA
	commitmentTxInfo.AmountB = fundingTransaction.AmountB

	commitmentTxInfo.CreateBy = user.PeerId
	commitmentTxInfo.CreateAt = time.Now()
	commitmentTxInfo.LastEditTime = time.Now()

	return commitmentTxInfo, nil
}
func createRDaTx(channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTxInfo, user *bean.User) (*dao.RevocableDeliveryTransaction, error) {
	rda := &dao.RevocableDeliveryTransaction{}
	rda.PeerIdA = channelInfo.PeerIdA
	rda.PeerIdB = channelInfo.PeerIdB
	rda.ChannelId = channelInfo.ChannelId
	rda.PropertyId = commitmentTxInfo.PropertyId
	rda.CreateSide = 0

	//input
	rda.InputTxid = commitmentTxInfo.Txid
	rda.InputVout = 0
	rda.InputAmount = commitmentTxInfo.AmountM
	//output
	rda.PubKeyA = channelInfo.PubKeyA
	rda.Sequnence = 1000
	rda.Amount = commitmentTxInfo.AmountM

	rda.CreateBy = user.PeerId
	rda.CreateAt = time.Now()
	rda.LastEditTime = time.Now()

	return rda, nil
}

func (service *fundingTransactionManager) FundingTransactionSign(jsonData string, user *bean.User) (signed *dao.FundingTransaction, err error) {
	data := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}

	if len(data.ChannelId) == 0 {
		return nil, errors.New("wrong ChannelId")
	}

	var fundingTransaction = &dao.FundingTransaction{}
	err = db.One("ChannelId", data.ChannelId, fundingTransaction)
	if err != nil {
		return nil, err
	}
	if data.Attitude {
		if len(data.FundeeSignature) == 0 {
			return nil, errors.New("wrong FundeeSignature")
		}
		fundingTransaction.AmountB = data.AmountB
		// temp storage private key of bob use to create C1
		fundingTransaction.FundeeSignature = data.FundeeSignature
		fundingTransaction.CurrState = dao.FundingTransactionState_Accept
	} else {
		fundingTransaction.CurrState = dao.FundingTransactionState_Defuse
	}
	fundingTransaction.FundeeSignAt = time.Now()

	tx, _ := db.Begin(true)
	defer tx.Rollback()

	if data.Attitude {
		channelInfo := &dao.ChannelInfo{}
		err = tx.Select(q.Eq("ChannelId", data.ChannelId)).First(channelInfo)
		if err != nil {
			return nil, err
		}

		fundingTransaction := &dao.FundingTransaction{}
		_ = db.Select(q.Eq("ChannelId", data.ChannelId)).First(fundingTransaction)

		// create C1a tx
		commitmentTxInfo, err := createCommitmentATx(channelInfo, fundingTransaction, user)
		if err != nil {
			return nil, err
		}

		txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
			channelInfo.ChannelPubKey,
			[]string{
				data.FundeeSignature,
			},
			[]rpc.TransactionInputItem{
				{commitmentTxInfo.InputTxid, commitmentTxInfo.InputVout, commitmentTxInfo.InputAmount},
			},
			[]rpc.TransactionOutputItem{
				{commitmentTxInfo.MultiAddress, commitmentTxInfo.AmountM},
				{commitmentTxInfo.PubKeyB, commitmentTxInfo.AmountB},
			},
			0,
			nil)
		if err != nil {
			return nil, err
		}
		commitmentTxInfo.Txid = txid
		commitmentTxInfo.TxHexFirstSign = hex
		commitmentTxInfo.FirstSignAt = time.Now()
		commitmentTxInfo.CurrState = dao.TxInfoState_OtherSign
		_ = tx.Save(commitmentTxInfo)

		// create RDa tx
		rdTransaction, err := createRDaTx(channelInfo, commitmentTxInfo, user)
		if err != nil {
			return nil, err
		}

		txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
			commitmentTxInfo.MultiAddress,
			[]string{
				data.FundeeSignature,
			},
			[]rpc.TransactionInputItem{
				{rdTransaction.InputTxid, rdTransaction.InputVout, rdTransaction.InputAmount},
			},
			[]rpc.TransactionOutputItem{
				{rdTransaction.PubKeyA, rdTransaction.Amount},
			},
			0,
			&rdTransaction.Sequnence)
		if err != nil {
			return nil, err
		}
		rdTransaction.Txid = txid
		rdTransaction.TxHexFirstSign = hex
		rdTransaction.FirstSignAt = time.Now()
		rdTransaction.CurrState = dao.TxInfoState_OtherSign
		_ = tx.Save(rdTransaction)
	}
	err = tx.Update(fundingTransaction)
	tx.Commit()
	return fundingTransaction, err
}

func (service *fundingTransactionManager) ItemByTempId(jsonData string) (node *dao.FundingTransaction, err error) {
	var tempChanId chainhash.Hash
	for index, item := range gjson.Parse(jsonData).Array() {
		tempChanId[index] = byte(item.Int())
	}
	return service.ItemByTempIdArray(tempChanId)
}

func (service *fundingTransactionManager) ItemByTempIdArray(tempId chainhash.Hash) (node *dao.FundingTransaction, err error) {
	var data = &dao.FundingTransaction{}
	err = db.One("TemporaryChannelId", tempId, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) AllItem(peerId string) (node []dao.FundingTransaction, err error) {
	var data []dao.FundingTransaction
	err = db.Select(q.Or(q.Eq("PeerIdB", peerId), q.Eq("PeerIdA", peerId))).OrderBy("CreateAt").Reverse().Find(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) ItemById(id int) (node *dao.FundingTransaction, err error) {
	var data = &dao.FundingTransaction{}
	err = db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) DelAll() (err error) {
	var data = &dao.FundingTransaction{}
	return db.Drop(data)
}

func (service *fundingTransactionManager) Del(id int) (err error) {
	var data = &dao.FundingTransaction{}
	count, err := db.Select(q.Eq("Id", id)).Count(data)
	if err == nil && count == 1 {
		err = db.DeleteStruct(data)
	}
	return err
}
func (service *fundingTransactionManager) TotalCount(peerId string) (count int, err error) {
	return db.Select(q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).Count(&dao.FundingTransaction{})
}
