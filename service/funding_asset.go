package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/chainhash"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
	"log"
	"time"
)

var tempAssetFundingCreatedData map[string]bean.FundingAssetOfP2p
var tempAssetFundingSignData map[string]map[string]interface{}
var tempAssetFundingAfterBobSignData map[string]string

// ********************* omni资产充值 *********************
// funder request to fund to the multiAddr (channel)

// 协议号：100034 充值者alice充值资产，请求创建C1a 需要增加Alice对C1a的签名
func (service *fundingTransactionManager) AssetFundingCreated(msg bean.RequestMessage, user *bean.User) (outputData interface{}, needSign bool, err error) {
	reqData := &bean.SendRequestAssetFunding{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "temporary_channel_id ")
		log.Println(err)
		return nil, false, err
	}

	if tool.CheckIsString(&reqData.FundingTxHex) == false {
		err = errors.New(enum.Tips_common_empty + " funding_tx_hex ")
		log.Println(err)
		return nil, false, err
	}

	testResult, _ := rpcClient.TestMemPoolAccept(reqData.FundingTxHex)
	if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
		return nil, false, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
	}

	if _, err := getAddressFromPubKey(reqData.TempAddressPubKey); err != nil {
		err = errors.New(enum.Tips_common_wrong + "temp_address_pub_key ")
		log.Println(err)
		return nil, false, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
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
		return nil, false, err
	}

	if channelInfo.CurrState != dao.ChannelState_WaitFundAsset {
		err = errors.New(enum.Tips_funding_notFundAssetState)
		log.Println(err)
		return nil, false, err
	}

	err = checkBtcFundFinish(channelInfo.ChannelAddress, true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	// if alice launch funding
	fundingTxHexDecode, err := rpcClient.OmniDecodeTransaction(reqData.FundingTxHex)
	if err != nil {
		err = errors.New(enum.Tips_funding_failDecodeRawTransaction + " : " + err.Error())
		log.Println(err)
		return nil, false, err
	}

	funder := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdB {
		funder = channelInfo.PeerIdB
	}
	fundingTxid, amountA, propertyId, err := checkOmniTxHex(fundingTxHexDecode, channelInfo, funder)
	if err != nil {
		log.Println(err)
		return nil, false, err
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

	// getProperty from omnicore
	// 验证PropertyId是否在omni存在
	_, err = rpcClient.OmniGetProperty(propertyId)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	fundingAssetTxHexDecode, err := rpcClient.DecodeRawTransaction(reqData.FundingTxHex)
	if err != nil {
		err = errors.New(enum.Tips_funding_failDecodeRawTransaction + " funding_tx_hex " + err.Error())
		log.Println(err)
		return nil, false, err
	}

	//get btc miner Fee data from transaction
	fundingTxid, _, fundingOutputIndex, err := checkBtcTxHex(fundingAssetTxHexDecode, channelInfo, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, false, err
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
			return nil, false, err
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

	fundingTransaction.CurrState = dao.FundingTransactionState_Create
	fundingTransaction.CreateBy = user.PeerId
	fundingTransaction.CreateAt = time.Now()
	err = tx.Save(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	channelInfo.ChannelId = fundingTransaction.ChannelId
	channelInfo.PropertyId = propertyId
	channelInfo.Amount = amountA
	channelInfo.FundingAddress = fundingTransaction.FunderAddress

	var commitmentTxInfo *dao.CommitmentTransaction

	var c1aTxData map[string]interface{}
	var usedTxid string
	needAliceSignC1aObj := bean.NeedClientSignHexData{}
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
			return nil, false, err
		}
		commitmentTxInfo.RSMCTempAddressIndex = reqData.TempAddressIndex

		if commitmentTxInfo.AmountToRSMC > 0 {
			c1aTxData, usedTxid, err = rpcClient.OmniCreateRawTransactionUseSingleInput(
				int(commitmentTxInfo.TxType),
				channelInfo.ChannelAddress,
				commitmentTxInfo.RSMCMultiAddress,
				fundingTransaction.PropertyId,
				commitmentTxInfo.AmountToRSMC,
				0,
				0, &channelInfo.ChannelAddressRedeemScript, "")
			if err != nil {
				log.Println(err)
				return nil, false, err
			}
			commitmentTxInfo.RsmcInputTxid = usedTxid
			commitmentTxInfo.RSMCTxHex = c1aTxData["hex"].(string)
		}

		commitmentTxInfo.CurrState = dao.TxInfoState_Create
		commitmentTxInfo.LastHash = ""
		commitmentTxInfo.CurrHash = ""
		err = tx.Save(commitmentTxInfo)
		if err != nil {
			log.Println(err)
			return nil, false, err
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
			return nil, false, errors.New(enum.Tips_common_notFound + ": CommitmentTransaction")
		}
		if commitmentTxInfo.LastHash != "" {
			return nil, false, errors.New(enum.Tips_common_wrong + "CommitmentTransaction")
		}

		// 如果没有完成alice对C1a的签名
		if len(commitmentTxInfo.RSMCTxid) == 0 {
			if commitmentTxInfo.AmountToRSMC > 0 {
				c1aTxData, usedTxid, err = rpcClient.OmniCreateRawTransactionUseSingleInput(
					int(commitmentTxInfo.TxType),
					channelInfo.ChannelAddress,
					commitmentTxInfo.RSMCMultiAddress,
					fundingTransaction.PropertyId,
					commitmentTxInfo.AmountToRSMC,
					0,
					0, &channelInfo.ChannelAddressRedeemScript, "")
				if err != nil {
					log.Println(err)
					return nil, false, err
				}
				commitmentTxInfo.RsmcInputTxid = usedTxid
				commitmentTxInfo.RSMCTxHex = c1aTxData["hex"].(string)
				_ = tx.Update(commitmentTxInfo)
			}
		} else {
			var inputs []map[string]interface{}
			node, err := rpcClient.GetInputInfo(channelInfo.ChannelAddress, commitmentTxInfo.RsmcInputTxid, channelInfo.ChannelAddressRedeemScript)
			if err != nil {
				return nil, false, err
			}
			inputs = append(inputs, node)
			needAliceSignC1aObj.Inputs = inputs
		}
	}

	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	if c1aTxData != nil {
		needAliceSignC1aObj.Hex = c1aTxData["hex"].(string)
		needAliceSignC1aObj.Inputs = c1aTxData["inputs"]
	} else {
		needAliceSignC1aObj.Hex = commitmentTxInfo.RSMCTxHex
	}
	needAliceSignC1aObj.TemporaryChannelId = reqData.TemporaryChannelId
	needAliceSignC1aObj.IsMultisig = true
	needAliceSignC1aObj.PubKeyA = channelInfo.PubKeyA
	needAliceSignC1aObj.PubKeyB = channelInfo.PubKeyB

	node := bean.FundingAssetOfP2p{}
	node.TemporaryChannelId = reqData.TemporaryChannelId
	node.FundingOmniHex = fundingTransaction.FundingTxHex
	node.C1aRsmcHex = commitmentTxInfo.RSMCTxHex
	node.RsmcTempAddressPubKey = commitmentTxInfo.RSMCTempAddressPubKey
	node.SignData = needAliceSignC1aObj
	node.FunderNodeAddress = msg.SenderNodePeerId
	node.FunderPeerId = msg.SenderUserPeerId

	// 如果需要签名，就发送待签名数据到alice客户端，否则就发送数据到bob所在的obd节点
	if c1aTxData != nil {
		if tempAssetFundingCreatedData == nil {
			tempAssetFundingCreatedData = make(map[string]bean.FundingAssetOfP2p)
		}
		tempAssetFundingCreatedData[user.PeerId+"_"+usedTxid] = node
		return needAliceSignC1aObj, true, nil
	}

	return node, false, err
}

// 协议号：101034 Alice对C1a签名（仅有ToRSMC）完成的响应
func (service *fundingTransactionManager) OnAliceSignC1a(msg bean.RequestMessage, user *bean.User) (outputData interface{}, err error) {

	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New("empty hex")
	}
	signedC1a := bean.AliceSignC1aOfAssetFunding{}
	_ = json.Unmarshal([]byte(msg.Data), &signedC1a)
	hex := signedC1a.SignedC1aHex
	if tool.CheckIsString(&hex) == false {
		return nil, errors.New(enum.Tips_common_wrong + "hex")
	}

	_, err = rpcClient.CheckMultiSign(true, hex, 1)
	if err != nil {
		return nil, err
	}

	resultDecode, err := rpcClient.DecodeRawTransaction(hex)
	txid := gjson.Get(resultDecode, "txid").Str
	inputTxId := gjson.Get(resultDecode, "vin").Array()[0].Get("txid").Str

	key := user.PeerId + "_" + inputTxId
	fundingAssetOfP2p := tempAssetFundingCreatedData[key]
	if len(fundingAssetOfP2p.TemporaryChannelId) == 0 {
		return nil, errors.New("not found the temp data, please send -100034 again")
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", fundingAssetOfP2p.TemporaryChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(channelInfo)
	if err != nil {
		err = errors.New(enum.Tips_funding_notFoundChannelByTempId + fundingAssetOfP2p.TemporaryChannelId)
		log.Println(err)
		return nil, err
	}

	commitmentTxInfo := &dao.CommitmentTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		First(commitmentTxInfo)
	if err != nil {
		return nil, errors.New(enum.Tips_common_notFound + ": CommitmentTransaction")
	}

	// 检测 输出地址，数量是否一致
	omniDecode, err := rpcClient.OmniDecodeTransaction(hex)
	if err != nil {
		return nil, err
	}
	sendingaddress := gjson.Get(omniDecode, "sendingaddress").Str
	referenceaddress := gjson.Get(omniDecode, "referenceaddress").Str
	amount := gjson.Get(omniDecode, "amount").Float()
	if amount != commitmentTxInfo.AmountToRSMC || sendingaddress != channelInfo.ChannelAddress || referenceaddress != commitmentTxInfo.RSMCMultiAddress {
		return nil, errors.New("err sign")
	}
	commitmentTxInfo.RSMCTxid = txid
	commitmentTxInfo.RSMCTxHex = hex
	_ = tx.Update(commitmentTxInfo)

	_ = tx.Commit()

	fundingAssetOfP2p.SignData.Hex = commitmentTxInfo.RSMCTxHex

	return fundingAssetOfP2p, nil
}

// 协议号：34 发送到bob所在obd的数据处理，然后再发给bob的客户端
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

// 协议号：100035 bob签收这次资产充值交易
func (service *fundingTransactionManager) AssetFundingSigned(jsonData string, signer *bean.User) (outputData map[string]interface{}, err error) {
	reqData := &bean.SignAssetFunding{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		return nil, errors.New(enum.Tips_common_empty + "temporary_channel_id")
	}

	if tool.CheckIsString(&reqData.SignedAliceRsmcHex) == false {
		return nil, errors.New(enum.Tips_common_empty + "signed_alice_rsmc_hex")
	}

	_, err = rpcClient.CheckMultiSign(true, reqData.SignedAliceRsmcHex, 2)
	if err != nil {
		return nil, err
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
		q.Eq("CurrState", dao.FundingTransactionState_Create),
	).
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
	fundingTransaction.FundeeSignAt = time.Now()
	channelInfo.ChannelId = fundingTransaction.ChannelId
	node["channel_id"] = channelInfo.ChannelId

	//region  sign C1 tx
	// 二次签名的验证
	signedRsmcHex := reqData.SignedAliceRsmcHex

	beforeSignAliceRsmcDecode, err := rpcClient.OmniDecodeTransaction(fundingTransaction.FunderRsmcHex)
	if err != nil {
		return nil, err
	}

	beforeSignAliceRsmcSendingaddress := gjson.Get(beforeSignAliceRsmcDecode, "sendingaddress").String()
	beforeSignAliceRsmcReferenceaddress := gjson.Get(beforeSignAliceRsmcDecode, "referenceaddress").Str
	beforeSignAliceRsmcAmount := gjson.Get(beforeSignAliceRsmcDecode, "amount").Float()

	omniDecode, err := rpcClient.OmniDecodeTransaction(signedRsmcHex)
	if err != nil {
		return nil, err
	}

	bodSignedRsmcTxId := gjson.Get(omniDecode, "txid").Str
	bodSignedSendingaddress := gjson.Get(omniDecode, "sendingaddress").Str
	bodSignedReferenceaddress := gjson.Get(omniDecode, "referenceaddress").Str
	amount := gjson.Get(omniDecode, "amount").Float()
	if beforeSignAliceRsmcSendingaddress != bodSignedSendingaddress ||
		bodSignedSendingaddress != channelInfo.ChannelAddress ||
		amount != beforeSignAliceRsmcAmount ||
		bodSignedReferenceaddress != beforeSignAliceRsmcReferenceaddress {
		return nil, errors.New("err rsmc sign")
	}
	//endregion

	//region create RD tx for alice
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{fundingTransaction.RsmcTempAddressPubKey, myPubKey})
	if err != nil {
		return nil, err
	}
	aliceRsmcMultiAddress := gjson.Get(multiAddr, "address").String()
	aliceRsmcRedeemScript := gjson.Get(multiAddr, "redeemScript").String()
	tempJson, err := rpcClient.GetAddressInfo(aliceRsmcMultiAddress)
	if err != nil {
		return nil, err
	}
	rsmcMultiAddressScriptPubKey := gjson.Get(tempJson, "scriptPubKey").String()

	inputs, err := getInputsForNextTxByParseTxHashVout(signedRsmcHex, aliceRsmcMultiAddress, rsmcMultiAddressScriptPubKey, aliceRsmcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	outputAddress := channelInfo.AddressA
	if funder == channelInfo.PeerIdB {
		outputAddress = channelInfo.AddressB
	}

	aliceRdHexDataMap, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		aliceRsmcMultiAddress,
		inputs,
		outputAddress,
		changeToAddress,
		fundingTransaction.PropertyId,
		fundingTransaction.AmountA,
		getBtcMinerAmount(channelInfo.BtcAmount),
		1000,
		&aliceRsmcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	aliceRdHexDataMap["is_multisig"] = true
	aliceRdHexDataMap["pub_key_a"] = fundingTransaction.RsmcTempAddressPubKey
	aliceRdHexDataMap["pub_key_b"] = myPubKey
	//endregion create RD tx for alice

	channelInfo.PropertyId = fundingTransaction.PropertyId
	channelInfo.Amount = funderAmount
	channelInfo.FundingAddress = fundingTransaction.FunderAddress

	// region create rawBR1b tx  for bob
	lastCommitmentTx := &dao.CommitmentTransaction{}
	lastCommitmentTx.Id = 0
	lastCommitmentTx.PropertyId = channelInfo.PropertyId
	lastCommitmentTx.RSMCTempAddressPubKey = fundingTransaction.RsmcTempAddressPubKey
	lastCommitmentTx.RSMCMultiAddress = aliceRsmcMultiAddress
	lastCommitmentTx.RSMCRedeemScript = aliceRsmcRedeemScript
	lastCommitmentTx.RSMCMultiAddressScriptPubKey = rsmcMultiAddressScriptPubKey
	lastCommitmentTx.RSMCTxHex = signedRsmcHex
	lastCommitmentTx.RSMCTxid = bodSignedRsmcTxId
	lastCommitmentTx.AmountToRSMC = funderAmount
	aliceBrHexDataMap, err := createCurrCommitmentTxRawBR(tx, dao.BRType_Rmsc, channelInfo, lastCommitmentTx, inputs, myAddress, *signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	aliceBrHexDataMap["is_multisig"] = true
	aliceBrHexDataMap["pub_key_a"] = fundingTransaction.RsmcTempAddressPubKey
	aliceBrHexDataMap["pub_key_b"] = myPubKey
	//endregion

	err = tx.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
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
	node["signedRsmcHex"] = signedRsmcHex
	node["aliceRdHexDataMap"] = aliceRdHexDataMap
	//bob缓存待返回给alice的数据
	if tempAssetFundingSignData == nil {
		tempAssetFundingSignData = make(map[string]map[string]interface{})
	}
	tempAssetFundingSignData[signer.PeerId+"_"+reqData.TemporaryChannelId] = node

	outputData = make(map[string]interface{})
	outputData["temporary_channel_id"] = reqData.TemporaryChannelId
	outputData["alice_rd_sign_data"] = aliceRdHexDataMap
	outputData["alice_br_sign_data"] = aliceBrHexDataMap

	return outputData, nil
}

// 协议号：101035 当bob完成了RD和BR的第一次签名
func (service *fundingTransactionManager) OnBobSignedRDAndBR(data string, user *bean.User) (aliceData, bobData map[string]interface{}, err error) {
	if tool.CheckIsString(&data) == false {
		return nil, nil, errors.New(enum.Tips_common_empty + " input data")
	}

	signRdAndBr := bean.SignRdAndBrOfAssetFunding{}
	_ = json.Unmarshal([]byte(data), &signRdAndBr)

	temporaryChannelId := signRdAndBr.TemporaryChannelId
	if tool.CheckIsString(&temporaryChannelId) == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + " temporary_channel_id")
	}

	brId := signRdAndBr.BrId
	if brId == 0 {
		return nil, nil, errors.New(enum.Tips_common_wrong + " br_id")
	}

	rdSignedHex := signRdAndBr.RdSignedHex
	if tool.CheckIsString(&rdSignedHex) == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + " rd_signed_hex")
	}
	_, err = rpcClient.CheckMultiSign(false, rdSignedHex, 1)
	if err != nil {
		return nil, nil, errors.New(enum.Tips_common_wrong + "rd_signed_hex")
	}

	brSignedHex := signRdAndBr.BrSignedHex
	if tool.CheckIsString(&brSignedHex) == false {
		return nil, nil, errors.New(enum.Tips_common_wrong + " br_signed_hex")
	}

	_, err = rpcClient.CheckMultiSign(false, brSignedHex, 1)
	if err != nil {
		return nil, nil, errors.New(enum.Tips_common_wrong + "br_signed_hex")
	}

	// 发送之前 bob的obd获取缓存待返回给alice的数据
	tempData := tempAssetFundingSignData[user.PeerId+"_"+temporaryChannelId]
	signedRsmcHex := tempData["signedRsmcHex"].(string)
	if len(signedRsmcHex) == 0 {
		return nil, nil, errors.New("wrong temporary_channel_id")
	}

	aliceRdHexDataMap := tempData["aliceRdHexDataMap"].(map[string]interface{})

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", temporaryChannelId),
		q.Eq("CurrState", dao.ChannelState_WaitFundAsset),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId)),
	).First(channelInfo)
	if err != nil {
		return nil, nil, err
	}

	channelInfo.CurrState = dao.ChannelState_CanUse

	err = tx.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	fundingTransaction := &dao.FundingTransaction{}
	err = tx.Select(
		q.Eq("TemporaryChannelId", temporaryChannelId),
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CurrState", dao.FundingTransactionState_Create),
	).
		OrderBy("CreateAt").Reverse().
		First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	fundingTransaction.CurrState = dao.FundingTransactionState_Accept
	err = tx.Update(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	err = updateCurrCommitmentTxRawBR(tx, brId, brSignedHex, *user)
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	// 把签名后的rd hex放入回传给alice的数据包里面
	aliceData = make(map[string]interface{})
	aliceRdHexDataMap["hex"] = rdSignedHex
	aliceData["temporary_channel_id"] = channelInfo.TemporaryChannelId
	aliceData["channel_id"] = channelInfo.ChannelId
	aliceData["rsmc_signed_hex"] = signedRsmcHex
	aliceData["alice_rd_hex_data"] = aliceRdHexDataMap

	bobData = make(map[string]interface{})
	bobData["channel_id"] = channelInfo.ChannelId

	return aliceData, bobData, nil
}

// 协议号：-35 p2p bob的obd发给alice的obd，当获取到bob的签名数据后，首先把收到的RD的部分签名的hex发给alice签名
func (service *fundingTransactionManager) OnGetBobSignedMsgAndSendDataToAlice(jsonData string, user *bean.User) (outputData map[string]interface{}, err error) {
	jsonObj := gjson.Parse(jsonData)

	temporaryChannelId := jsonObj.Get("temporary_channel_id").String()
	channelId := jsonObj.Get("channel_id").String()

	aliceRdHexDataMap := make(map[string]interface{})
	aliceRdHex := jsonObj.Get("alice_rd_hex_data").String()
	err = json.Unmarshal([]byte(aliceRdHex), &aliceRdHexDataMap)
	if err != nil {
		log.Println(err)
	}
	aliceRdHexDataMap["temporary_channel_id"] = temporaryChannelId
	aliceRdHexDataMap["channel_id"] = channelId

	if tempAssetFundingAfterBobSignData == nil {
		tempAssetFundingAfterBobSignData = make(map[string]string)
	}
	tempAssetFundingAfterBobSignData[user.PeerId+"_"+channelId] = jsonData

	return aliceRdHexDataMap, nil
}

// 协议号：-101134 充值者alice在获得bob的签名数据后的业务逻辑，验证，更新和保存C1a的ToBob，ToRsmc，以及ToRsmc的RD和BR交易（这里需要增加一步临时私钥签名的过程）
func (service *fundingTransactionManager) OnAliceSignedRdAtAliceSide(data string, user *bean.User) (outData interface{}, err error) {

	signedRD := bean.AliceSignRDOfAssetFunding{}
	_ = json.Unmarshal([]byte(data), &signedRD)

	channelId := signedRD.ChannelId

	signedRdHex := signedRD.RdSignedHex
	_, err = rpcClient.CheckMultiSign(false, signedRdHex, 2)
	if err != nil {
		return nil, errors.New(enum.Tips_common_wrong + "rd_signed_hex")
	}

	cacheData := tempAssetFundingAfterBobSignData[user.PeerId+"_"+channelId]
	rsmcSignedHex := gjson.Get(cacheData, "rsmc_signed_hex").Str

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
	err = tx.Select(
		q.Eq("TemporaryChannelId", channelInfo.TemporaryChannelId),
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CurrState", dao.FundingTransactionState_Create)).First(fundingTransaction)
	if err != nil {
		err = errors.New("not found the fundingTransaction")
		log.Println(err)
		return nil, err
	}

	node := make(map[string]interface{})
	fundingTransaction.FundeeSignAt = time.Now()

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

	node["temporary_channel_id"] = channelInfo.TemporaryChannelId
	node["channel_id"] = channelInfo.ChannelId
	return node, nil
}
