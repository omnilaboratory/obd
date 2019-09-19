package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/config"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"sync"
	"time"
)

type fundingTransactionManager struct {
	operateFlag sync.Mutex
}

var FundingTransactionService fundingTransactionManager

func (service *fundingTransactionManager) CreateFundingBtcTxRequest(jsonData string, user *bean.User) (fundingTransaction map[string]interface{}, err error) {
	reqData := &bean.FundingBtcCreated{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if len(reqData.TemporaryChannelId) == 0 {
		err = errors.New("wrong TemporaryChannelId ")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("wrong ChannelAddressPrivateKey ")
		log.Println(err)
		return nil, err
	}

	btcFeeTxHexDecode, err := rpcClient.DecodeRawTransaction(reqData.FundingTxHex)
	if err != nil {
		err = errors.New("BtcFeeTxHex  parse fail " + err.Error())
		log.Println(err)
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("CurrState", dao.ChannelState_Accept), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//get btc miner Fee data from transaction
	fundingTxid, amount, _, err := checkBtcTxHex(btcFeeTxHexDecode, channelInfo, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	out, _ := decimal.NewFromFloat(rpc.GetMinerFee()).Add(decimal.NewFromFloat(config.Dust)).Float64()
	if amount < out {
		err = errors.New("error btc amount")
		log.Println(err)
		return nil, err
	}

	tx, _ := db.Begin(true)
	defer tx.Rollback()

	fundingBtcRequest := &dao.FundingBtcRequest{}
	count, _ := tx.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("TxId", fundingTxid), q.Eq("Owner", user.PeerId), q.Eq("IsEnable", true), q.Eq("IsFinish", true)).Count(fundingBtcRequest)
	if count != 0 {
		err = errors.New("have funding btc fee")
		log.Println(err)
		return nil, err
	}

	pubKey := channelInfo.PubKeyA
	if user.PeerId == channelInfo.PeerIdB {
		pubKey = channelInfo.PubKeyB
	}
	tempAddrPrivateKeyMap[pubKey] = reqData.ChannelAddressPrivateKey

	_ = tx.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("TxId", fundingTxid), q.Eq("Owner", user.PeerId), q.Eq("IsEnable", true)).Find(fundingBtcRequest)
	if fundingBtcRequest.Id == 0 {
		fundingBtcRequest = &dao.FundingBtcRequest{}
		fundingBtcRequest.Owner = user.PeerId
		fundingBtcRequest.TemporaryChannelId = reqData.TemporaryChannelId
		fundingBtcRequest.TxHash = reqData.FundingTxHex
		fundingBtcRequest.TxId = fundingTxid
		fundingBtcRequest.IsEnable = true
		fundingBtcRequest.CreateAt = time.Now()
		fundingBtcRequest.Amount = amount
		fundingBtcRequest.IsFinish = false
		err = tx.Save(fundingBtcRequest)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	tx.Commit()
	node := make(map[string]interface{})
	node["temporary_channel_id"] = reqData.TemporaryChannelId
	node["amount"] = amount
	node["fundingTxid"] = fundingTxid
	return node, nil
}

func (service *fundingTransactionManager) FundingBtcTxSign(jsonData string, signer *bean.User) (outData interface{}, err error) {
	reqData := &bean.FundingBtcSigned{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if len(reqData.TemporaryChannelId) == 0 {
		err = errors.New("wrong TemporaryChannelId ")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("wrong ChannelAddressPrivateKey ")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&reqData.FundingTxid) == false {
		err = errors.New("wrong ChannelAddressPrivateKey ")
		log.Println(err)
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("CurrState", dao.ChannelState_Accept), q.Or(q.Eq("PeerIdA", signer.PeerId), q.Eq("PeerIdB", signer.PeerId))).OrderBy("CreateAt").Reverse().First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	toAddress := channelInfo.AddressA
	creator := channelInfo.PeerIdA
	pubKey := channelInfo.PubKeyB
	if signer.PeerId == channelInfo.PeerIdA {
		creator = channelInfo.PeerIdB
		pubKey = channelInfo.PubKeyA
		toAddress = channelInfo.AddressB
	}

	funderPrivateKey := tempAddrPrivateKeyMap[pubKey]
	if tool.CheckIsString(&funderPrivateKey) == false {
		err = errors.New("wrong funderPrivateKey ")
		log.Println(err)
		return nil, err
	}

	fundingBtcRequest := &dao.FundingBtcRequest{}
	err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("TxId", reqData.FundingTxid), q.Eq("Owner", creator), q.Eq("IsEnable", true), q.Eq("IsFinish", false)).Find(fundingBtcRequest)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if reqData.Approval == false {
		fundingBtcRequest.FinishAt = time.Now()
		_ = db.UpdateField(fundingBtcRequest, "IsFinish", true)
		_ = db.UpdateField(fundingBtcRequest, "IsEnable", false)
		_ = db.Update(fundingBtcRequest)
		return fundingBtcRequest, nil
	}

	//TODO create miner fee redeem transaction
	btcFeeTxHexDecode, err := rpcClient.DecodeRawTransaction(fundingBtcRequest.TxHash)
	if err != nil {
		err = errors.New("BtcFeeTxHex  parse fail " + err.Error())
		log.Println(err)
		return nil, err
	}
	_, _, vout, err := checkBtcTxHex(btcFeeTxHexDecode, channelInfo, creator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result, err := rpcClient.SendRawTransaction(fundingBtcRequest.TxHash)
	if err != nil {
		return nil, err
	}
	log.Println(result)

	minerFeeRedeemTransaction := &dao.MinerFeeRedeemTransaction{}
	txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
		channelInfo.ChannelAddress,
		[]string{
			funderPrivateKey,
			reqData.ChannelAddressPrivateKey},
		[]rpc.TransactionInputItem{
			{
				Txid:         fundingBtcRequest.TxId,
				Vout:         vout,
				ScriptPubKey: channelInfo.ChannelAddressScriptPubKey},
		},
		[]rpc.TransactionOutputItem{
			{
				ToBitCoinAddress: toAddress,
				Amount:           fundingBtcRequest.Amount},
		},
		0,
		0,
		&channelInfo.ChannelAddressRedeemScript)
	if err != nil {
		return nil, err
	}
	minerFeeRedeemTransaction.Txid = txid
	minerFeeRedeemTransaction.TxHash = hex
	minerFeeRedeemTransaction.CreateAt = time.Now()
	minerFeeRedeemTransaction.Owner = creator
	minerFeeRedeemTransaction.TemporaryChannelId = reqData.TemporaryChannelId
	_ = db.Save(minerFeeRedeemTransaction)

	return minerFeeRedeemTransaction, nil
}

//funder request to fund to the multiAddr (channel)
func (service *fundingTransactionManager) CreateFundingOmniTxRequest(jsonData string, user *bean.User) (fundingTransaction *dao.FundingTransaction, err error) {
	reqData := &bean.FundingCreated{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if len(reqData.TemporaryChannelId) == 0 {
		err = errors.New("wrong TemporaryChannelId ")
		log.Println(err)
		return nil, err
	}

	if _, err := getAddressFromPubKey(reqData.TempAddressPubKey); err != nil {
		err = errors.New("wrong TempAddressPubKey ")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.TempAddressPrivateKey) == false {
		err = errors.New("wrong TempAddressPrivateKey ")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("wrong ChannelAddressPrivateKey ")
		log.Println(err)
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("CurrState", dao.ChannelState_Accept), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = checkBtcFundFinish(channelInfo.ChannelAddress)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// if alice launch funding
	if user.PeerId == channelInfo.PeerIdA {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = reqData.ChannelAddressPrivateKey
	} else { // if bob launch funding
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = reqData.ChannelAddressPrivateKey
	}
	tempAddrPrivateKeyMap[reqData.TempAddressPubKey] = reqData.TempAddressPrivateKey

	count, _ := db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId)).Count(&dao.FundingTransaction{})
	if count == 0 {
		if tool.CheckIsString(&reqData.FundingTxHex) == false {
			err = errors.New("wrong TxHash ")
			log.Println(err)
			return nil, err
		}

		fundingTxHexDecode, err := rpcClient.OmniDecodeTransaction(reqData.FundingTxHex)
		if err != nil {
			err = errors.New("TxHash  parse fail " + err.Error())
			log.Println(err)
			return nil, err
		}

		// sync locker
		service.operateFlag.Lock()
		defer service.operateFlag.Unlock()

		if bean.ChannelIdService.IsEmpty(channelInfo.ChannelId) == false {
			err = errors.New("channel is used, can not funding again")
			log.Println(err)
			return nil, err
		}

		fundingTxid, amountA, propertyId, err := checkOmniTxHex(fundingTxHexDecode, channelInfo, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		reqData.PropertyId = propertyId
		// getProperty from omnicore
		result, err := rpcClient.OmniGetProperty(reqData.PropertyId)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(result)

		btcTxHashDecode, err := rpcClient.DecodeRawTransaction(reqData.FundingTxHex)
		if err != nil {
			err = errors.New("BtcFeeTxHex  parse fail " + err.Error())
			log.Println(err)
			return nil, err
		}

		//get btc miner Fee data from transaction
		fundingTxid, _, fundingOutputIndex, err := checkBtcTxHex(btcTxHashDecode, channelInfo, user.PeerId)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		hash, _ := chainhash.NewHashFromStr(fundingTxid)
		op := &bean.OutPoint{
			Hash:  *hash,
			Index: fundingOutputIndex,
		}

		fundingTransaction = &dao.FundingTransaction{}
		fundingTransaction.ChannelId = bean.ChannelIdService.NewChanIDFromOutPoint(op)
		channelInfo.ChannelId = fundingTransaction.ChannelId

		count, err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId)).Count(channelInfo)
		if err != nil || count != 0 {
			err = errors.New("fundingTx have been used")
			log.Println(err)
			return nil, err
		}

		fundingTransaction.ChannelInfoId = channelInfo.Id
		fundingTransaction.PropertyId = reqData.PropertyId
		fundingTransaction.PeerIdA = channelInfo.PeerIdA
		fundingTransaction.PeerIdB = channelInfo.PeerIdB

		// if alice launch funding
		if user.PeerId == channelInfo.PeerIdA {
			fundingTransaction.AmountA = amountA
			fundingTransaction.FunderAddress = channelInfo.AddressA
			tempAddrPrivateKeyMap[channelInfo.PubKeyA] = reqData.ChannelAddressPrivateKey
		} else { // if bob launch funding
			fundingTransaction.AmountB = amountA
			fundingTransaction.FunderAddress = channelInfo.AddressB
			tempAddrPrivateKeyMap[channelInfo.PubKeyB] = reqData.ChannelAddressPrivateKey
		}
		fundingTransaction.FundingTxHex = reqData.FundingTxHex
		fundingTransaction.FundingTxid = fundingTxid
		fundingTransaction.FundingOutputIndex = fundingOutputIndex
		fundingTransaction.FunderPubKey2ForCommitment = reqData.TempAddressPubKey
		tempAddrPrivateKeyMap[fundingTransaction.FunderPubKey2ForCommitment] = reqData.TempAddressPrivateKey

		tx, err := db.Begin(true)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer tx.Rollback()

		err = tx.Update(channelInfo)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		fundingTransaction.CurrState = dao.FundingTransactionState_Create
		fundingTransaction.CreateBy = user.PeerId
		fundingTransaction.CreateAt = time.Now()
		err = tx.Save(fundingTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId)).First(fundingTransaction)
		log.Println(err)
	}
	return fundingTransaction, err
}

func checkBtcFundFinish(address string) error {
	result, err := rpcClient.ListUnspent(address)
	if err != nil {
		return err
	}
	array := gjson.Parse(result).Array()
	if len(array) == 0 {
		return errors.New("empty balance")
	}
	if len(array) < 2 {
		return errors.New("btc fund have been not finished")
	}
	log.Println("listunspent", array)

	out, _ := decimal.NewFromFloat(rpc.GetMinerFee()).Add(decimal.NewFromFloat(0.00000546)).Float64()
	for _, item := range array {
		amount := item.Get("amount").Float()
		if amount < out {
			return errors.New("btc amount error")
		}
	}
	return nil
}

func (service *fundingTransactionManager) FundingOmniTxSign(jsonData string, signer *bean.User) (signed *dao.FundingTransaction, err error) {
	reqData := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return nil, err
	}

	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		return nil, errors.New("wrong ChannelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		log.Println("channel not find")
		return nil, err
	}

	// default if alice launch the funding,signer is bob
	var owner = channelInfo.PeerIdA
	if signer.PeerId == channelInfo.PeerIdA {
		owner = channelInfo.PeerIdB
	}

	var fundingTransaction = &dao.FundingTransaction{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.FundingTransactionState_Create)).First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if reqData.Approval {
		if tool.CheckIsString(&reqData.FundeeChannelAddressPrivateKey) == false {
			return nil, errors.New("wrong FundeeChannelAddressPrivateKey")
		}
		fundingTransaction.CurrState = dao.FundingTransactionState_Accept
	} else {
		fundingTransaction.CurrState = dao.FundingTransactionState_Defuse
		channelInfo.CurrState = dao.ChannelState_FundingDefuse
	}
	fundingTransaction.FundeeSignAt = time.Now()

	tx, err := db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	if reqData.Approval {
		funderChannelAddressPrivateKey := ""
		if owner == channelInfo.PeerIdA {
			fundingTransaction.AmountB = reqData.AmountB
			funderChannelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyA]
			defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
		} else {
			fundingTransaction.AmountA = reqData.AmountB
			funderChannelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyB]
			defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
		}
		if tool.CheckIsString(&funderChannelAddressPrivateKey) == false {
			err = errors.New("fail to get the funder's channel address private key ")
			log.Println(err)
			return nil, err
		}

		funderTempAddressPrivateKey := tempAddrPrivateKeyMap[fundingTransaction.FunderPubKey2ForCommitment]
		defer delete(tempAddrPrivateKeyMap, fundingTransaction.FunderPubKey2ForCommitment)
		if tool.CheckIsString(&funderTempAddressPrivateKey) == false {
			err = errors.New("fail to get the funder's tmep address private key ")
			log.Println(err)
			return nil, err
		}

		// create C1 tx
		var outputBean = commitmentOutputBean{}
		outputBean.TempPubKey = fundingTransaction.FunderPubKey2ForCommitment
		if owner == channelInfo.PeerIdA {
			outputBean.ToPubKey = channelInfo.PubKeyB
			outputBean.ToAddress = channelInfo.AddressB
			outputBean.AmountM = fundingTransaction.AmountA
			outputBean.AmountB = fundingTransaction.AmountB
		} else {
			outputBean.ToPubKey = channelInfo.PubKeyA
			outputBean.ToAddress = channelInfo.AddressA
			outputBean.AmountM = fundingTransaction.AmountB
			outputBean.AmountB = fundingTransaction.AmountA
		}

		commitmentTxInfo, err := createCommitmentTx(owner, channelInfo, fundingTransaction, outputBean, signer)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		txid, hex, inputVoutForBob, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTx(
			channelInfo.ChannelAddress,
			[]string{
				funderChannelAddressPrivateKey,
				reqData.FundeeChannelAddressPrivateKey,
			},
			commitmentTxInfo.MultiAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountM,
			0,
			0)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(inputVoutForBob)
		commitmentTxInfo.TxidToTempMultiAddress = txid
		commitmentTxInfo.TransactionSignHexToTempMultiAddress = hex

		//create to Bob tx
		toAddress := channelInfo.AddressB
		changeToAddress := channelInfo.AddressA
		if signer.PeerId == channelInfo.PeerIdA {
			changeToAddress = channelInfo.AddressB
			toAddress = channelInfo.AddressA
		}
		txid, hex, err = rpcClient.OmniCreateAndSignRawTransactionForCommitmentTxToBob(
			channelInfo.ChannelAddress,
			inputVoutForBob,
			[]string{
				funderChannelAddressPrivateKey,
				reqData.FundeeChannelAddressPrivateKey,
			},
			toAddress,
			changeToAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountB,
			0,
			0)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		commitmentTxInfo.TxidToOther = txid
		commitmentTxInfo.TransactionSignHexToOther = hex

		commitmentTxInfo.SignAt = time.Now()
		commitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
		commitmentTxInfo.LastHash = ""
		commitmentTxInfo.CurrHash = ""
		err = tx.Save(commitmentTxInfo)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		bytes, err := json.Marshal(commitmentTxInfo)
		msgHash := tool.SignMsg(bytes)
		commitmentTxInfo.CurrHash = msgHash
		err = tx.Update(commitmentTxInfo)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		// create RD tx
		outputAddress := channelInfo.AddressA
		if owner == channelInfo.PeerIdB {
			outputAddress = channelInfo.AddressB
		}
		rdTransaction, _ := createRDTx(owner, channelInfo, commitmentTxInfo, outputAddress, signer)

		inputs, err := getRdInputsFromCommitmentTx(commitmentTxInfo.TransactionSignHexToTempMultiAddress, commitmentTxInfo.MultiAddress, commitmentTxInfo.ScriptPubKey)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		txid, hex, err = rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
			commitmentTxInfo.MultiAddress,
			[]string{
				funderTempAddressPrivateKey,
				reqData.FundeeChannelAddressPrivateKey,
			},
			inputs,
			rdTransaction.OutputAddress,
			changeToAddress,
			fundingTransaction.PropertyId,
			rdTransaction.Amount,
			0,
			rdTransaction.Sequence,
			&commitmentTxInfo.RedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		rdTransaction.Txid = txid
		rdTransaction.TransactionSignHex = hex
		rdTransaction.SignAt = time.Now()
		rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(rdTransaction)
		if err != nil {
			return nil, err
		}
	}

	if reqData.Approval {
		// if agree,send the fundingtx to chain network
		_, err := rpcClient.SendRawTransaction(fundingTransaction.FundingTxHex)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	if reqData.Approval == false {
		err = tx.Update(channelInfo)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	err = tx.Update(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, err
	}

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
