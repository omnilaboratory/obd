package service

import (
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
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
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{h, henxTx.PayeeChannelPubKey})
	if err != nil {
		return nil, err
	}
	henxTx.OutputAddress = gjson.Get(multiAddr, "address").String()
	henxTx.RedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	jsonData, err := rpcClient.GetAddressInfo(henxTx.OutputAddress)
	if err != nil {
		return nil, err
	}
	henxTx.ScriptPubKey = gjson.Get(jsonData, "scriptPubKey").String()
	henxTx.OutAmount = outputBean["amount"].(float64)
	henxTx.Timeout = timeout
	henxTx.CreateBy = user.PeerId
	henxTx.CreateAt = time.Now()
	return henxTx, nil
}

func createHtlcTimeoutTxObj(tx storm.Node, owner string, channelInfo dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTransaction, outputBean commitmentOutputBean, timeout int, user bean.User) (*dao.HTLCTimeoutTxForAAndExecutionForB, error) {
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
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{htlcTimeoutTx.RSMCTempAddressPubKey, outputBean.OppositeSideChannelPubKey})
	if err != nil {
		return nil, err
	}
	htlcTimeoutTx.RSMCMultiAddress = gjson.Get(multiAddr, "address").String()
	htlcTimeoutTx.RSMCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	jsonData, err := rpcClient.GetAddressInfo(htlcTimeoutTx.RSMCMultiAddress)
	if err != nil {
		return nil, err
	}
	htlcTimeoutTx.RSMCMultiAddressScriptPubKey = gjson.Get(jsonData, "scriptPubKey").String()
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
func createHT1aForAlice(aliceDataJson bean.AliceRequestAddHtlc, signedHtlcHex string,
	bobChannelPubKey string, bobChannelAddressPrivateKey string,
	propertyId int64, amountToHtlc float64, htlcTimeOut int) (*string, error) {
	aliceHtlcMultiAddr, err := rpcClient.CreateMultiSig(2, []string{aliceDataJson.CurrHtlcTempAddressPubKey, bobChannelPubKey})
	if err != nil {
		return nil, err
	}
	aliceHtlcMultiAddress := gjson.Get(aliceHtlcMultiAddr, "address").String()
	aliceHtlcRedeemScript := gjson.Get(aliceHtlcMultiAddr, "redeemScript").String()
	tempJson, err := rpcClient.GetAddressInfo(aliceHtlcMultiAddress)
	if err != nil {
		return nil, err
	}
	aliceHtlcMultiAddressScriptPubKey := gjson.Get(tempJson, "scriptPubKey").String()

	htlcInputs, err := getInputsForNextTxByParseTxHashVout(signedHtlcHex, aliceHtlcMultiAddress, aliceHtlcMultiAddressScriptPubKey, aliceHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	aliceHt1aMultiAddr, err := rpcClient.CreateMultiSig(2, []string{aliceDataJson.CurrHtlcTempAddressForHt1aPubKey, bobChannelPubKey})
	if err != nil {
		return nil, err
	}
	aliceHt1aMultiAddress := gjson.Get(aliceHt1aMultiAddr, "address").String()
	_, aliceHt1ahex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		aliceHtlcMultiAddress,
		[]string{
			bobChannelAddressPrivateKey,
		},
		htlcInputs,
		aliceHt1aMultiAddress,
		aliceHt1aMultiAddress,
		propertyId,
		amountToHtlc,
		0,
		htlcTimeOut,
		&aliceHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &aliceHt1ahex, nil
}

func signHT1aForAlice(tx storm.Node, channelInfo dao.ChannelInfo, commitmentTransaction *dao.CommitmentTransaction,
	unsignedHt1aHex string, htlcTempPubKey string, payeePubKey string, htlaTempPubKey string, htlcTimeOut int, user bean.User) (htlcTimeoutTx *dao.HTLCTimeoutTxForAAndExecutionForB, err error) {

	outputBean := commitmentOutputBean{}
	outputBean.AmountToRsmc = commitmentTransaction.AmountToHtlc
	outputBean.RsmcTempPubKey = htlaTempPubKey
	outputBean.OppositeSideChannelPubKey = payeePubKey
	htlcTimeoutTx, err = createHtlcTimeoutTxObj(tx, user.PeerId, channelInfo, commitmentTransaction, outputBean, htlcTimeOut, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	payerHt1aInputsFromHtlc, err := getInputsForNextTxByParseTxHashVout(
		commitmentTransaction.HtlcTxHex,
		commitmentTransaction.HTLCMultiAddress,
		commitmentTransaction.HTLCMultiAddressScriptPubKey,
		commitmentTransaction.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, signedHtlaHex, err := rpcClient.OmniSignRawTransactionForUnsend(unsignedHt1aHex, payerHt1aInputsFromHtlc, tempAddrPrivateKeyMap[htlcTempPubKey])
	if err != nil {
		return nil, err
	}
	result, err := rpcClient.TestMemPoolAccept(signedHtlaHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
			return nil, errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
		}
	}

	htlcTimeoutTx.InputTxid = payerHt1aInputsFromHtlc[0].Txid

	htlcTimeoutTx.RSMCTxid = txid
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
func createHtlcLockByHForBobAtPayeeSide(aliceDataJson bean.AliceRequestAddHtlc, signedHtlcHex string,
	bobChannelPubKey string, bobChannelAddressPrivateKey string,
	propertyId int64, amountToHtlc float64) (*string, error) {
	aliceHtlcMultiAddr, err := rpcClient.CreateMultiSig(2, []string{aliceDataJson.CurrHtlcTempAddressPubKey, bobChannelPubKey})
	if err != nil {
		return nil, err
	}
	aliceHtlcMultiAddress := gjson.Get(aliceHtlcMultiAddr, "address").String()
	aliceHtlcRedeemScript := gjson.Get(aliceHtlcMultiAddr, "redeemScript").String()
	tempJson, err := rpcClient.GetAddressInfo(aliceHtlcMultiAddress)
	if err != nil {
		return nil, err
	}
	aliceHtlcMultiAddressScriptPubKey := gjson.Get(tempJson, "scriptPubKey").String()

	htlcInputs, err := getInputsForNextTxByParseTxHashVout(signedHtlcHex, aliceHtlcMultiAddress, aliceHtlcMultiAddressScriptPubKey, aliceHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcLockByHMultiAddr, err := rpcClient.CreateMultiSig(2, []string{aliceDataJson.H, bobChannelPubKey})
	if err != nil {
		return nil, err
	}
	htlcLockByHMultiAddress := gjson.Get(htlcLockByHMultiAddr, "address").String()
	_, hlockHex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		aliceHtlcMultiAddress,
		[]string{
			bobChannelAddressPrivateKey,
		},
		htlcInputs,
		htlcLockByHMultiAddress,
		htlcLockByHMultiAddress,
		propertyId,
		amountToHtlc,
		0,
		0,
		&aliceHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &hlockHex, nil
}

// 付款方在42号协议，签名Hlock交易，用H+收款方地址构建的多签地址，锁住给收款方的钱
func signHtlcLockByHTxAtPayerSide(tx storm.Node, channelInfo dao.ChannelInfo,
	commitmentTransaction *dao.CommitmentTransaction, lockByHForBobHex string, user bean.User) (henxTx *dao.HtlcLockTxByH, err error) {
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
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, signedHLockHex, err := rpcClient.OmniSignRawTransactionForUnsend(lockByHForBobHex, htlcOutputs, tempAddrPrivateKeyMap[commitmentTransaction.HTLCTempAddressPubKey])
	if err != nil {
		return nil, err
	}
	result, err := rpcClient.TestMemPoolAccept(signedHLockHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
			return nil, errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
		}
	}

	hlockTx.Txid = txid
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

// 付款方在42号协议，用签名完成toHtlc的Hex，就用这个完整交易Hex，构建C3b方的Hlock交易
func createHtlcLockByHForBobAtPayerSide(bobDataJson gjson.Result, signedHtlcHex string, h string, payeeChannelPubKey string,
	aliceChannelPubKey string, aliceChannelAddressPrivateKey string,
	propertyId int64, amountToHtlc float64) (*string, error) {
	bobHtlcMultiAddr, err := rpcClient.CreateMultiSig(2, []string{bobDataJson.Get("currHtlcTempAddressPubKey").String(), aliceChannelPubKey})
	if err != nil {
		return nil, err
	}
	bobHtlcMultiAddress := gjson.Get(bobHtlcMultiAddr, "address").String()
	bobHtlcRedeemScript := gjson.Get(bobHtlcMultiAddr, "redeemScript").String()
	tempJson, err := rpcClient.GetAddressInfo(bobHtlcMultiAddress)
	if err != nil {
		return nil, err
	}
	bobHtlcMultiAddressScriptPubKey := gjson.Get(tempJson, "scriptPubKey").String()

	htlcInputs, err := getInputsForNextTxByParseTxHashVout(signedHtlcHex, bobHtlcMultiAddress, bobHtlcMultiAddressScriptPubKey, bobHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcLockByHMultiAddr, err := rpcClient.CreateMultiSig(2, []string{h, payeeChannelPubKey})
	if err != nil {
		return nil, err
	}
	htlcLockByHMultiAddress := gjson.Get(htlcLockByHMultiAddr, "address").String()
	_, hlockHex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		bobHtlcMultiAddress,
		[]string{
			aliceChannelAddressPrivateKey,
		},
		htlcInputs,
		htlcLockByHMultiAddress,
		htlcLockByHMultiAddress,
		propertyId,
		amountToHtlc,
		0,
		0,
		&bobHtlcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &hlockHex, nil
}

// 收款方在43号协议，签名Hlock交易
func signHtlcLockByHForBobAtPayeeSide(tx storm.Node, channelInfo dao.ChannelInfo,
	commitmentTransaction *dao.CommitmentTransaction, lockByHForBobHex string, user bean.User) (henxTx *dao.HtlcLockTxByH, err error) {
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

	htlcOutputs, err := getInputsForNextTxByParseTxHashVout(
		commitmentTransaction.HtlcTxHex,
		commitmentTransaction.HTLCMultiAddress,
		commitmentTransaction.HTLCMultiAddressScriptPubKey,
		commitmentTransaction.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, signedHLockHex, err := rpcClient.OmniSignRawTransactionForUnsend(lockByHForBobHex, htlcOutputs, tempAddrPrivateKeyMap[commitmentTransaction.HTLCTempAddressPubKey])
	if err != nil {
		return nil, err
	}
	result, err := rpcClient.TestMemPoolAccept(signedHLockHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
			return nil, errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
		}
	}

	hlock.Txid = txid
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

func signHtRD1a(tx storm.Node, aliceHt1aRDhex string, latestCommitmentTransaction dao.CommitmentTransaction, user bean.User) (rdTransaction *dao.RevocableDeliveryTransaction, err error) {
	//签名并保存HTRD1a
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
	rdTransaction, err = createHtlcRDTxObj(user.PeerId, channelInfo, ht1a, outAddress, &user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ht1aOutputs, err := getInputsForNextTxByParseTxHashVout(ht1a.RSMCTxHex, ht1a.RSMCMultiAddress, ht1a.RSMCMultiAddressScriptPubKey, ht1a.RSMCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htrdTxid, htrdHex, err := rpcClient.OmniSignRawTransactionForUnsend(aliceHt1aRDhex, ht1aOutputs, tempAddrPrivateKeyMap[ht1a.RSMCTempAddressPubKey])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rdTransaction.Txid = htrdTxid
	rdTransaction.TxHex = htrdHex
	rdTransaction.SignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(rdTransaction)
	return rdTransaction, nil
}
