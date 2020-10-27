package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/chainhash"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/tool"
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

// 缓存btc矿工费充值的需要发送给bob的数据
var tempBtcFundingCreatedData map[string]bean.FundingBtcOfP2p

// btc矿工费充值交易的创建 -100340
func (service *fundingTransactionManager) BtcFundingCreated(msg bean.RequestMessage, user *bean.User) (fundingTransaction interface{}, targetUser string, err error) {
	reqData := &bean.SendRequestFundingBtc{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		err = errors.New(enum.Tips_common_wrong + " temporary_channel_id")
		log.Println(err)
		return nil, "", err
	}

	btcFeeTxHexDecode, err := rpcClient.DecodeRawTransaction(reqData.FundingTxHex)
	if err != nil {
		err = errors.New(enum.Tips_funding_failDecodeRawTransaction + " funding_tx_hex: " + err.Error())
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
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", errors.New(enum.Tips_funding_notFoundChannelByTempId + reqData.TemporaryChannelId)
	}

	if channelInfo.CurrState != dao.ChannelState_WaitFundAsset {
		return nil, "", errors.New(enum.Tips_funding_notFundBtcState)
	}

	//check btc funding time
	result, err := rpcClient.ListReceivedByAddress(channelInfo.ChannelAddress)
	if err == nil {
		if len(gjson.Parse(result).Array()) > 0 {
			btcFundingTimes := len(gjson.Parse(result).Array()[0].Get("txids").Array())
			if btcFundingTimes >= config.BtcNeedFundTimes {
				return nil, "", errors.New(enum.Tips_funding_enoughBtcFundingTime)
			}
		}
	}

	targetUser = channelInfo.PeerIdB
	redeemToAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		redeemToAddress = channelInfo.AddressB
		targetUser = channelInfo.PeerIdA
	}

	if msg.RecipientUserPeerId != targetUser {
		return nil, "", errors.New(enum.Tips_common_wrong + "recipient_user_peer_id")
	}

	//get btc miner Fee data from transaction
	fundingTxid, amount, vout, err := checkBtcTxHex(btcFeeTxHexDecode, channelInfo, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	out := GetBtcMinerFundMiniAmount()
	if amount < out {
		err = errors.New(enum.Tips_funding_btcAmountMustGreater + tool.FloatToString(out, 8))
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
		err = errors.New(enum.Tips_funding_btcTxBeenSend)
		log.Println(err)
		return nil, "", err
	}

	latestBtcFundingRequest := &dao.FundingBtcRequest{}
	_ = tx.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("Owner", user.PeerId),
		q.Or(
			q.Eq("IsFinish", false),
			q.And(
				q.Eq("IsFinish", true),
				q.Eq("SignApproval", false)))).
		OrderBy("CreateAt").
		Reverse().
		First(latestBtcFundingRequest)
	if latestBtcFundingRequest.Id > 0 && latestBtcFundingRequest.TxId != fundingTxid {
		result, err = rpcClient.TestMemPoolAccept(latestBtcFundingRequest.TxHash)
		if err == nil {
			allowed := gjson.Parse(result).Get("allowed").Bool()
			if allowed == false {
				latestBtcFundingRequest.IsFinish = true
				_ = tx.Update(latestBtcFundingRequest)
			} else {
				return nil, "", errors.New(enum.Tips_funding_fundTxIsRunning)
			}
		}
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
	var needAliceSignData map[string]interface{}
	clientSignHexData := bean.NeedClientSignHexData{}

	if fundingBtcRequest.Id == 0 {
		fundingBtcRequest = &dao.FundingBtcRequest{
			Owner:              user.PeerId,
			TemporaryChannelId: reqData.TemporaryChannelId,
			TxHash:             reqData.FundingTxHex,
			TxId:               fundingTxid,
			Amount:             amount,
			CreateAt:           time.Now(),
			IsFinish:           false,
		}
		err = tx.Save(fundingBtcRequest)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}

		// 创建一个btc赎回交易 ，alice首先签名
		needAliceSignData, err = rpcClient.BtcCreateRawTransactionForUnsendInputTx(
			channelInfo.ChannelAddress,
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

		minerFeeRedeemTransaction := &dao.MinerFeeRedeemTransaction{}
		minerFeeRedeemTransaction.TemporaryChannelId = reqData.TemporaryChannelId
		minerFeeRedeemTransaction.FundingTxId = fundingTxid
		minerFeeRedeemTransaction.Hex = needAliceSignData["hex"].(string)
		minerFeeRedeemTransaction.IsFinish = false
		minerFeeRedeemTransaction.Amount = needAliceSignData["totalAmount"].(float64)
		minerFeeRedeemTransaction.CreateAt = time.Now()
		minerFeeRedeemTransaction.Owner = user.PeerId
		_ = tx.Save(minerFeeRedeemTransaction)

	} else {
		if fundingBtcRequest.IsFinish {
			fundingBtcRequest.IsFinish = false
			_ = tx.UpdateField(fundingBtcRequest, "IsFinish", false)
		}

		err = tx.Select(
			q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
			q.Eq("FundingTxId", fundingTxid),
		).First(minerFeeRedeemTransaction)
		if err != nil {
			return nil, "", err
		}

		// 如果在前面的步骤，已经创建了minerFeeRedeemTransaction，但是没有完成签名，需要继续发起签名逻辑
		if len(minerFeeRedeemTransaction.Txid) == 0 {
			needAliceSignData, err = rpcClient.BtcCreateRawTransactionForUnsendInputTx(
				channelInfo.ChannelAddress,
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
		} else {
			var inputs []map[string]interface{}
			node := make(map[string]interface{})
			node["txid"] = fundingBtcRequest.TxId
			node["vout"] = vout
			node["amount"] = amount
			node["redeemScript"] = channelInfo.ChannelAddressRedeemScript
			node["scriptPubKey"] = channelInfo.ChannelAddressScriptPubKey
			inputs = append(inputs, node)
			clientSignHexData.Inputs = inputs
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if needAliceSignData != nil {
		clientSignHexData.Inputs = needAliceSignData["inputs"]
	}
	clientSignHexData.Hex = minerFeeRedeemTransaction.Hex
	clientSignHexData.IsMultisig = true
	clientSignHexData.PubKeyA = channelInfo.PubKeyA
	clientSignHexData.PubKeyB = channelInfo.PubKeyB

	fundingBtcOfP2p := bean.FundingBtcOfP2p{
		TemporaryChannelId: reqData.TemporaryChannelId,
		FundingTxid:        fundingTxid,
		FundingBtcHex:      reqData.FundingTxHex,
		FundingRedeemHex:   minerFeeRedeemTransaction.Hex,
		SignData:           clientSignHexData,
		FunderNodeAddress:  msg.SenderNodePeerId,
		FunderPeerId:       msg.SenderUserPeerId,
	}
	// 如果需要签名，就发送待签名数据到客户端，否则就发送数据到bob所在的obd节点
	if needAliceSignData != nil {
		targetUser = user.PeerId
		if tempBtcFundingCreatedData == nil {
			tempBtcFundingCreatedData = make(map[string]bean.FundingBtcOfP2p)
		}
		tempBtcFundingCreatedData[user.PeerId+"_"+fundingTxid] = fundingBtcOfP2p
		return clientSignHexData, targetUser, nil
	}

	return fundingBtcOfP2p, targetUser, nil
}

// 响应alice对btc矿工费充值的赎回交易的签名 -100341
func (service *fundingTransactionManager) AliceSignBtcFundingMinerFeeRedeemTx(jsonObj string, user *bean.User) (retData interface{}, targetUser string, err error) {

	if tool.CheckIsString(&jsonObj) == false {
		return nil, "", errors.New("empty hex")
	}
	hex := gjson.Get(jsonObj, "hex").Str
	resultDecode, err := rpcClient.DecodeRawTransaction(hex)
	inputTxId := gjson.Get(resultDecode, "vin").Array()[0].Get("txid").Str

	key := user.PeerId + "_" + inputTxId
	fundingBtcOfP2p := tempBtcFundingCreatedData[key]
	if len(fundingBtcOfP2p.FundingTxid) == 0 {
		return nil, "", errors.New("not found the temp data, please send -100340 again")
	}

	result, err := rpcClient.TestMemPoolAccept(hex)
	if err != nil {
		return nil, "", err
	}
	txid := gjson.Parse(result).Array()[0].Get("txid").Str
	if len(txid) == 0 {
		return nil, "", errors.New("fail to test the hex")
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", fundingBtcOfP2p.TemporaryChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", errors.New(enum.Tips_funding_notFoundChannelByTempId + fundingBtcOfP2p.TemporaryChannelId)
	}

	if channelInfo.CurrState != dao.ChannelState_WaitFundAsset {
		return nil, "", errors.New(enum.Tips_funding_notFundBtcState)
	}

	redeemToAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		redeemToAddress = channelInfo.AddressB
	}
	minerFeeRedeemTransaction := &dao.MinerFeeRedeemTransaction{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", fundingBtcOfP2p.TemporaryChannelId),
		q.Eq("FundingTxId", fundingBtcOfP2p.FundingTxid),
	).First(minerFeeRedeemTransaction)

	if err != nil {
		return nil, "", err
	}

	if checkSignMinerFeeRedeemTx(resultDecode, minerFeeRedeemTransaction.FundingTxId, redeemToAddress, minerFeeRedeemTransaction.Amount) == false {
		return nil, "", errors.New("error sign")
	}

	minerFeeRedeemTransaction.Hex = hex
	minerFeeRedeemTransaction.Txid = txid
	_ = tx.Update(minerFeeRedeemTransaction)
	_ = tx.Commit()

	delete(tempBtcFundingCreatedData, key)

	fundingBtcOfP2p.FundingRedeemHex = minerFeeRedeemTransaction.Hex

	return fundingBtcOfP2p, targetUser, nil
}

// 检测签名后的hex是否和原有数据一致，防止客户端篡改
func checkSignMinerFeeRedeemTx(result string, inTxid string, outAddress string, amount float64) bool {
	vinArr := gjson.Get(result, "vin").Array()
	if len(vinArr) != 1 {
		return false
	}

	txid := vinArr[0].Get("txid").Str
	if txid != inTxid {
		return false
	}

	voutArr := gjson.Get(result, "vout").Array()
	if len(voutArr) != 1 {
		return false
	}
	address := voutArr[0].Get("scriptPubKey").Get("addresses").Array()[0].Str
	if address != outAddress {
		return false
	}

	outAmount := voutArr[0].Get("value").Float()
	if outAmount != amount {
		return false
	}

	return true
}

// Bob签收btc矿工费充值交易之前的obd的操作
func (service *fundingTransactionManager) BeforeSignBtcFundingCreatedAtBobSide(data string, user *bean.User) (outData interface{}, err error) {
	fundingBtcOfP2p := bean.FundingBtcOfP2p{}
	_ = json.Unmarshal([]byte(data), &fundingBtcOfP2p)
	temporaryChannelId := fundingBtcOfP2p.TemporaryChannelId
	fundingBtcHex := fundingBtcOfP2p.FundingBtcHex
	fundingRedeemHex := fundingBtcOfP2p.FundingRedeemHex
	if tool.CheckIsString(&temporaryChannelId) == false {
		err = errors.New(enum.Tips_common_wrong + "temporary_channel_id")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&fundingBtcHex) == false {
		err = errors.New(enum.Tips_common_wrong + "funding_btc_hex")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&fundingRedeemHex) == false {
		err = errors.New(enum.Tips_common_wrong + "funding_redeem_hex")
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
		err = errors.New(enum.Tips_funding_failDecodeRawTransaction + " funding_btc_hex " + err.Error())
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

	//if get the request at first time
	if fundingBtcRequest.Id == 0 {
		fundingBtcRequest = &dao.FundingBtcRequest{
			Owner:              funder,
			TemporaryChannelId: temporaryChannelId,
			RedeemHex:          fundingRedeemHex,
			TxHash:             fundingBtcHex,
			TxId:               fundingTxid,
			Amount:             amount,
			CreateAt:           time.Now(),
			IsFinish:           false,
		}
		_ = tx.Save(fundingBtcRequest)
	} else {
		if fundingBtcRequest.IsFinish {
			_ = tx.UpdateField(fundingBtcRequest, "IsFinish", false)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return nil, nil
}

// Bob签收Alice的btc矿工费充值交易 -100350
func (service *fundingTransactionManager) FundingBtcTxSigned(msg bean.RequestMessage, user *bean.User) (outData interface{}, funder string, err error) {
	reqData := bean.SendSignFundingBtc{}
	err = json.Unmarshal([]byte(msg.Data), &reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		err = errors.New(enum.Tips_common_wrong + " temporary_channel_id ")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.FundingTxid) == false {
		err = errors.New(enum.Tips_common_wrong + "funding_txid ")
		log.Println(err)
		return nil, "", err
	}
	if reqData.Approval {
		if tool.CheckIsString(&reqData.SignedMinerRedeemTransactionHex) == false {
			err = errors.New(enum.Tips_common_wrong + "signed_miner_redeem_transaction_hex ")
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
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").
		Reverse().
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", errors.New(enum.Tips_funding_notFoundChannelByTempId + reqData.TemporaryChannelId)
	}

	if channelInfo.CurrState != dao.ChannelState_WaitFundAsset {
		return nil, "", errors.New(enum.Tips_funding_notFundAssetState)
	}

	funder = channelInfo.PeerIdA
	funderAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdA {
		funder = channelInfo.PeerIdB
		funderAddress = channelInfo.AddressB
	}

	if funder != msg.RecipientUserPeerId {
		return nil, "", errors.New(enum.Tips_common_wrong + "recipient_user_peer_id")
	}

	fundingBtcRequest := &dao.FundingBtcRequest{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("TxId", reqData.FundingTxid),
		q.Eq("Owner", funder),
		q.Eq("IsFinish", false),
	).
		First(fundingBtcRequest)
	if err != nil {
		err = errors.New(enum.Tips_funding_notFoundFundBtcTx)
		log.Println(err)
		return nil, "", err
	}

	channelInfo.BtcAmount = fundingBtcRequest.Amount
	_ = tx.Update(channelInfo)

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

	result, err := rpcClient.DecodeRawTransaction(fundingBtcRequest.RedeemHex)
	if err != nil {
		return nil, "", err
	}
	minerOutAmount := gjson.Get(result, "vout").Array()[0].Get("value").Float()

	resultDecode, err := rpcClient.DecodeRawTransaction(reqData.SignedMinerRedeemTransactionHex)
	if err != nil {
		return nil, "", err
	}
	if checkSignMinerFeeRedeemTx(resultDecode, reqData.FundingTxid, funderAddress, minerOutAmount) == false {
		return nil, "", errors.New("wrong signed hex")
	}

	//赎回交易签名成功后，广播交易
	_, err = rpcClient.SendRawTransaction(fundingBtcRequest.TxHash)
	if err != nil {
		if strings.Contains(err.Error(), "Transaction already in block chain") == false {
			return nil, funder, err
		}
	}

	fundingBtcRequest.RedeemHex = reqData.SignedMinerRedeemTransactionHex
	fundingBtcRequest.FinishAt = time.Now()
	fundingBtcRequest.IsFinish = true
	_ = tx.Update(fundingBtcRequest)

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	node["funding_redeem_hex"] = fundingBtcRequest.RedeemHex
	return node, funder, nil
}

// 操作人：alice所在的obd节点：bob签名完成，返回到alice的obd节点，需要先对alice的数据进行更新
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

	channelInfo.BtcAmount = fundingBtcRequest.Amount
	_ = tx.Update(channelInfo)

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
		minerFeeRedeemTransaction.IsFinish = true
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

// ********************* omni资产充值 *********************
// funder request to fund to the multiAddr (channel)
func (service *fundingTransactionManager) AssetFundingCreated(msg bean.RequestMessage, user *bean.User) (outputData interface{}, err error) {
	reqData := &bean.SendRequestAssetFunding{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "temporary_channel_id ")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.FundingTxHex) == false {
		err = errors.New(enum.Tips_common_empty + " funding_tx_hex ")
		log.Println(err)
		return nil, err
	}

	if _, err := getAddressFromPubKey(reqData.TempAddressPubKey); err != nil {
		err = errors.New(enum.Tips_common_wrong + "temp_address_pub_key ")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.TempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_wrong + "temp_address_private_key ")
		log.Println(err)
		return nil, err
	}
	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_wrong + "channel_address_private_key ")
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
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		err = errors.New(enum.Tips_funding_notFoundChannelByTempId + reqData.TemporaryChannelId)
		log.Println(err)
		return nil, err
	}

	if channelInfo.CurrState != dao.ChannelState_WaitFundAsset {
		err = errors.New(enum.Tips_funding_notFundAssetState)
		log.Println(err)
		return nil, err
	}

	err = checkBtcFundFinish(channelInfo.ChannelAddress, true)
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
		err = errors.New(enum.Tips_funding_failDecodeRawTransaction + " : " + err.Error())
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

	fundingAssetTxHexDecode, err := rpcClient.DecodeRawTransaction(reqData.FundingTxHex)
	if err != nil {
		err = errors.New(enum.Tips_funding_failDecodeRawTransaction + " funding_tx_hex " + err.Error())
		log.Println(err)
		return nil, err
	}

	//get btc miner Fee data from transaction
	fundingTxid, _, fundingOutputIndex, err := checkBtcTxHex(fundingAssetTxHexDecode, channelInfo, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//生成通道id
	hash, _ := chainhash.NewHashFromStr(fundingTxid)
	op := &bean.OutPoint{
		Hash:  *hash,
		Index: fundingOutputIndex,
	}

	needCreateC1a := false
	fundingTransaction := &dao.FundingTransaction{}
	fundingTransaction.TemporaryChannelId = reqData.TemporaryChannelId
	fundingTransaction.ChannelId = bean.ChannelIdService.NewChanIDFromOutPoint(op)
	if tool.CheckIsString(&channelInfo.ChannelId) { //不是第一次发送请求
		if fundingTransaction.ChannelId != channelInfo.ChannelId { //不同的充值交易id
			needCreateC1a = true
		}
	} else { //第一次发送请求
		needCreateC1a = true
	}

	if needCreateC1a {
		flag := httpGetChannelStateFromTracker(fundingTransaction.ChannelId)
		if flag != 0 && flag != int(dao.ChannelState_WaitFundAsset) {
			err = errors.New(enum.Tips_funding_needChangeFundTx)
			log.Println(err)
			return nil, err
		}
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
	channelInfo.Amount = amountA
	channelInfo.FundingAddress = fundingTransaction.FunderAddress

	var commitmentTxInfo *dao.CommitmentTransaction
	if needCreateC1a {
		//region  create C1 tx
		funder := user.PeerId
		var outputBean = commitmentTxOutputBean{}
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
		commitmentTxInfo.RSMCTempAddressIndex = reqData.TempAddressIndex

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
			return nil, errors.New(enum.Tips_common_notFound + ": CommitmentTransaction")
		}
		if commitmentTxInfo.LastHash != "" {
			return nil, errors.New(enum.Tips_common_wrong + "CommitmentTransaction")
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

	node := bean.FundingAssetOfP2p{}
	node.TemporaryChannelId = reqData.TemporaryChannelId
	node.FundingOmniHex = fundingTransaction.FundingTxHex
	node.C1aRsmcHex = commitmentTxInfo.RSMCTxHex
	node.RsmcTempAddressPubKey = commitmentTxInfo.RSMCTempAddressPubKey
	node.FunderNodeAddress = msg.SenderNodePeerId
	node.FunderPeerId = msg.SenderUserPeerId
	return node, err
}

func (service *fundingTransactionManager) BeforeSignAssetFundingCreateAtBobSide(data string, user *bean.User) (outData interface{}, err error) {
	jsonObj := bean.FundingAssetOfP2p{}
	_ = json.Unmarshal([]byte(data), &jsonObj)
	temporaryChannelId := jsonObj.TemporaryChannelId
	fundingTxHex := jsonObj.FundingOmniHex
	c1aRemcHex := jsonObj.C1aRsmcHex
	rsmcTempAddressPubKey := jsonObj.RsmcTempAddressPubKey
	if tool.CheckIsString(&temporaryChannelId) == false {
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

	//生成通道id
	hash, _ := chainhash.NewHashFromStr(fundingTxid)
	op := &bean.OutPoint{
		Hash:  *hash,
		Index: fundingOutputIndex,
	}
	channelId := bean.ChannelIdService.NewChanIDFromOutPoint(op)

	fundingTransaction := &dao.FundingTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("TemporaryChannelId", temporaryChannelId),
		q.Eq("FundingTxid", fundingTxid)).
		OrderBy("CreateAt").Reverse().
		First(fundingTransaction)
	if fundingTransaction.Id == 0 {
		fundingTransaction.ChannelId = channelId
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
		fundingTransaction.TemporaryChannelId = temporaryChannelId
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
		if tool.CheckIsString(&fundingTransaction.TemporaryChannelId) == false {
			fundingTransaction.TemporaryChannelId = temporaryChannelId
			_ = tx.Update(fundingTransaction)

		}
	}

	if tool.CheckIsString(&channelInfo.ChannelId) == false {
		channelInfo.ChannelId = fundingTransaction.ChannelId
		err = tx.Update(channelInfo)
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
	//node := make(map[string]interface{})
	//node["temporary_channel_id"] = temporaryChannelId
	//node["funding_omni_hex"] = fundingTransaction.FundingTxHex
	//node["c1a_rsmc_hex"] = c1aRemcHex
	return nil, nil
}

func (service *fundingTransactionManager) AssetFundingSigned(jsonData string, signer *bean.User) (outputData interface{}, err error) {
	reqData := &bean.SignAssetFunding{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		return nil, errors.New(enum.Tips_common_empty + "temporary_channel_id")
	}

	tx, err := signer.Db.Begin(true)
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
			q.Eq("PeerIdA", signer.PeerId),
			q.Eq("PeerIdB", signer.PeerId)),
	).
		First(channelInfo)
	if err != nil {
		channelInfo = nil
	}

	if channelInfo == nil {
		err = errors.New(enum.Tips_funding_notFoundChannelByTempId + reqData.TemporaryChannelId)
		log.Println(err)
		return nil, err
	}

	fundingTransaction := &dao.FundingTransaction{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("ChannelId", channelInfo.ChannelId),
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
	node["temporary_channel_id"] = channelInfo.TemporaryChannelId
	//node["approval"] = reqData.Approval
	fundingTransaction.FundeeSignAt = time.Now()
	//如果不同意这次充值
	//if reqData.Approval == false {
	//	fundingTransaction.CurrState = dao.FundingTransactionState_Defuse
	//	err = tx.Update(fundingTransaction)
	//	if err != nil {
	//		return nil, err
	//	}
	//	_ = tx.Commit()
	//	return node, nil
	//}
	channelInfo.ChannelId = fundingTransaction.ChannelId
	node["channel_id"] = channelInfo.ChannelId
	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		return nil, errors.New(enum.Tips_common_empty + " fundee_channel_address_private_key")
	}
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, myPubKey)
	if err != nil {
		return nil, err
	}

	//region  sign C1 tx
	// 二次签名
	rsmcTxId, signedRsmcHex, err := rpcClient.BtcSignRawTransaction(fundingTransaction.FunderRsmcHex, reqData.ChannelAddressPrivateKey)
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

	inputs, err := getInputsForNextTxByParseTxHashVout(signedRsmcHex, rsmcMultiAddress, rsmcMultiAddressScriptPubKey, rsmcRedeemScript)
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
			reqData.ChannelAddressPrivateKey,
		},
		inputs,
		outputAddress,
		changeToAddress,
		fundingTransaction.PropertyId,
		fundingTransaction.AmountA,
		getBtcMinerAmount(channelInfo.BtcAmount),
		1000,
		&rsmcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//endregion create RD tx for alice

	channelInfo.PropertyId = fundingTransaction.PropertyId
	channelInfo.Amount = funderAmount
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
	err = createCurrCommitmentTxBR(tx, dao.BRType_Rmsc, channelInfo, lastCommitmentTx, inputs, myAddress, reqData.ChannelAddressPrivateKey, *signer)
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

func (service *fundingTransactionManager) AfterBobSignAssetFundingAtAliceSide(data string, user *bean.User) (outData interface{}, err error) {
	jsonObj := gjson.Parse(data)
	temporaryChannelId := jsonObj.Get("temporary_channel_id").String()
	rsmcSignedHex := jsonObj.Get("rsmc_signed_hex").String()
	rdHex := jsonObj.Get("rd_hex").String()
	//approval := jsonObj.Get("approval").Bool()

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
		err = errors.New("not found the channelInfo ")
		log.Println(err)
		return nil, err
	}
	fundingTransaction := &dao.FundingTransaction{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", temporaryChannelId),
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CurrState", dao.FundingTransactionState_Create)).First(fundingTransaction)
	if err != nil {
		err = errors.New("not found the fundingTransaction")
		log.Println(err)
		return nil, err
	}

	node := make(map[string]interface{})
	fundingTransaction.FundeeSignAt = time.Now()
	node["temporary_channel_id"] = channelInfo.TemporaryChannelId
	//node["approval"] = approval
	//if approval == false {
	//	fundingTransaction.CurrState = dao.FundingTransactionState_Defuse
	//	_ = tx.Update(fundingTransaction)
	//	_ = tx.Commit()
	//	return node, nil
	//}

	tempPrivKey := tempAddrPrivateKeyMap[fundingTransaction.FunderPubKey2ForCommitment]
	if tool.CheckIsString(&tempPrivKey) == false {
		err = errors.New("not found the FunderPrivKey2ForCommitment")
		log.Println(err)
		return nil, err
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
	inputs, err := getInputsForNextTxByParseTxHashVout(rsmcSignedHex, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey, commitmentTxInfo.RSMCRedeemScript)
	if err != nil || len(inputs) == 0 {
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
	rdTransaction.SignAt = time.Now()
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

	//同步通道信息到tracker
	sendChannelStateToTracker(*channelInfo, *commitmentTxInfo)

	node["channel_id"] = channelInfo.ChannelId
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
	var items []dao.FundingBtcRequest
	err = user.Db.Select(
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		Find(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
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
	var items []dao.FundingBtcRequest
	err = user.Db.Select(q.Eq("TemporaryChannelId", tempChanId)).OrderBy("CreateAt").Reverse().Find(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (service *fundingTransactionManager) BtcFundingRDAllItem(user bean.User) (node []dao.MinerFeeRedeemTransaction, err error) {
	var items []dao.MinerFeeRedeemTransaction
	err = user.Db.Select(
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		Find(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
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
	var items []dao.MinerFeeRedeemTransaction
	err = user.Db.Select(q.Eq("TemporaryChannelId", tempChanId)).OrderBy("CreateAt").Reverse().Find(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (service *fundingTransactionManager) BtcFundingItemByChannelId(channelId string, user bean.User) (node []dao.FundingBtcRequest, err error) {
	if tool.CheckIsString(&channelId) == false {
		return nil, errors.New(enum.Tips_common_wrong + "channelId")
	}
	channelInfo := dao.ChannelInfo{}
	_ = user.Db.Select(q.Eq("ChannelId", channelId)).First(&channelInfo)
	if channelInfo.Id == 0 {
		return nil, errors.New(enum.Tips_funding_notFoundChannelByChannelId + channelId)
	}
	var itemes []dao.FundingBtcRequest
	err = user.Db.Select(q.Eq("TemporaryChannelId", channelInfo.TemporaryChannelId)).OrderBy("CreateAt").Reverse().Find(&itemes)
	if err != nil {
		return nil, err
	}
	return itemes, nil
}

func (service *fundingTransactionManager) BtcFundingItemRDByTempChannelIdAndFundingTxid(jsonData string, user bean.User) (node *dao.MinerFeeRedeemTransaction, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New(enum.Tips_common_wrong + "request data")
	}
	var tempChanId = gjson.Parse(jsonData).Get("temporaryChannelId").String()
	if tool.CheckIsString(&tempChanId) == false {
		return nil, errors.New(enum.Tips_common_wrong + "temporaryChannelId")
	}
	var fundingTxid = gjson.Parse(jsonData).Get("fundingTxid").String()
	if tool.CheckIsString(&fundingTxid) == false {
		return nil, errors.New(enum.Tips_common_wrong + "fundingTxid")
	}

	var item = &dao.MinerFeeRedeemTransaction{}
	err = user.Db.Select(q.Eq("TemporaryChannelId", tempChanId), q.Eq("FundingTxId", fundingTxid)).OrderBy("CreateAt").Reverse().First(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (service *fundingTransactionManager) AssetFundingAllItem(user bean.User) (node []dao.FundingTransaction, err error) {
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

func (service *fundingTransactionManager) AssetFundingItemById(id int, user bean.User) (node *dao.FundingTransaction, err error) {
	var data = &dao.FundingTransaction{}
	err = user.Db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) AssetFundingItemByChannelId(chanId string, user bean.User) (node *dao.FundingTransaction, err error) {
	var item = &dao.FundingTransaction{}
	err = user.Db.Select(q.Eq("ChannelId", chanId)).First(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (service *fundingTransactionManager) AssetFundingTotalCount(user bean.User) (count int, err error) {
	return user.Db.Select(
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		Count(&dao.FundingTransaction{})
}
