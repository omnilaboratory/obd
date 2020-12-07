package service

import (
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/tidwall/gjson"
	"log"
	"time"
)

func createHtlcHLockTxObj(tx storm.Node, owner string, channelInfo dao.ChannelInfo, h string, commitmentTxInfo dao.CommitmentTransaction, outputBean map[string]interface{}, timeout int, user bean.User) (henxTx *dao.HtlcLockTxByH, err error) {
	henxTx = &dao.HtlcLockTxByH{}
	henxTx.ChannelId = channelInfo.ChannelId
	henxTx.CommitmentTxId = commitmentTxInfo.Id
	henxTx.PropertyId = commitmentTxInfo.PropertyId
	henxTx.HtlcH = commitmentTxInfo.HtlcH
	henxTx.Owner = owner
	count, err := tx.Select(
		q.Eq("ChannelId", henxTx.ChannelId),
		q.Eq("CommitmentTxId", henxTx.CommitmentTxId),
		q.Eq("Owner", owner)).
		Count(henxTx)
	if err == nil {
		if count > 0 {
			return nil, errors.New("already exist")
		}
	}
	//input
	henxTx.InputHex = commitmentTxInfo.HtlcTxHex
	henxTx.InputTxid = commitmentTxInfo.HTLCTxid

	henxTx.PayeeChannelPubKey = outputBean["otherSideChannelPubKey"].(string)
	multiAddr, err := omnicore.CreateMultiSig(2, []string{h, henxTx.PayeeChannelPubKey})
	if err != nil {
		return nil, err
	}
	henxTx.OutputAddress = gjson.Get(multiAddr, "address").String()
	henxTx.RedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	henxTx.ScriptPubKey = gjson.Get(multiAddr, "scriptPubKey").String()
	henxTx.OutAmount = outputBean["amount"].(float64)
	henxTx.Timeout = timeout
	henxTx.CreateBy = user.PeerId
	henxTx.CreateAt = time.Now()
	return henxTx, nil
}

func createHtlcTimeoutTxObj(tx storm.Node, owner string, channelInfo dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTransaction, outputBean commitmentTxOutputBean, timeout int, user bean.User) (*dao.HTLCTimeoutTxForAAndExecutionForB, error) {
	htlcTimeoutTx := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	htlcTimeoutTx.ChannelId = channelInfo.ChannelId
	htlcTimeoutTx.CommitmentTxId = commitmentTxInfo.Id
	htlcTimeoutTx.PropertyId = commitmentTxInfo.PropertyId
	htlcTimeoutTx.Owner = owner
	count, err := tx.Select(
		q.Eq("ChannelId", htlcTimeoutTx.ChannelId),
		q.Eq("CommitmentTxId", htlcTimeoutTx.CommitmentTxId),
		q.Eq("Owner", owner)).
		Count(htlcTimeoutTx)
	if err == nil {
		if count > 0 {
			return nil, errors.New("already exist")
		}
	}
	//input
	htlcTimeoutTx.InputHex = commitmentTxInfo.HtlcTxHex
	htlcTimeoutTx.InputAmount = commitmentTxInfo.AmountToHtlc

	//output to rsmc
	htlcTimeoutTx.RSMCTempAddressPubKey = outputBean.RsmcTempPubKey
	multiAddr, err := omnicore.CreateMultiSig(2, []string{htlcTimeoutTx.RSMCTempAddressPubKey, outputBean.OppositeSideChannelPubKey})
	if err != nil {
		return nil, err
	}
	htlcTimeoutTx.RSMCMultiAddress = gjson.Get(multiAddr, "address").String()
	htlcTimeoutTx.RSMCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	htlcTimeoutTx.RSMCMultiAddressScriptPubKey = gjson.Get(multiAddr, "scriptPubKey").String()
	htlcTimeoutTx.RSMCOutAmount = outputBean.AmountToRsmc
	htlcTimeoutTx.Timeout = timeout
	htlcTimeoutTx.CreateBy = user.PeerId
	htlcTimeoutTx.CreateAt = time.Now()
	return htlcTimeoutTx, nil
}

func createHtlcRDTxObj(owner string, channelInfo *dao.ChannelInfo, htlcTimeoutTx *dao.HTLCTimeoutTxForAAndExecutionForB, toAddress string,
	user *bean.User) (*dao.RevocableDeliveryTransaction, error) {
	htrd := &dao.RevocableDeliveryTransaction{}
	htrd.ChannelId = channelInfo.ChannelId
	htrd.CommitmentTxId = htlcTimeoutTx.Id
	htrd.PeerIdA = channelInfo.PeerIdA
	htrd.PeerIdB = channelInfo.PeerIdB
	htrd.PropertyId = htlcTimeoutTx.PropertyId
	htrd.Owner = owner
	htrd.RDType = 1

	//input
	htrd.InputTxHex = htlcTimeoutTx.RSMCTxHex
	htrd.InputTxid = htlcTimeoutTx.RSMCTxid
	htrd.InputVout = 0
	htrd.InputAmount = htlcTimeoutTx.RSMCOutAmount
	//output
	htrd.OutputAddress = toAddress
	htrd.Sequence = 1000
	htrd.Amount = htlcTimeoutTx.RSMCOutAmount

	htrd.CreateBy = user.PeerId
	htrd.CreateAt = time.Now()
	htrd.LastEditTime = time.Now()

	return htrd, nil
}

//为alice生成HT1a
func createHT1aForAlice(channelInfo dao.ChannelInfo, aliceDataJson bean.CreateHtlcTxForC3aOfP2p, signedHtlcHex string,
	bobChannelPubKey string, propertyId int64, amountToHtlc float64, htlcTimeOut int) (*bean.NeedClientSignTxData, error) {
	aliceHtlcMultiAddr, err := omnicore.CreateMultiSig(2, []string{aliceDataJson.CurrHtlcTempAddressPubKey, bobChannelPubKey})
	if err != nil {
		return nil, err
	}
	aliceHtlcMultiAddress := gjson.Get(aliceHtlcMultiAddr, "address").String()
	aliceHtlcRedeemScript := gjson.Get(aliceHtlcMultiAddr, "redeemScript").String()
	aliceHtlcMultiAddressScriptPubKey := gjson.Get(aliceHtlcMultiAddr, "scriptPubKey").String()

	htlcInputs, err := getInputsForNextTxByParseTxHashVout(signedHtlcHex, aliceHtlcMultiAddress, aliceHtlcMultiAddressScriptPubKey, aliceHtlcRedeemScript)
	if err != nil || len(htlcInputs) == 0 {
		log.Println(err)
		return nil, err
	}

	aliceHt1aMultiAddr, err := omnicore.CreateMultiSig(2, []string{aliceDataJson.CurrHtlcTempAddressForHt1aPubKey, bobChannelPubKey})
	if err != nil {
		return nil, err
	}
	aliceHt1aMultiAddress := gjson.Get(aliceHt1aMultiAddr, "address").String()
	aliceHt1aTxData, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		aliceHtlcMultiAddress,
		htlcInputs,
		aliceHt1aMultiAddress,
		aliceHt1aMultiAddress,
		propertyId,
		amountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		htlcTimeOut,
		&aliceHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	ht1aTxData := bean.NeedClientSignTxData{}
	ht1aTxData.Hex = aliceHt1aTxData["hex"].(string)
	ht1aTxData.Inputs = aliceHt1aTxData["inputs"]
	ht1aTxData.IsMultisig = true
	return &ht1aTxData, nil
}

func saveHT1aForAlice(tx storm.Node, channelInfo dao.ChannelInfo, commitmentTransaction *dao.CommitmentTransaction,
	signedHtlaHex string, htlcRequestInfo dao.AddHtlcRequestInfo, payeePubKey string, htlcTimeOut int, user bean.User) (htlcTimeoutTx *dao.HTLCTimeoutTxForAAndExecutionForB, err error) {

	outputBean := commitmentTxOutputBean{}
	outputBean.AmountToRsmc = commitmentTransaction.AmountToHtlc
	outputBean.RsmcTempPubKey = htlcRequestInfo.CurrHtlcTempAddressForHt1aPubKey
	outputBean.OppositeSideChannelPubKey = payeePubKey
	htlcTimeoutTx, err = createHtlcTimeoutTxObj(tx, user.PeerId, channelInfo, commitmentTransaction, outputBean, htlcTimeOut, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	htlcTimeoutTx.RSMCTempAddressIndex = htlcRequestInfo.CurrHtlcTempAddressForHt1aIndex

	payerHt1aInputsFromHtlc, err := getInputsForNextTxByParseTxHashVout(
		commitmentTransaction.HtlcTxHex,
		commitmentTransaction.HTLCMultiAddress,
		commitmentTransaction.HTLCMultiAddressScriptPubKey,
		commitmentTransaction.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	htlcTimeoutTx.InputTxid = payerHt1aInputsFromHtlc[0].Txid

	htlcTimeoutTx.RSMCTxid = rpcClient.GetTxId(signedHtlaHex)
	htlcTimeoutTx.RSMCTxHex = signedHtlaHex
	htlcTimeoutTx.SignAt = time.Now()
	htlcTimeoutTx.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(htlcTimeoutTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return htlcTimeoutTx, nil
}

// 收款方在41号协议用签名完成toHtlc的Hex，就用这个完整交易Hex，构建C3a方的Hlock交易
func createHtlcLockByHForBobAtPayeeSide(channelInfo dao.ChannelInfo, aliceDataJson bean.CreateHtlcTxForC3aOfP2p, signedHtlcHex string,
	bobChannelPubKey string, propertyId int64, amountToHtlc float64) (*bean.NeedClientSignTxData, error) {
	aliceHtlcMultiAddr, err := omnicore.CreateMultiSig(2, []string{aliceDataJson.CurrHtlcTempAddressPubKey, bobChannelPubKey})
	if err != nil {
		return nil, err
	}
	aliceHtlcMultiAddress := gjson.Get(aliceHtlcMultiAddr, "address").String()
	aliceHtlcRedeemScript := gjson.Get(aliceHtlcMultiAddr, "redeemScript").String()
	aliceHtlcMultiAddressScriptPubKey := gjson.Get(aliceHtlcMultiAddr, "scriptPubKey").String()

	htlcInputs, err := getInputsForNextTxByParseTxHashVout(signedHtlcHex, aliceHtlcMultiAddress, aliceHtlcMultiAddressScriptPubKey, aliceHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcLockByHMultiAddr, err := omnicore.CreateMultiSig(2, []string{aliceDataJson.H, bobChannelPubKey})
	if err != nil {
		return nil, err
	}
	htlcLockByHMultiAddress := gjson.Get(htlcLockByHMultiAddr, "address").String()
	hlockHexData, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		aliceHtlcMultiAddress,
		htlcInputs,
		htlcLockByHMultiAddress,
		htlcLockByHMultiAddress,
		propertyId,
		amountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		0,
		&aliceHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c2bBrRawData := bean.NeedClientSignTxData{}
	c2bBrRawData.Hex = hlockHexData["hex"].(string)
	c2bBrRawData.Inputs = hlockHexData["inputs"]
	c2bBrRawData.IsMultisig = true

	return &c2bBrRawData, nil
}

func saveHtlcLockByHTxAtPayerSide(tx storm.Node, channelInfo dao.ChannelInfo,
	commitmentTransaction *dao.CommitmentTransaction, signedHLockHex string, user bean.User) (henxTx *dao.HtlcLockTxByH, err error) {
	payeePubKey := channelInfo.PubKeyB
	if user.PeerId == channelInfo.PeerIdB {
		payeePubKey = channelInfo.PubKeyA
	}
	outputBean := make(map[string]interface{})
	outputBean["amount"] = commitmentTransaction.AmountToHtlc
	outputBean["otherSideChannelPubKey"] = payeePubKey

	hlockTx, err := createHtlcHLockTxObj(tx, user.PeerId, channelInfo, commitmentTransaction.HtlcH, *commitmentTransaction, outputBean, 0, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcOutputs, err := getInputsForNextTxByParseTxHashVout(
		commitmentTransaction.HtlcTxHex,
		commitmentTransaction.HTLCMultiAddress,
		commitmentTransaction.HTLCMultiAddressScriptPubKey,
		commitmentTransaction.HTLCRedeemScript)
	if err != nil || len(htlcOutputs) == 0 {
		log.Println(err)
		return nil, err
	}

	hlockTx.Txid = rpcClient.GetTxId(signedHLockHex)
	hlockTx.TxHex = signedHLockHex
	hlockTx.CreateAt = time.Now()
	hlockTx.CurrState = dao.TxInfoState_Create
	err = tx.Save(hlockTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return hlockTx, err
}

// 创建和保存收款方的He的Raw交易
func saveHtlcHeTxForPayee(tx storm.Node, channelInfo dao.ChannelInfo, commitmentTransaction *dao.CommitmentTransaction, reqData bean.BobSignedHtlcSubTxOfC3b, c3bHlockHex string, user bean.User) (c3bHeRawData *bean.NeedClientSignTxData, err error) {
	he1b := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTransaction.Id),
		q.Eq("Owner", user.PeerId)).First(he1b)

	payeeChannelPubKey := channelInfo.PubKeyB
	payerChannelPubkey := channelInfo.PubKeyA
	if user.PeerId == channelInfo.PeerIdA {
		payeeChannelPubKey = channelInfo.PubKeyA
		payerChannelPubkey = channelInfo.PubKeyB
	}

	c3bHlockMultiAddress, c3bHlockRedeemScript, c3bHlockAddrScriptPubKey, err := createMultiSig(commitmentTransaction.HtlcH, payeeChannelPubKey)
	if err != nil {
		return nil, err
	}

	c3bHeMultiAddress, redeemScript, scriptPubKey, err := createMultiSig(reqData.CurrHtlcTempAddressForHePubKey, payerChannelPubkey)
	if err != nil {
		return nil, err
	}

	c3bHlocOutputs, err := getInputsForNextTxByParseTxHashVout(c3bHlockHex, c3bHlockMultiAddress, c3bHlockAddrScriptPubKey, c3bHlockRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c3bHeTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		c3bHlockMultiAddress,
		c3bHlocOutputs,
		c3bHeMultiAddress,
		c3bHeMultiAddress,
		channelInfo.PropertyId,
		commitmentTransaction.AmountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		0,
		&c3bHlockRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create HTD1b for C3b")
	}

	c3bHeRawData = &bean.NeedClientSignTxData{}
	c3bHeRawData.Hex = c3bHeTx["hex"].(string)
	c3bHeRawData.Inputs = c3bHeTx["inputs"]
	c3bHeRawData.IsMultisig = true
	c3bHeRawData.PubKeyA = commitmentTransaction.HtlcH
	c3bHeRawData.PubKeyB = payeeChannelPubKey

	if he1b.Id == 0 {
		he1b.InputHex = c3bHlockHex
		he1b.InputTxid = c3bHlocOutputs[0].Txid
		he1b.InputAmount = commitmentTransaction.AmountToHtlc

		he1b.RSMCTempAddressIndex = reqData.CurrHtlcTempAddressForHeIndex
		he1b.RSMCTempAddressPubKey = reqData.CurrHtlcTempAddressForHePubKey
		he1b.RSMCMultiAddress = c3bHeMultiAddress
		he1b.RSMCRedeemScript = redeemScript
		he1b.RSMCMultiAddressScriptPubKey = scriptPubKey
		he1b.RSMCOutAmount = commitmentTransaction.AmountToHtlc
		he1b.RSMCTxHex = c3bHeRawData.Hex

		he1b.ChannelId = channelInfo.ChannelId
		he1b.CommitmentTxId = commitmentTransaction.Id
		he1b.Owner = user.PeerId
		he1b.CreateBy = user.PeerId
		he1b.CreateAt = time.Now()
		he1b.CurrState = dao.TxInfoState_Init
		err = tx.Save(he1b)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	return c3bHeRawData, nil
}

// bob对He完成第一次签名，更新收款方的He交易
func updateHtlcHeTxForPayee(tx storm.Node, channelInfo dao.ChannelInfo, commitmentTransaction *dao.CommitmentTransaction, signedHex string) (err error) {
	he1b := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CurrState", dao.TxInfoState_Init),
		q.Eq("CommitmentTxId", commitmentTransaction.Id)).First(he1b)
	if he1b.Id == 0 {
		return errors.New("Not found he1b in the database")
	}
	he1b.RSMCTxid = rpcClient.GetTxId(signedHex)
	he1b.RSMCTxHex = signedHex
	he1b.CurrState = dao.TxInfoState_Create
	err = tx.Update(he1b)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// 付款方在42号协议，用签名完成toHtlc的Hex，就用这个完整交易Hex，构建C3b方的Hlock交易
func createHtlcLockByHForBobAtPayerSide(channelInfo dao.ChannelInfo, bobDataJson bean.NeedAliceSignHtlcTxOfC3bP2p, signedHtlcHex string, h string, payeeChannelPubKey string,
	aliceChannelPubKey string, propertyId int64, amountToHtlc float64) (txData bean.NeedClientSignTxData, err error) {
	bobHtlcMultiAddr, err := omnicore.CreateMultiSig(2, []string{bobDataJson.PayeeCurrHtlcTempAddressPubKey, aliceChannelPubKey})
	if err != nil {
		return txData, err
	}
	bobHtlcMultiAddress := gjson.Get(bobHtlcMultiAddr, "address").String()
	bobHtlcRedeemScript := gjson.Get(bobHtlcMultiAddr, "redeemScript").String()
	bobHtlcMultiAddressScriptPubKey := gjson.Get(bobHtlcMultiAddr, "scriptPubKey").String()

	htlcInputs, err := getInputsForNextTxByParseTxHashVout(signedHtlcHex, bobHtlcMultiAddress, bobHtlcMultiAddressScriptPubKey, bobHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return txData, err
	}

	htlcLockByHMultiAddr, err := omnicore.CreateMultiSig(2, []string{h, payeeChannelPubKey})
	if err != nil {
		return txData, err
	}
	htlcLockByHMultiAddress := gjson.Get(htlcLockByHMultiAddr, "address").String()
	hlockTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
		bobHtlcMultiAddress,
		htlcInputs,
		htlcLockByHMultiAddress,
		htlcLockByHMultiAddress,
		propertyId,
		amountToHtlc,
		getBtcMinerAmount(channelInfo.BtcAmount),
		0,
		&bobHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return txData, err
	}

	c3bHlockRawData := bean.NeedClientSignTxData{}
	c3bHlockRawData.Hex = hlockTx["hex"].(string)
	c3bHlockRawData.Inputs = hlockTx["inputs"]
	c3bHlockRawData.IsMultisig = true
	c3bHlockRawData.PubKeyA = bobDataJson.PayeeCurrHtlcTempAddressPubKey
	c3bHlockRawData.PubKeyB = aliceChannelPubKey
	return c3bHlockRawData, nil
}

func saveHtlcLockByHForBobAtPayeeSide(tx storm.Node, channelInfo dao.ChannelInfo,
	commitmentTransaction *dao.CommitmentTransaction, signedHLockHex string, user bean.User) (henxTx *dao.HtlcLockTxByH, err error) {
	payeePubKey := channelInfo.PubKeyB
	if user.PeerId == channelInfo.PeerIdA {
		payeePubKey = channelInfo.PubKeyA
	}
	outputBean := make(map[string]interface{})
	outputBean["amount"] = commitmentTransaction.AmountToHtlc
	outputBean["otherSideChannelPubKey"] = payeePubKey

	hlock, err := createHtlcHLockTxObj(tx, user.PeerId, channelInfo, commitmentTransaction.HtlcH, *commitmentTransaction, outputBean, 0, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	hlock.Txid = rpcClient.GetTxId(signedHLockHex)
	hlock.TxHex = signedHLockHex
	hlock.CreateAt = time.Now()
	hlock.CurrState = dao.TxInfoState_Create
	err = tx.Save(hlock)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return hlock, err
}

func saveHtRD1a(tx storm.Node, aliceHt1aRDhex string, latestCommitmentTransaction dao.CommitmentTransaction, user bean.User) (rdTransaction *dao.RevocableDeliveryTransaction, err error) {
	ht1a := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = tx.Select(
		q.Eq("ChannelId", latestCommitmentTransaction.ChannelId),
		q.Eq("CommitmentTxId", latestCommitmentTransaction.Id),
		q.Eq("Owner", user.PeerId)).
		First(ht1a)
	if err != nil {
		err = errors.New("not found ht1a")
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(q.Eq("ChannelId", latestCommitmentTransaction.ChannelId)).First(channelInfo)
	if err != nil {
		err = errors.New("not found channelInfo when save htrd1a")
		return nil, err
	}

	outAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		outAddress = channelInfo.AddressB
	}

	htrd := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", latestCommitmentTransaction.ChannelId),
		q.Eq("CommitmentTxId", ht1a.Id),
		q.Eq("Owner", user.PeerId)).
		First(htrd)
	if htrd.Id == 0 {
		htrd, err = createHtlcRDTxObj(user.PeerId, channelInfo, ht1a, outAddress, &user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		htrd.Txid = rpcClient.GetTxId(aliceHt1aRDhex)
		htrd.TxHex = aliceHt1aRDhex
		htrd.SignAt = time.Now()
		htrd.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(htrd)
	}
	return htrd, nil
}
