package service

import (
	"errors"
	"fmt"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/tidwall/gjson"
	"log"
	"time"
)

//创建BR
func createCurrCommitmentTxBR(tx storm.Node, brType dao.BRType, channelInfo *dao.ChannelInfo, commitmentTx *dao.CommitmentTransaction, inputs []rpc.TransactionInputItem,
	outputAddress, channelPrivateKey string, user bean.User) (err error) {
	if len(inputs) == 0 {
		return nil
	}
	breachRemedyTransaction := &dao.BreachRemedyTransaction{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("InputTxid", commitmentTx.RSMCTxid),
		q.Eq("Type", brType),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(breachRemedyTransaction)
	if breachRemedyTransaction.Id == 0 {
		breachRemedyTransaction, err = createBRTxObj(user.PeerId, channelInfo, brType, commitmentTx, &user)
		if err != nil {
			log.Println(err)
			return err
		}
		if breachRemedyTransaction.Amount > 0 {
			txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
				commitmentTx.RSMCMultiAddress,
				[]string{
					channelPrivateKey,
				},
				inputs,
				outputAddress,
				channelInfo.FundingAddress,
				channelInfo.PropertyId,
				breachRemedyTransaction.Amount,
				getBtcMinerAmount(channelInfo.BtcAmount),
				0,
				&commitmentTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return err
			}
			breachRemedyTransaction.OutAddress = outputAddress
			breachRemedyTransaction.Txid = txid
			breachRemedyTransaction.BrTxHex = hex
			breachRemedyTransaction.CurrState = dao.TxInfoState_Create
			_ = tx.Save(breachRemedyTransaction)
		}
	}
	return nil
}

//创建RawBR
func createCurrCommitmentTxRawBR(tx storm.Node, brType dao.BRType, channelInfo *dao.ChannelInfo,
	commitmentTx *dao.CommitmentTransaction, inputs []rpc.TransactionInputItem,
	outputAddress string, user bean.User) (retMap map[string]interface{}, err error) {
	if len(inputs) == 0 {
		return nil, nil
	}
	breachRemedyTransaction := &dao.BreachRemedyTransaction{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("InputTxid", commitmentTx.RSMCTxid),
		q.Eq("Type", brType),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(breachRemedyTransaction)
	if breachRemedyTransaction.Id == 0 {
		breachRemedyTransaction, err = createBRTxObj(user.PeerId, channelInfo, brType, commitmentTx, &user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if breachRemedyTransaction.Amount > 0 {
			retMap, err = rpcClient.OmniCreateRawTransactionUseUnsendInput(
				commitmentTx.RSMCMultiAddress,
				inputs,
				outputAddress,
				channelInfo.FundingAddress,
				channelInfo.PropertyId,
				breachRemedyTransaction.Amount,
				getBtcMinerAmount(channelInfo.BtcAmount),
				0,
				&commitmentTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			breachRemedyTransaction.OutAddress = outputAddress
			breachRemedyTransaction.BrTxHex = retMap["hex"].(string)
			breachRemedyTransaction.CurrState = dao.TxInfoState_Create
			_ = tx.Save(breachRemedyTransaction)
			retMap["br_id"] = breachRemedyTransaction.Id
		}
	} else {
		retMap, err = rpcClient.OmniCreateRawTransactionUseUnsendInput(
			commitmentTx.RSMCMultiAddress,
			inputs,
			outputAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			breachRemedyTransaction.Amount,
			getBtcMinerAmount(channelInfo.BtcAmount),
			0,
			&commitmentTx.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		retMap["br_id"] = breachRemedyTransaction.Id
	}

	return retMap, nil
}

// 第一次签名完成，更新RawBR
func updateCurrCommitmentTxRawBR(tx storm.Node, id int64, firstSignedBrHex string, user bean.User) (err error) {
	breachRemedyTransaction := &dao.BreachRemedyTransaction{}
	_ = tx.Select(
		q.Eq("Id", id),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(breachRemedyTransaction)
	if breachRemedyTransaction.Id > 0 {
		if len(breachRemedyTransaction.Txid) == 0 {
			omniDecode, err := rpcClient.OmniDecodeTransaction(firstSignedBrHex)
			if err != nil {
				return nil
			}
			brTxId := gjson.Get(omniDecode, "txid").Str

			breachRemedyTransaction.Txid = brTxId
			breachRemedyTransaction.BrTxHex = firstSignedBrHex
			_ = tx.Update(breachRemedyTransaction)

		}
	}
	return errors.New("not found br by id")
}

//对上一个承诺交易的br进行签名
func signLastBR(tx storm.Node, brType dao.BRType, channelInfo dao.ChannelInfo, userPeerId string, lastTempAddressPrivateKey string, lastCommitmentTxid int) (err error) {
	lastBreachRemedyTransaction := &dao.BreachRemedyTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", lastCommitmentTxid),
		q.Eq("Type", brType),
		q.Or(
			q.Eq("PeerIdA", userPeerId),
			q.Eq("PeerIdB", userPeerId))).
		OrderBy("CreateAt").
		Reverse().First(lastBreachRemedyTransaction)
	if lastBreachRemedyTransaction != nil && lastBreachRemedyTransaction.Id > 0 {
		inputs, err := getInputsForNextTxByParseTxHashVout(
			lastBreachRemedyTransaction.InputTxHex,
			lastBreachRemedyTransaction.InputAddress,
			lastBreachRemedyTransaction.InputAddressScriptPubKey,
			lastBreachRemedyTransaction.InputRedeemScript)
		if err != nil {
			log.Println(err)
			return errors.New(fmt.Sprintf(enum.Tips_rsmc_failToGetInput, "breachRemedyTransaction"))
		}
		signedBRTxid, signedBRHex, err := rpcClient.OmniSignRawTransactionForUnsend(lastBreachRemedyTransaction.BrTxHex, inputs, lastTempAddressPrivateKey)
		if err != nil {
			return errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "breachRemedyTransaction"))
		}
		result, err := rpcClient.TestMemPoolAccept(signedBRHex)
		if err != nil {
			return errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "breachRemedyTransaction"))
		}
		if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
			if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
				return errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
			}
		}
		lastBreachRemedyTransaction.Txid = signedBRTxid
		lastBreachRemedyTransaction.BrTxHex = signedBRHex
		lastBreachRemedyTransaction.SignAt = time.Now()
		lastBreachRemedyTransaction.CurrState = dao.TxInfoState_CreateAndSign
		return tx.Update(lastBreachRemedyTransaction)
	}
	return nil
}

func checkBobRemcData(rsmcHex string, commitmentTransaction *dao.CommitmentTransaction) error {
	result, err := rpcClient.OmniDecodeTransaction(rsmcHex)
	if err != nil {
		return errors.New("error rsmcHex")
	}
	parse := gjson.Parse(result)
	if parse.Get("propertyid").Int() != commitmentTransaction.PropertyId {
		return errors.New("error propertyId in rsmcHex ")
	}
	if parse.Get("amount").Float() != commitmentTransaction.AmountToCounterparty {
		return errors.New("error amount in rsmcHex ")
	}
	return nil
}
