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
	"log"
	"sync"
	"time"
)

type fundingTransactionManager struct {
	operateFlag sync.Mutex
}

var FundingTransactionService fundingTransactionManager

//funder request to fund to the multiAddr (channel)
func (service *fundingTransactionManager) CreateFundingTxRequest(jsonData string, user *bean.User) (fundingTransaction *dao.FundingTransaction, err error) {
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

	fundingTransaction = &dao.FundingTransaction{}
	count, _ := db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId)).Count(fundingTransaction)
	if count == 0 {
		if tool.CheckIsString(&reqData.FundingTxHex) == false {
			err = errors.New("wrong FundingTxHex ")
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

		fundingTxHexDecode, err := rpcClient.DecodeRawTransaction(reqData.FundingTxHex)
		if err != nil {
			err = errors.New("FundingTxHex  parse fail " + err.Error())
			log.Println(err)
			return nil, err
		}

		// sync locker
		service.operateFlag.Lock()
		defer service.operateFlag.Unlock()

		channelInfo := &dao.ChannelInfo{}
		err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("CurrState", dao.ChannelState_Accept), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(channelInfo)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if bean.ChannelIdService.IsEmpty(channelInfo.ChannelId) == false {
			err = errors.New("channel is used, can not funding again")
			log.Println(err)
			return nil, err
		}

		fundingTxid, amountA, fundingOutputIndex, err := checkFundingTxHex(fundingTxHexDecode, channelInfo, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		hash, _ := chainhash.NewHashFromStr(fundingTxid)
		op := &bean.OutPoint{
			Hash:  *hash,
			Index: fundingOutputIndex,
		}
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
			tempAddrPrivateKeyMap[channelInfo.PubKeyA] = reqData.ChannelAddressPrivateKey
		} else { // if bob launch funding
			fundingTransaction.AmountB = amountA
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

func checkFundingTxHex(fundingTxHexDecode string, channelInfo *dao.ChannelInfo, user *bean.User) (fundingTxid string, amountA float64, fundingOutputIndex uint32, err error) {
	jsonFundingTxHexDecode := gjson.Parse(fundingTxHexDecode)
	fundingTxid = jsonFundingTxHexDecode.Get("txid").String()

	//vin
	if jsonFundingTxHexDecode.Get("vin").IsArray() == false {
		err = errors.New("wrong Tx input vin")
		log.Println(err)
		return "", 0, 0, err
	}
	inTxid := jsonFundingTxHexDecode.Get("vin").Array()[0].Get("txid").String()
	inputTx, err := rpcClient.GetTransactionById(inTxid)
	if err != nil {
		err = errors.New("wrong input: " + err.Error())
		log.Println(err)
		return "", 0, 0, err
	}

	jsonInputTxDecode := gjson.Parse(inputTx)
	flag := false
	inputHexDecode, err := rpcClient.DecodeRawTransaction(jsonInputTxDecode.Get("hex").String())
	if err != nil {
		err = errors.New("wrong input: " + err.Error())
		log.Println(err)
		return "", 0, 0, err
	}

	funderAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		funderAddress = channelInfo.AddressB
	}
	jsonInputHexDecode := gjson.Parse(inputHexDecode)
	if jsonInputHexDecode.Get("vout").IsArray() {
		for _, item := range jsonInputHexDecode.Get("vout").Array() {
			addresses := item.Get("scriptPubKey").Get("addresses").Array()
			for _, subItem := range addresses {
				if subItem.String() == funderAddress {
					flag = true
					break
				}
			}
			if flag {
				break
			}
		}
	}

	if flag == false {
		err = errors.New("wrong vin " + jsonFundingTxHexDecode.Get("vin").String())
		log.Println(err)
		return "", 0, 0, err
	}

	//vout
	flag = false
	if jsonFundingTxHexDecode.Get("vout").IsArray() == false {
		err = errors.New("wrong Tx vout")
		log.Println(err)
		return "", 0, 0, err
	}
	for _, item := range jsonFundingTxHexDecode.Get("vout").Array() {
		addresses := item.Get("scriptPubKey").Get("addresses").Array()
		for _, subItem := range addresses {
			if subItem.String() == channelInfo.ChannelAddress {
				amountA = item.Get("value").Float()
				fundingOutputIndex = uint32(item.Get("n").Int())
				flag = true
				break
			}
		}
		if flag {
			break
		}
	}
	if flag == false {
		err = errors.New("wrong vout " + jsonFundingTxHexDecode.Get("vout").String())
		log.Println(err)
		return "", 0, 0, err
	}
	return fundingTxid, amountA, fundingOutputIndex, err
}

func (service *fundingTransactionManager) FundingTxSign(jsonData string, signer *bean.User) (signed *dao.FundingTransaction, err error) {
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

	if reqData.Attitude {
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

	if reqData.Attitude {

		funderChannelAddressPrivateKey := ""
		if owner == channelInfo.PeerIdA {
			fundingTransaction.AmountB = reqData.AmountB
			funderChannelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyA]
		} else {
			fundingTransaction.AmountA = reqData.AmountB
			funderChannelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyB]
		}
		if tool.CheckIsString(&funderChannelAddressPrivateKey) == false {
			err = errors.New("fail to get the funder's channel address private key ")
			log.Println(err)
			return nil, err
		}

		funderTempAddressPrivateKey := tempAddrPrivateKeyMap[fundingTransaction.FunderPubKey2ForCommitment]
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

		txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
			channelInfo.ChannelAddress,
			[]string{
				funderChannelAddressPrivateKey,
				reqData.FundeeChannelAddressPrivateKey,
			},
			[]rpc.TransactionInputItem{
				{
					commitmentTxInfo.InputTxid,
					channelInfo.ChannelAddressScriptPubKey,
					commitmentTxInfo.InputVout,
					commitmentTxInfo.InputAmount},
			},
			[]rpc.TransactionOutputItem{
				{commitmentTxInfo.MultiAddress, commitmentTxInfo.AmountM},
				{outputBean.ToAddress, commitmentTxInfo.AmountB},
			},
			0,
			0,
			&channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		commitmentTxInfo.Txid = txid
		commitmentTxInfo.TransactionSignHex = hex
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
		rdTransaction, err := createRDTx(owner, channelInfo, commitmentTxInfo, outputAddress, signer)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
			commitmentTxInfo.MultiAddress,
			[]string{
				funderTempAddressPrivateKey,
				reqData.FundeeChannelAddressPrivateKey,
			},
			[]rpc.TransactionInputItem{
				{rdTransaction.InputTxid,
					commitmentTxInfo.ScriptPubKey,
					rdTransaction.InputVout,
					rdTransaction.InputAmount},
			},
			[]rpc.TransactionOutputItem{
				{rdTransaction.OutputAddress, rdTransaction.Amount},
			},
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

	if reqData.Attitude {
		// if agree,send the fundingtx to chain network
		_, err := rpcClient.SendRawTransaction(fundingTransaction.FundingTxHex)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	if reqData.Attitude == false {
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
