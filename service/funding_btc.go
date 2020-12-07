package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
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

	btcFeeTxHexDecode, err := omnicore.DecodeBtcRawTransaction(reqData.FundingTxHex)
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
	result := conn2tracker.ListReceivedByAddress(channelInfo.ChannelAddress)

	if result != "" {
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
		result = conn2tracker.TestMemPoolAccept(latestBtcFundingRequest.TxHash)
		if result != "" {
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
		needAliceSignData, err = omnicore.BtcCreateRawTransactionForUnsendInputTx(
			channelInfo.ChannelAddress,
			[]bean.TransactionInputItem{
				{
					Txid:         fundingBtcRequest.TxId,
					Vout:         vout,
					Amount:       amount,
					ScriptPubKey: channelInfo.ChannelAddressScriptPubKey},
			},
			[]bean.TransactionOutputItem{
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
		minerFeeRedeemTransaction.Amount = needAliceSignData["total_out_amount"].(float64)
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
			needAliceSignData, err = omnicore.BtcCreateRawTransactionForUnsendInputTx(
				channelInfo.ChannelAddress,
				[]bean.TransactionInputItem{
					{
						Txid:         fundingBtcRequest.TxId,
						Vout:         vout,
						Amount:       amount,
						ScriptPubKey: channelInfo.ChannelAddressScriptPubKey},
				},
				[]bean.TransactionOutputItem{
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
			clientSignHexData.TotalInAmount = amount
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if needAliceSignData != nil {
		clientSignHexData.Hex = needAliceSignData["hex"].(string)
		clientSignHexData.Inputs = needAliceSignData["inputs"]
		clientSignHexData.TotalInAmount = needAliceSignData["total_in_amount"].(float64)
		clientSignHexData.TotalOutAmount = needAliceSignData["total_out_amount"].(float64)
	} else {
		clientSignHexData.Hex = minerFeeRedeemTransaction.Hex
	}

	clientSignHexData.TemporaryChannelId = reqData.TemporaryChannelId
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
func (service *fundingTransactionManager) OnAliceSignBtcFundingMinerFeeRedeemTx(jsonObj string, user *bean.User) (retData interface{}, targetUser string, err error) {

	if tool.CheckIsString(&jsonObj) == false {
		return nil, "", errors.New("empty hex")
	}
	hex := gjson.Get(jsonObj, "hex").Str
	resultDecode, err := omnicore.DecodeBtcRawTransaction(hex)
	if err != nil {
		return nil, "", err
	}

	inputTxId := gjson.Get(resultDecode, "vin").Array()[0].Get("txid").Str

	key := user.PeerId + "_" + inputTxId
	fundingBtcOfP2p := tempBtcFundingCreatedData[key]
	if len(fundingBtcOfP2p.FundingTxid) == 0 {
		return nil, "", errors.New("not found the temp data, please send -100340 again")
	}

	_, err = rpcClient.CheckMultiSign(false, hex, 1)
	if err != nil {
		return nil, "", err
	}

	txid := rpcClient.GetTxId(hex)
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
	fundingBtcOfP2p.SignData.Hex = minerFeeRedeemTransaction.Hex

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

	btcFeeTxHexDecode, err := omnicore.DecodeBtcRawTransaction(fundingBtcHex)
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

		_, err = rpcClient.CheckMultiSign(false, reqData.SignedMinerRedeemTransactionHex, 2)
		if err != nil {
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

	result, err := omnicore.DecodeBtcRawTransaction(fundingBtcRequest.RedeemHex)
	if err != nil {
		return nil, "", err
	}
	minerOutAmount := gjson.Get(result, "vout").Array()[0].Get("value").Float()

	resultDecode, err := omnicore.DecodeBtcRawTransaction(reqData.SignedMinerRedeemTransactionHex)
	if err != nil {
		return nil, "", err
	}
	if checkSignMinerFeeRedeemTx(resultDecode, reqData.FundingTxid, funderAddress, minerOutAmount) == false {
		return nil, "", errors.New("wrong signed hex")
	}

	//赎回交易签名成功后，广播交易
	_, err = conn2tracker.SendRawTransaction(fundingBtcRequest.TxHash)
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

func checkHexOutputAddressFromOmniDecode(hexDecode string, toAddress string) string {
	_, err := omnicore.VerifyOmniTxHexOutAddress(hexDecode, toAddress)
	if err != nil {
		return ""
	}
	return rpcClient.GetTxId(hexDecode)
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
