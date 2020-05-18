package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/chainhash"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/tool"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
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

	if data.RecipientUserPeerId != targetUser {
		return nil, "", errors.New("error RecipientUserPeerId")
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

	out, _ := decimal.NewFromFloat(rpc.GetMinerFee()).Add(decimal.NewFromFloat(rpc.GetOmniDustBtc())).Mul(decimal.NewFromFloat(4.0)).Float64()
	if amount < out {
		err = errors.New("error btc amount, must greater " + tool.FloatToString(out, 8))
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
func (service *fundingTransactionManager) BeforeBobSignBtcFundingAtBobSide(data string, user *bean.User) (outData interface{}, err error) {
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
			_ = tx.Update(fundingBtcRequest)
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
		_ = tx.Commit()
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
func (service *fundingTransactionManager) AfterBobSignBtcFundingAtAliceSide(data string, user *bean.User) (outData interface{}, err error) {
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
func (service *fundingTransactionManager) AssetFundingCreated(jsonData string, user *bean.User) (outputData interface{}, err error) {
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
		err = errors.New("wrong TxHex ")
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

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
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
		err = errors.New("not found the channelInfo " + reqData.TemporaryChannelId)
		log.Println(err)
		return nil, err
	}

	err = checkBtcFundFinish(channelInfo.ChannelAddress)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//检测充值用户的私钥和公钥的一致性
	myPubKey := channelInfo.PubKeyA
	if channelInfo.PeerIdB == user.PeerId {
		myPubKey = channelInfo.PubKeyB
	}
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, myPubKey)
	if err != nil {
		return nil, err
	}

	// if alice launch funding
	fundingTxHexDecode, err := rpcClient.OmniDecodeTransaction(reqData.FundingTxHex)
	if err != nil {
		err = errors.New("TxHex  parse fail " + err.Error())
		log.Println(err)
		return nil, err
	}

	funder := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdB {
		funder = channelInfo.PeerIdB
	}
	fundingTxid, amountA, propertyId, err := checkOmniTxHex(fundingTxHexDecode, channelInfo, funder)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// 如果通道的通道id已经存在，就是这个临时地址的通道id已经在上一次的充值请求中生成了，
	// 因为这次的请求，就把之前对应的充值请求删除掉
	if tool.CheckIsString(&channelInfo.ChannelId) {
		var list []dao.FundingTransaction
		err = tx.Select(
			q.Eq("ChannelId", channelInfo.ChannelId),
		).Find(&list)
		for _, item := range list {
			_ = tx.DeleteStruct(&item)
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

	txHexDecode, err := rpcClient.DecodeRawTransaction(reqData.FundingTxHex)
	if err != nil {
		err = errors.New("TxHex  parse fail " + err.Error())
		log.Println(err)
		return nil, err
	}

	//get btc miner Fee data from transaction
	fundingTxid, _, fundingOutputIndex, err := checkBtcTxHex(txHexDecode, channelInfo, user.PeerId)
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

	needCreateC1a := false
	fundingTransaction := &dao.FundingTransaction{}
	fundingTransaction.ChannelId = bean.ChannelIdService.NewChanIDFromOutPoint(op)
	if tool.CheckIsString(&channelInfo.ChannelId) { //不是第一次发送请求
		if fundingTransaction.ChannelId != channelInfo.ChannelId { //不同的充值交易id
			//如果不是相同的充值交易，则会生成不同的通道id，这个通道id就需要去检测唯一性
			count, err := tx.Select(q.Eq("ChannelId", fundingTransaction.ChannelId)).Count(channelInfo)
			if err != nil || count != 0 {
				err = errors.New("fundingTx have been used")
				log.Println(err)
				return nil, err
			}
			needCreateC1a = true
		}
	} else { //第一次发送请求
		needCreateC1a = true
	}

	fundingTransaction.ChannelInfoId = channelInfo.Id
	fundingTransaction.PropertyId = propertyId
	fundingTransaction.PeerIdA = channelInfo.PeerIdA
	fundingTransaction.PeerIdB = channelInfo.PeerIdB

	// if alice launch funding
	if user.PeerId == channelInfo.PeerIdA {
		fundingTransaction.AmountA = amountA
		fundingTransaction.FunderAddress = channelInfo.AddressA
	} else { // if bob launch funding
		fundingTransaction.AmountB = amountA
		fundingTransaction.FunderAddress = channelInfo.AddressB
	}
	fundingTransaction.FundingTxHex = reqData.FundingTxHex
	fundingTransaction.FundingTxid = fundingTxid
	fundingTransaction.FundingOutputIndex = fundingOutputIndex
	fundingTransaction.FunderPubKey2ForCommitment = reqData.TempAddressPubKey
	tempAddrPrivateKeyMap[fundingTransaction.FunderPubKey2ForCommitment] = reqData.TempAddressPrivateKey

	fundingTransaction.CurrState = dao.FundingTransactionState_Create
	fundingTransaction.CreateBy = user.PeerId
	fundingTransaction.CreateAt = time.Now()
	err = tx.Save(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	channelInfo.ChannelId = fundingTransaction.ChannelId
	channelInfo.PropertyId = propertyId
	channelInfo.FundingAddress = fundingTransaction.FunderAddress

	var commitmentTxInfo *dao.CommitmentTransaction
	if needCreateC1a {
		// TODO 创建C1a 充值交易的output是通道地址的input，这个input需要在双方签名的基础上才能开销，
		//  C1a的资产分配，就是把充值的额度全部给alice，
		//  不过alice不能直接获取到这个钱，需要用一个临时多签去锁住，且赎回条件是bob同意并且需要等待1000个区块高度
		//region  create C1 tx
		funder := user.PeerId
		var outputBean = commitmentOutputBean{}
		outputBean.RsmcTempPubKey = fundingTransaction.FunderPubKey2ForCommitment
		// if alice funding
		if funder == channelInfo.PeerIdA {
			outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
			outputBean.OppositeSideChannelAddress = channelInfo.AddressB
			outputBean.AmountToRsmc = fundingTransaction.AmountA
			outputBean.AmountToCounterparty = fundingTransaction.AmountB
		} else { // if bob funding
			outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
			outputBean.OppositeSideChannelAddress = channelInfo.AddressA
			outputBean.AmountToRsmc = fundingTransaction.AmountB
			outputBean.AmountToCounterparty = fundingTransaction.AmountA
		}

		commitmentTxInfo, err = createCommitmentTx(funder, channelInfo, fundingTransaction, outputBean, user)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if commitmentTxInfo.AmountToRSMC > 0 {
			txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUseSingleInput(
				int(commitmentTxInfo.TxType),
				channelInfo.ChannelAddress,
				[]string{
					reqData.ChannelAddressPrivateKey,
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
			commitmentTxInfo.RsmcInputTxid = usedTxid
			commitmentTxInfo.RSMCTxid = txid
			commitmentTxInfo.RSMCTxHex = hex
		}

		commitmentTxInfo.CurrState = dao.TxInfoState_Create
		commitmentTxInfo.LastHash = ""
		commitmentTxInfo.CurrHash = ""
		err = tx.Save(commitmentTxInfo)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		//endregion
	} else {
		commitmentTxInfo = &dao.CommitmentTransaction{}
		err = tx.Select(
			q.Eq("ChannelId", channelInfo.ChannelId),
			q.Eq("Owner", user.PeerId)).
			OrderBy("CreateAt").Reverse().
			First(commitmentTxInfo)
		if err != nil {
			return nil, errors.New("not found C1a")
		}
		if commitmentTxInfo.LastHash != "" {
			return nil, errors.New("error C1a")
		}
	}

	err = tx.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	node := make(map[string]interface{})
	node["temporary_channel_id"] = reqData.TemporaryChannelId
	node["channel_id"] = fundingTransaction.ChannelId
	node["funding_omni_hex"] = fundingTransaction.FundingTxHex
	node["c1a_rsmc_hex"] = commitmentTxInfo.RSMCTxHex
	node["rsmc_temp_address_pub_key"] = commitmentTxInfo.RSMCTempAddressPubKey
	return node, err
}

func (service *fundingTransactionManager) BeforeBobSignOmniFundingAtBobSide(data string, user *bean.User) (outData interface{}, err error) {
	jsonObj := gjson.Parse(data)
	temporaryChannelId := jsonObj.Get("temporary_channel_id").String()
	channelId := jsonObj.Get("channel_id").String()
	fundingTxHex := jsonObj.Get("funding_omni_hex").String()
	c1aRemcHex := jsonObj.Get("c1a_rsmc_hex").String()
	rsmcTempAddressPubKey := jsonObj.Get("rsmc_temp_address_pub_key").String()
	if tool.CheckIsString(&channelId) == false {
		err = errors.New("wrong temporaryChannelId ")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&fundingTxHex) == false {
		err = errors.New("wrong fundingTxHex ")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&c1aRemcHex) == false {
		err = errors.New("wrong c1aRemcHex ")
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
		err = errors.New("not found the channel")
		log.Println(err)
		return nil, err
	}

	txHexDecode, err := rpcClient.DecodeRawTransaction(fundingTxHex)
	if err != nil {
		err = errors.New("TxHex  parse fail " + err.Error())
		log.Println(err)
		return nil, err
	}

	funder := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		funder = channelInfo.PeerIdB
	}
	//get btc miner Fee data from transaction
	fundingTxid, _, fundingOutputIndex, err := checkBtcTxHex(txHexDecode, channelInfo, funder)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fundingTransaction := &dao.FundingTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("FundingTxid", fundingTxid)).
		OrderBy("CreateAt").Reverse().
		First(fundingTransaction)
	if fundingTransaction.Id == 0 {
		//为了生成通道id
		hash, _ := chainhash.NewHashFromStr(fundingTxid)
		op := &bean.OutPoint{
			Hash:  *hash,
			Index: fundingOutputIndex,
		}
		fundingTransaction.ChannelId = bean.ChannelIdService.NewChanIDFromOutPoint(op)
		if fundingTransaction.ChannelId != channelId {
			return nil, errors.New("error funding tx hex")
		}

		fundingTxHexDecode, err := rpcClient.OmniDecodeTransaction(fundingTxHex)
		if err != nil {
			err = errors.New("TxHex  parse fail " + err.Error())
			log.Println(err)
			return nil, err
		}

		// if alice launch funding
		funder := channelInfo.PeerIdA
		if user.PeerId == channelInfo.PeerIdA {
			funder = channelInfo.PeerIdB
		}

		fundingTxid, amountA, propertyId, err := checkOmniTxHex(fundingTxHexDecode, channelInfo, funder)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		//如果不是相同的充值交易，则会生成不同的通道id，这个通道id就需要去检测唯一性
		if tool.CheckIsString(&channelInfo.ChannelId) {
			if fundingTransaction.ChannelId != channelInfo.ChannelId {
				channelInfo.ChannelId = fundingTransaction.ChannelId
				count, err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId)).Count(channelInfo)
				if err != nil || count != 0 {
					err = errors.New("fundingTx have been used")
					log.Println(err)
					return nil, err
				}
			}
		}
		fundingTransaction.ChannelInfoId = channelInfo.Id
		fundingTransaction.PropertyId = propertyId
		fundingTransaction.PeerIdA = channelInfo.PeerIdA
		fundingTransaction.PeerIdB = channelInfo.PeerIdB

		// if alice launch funding
		if user.PeerId == channelInfo.PeerIdB {
			fundingTransaction.AmountA = amountA
			fundingTransaction.FunderAddress = channelInfo.AddressA
		} else { // if bob launch funding
			fundingTransaction.AmountB = amountA
			fundingTransaction.FunderAddress = channelInfo.AddressB
		}
		fundingTransaction.FundingTxHex = fundingTxHex
		fundingTransaction.FundingTxid = fundingTxid
		fundingTransaction.FundingOutputIndex = fundingOutputIndex
		fundingTransaction.RsmcTempAddressPubKey = rsmcTempAddressPubKey
		fundingTransaction.FunderRsmcHex = c1aRemcHex

		err = tx.Update(channelInfo)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		fundingTransaction.CurrState = dao.FundingTransactionState_Create
		fundingTransaction.CreateBy = user.PeerId
		fundingTransaction.Owner = funder
		fundingTransaction.CreateAt = time.Now()
		err = tx.Save(fundingTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		if fundingTransaction.CurrState != dao.FundingTransactionState_Create {
			fundingTransaction.CurrState = dao.FundingTransactionState_Create
			_ = tx.Update(fundingTransaction)
		}

		if tool.CheckIsString(&fundingTransaction.RsmcTempAddressPubKey) == false {
			fundingTransaction.RsmcTempAddressPubKey = rsmcTempAddressPubKey
			_ = tx.Update(fundingTransaction)
		}
	}

	if tool.CheckIsString(&channelInfo.ChannelId) == false {
		channelInfo.ChannelId = fundingTransaction.ChannelId
		err := tx.Update(channelInfo)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	node := make(map[string]interface{})
	node["channel_id"] = fundingTransaction.ChannelId
	node["funding_omni_hex"] = fundingTransaction.FundingTxHex
	node["c1a_rsmc_hex"] = c1aRemcHex
	return node, err
}

func (service *fundingTransactionManager) AssetFundingSigned(jsonData string, signer *bean.User) (outputData interface{}, err error) {
	reqData := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&reqData.ChannelId) == false {
		return nil, errors.New("wrong ChannelId")
	}

	tx, err := signer.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Or(
			q.Eq("PeerIdA", signer.PeerId),
			q.Eq("PeerIdB", signer.PeerId)),
		q.Eq("CurrState", dao.ChannelState_WaitFundAsset)).
		First(channelInfo)
	if err != nil {
		channelInfo = nil
	}
	if channelInfo == nil {
		err = errors.New("not found channel " + reqData.ChannelId)
		log.Println(err)
		return nil, err
	}

	fundingTransaction := &dao.FundingTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", dao.FundingTransactionState_Create)).
		OrderBy("CreateAt").Reverse().
		First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var funder = channelInfo.PeerIdA
	myPubKey := channelInfo.PubKeyB
	myAddress := channelInfo.AddressB
	changeToAddress := channelInfo.AddressA
	funderAmount := fundingTransaction.AmountA
	if signer.PeerId == channelInfo.PeerIdA {
		funder = channelInfo.PeerIdB
		myPubKey = channelInfo.PubKeyA
		myAddress = channelInfo.AddressA
		changeToAddress = channelInfo.AddressB
		funderAmount = fundingTransaction.AmountB
	}

	node := make(map[string]interface{})
	node["channel_id"] = channelInfo.ChannelId
	node["approval"] = reqData.Approval
	fundingTransaction.FundeeSignAt = time.Now()
	//如果不同意这次充值
	if reqData.Approval == false {
		fundingTransaction.CurrState = dao.FundingTransactionState_Defuse
		err = db.Update(fundingTransaction)
		if err != nil {
			return nil, err
		}
		return node, nil
	}

	if tool.CheckIsString(&reqData.FundeeChannelAddressPrivateKey) == false {
		return nil, errors.New("wrong FundeeChannelAddressPrivateKey")
	}
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.FundeeChannelAddressPrivateKey, myPubKey)
	if err != nil {
		return nil, err
	}

	//region  sign C1 tx
	// 二次签名
	rsmcTxId, signedRsmcHex, err := rpcClient.BtcSignRawTransaction(fundingTransaction.FunderRsmcHex, reqData.FundeeChannelAddressPrivateKey)
	if err != nil {
		return nil, err
	}
	result, err := rpcClient.TestMemPoolAccept(signedRsmcHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		return nil, errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
	}
	//endregion

	//region create RD tx for alice
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{fundingTransaction.RsmcTempAddressPubKey, myPubKey})
	if err != nil {
		return nil, err
	}
	rsmcMultiAddress := gjson.Get(multiAddr, "address").String()
	rsmcRedeemScript := gjson.Get(multiAddr, "redeemScript").String()
	tempJson, err := rpcClient.GetAddressInfo(rsmcMultiAddress)
	if err != nil {
		return nil, err
	}
	rsmcMultiAddressScriptPubKey := gjson.Get(tempJson, "scriptPubKey").String()

	inputs, err := getInputsForNextTxByParseTxHashVout(signedRsmcHex, rsmcMultiAddress, rsmcMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	outputAddress := channelInfo.AddressA
	if funder == channelInfo.PeerIdB {
		outputAddress = channelInfo.AddressB
	}
	_, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		rsmcMultiAddress,
		[]string{
			reqData.FundeeChannelAddressPrivateKey,
		},
		inputs,
		outputAddress,
		changeToAddress,
		fundingTransaction.PropertyId,
		fundingTransaction.AmountA,
		0,
		1000,
		&rsmcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//endregion create RD tx for alice

	channelInfo.PropertyId = fundingTransaction.PropertyId
	channelInfo.FundingAddress = fundingTransaction.FunderAddress
	// region create BR1b tx  for bob
	lastCommitmentTx := &dao.CommitmentTransaction{}
	lastCommitmentTx.Id = 0
	lastCommitmentTx.PropertyId = channelInfo.PropertyId
	lastCommitmentTx.RSMCTempAddressPubKey = fundingTransaction.RsmcTempAddressPubKey
	lastCommitmentTx.RSMCMultiAddress = rsmcMultiAddress
	lastCommitmentTx.RSMCRedeemScript = rsmcRedeemScript
	lastCommitmentTx.RSMCMultiAddressScriptPubKey = rsmcMultiAddressScriptPubKey
	lastCommitmentTx.RSMCTxHex = signedRsmcHex
	lastCommitmentTx.RSMCTxid = rsmcTxId
	lastCommitmentTx.AmountToRSMC = funderAmount
	err = createCurrCommitmentTxBR(tx, dao.BRType_Rmsc, channelInfo, lastCommitmentTx, inputs, myAddress, reqData.FundeeChannelAddressPrivateKey, *signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//endregion

	channelInfo.CurrState = dao.ChannelState_CanUse

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

	node["rsmc_signed_hex"] = signedRsmcHex
	node["rd_hex"] = hex
	return node, err
}

func (service *fundingTransactionManager) AfterBobSignOmniFundingAtAilceSide(data string, user *bean.User) (outData interface{}, err error) {
	jsonObj := gjson.Parse(data)
	channelId := jsonObj.Get("channel_id").String()
	rsmcSignedHex := jsonObj.Get("rsmc_signed_hex").String()
	rdHex := jsonObj.Get("rd_hex").String()
	approval := jsonObj.Get("approval").Bool()

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("CurrState", dao.ChannelState_WaitFundAsset),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		err = errors.New("not found the channelInfo ")
		log.Println(err)
		return nil, err
	}
	fundingTransaction := &dao.FundingTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelId), q.Eq("CurrState", dao.FundingTransactionState_Create)).First(fundingTransaction)
	if err != nil {
		err = errors.New("not found the fundingTransaction")
		log.Println(err)
		return nil, err
	}

	tempPrivKey := tempAddrPrivateKeyMap[fundingTransaction.FunderPubKey2ForCommitment]
	if tool.CheckIsString(&tempPrivKey) == false {
		err = errors.New("not found the FunderPrivKey2ForCommitment")
		log.Println(err)
		return nil, err
	}

	node := make(map[string]interface{})
	fundingTransaction.FundeeSignAt = time.Now()
	node["temporary_channel_id"] = channelInfo.TemporaryChannelId
	node["approval"] = approval
	if approval == false {
		fundingTransaction.CurrState = dao.FundingTransactionState_Defuse
		_ = tx.Update(fundingTransaction)
		_ = tx.Commit()
		return node, nil
	}

	commitmentTxInfo := &dao.CommitmentTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(commitmentTxInfo)
	if err != nil {
		err = errors.New("not found the commitmentTxInfo")
		return nil, err
	}

	fundingTxHexDecode, err := rpcClient.OmniDecodeTransaction(rsmcSignedHex)
	if err != nil {
		err = errors.New("TxHex  parse fail " + err.Error())
		log.Println(err)
		return nil, err
	}

	txid := gjson.Get(fundingTxHexDecode, "txid").String()
	amountA := gjson.Get(fundingTxHexDecode, "amount").Float()
	propertyId := gjson.Get(fundingTxHexDecode, "propertyid").Int()
	sendingAddress := gjson.Get(fundingTxHexDecode, "sendingaddress").String()
	referenceAddress := gjson.Get(fundingTxHexDecode, "referenceaddress").String()
	if sendingAddress != channelInfo.ChannelAddress && referenceAddress != commitmentTxInfo.RSMCMultiAddress {
		return nil, errors.New("error address from signed hex")
	}
	if commitmentTxInfo.AmountToRSMC != amountA {
		return nil, errors.New("error amount from signed hex")
	}
	if commitmentTxInfo.PropertyId != propertyId {
		return nil, errors.New("error propertyId from signed hex")
	}
	commitmentTxInfo.RSMCTxid = txid
	commitmentTxInfo.RSMCTxHex = rsmcSignedHex

	commitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign

	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = tx.Update(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	owner := channelInfo.PeerIdA
	toAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		owner = channelInfo.PeerIdB
		toAddress = channelInfo.AddressB
	}

	// RD 二次签名
	inputs, err := getInputsForNextTxByParseTxHashVout(rsmcSignedHex, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	_, signedRdHex, err := rpcClient.OmniSignRawTransactionForUnsend(rdHex, inputs, tempAddrPrivateKeyMap[fundingTransaction.FunderPubKey2ForCommitment])
	if err != nil {
		return nil, err
	}
	result, err := rpcClient.TestMemPoolAccept(signedRdHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
			return nil, errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
		}
	}
	txid = checkHexOutputAddressFromOmniDecode(signedRdHex, inputs, toAddress)
	if txid == "" {
		return nil, errors.New("rdtx has wrong output address")
	}
	rdTransaction, err := createRDTx(owner, channelInfo, commitmentTxInfo, toAddress, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rdTransaction.RDType = 0
	rdTransaction.TxHex = signedRdHex
	rdTransaction.Txid = txid
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(rdTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	channelInfo.CurrState = dao.ChannelState_CanUse
	err = tx.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = rpcClient.SendRawTransaction(fundingTransaction.FundingTxHex)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fundingTransaction.CurrState = dao.FundingTransactionState_Accept
	_ = tx.Update(fundingTransaction)

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	node["channelId"] = channelInfo.ChannelId
	return node, nil
}

func checkHexOutputAddressFromOmniDecode(hexDecode string, inputs []rpc.TransactionInputItem, toAddress string) string {
	result, err := rpcClient.OmniDecodeTransactionWithPrevTxs(hexDecode, inputs)
	if err != nil {
		return ""
	}
	if gjson.Parse(result).Get("referenceaddress").String() == toAddress {
		return gjson.Parse(result).Get("txid").String()
	}
	return ""
}

func (service *fundingTransactionManager) BtcFundingAllItem(user bean.User) (node []dao.FundingBtcRequest, err error) {
	var itemes []dao.FundingBtcRequest
	err = user.Db.Select(
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		Find(&itemes)
	if err != nil {
		return nil, err
	}
	return itemes, nil
}

func (service *fundingTransactionManager) BtcFundingItemById(id int, user bean.User) (node *dao.FundingBtcRequest, err error) {
	var data = &dao.FundingBtcRequest{}
	err = user.Db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) BtcFundingItemByTempChannelId(tempChanId string, user bean.User) (node []dao.FundingBtcRequest, err error) {
	var itemes []dao.FundingBtcRequest
	err = user.Db.Select(q.Eq("TemporaryChannelId", tempChanId)).OrderBy("CreateAt").Reverse().Find(&itemes)
	if err != nil {
		return nil, err
	}
	return itemes, nil
}

func (service *fundingTransactionManager) BtcFundingRDAllItem(user bean.User) (node []dao.MinerFeeRedeemTransaction, err error) {
	var itemes []dao.MinerFeeRedeemTransaction
	err = user.Db.Select(
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		Find(&itemes)
	if err != nil {
		return nil, err
	}
	return itemes, nil
}

func (service *fundingTransactionManager) BtcFundingRDItemById(id int, user bean.User) (node *dao.MinerFeeRedeemTransaction, err error) {
	var data = &dao.MinerFeeRedeemTransaction{}
	err = user.Db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) BtcFundingItemRDByTempChannelId(tempChanId string, user bean.User) (node []dao.MinerFeeRedeemTransaction, err error) {
	var itemes []dao.MinerFeeRedeemTransaction
	err = user.Db.Select(q.Eq("TemporaryChannelId", tempChanId)).OrderBy("CreateAt").Reverse().Find(&itemes)
	if err != nil {
		return nil, err
	}
	return itemes, nil
}

func (service *fundingTransactionManager) BtcFundingItemRDByTempChannelIdAndFundingTxid(jsonData string, user bean.User) (node *dao.MinerFeeRedeemTransaction, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("wrong request data")
	}
	var tempChanId = gjson.Parse(jsonData).Get("temporaryChannelId").String()
	if tool.CheckIsString(&tempChanId) == false {
		return nil, errors.New("wrong temporaryChannelId")
	}
	var fundingTxid = gjson.Parse(jsonData).Get("fundingTxid").String()
	if tool.CheckIsString(&fundingTxid) == false {
		return nil, errors.New("wrong fundingTxid")
	}

	var item = &dao.MinerFeeRedeemTransaction{}
	err = user.Db.Select(q.Eq("TemporaryChannelId", tempChanId), q.Eq("FundingTxId", fundingTxid)).OrderBy("CreateAt").Reverse().First(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (service *fundingTransactionManager) OmniFundingAllItem(user bean.User) (node []dao.FundingTransaction, err error) {
	var data []dao.FundingTransaction
	err = user.Db.Select(
		q.Or(q.Eq("PeerIdB", user.PeerId),
			q.Eq("PeerIdA", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		Find(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) OmniFundingItemById(id int, user bean.User) (node *dao.FundingTransaction, err error) {
	var data = &dao.FundingTransaction{}
	err = user.Db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) OmniFundingItemByChannelId(chanId string, user bean.User) (node *dao.FundingTransaction, err error) {
	var item = &dao.FundingTransaction{}
	err = user.Db.Select(q.Eq("ChannelId", chanId)).First(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (service *fundingTransactionManager) OmniFundingTotalCount(user bean.User) (count int, err error) {
	return user.Db.Select(
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		Count(&dao.FundingTransaction{})
}
