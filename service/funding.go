package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"obd/bean"
	"obd/bean/chainhash"
	"obd/config"
	"obd/dao"
	"obd/rpc"
	"obd/tool"
	"strings"
	"sync"
	"time"
)

type fundingTransactionManager struct {
	operateFlag sync.Mutex
}

var FundingTransactionService fundingTransactionManager

func (service *fundingTransactionManager) BTCFundingCreated(data bean.RequestMessage, user *bean.User) (fundingTransaction map[string]interface{}, targetUser string, err error) {
	reqData := &bean.FundingBtcCreated{}
	err = json.Unmarshal([]byte(data.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		err = errors.New("wrong TemporaryChannelId")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("wrong ChannelAddressPrivateKey ")
		log.Println(err)
		return nil, "", err
	}

	btcFeeTxHexDecode, err := rpcClient.DecodeRawTransaction(reqData.FundingTxHex)
	if err != nil {
		err = errors.New("BtcFeeTxHex  parse fail " + err.Error())
		log.Println(err)
		return nil, "", err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("CurrState", dao.ChannelState_WaitFundAsset),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	targetUser = channelInfo.PeerIdB
	pubKey := channelInfo.PubKeyA
	redeemToAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		pubKey = channelInfo.PubKeyB
		redeemToAddress = channelInfo.AddressB
		targetUser = channelInfo.PeerIdA
	}

	if data.RecipientPeerId != targetUser {
		return nil, "", errors.New("error RecipientPeerId")
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, pubKey)
	if err != nil {
		return nil, "", err
	}

	//get btc miner Fee data from transaction
	fundingTxid, amount, vout, err := checkBtcTxHex(btcFeeTxHexDecode, channelInfo, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	out, _ := decimal.NewFromFloat(rpc.GetMinerFee()).Add(decimal.NewFromFloat(config.Dust)).Float64()
	if amount < out {
		err = errors.New("error btc amount")
		log.Println(err)
		return nil, "", err
	}

	count, _ := tx.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("TxId", fundingTxid),
		q.Eq("Owner", user.PeerId),
		q.Eq("IsFinish", true),
		q.Eq("SignApproval", true)).
		Count(&dao.FundingBtcRequest{})
	if count != 0 {
		err = errors.New("the tx have been send")
		log.Println(err)
		return nil, "", err
	}

	fundingBtcRequest := &dao.FundingBtcRequest{}
	_ = tx.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("TxId", fundingTxid),
		q.Eq("Owner", user.PeerId),
		q.Or(
			q.Eq("IsFinish", false),
			q.And(
				q.Eq("IsFinish", true),
				q.Eq("SignApproval", false)))).
		First(fundingBtcRequest)

	minerFeeRedeemTransaction := &dao.MinerFeeRedeemTransaction{}
	//如果这个交易是第一次请求，那就创建，并且创建一个赎回交易
	if fundingBtcRequest.Id == 0 {
		fundingBtcRequest = &dao.FundingBtcRequest{}
		fundingBtcRequest.Owner = user.PeerId
		fundingBtcRequest.TemporaryChannelId = reqData.TemporaryChannelId
		fundingBtcRequest.TxHash = reqData.FundingTxHex
		fundingBtcRequest.TxId = fundingTxid
		fundingBtcRequest.CreateAt = time.Now()
		fundingBtcRequest.Amount = amount
		fundingBtcRequest.IsFinish = false
		err = tx.Save(fundingBtcRequest)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}

		// 创建一个btc赎回交易 ，alice首先签名
		txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
			channelInfo.ChannelAddress,
			[]string{
				reqData.ChannelAddressPrivateKey},
			[]rpc.TransactionInputItem{
				{
					Txid:         fundingBtcRequest.TxId,
					Vout:         vout,
					Amount:       amount,
					ScriptPubKey: channelInfo.ChannelAddressScriptPubKey},
			},
			[]rpc.TransactionOutputItem{
				{
					ToBitCoinAddress: redeemToAddress,
					Amount:           fundingBtcRequest.Amount},
			},
			0,
			0,
			&channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			return nil, "", err
		}

		minerFeeRedeemTransaction.Txid = txid
		minerFeeRedeemTransaction.Hex = hex
		minerFeeRedeemTransaction.FundingTxId = fundingTxid
		minerFeeRedeemTransaction.CreateAt = time.Now()
		minerFeeRedeemTransaction.Owner = user.PeerId
		minerFeeRedeemTransaction.TemporaryChannelId = reqData.TemporaryChannelId
		_ = tx.Save(minerFeeRedeemTransaction)
	} else {
		if fundingBtcRequest.IsFinish {
			fundingBtcRequest.IsFinish = false
			tx.Update(fundingBtcRequest)
		}

		err := tx.Select(
			q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
			q.Eq("FundingTxId", fundingTxid),
		).First(minerFeeRedeemTransaction)
		if err != nil {
			return nil, "", err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	node := make(map[string]interface{})
	node["temporary_channel_id"] = reqData.TemporaryChannelId
	node["funding_txid"] = fundingTxid
	node["funding_btc_hex"] = reqData.FundingTxHex
	node["funding_redeem_hex"] = minerFeeRedeemTransaction.Hex
	return node, targetUser, nil
}

//bob签收btc充值之前的obd的操作
func (service *fundingTransactionManager) BeforeBobSignBtcFunding(data string, user *bean.User) (outData interface{}, err error) {
	jsonObj := gjson.Parse(data)
	temporaryChannelId := jsonObj.Get("temporary_channel_id").String()
	fundingBtcHex := jsonObj.Get("funding_btc_hex").String()
	fundingRedeemHex := jsonObj.Get("funding_redeem_hex").String()
	if tool.CheckIsString(&temporaryChannelId) == false {
		err = errors.New("wrong temporaryChannelId ")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&fundingBtcHex) == false {
		err = errors.New("wrong fundingBtcHex ")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&fundingRedeemHex) == false {
		err = errors.New("wrong fundingRedeemHex ")
		log.Println(err)
		return nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", temporaryChannelId),
		q.Eq("CurrState", dao.ChannelState_WaitFundAsset),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	funder := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		funder = channelInfo.PeerIdB
	}

	btcFeeTxHexDecode, err := rpcClient.DecodeRawTransaction(fundingBtcHex)
	if err != nil {
		err = errors.New("BtcFeeTxHex  parse fail " + err.Error())
		log.Println(err)
		return nil, err
	}
	//get btc miner Fee data from transaction
	fundingTxid, amount, _, err := checkBtcTxHex(btcFeeTxHexDecode, channelInfo, funder)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fundingBtcRequest := &dao.FundingBtcRequest{}
	_ = tx.Select(
		q.Eq("TemporaryChannelId", temporaryChannelId),
		q.Eq("TxId", fundingTxid),
		q.Eq("Owner", funder),
		q.Or(
			q.Eq("IsFinish", false),
			q.And(
				q.Eq("IsFinish", true),
				q.Eq("SignApproval", false)))).
		First(fundingBtcRequest)

	//如果这个交易是第一次请求，那就创建，并且创建一个赎回交易
	if fundingBtcRequest.Id == 0 {
		fundingBtcRequest = &dao.FundingBtcRequest{}
		fundingBtcRequest.Owner = funder
		fundingBtcRequest.TemporaryChannelId = temporaryChannelId
		fundingBtcRequest.TxHash = fundingBtcHex
		fundingBtcRequest.RedeemHex = fundingRedeemHex
		fundingBtcRequest.TxId = fundingTxid
		fundingBtcRequest.CreateAt = time.Now()
		fundingBtcRequest.Amount = amount
		fundingBtcRequest.IsFinish = false
		err = tx.Save(fundingBtcRequest)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		if fundingBtcRequest.IsFinish {
			fundingBtcRequest.IsFinish = false
			tx.Update(fundingBtcRequest)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return nil, nil
}

func (service *fundingTransactionManager) FundingBtcTxSigned(jsonData string, user *bean.User) (outData interface{}, funder string, err error) {
	reqData := &bean.FundingBtcSigned{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		err = errors.New("wrong TemporaryChannelId ")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.FundingTxid) == false {
		err = errors.New("wrong FundingTxid ")
		log.Println(err)
		return nil, "", err
	}

	if reqData.Approval {
		if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
			err = errors.New("wrong ChannelAddressPrivateKey ")
			log.Println(err)
			return nil, "", err
		}
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("CurrState", dao.ChannelState_WaitFundAsset),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").
		Reverse().
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	funder = channelInfo.PeerIdA
	myPubKey := channelInfo.PubKeyB
	if user.PeerId == channelInfo.PeerIdA {
		funder = channelInfo.PeerIdB
		myPubKey = channelInfo.PubKeyA
	}

	fundingBtcRequest := &dao.FundingBtcRequest{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("TxId", reqData.FundingTxid),
		q.Eq("Owner", funder),
		q.Eq("IsFinish", false)).
		First(fundingBtcRequest)
	if err != nil {
		err = errors.New("not found the btc fund tx")
		log.Println(err)
		return nil, "", err
	}

	node := make(map[string]interface{})
	node["temporary_channel_id"] = fundingBtcRequest.TemporaryChannelId
	node["funding_txid"] = fundingBtcRequest.TxId
	node["approval"] = reqData.Approval

	fundingBtcRequest.SignApproval = reqData.Approval
	fundingBtcRequest.SignAt = time.Now()
	if reqData.Approval == false {
		_ = tx.Update(fundingBtcRequest)
		return node, funder, nil
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, myPubKey)
	if err != nil {
		return nil, "", err
	}

	btcFeeTxHexDecode, err := rpcClient.DecodeRawTransaction(fundingBtcRequest.TxHash)
	if err != nil {
		err = errors.New("BtcFeeTxHex  parse fail " + err.Error())
		log.Println(err)
		return nil, funder, err
	}
	fundingTxid, amount, vout, err := checkBtcTxHex(btcFeeTxHexDecode, channelInfo, funder)
	if err != nil {
		log.Println(err)
		return nil, funder, err
	}

	// 二次签名
	inputItems := make([]rpc.TransactionInputItem, 0)
	inputItems = append(inputItems, rpc.TransactionInputItem{
		Txid:         fundingTxid,
		ScriptPubKey: channelInfo.ChannelAddressScriptPubKey,
		Vout:         vout,
		Amount:       amount})
	_, signedHex, err := rpcClient.BtcSignRawTransactionForUnsend(fundingBtcRequest.RedeemHex, inputItems, reqData.ChannelAddressPrivateKey)
	if err != nil {
		return nil, funder, err
	}

	//赎回交易签名成功后，广播交易
	result, err := rpcClient.SendRawTransaction(fundingBtcRequest.TxHash)
	if err != nil {
		if strings.Contains(err.Error(), "Transaction already in block chain") == false {
			return nil, funder, err
		}
	}
	log.Println(result)

	fundingBtcRequest.FinishAt = time.Now()
	fundingBtcRequest.IsFinish = true
	_ = tx.Update(fundingBtcRequest)

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	node["funding_redeem_hex"] = signedHex
	return node, funder, nil
}

//alice 所在的节点 bob签名完成，返回到alice的obd节点，需要先对alice的数据进行更新
func (service *fundingTransactionManager) AfterBobSignBtcFunding(data string, user *bean.User) (outData interface{}, err error) {
	jsonObj := gjson.Parse(data)
	temporaryChannelId := jsonObj.Get("temporary_channel_id").String()
	approval := jsonObj.Get("approval").Bool()
	fundingTxid := jsonObj.Get("funding_txid").String()
	fundingRedeemHex := ""

	if &approval == nil {
		err = errors.New("wrong approval ")
		log.Println(err)
		return nil, err
	}
	if approval {
		fundingRedeemHex = jsonObj.Get("funding_redeem_hex").String()
		if tool.CheckIsString(&fundingRedeemHex) == false {
			err = errors.New("wrong fundingRedeemHex ")
			log.Println(err)
			return nil, err
		}
	}

	if tool.CheckIsString(&temporaryChannelId) == false {
		err = errors.New("wrong temporaryChannelId ")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&fundingTxid) == false {
		err = errors.New("wrong fundingTxid ")
		log.Println(err)
		return nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", temporaryChannelId),
		q.Eq("CurrState", dao.ChannelState_WaitFundAsset),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	funder := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdB {
		funder = channelInfo.PeerIdB
	}

	fundingBtcRequest := &dao.FundingBtcRequest{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", temporaryChannelId),
		q.Eq("TxId", fundingTxid),
		q.Eq("Owner", funder),
		q.Eq("IsFinish", false)).
		First(fundingBtcRequest)
	if err != nil {
		log.Println(err)
	}

	node := make(map[string]interface{})
	node["temporary_channel_id"] = fundingBtcRequest.TemporaryChannelId
	node["funding_txid"] = fundingBtcRequest.TxId
	node["approval"] = approval

	fundingBtcRequest.SignApproval = approval
	fundingBtcRequest.SignAt = time.Now()
	fundingBtcRequest.FinishAt = time.Now()
	fundingBtcRequest.IsFinish = true
	_ = tx.Update(fundingBtcRequest)

	if approval {
		minerFeeRedeemTransaction := &dao.MinerFeeRedeemTransaction{}
		err = tx.Select(
			q.Eq("TemporaryChannelId", temporaryChannelId),
			q.Eq("FundingTxId", fundingTxid),
		).First(minerFeeRedeemTransaction)
		if err != nil {
			return nil, err
		}
		minerFeeRedeemTransaction.Hex = fundingRedeemHex
		err = tx.Update(minerFeeRedeemTransaction)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return fundingBtcRequest, nil
}

//funder request to fund to the multiAddr (channel)
func (service *fundingTransactionManager) AssetFundingCreated(jsonData string, user *bean.User) (fundingTransaction *dao.FundingTransaction, err error) {
	reqData := &bean.FundingCreated{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		err = errors.New("wrong TemporaryChannelId ")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.FundingTxHex) == false {
		err = errors.New("wrong TxHash ")
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

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.TempAddressPrivateKey, reqData.TempAddressPubKey)
	if err != nil {
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("CurrState", dao.ChannelState_WaitFundAsset),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		err = errors.New("not found the channelInfo " + reqData.TemporaryChannelId)
		log.Println(err)
		return nil, err
	}

	//目标用户是否在线
	targetUser := channelInfo.PeerIdB
	myPubKey := channelInfo.PubKeyA
	if targetUser == user.PeerId {
		targetUser = channelInfo.PeerIdA
		myPubKey = channelInfo.PubKeyB
	}
	if err = FindUserIsOnline(targetUser); err != nil {
		return nil, err
	}

	err = checkBtcFundFinish(channelInfo.ChannelAddress)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//检测充值用户的私钥和公钥的一致性
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, myPubKey)
	if err != nil {
		return nil, err
	}

	// if alice launch funding
	tempAddrPrivateKeyMap[myPubKey] = reqData.ChannelAddressPrivateKey
	tempAddrPrivateKeyMap[reqData.TempAddressPubKey] = reqData.TempAddressPrivateKey

	fundingTxHexDecode, err := rpcClient.OmniDecodeTransaction(reqData.FundingTxHex)
	if err != nil {
		err = errors.New("TxHash  parse fail " + err.Error())
		log.Println(err)
		return nil, err
	}
	fundingTxid, amountA, propertyId, err := checkOmniTxHex(fundingTxHexDecode, channelInfo, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&channelInfo.ChannelId) {
		var list []dao.FundingTransaction
		err = db.Select(
			q.Eq("ChannelId", channelInfo.ChannelId),
		).Find(&list)
		for _, item := range list {
			_ = db.DeleteStruct(&item)
		}
	}

	// sync locker
	service.operateFlag.Lock()
	defer service.operateFlag.Unlock()

	reqData.PropertyId = propertyId

	// getProperty from omnicore
	// 验证PropertyId是否在omni存在
	_, err = rpcClient.OmniGetProperty(reqData.PropertyId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

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

	//为了生成通道id
	hash, _ := chainhash.NewHashFromStr(fundingTxid)
	op := &bean.OutPoint{
		Hash:  *hash,
		Index: fundingOutputIndex,
	}

	fundingTransaction = &dao.FundingTransaction{}
	fundingTransaction.ChannelId = bean.ChannelIdService.NewChanIDFromOutPoint(op)
	if fundingTransaction.ChannelId != channelInfo.ChannelId {
		channelInfo.ChannelId = fundingTransaction.ChannelId
		count, err := db.Select(q.Eq("ChannelId", channelInfo.ChannelId)).Count(channelInfo)
		if channelInfo.PropertyId != propertyId {
			if err != nil || count != 0 {
				err = errors.New("fundingTx have been used")
				log.Println(err)
				return nil, err
			}
		}
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
	//tempAddrPrivateKeyMap[fundingTransaction.FunderPubKey2ForCommitment] = reqData.TempAddressPrivateKey

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

	return fundingTransaction, err
}

func checkBtcFundFinish(address string) error {
	result, err := rpcClient.ListUnspent(address)
	if err != nil {
		return err
	}
	array := gjson.Parse(result).Array()
	log.Println("listunspent", array)
	if len(array) < 3 {
		return errors.New("btc fund have been not finished")
	}

	pMoney := rpc.GetOmniDustBtc()
	out, _ := decimal.NewFromFloat(rpc.GetMinerFee()).Add(decimal.NewFromFloat(pMoney)).Float64()
	for _, item := range array {
		amount := item.Get("amount").Float()
		if amount != pMoney {
			if amount < out {
				return errors.New("btc amount error")
			}
		}
	}
	return nil
}

func (service *fundingTransactionManager) AssetFundingSigned(jsonData string, signer *bean.User) (fundingTransaction *dao.FundingTransaction, err error) {
	reqData := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&reqData.ChannelId) == false {
		return nil, errors.New("wrong ChannelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", dao.ChannelState_WaitFundAsset),
		q.Or(
			q.Eq("PeerIdA", signer.PeerId),
			q.Eq("PeerIdB", signer.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		err = errors.New("not found the channel")
		log.Println(err)
		return nil, err
	}

	// default if alice launch the funding,signer is bob
	var funder = channelInfo.PeerIdA
	myPubKey := channelInfo.PubKeyB
	if signer.PeerId == channelInfo.PeerIdA {
		funder = channelInfo.PeerIdB
		myPubKey = channelInfo.PubKeyA
	}

	fundingTransaction = &dao.FundingTransaction{}
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", dao.FundingTransactionState_Create)).
		OrderBy("CreateAt").Reverse().
		First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fundingTransaction.FundeeSignAt = time.Now()
	//如果不同意这次充值
	if reqData.Approval == false {
		fundingTransaction.CurrState = dao.FundingTransactionState_Defuse
		err = db.Update(fundingTransaction)
		if err != nil {
			return nil, err
		}
		return fundingTransaction, nil
	}

	tx, err := db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	if tool.CheckIsString(&reqData.FundeeChannelAddressPrivateKey) == false {
		return nil, errors.New("wrong FundeeChannelAddressPrivateKey")
	}
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.FundeeChannelAddressPrivateKey, myPubKey)
	if err != nil {
		return nil, err
	}

	funderChannelAddressPrivateKey := ""
	if funder == channelInfo.PeerIdA {
		fundingTransaction.AmountB = reqData.AmountB
		funderChannelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyA]
		delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	} else {
		fundingTransaction.AmountA = reqData.AmountB
		funderChannelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyB]
		delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	}

	if tool.CheckIsString(&funderChannelAddressPrivateKey) == false {
		err = errors.New("fail to get the funder's channel address private key ")
		log.Println(err)
		return nil, err
	}

	funderTempAddressPrivateKey := tempAddrPrivateKeyMap[fundingTransaction.FunderPubKey2ForCommitment]
	delete(tempAddrPrivateKeyMap, fundingTransaction.FunderPubKey2ForCommitment)
	if tool.CheckIsString(&funderTempAddressPrivateKey) == false {
		err = errors.New("fail to get the funder's tmep address private key ")
		log.Println(err)
		return nil, err
	}

	//region  create C1 tx
	var outputBean = commitmentOutputBean{}
	outputBean.RsmcTempPubKey = fundingTransaction.FunderPubKey2ForCommitment
	// if alice funding
	if funder == channelInfo.PeerIdA {
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
		outputBean.OppositeSideChannelAddress = channelInfo.AddressB
		outputBean.AmountToRsmc = fundingTransaction.AmountA
		outputBean.AmountToOther = fundingTransaction.AmountB
	} else { // if bob funding
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
		outputBean.OppositeSideChannelAddress = channelInfo.AddressA
		outputBean.AmountToRsmc = fundingTransaction.AmountB
		outputBean.AmountToOther = fundingTransaction.AmountA
	}

	commitmentTxInfo, err := createCommitmentTx(funder, channelInfo, fundingTransaction, outputBean, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	usedTxidTemp := ""
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUserSingleInput(
			int(commitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			[]string{
				funderChannelAddressPrivateKey,
				reqData.FundeeChannelAddressPrivateKey,
			},
			commitmentTxInfo.RSMCMultiAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(usedTxid)
		usedTxidTemp = usedTxid
		commitmentTxInfo.RSMCTxid = txid
		commitmentTxInfo.RSMCTxHash = hex
	}

	changeToAddress := channelInfo.AddressA
	if signer.PeerId == channelInfo.PeerIdA {
		changeToAddress = channelInfo.AddressB
	}
	//endregion

	//region create to Bob tx
	if commitmentTxInfo.AmountToOther > 0 {
		toAddress := channelInfo.AddressB
		if signer.PeerId == channelInfo.PeerIdA {
			toAddress = channelInfo.AddressA
		}
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseRestInput(
			int(commitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			usedTxidTemp,
			[]string{
				funderChannelAddressPrivateKey,
				reqData.FundeeChannelAddressPrivateKey,
			},
			toAddress,
			changeToAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToOther,
			0,
			0, &channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		commitmentTxInfo.ToOtherTxid = txid
		commitmentTxInfo.ToOtherTxHash = hex
	}

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
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = tx.Update(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//endregion

	//region create RD tx
	outputAddress := channelInfo.AddressA
	if funder == channelInfo.PeerIdB {
		outputAddress = channelInfo.AddressB
	}
	rdTransaction, _ := createRDTx(funder, channelInfo, commitmentTxInfo, outputAddress, signer)

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.RSMCTxHash, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		commitmentTxInfo.RSMCMultiAddress,
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
		&commitmentTxInfo.RSMCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rdTransaction.Txid = txid
	rdTransaction.TxHash = hex
	rdTransaction.SignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(rdTransaction)
	if err != nil {
		return nil, err
	}
	//endregion

	_, err = rpcClient.SendRawTransaction(fundingTransaction.FundingTxHex)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	channelInfo.CurrState = dao.ChannelState_CanUse
	channelInfo.PropertyId = fundingTransaction.PropertyId
	err = tx.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fundingTransaction.CurrState = dao.FundingTransactionState_Accept
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

	return fundingTransaction, nil
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
	err = db.Select(
		q.Or(q.Eq("PeerIdB", peerId),
			q.Eq("PeerIdA", peerId))).
		OrderBy("CreateAt").Reverse().
		Find(&data)
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
	return db.Select(
		q.Or(
			q.Eq("PeerIdA", peerId),
			q.Eq("PeerIdB", peerId))).
		Count(&dao.FundingTransaction{})
}
