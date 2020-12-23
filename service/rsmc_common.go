package service

import (
	"errors"
	"fmt"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"time"
)

//创建RawBR
func createCurrCommitmentTxRawBR(tx storm.Node, brType dao.BRType, channelInfo *dao.ChannelInfo,
	commitmentTx *dao.CommitmentTransaction, inputs []bean.TransactionInputItem,
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
			retMap, err = omnicore.OmniCreateRawTransactionUseUnsendInput(
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
		retMap, err = omnicore.OmniCreateRawTransactionUseUnsendInput(
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

func createCurrCommitmentTxPartialSignedBR(tx storm.Node, brType dao.BRType, channelInfo *dao.ChannelInfo,
	commitmentTx *dao.CommitmentTransaction, inputs []bean.TransactionInputItem,
	outputAddress string, brHex string, user bean.User) (err error) {
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
			breachRemedyTransaction.OutAddress = outputAddress
			breachRemedyTransaction.BrTxHex = brHex
			breachRemedyTransaction.Txid = omnicore.GetTxId(brHex)
			breachRemedyTransaction.CurrState = dao.TxInfoState_Create
			_ = tx.Save(breachRemedyTransaction)
		}
	}
	return nil
}

//创建RawBR obj
func createRawBR(brType dao.BRType, channelInfo *dao.ChannelInfo, commitmentTx *dao.CommitmentTransaction, inputs []bean.TransactionInputItem,
	outputAddress string, user bean.User) (retMap bean.NeedClientSignTxData, err error) {
	if len(inputs) == 0 {
		return retMap, errors.New("empty inputs")
	}

	breachRemedyTransaction, err := createBRTxObj(user.PeerId, channelInfo, brType, commitmentTx, &user)
	if err != nil {
		log.Println(err)
		return retMap, err
	}
	if breachRemedyTransaction.Amount > 0 {
		brTxData, err := omnicore.OmniCreateRawTransactionUseUnsendInput(
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
			return retMap, err
		}
		c2bBrRawData := bean.NeedClientSignTxData{}
		c2bBrRawData.Hex = brTxData["hex"].(string)
		c2bBrRawData.Inputs = brTxData["inputs"]
		c2bBrRawData.IsMultisig = true
		return c2bBrRawData, nil
	}

	return retMap, errors.New("fail to createRawBR brType " + strconv.Itoa(int(brType)))
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
			if pass, txid := compareBR(breachRemedyTransaction.BrTxHex, firstSignedBrHex); pass == true {
				breachRemedyTransaction.Txid = txid
				breachRemedyTransaction.BrTxHex = firstSignedBrHex
				_ = tx.Update(breachRemedyTransaction)
				return nil
			}
		}
	}
	return errors.New("not found br by id")
}

func compareBR(hex1 string, hex2 string) (bool, string) {

	hex1Decode, err := omnicore.DecodeBtcRawTransaction(hex1)
	if err != nil {
		return false, ""
	}
	hex2Decode, err := omnicore.DecodeBtcRawTransaction(hex2)
	if err != nil {
		return false, ""
	}
	hex1Vin := gjson.Get(hex1Decode, "vin").Array()
	hex2Vin := gjson.Get(hex2Decode, "vin").Array()
	if len(hex1Vin) != len(hex2Vin) {
		return false, ""
	}
	for i := 0; i < len(hex1Vin); i++ {
		vin1txId := hex1Vin[i].Get("txid").Str
		vin2txId := hex2Vin[i].Get("txid").Str
		if vin1txId != vin2txId {
			return false, ""
		}
	}
	return true, gjson.Get(hex2Decode, "txid").Str
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
	if lastBreachRemedyTransaction.Id > 0 {
		inputs, err := getInputsForNextTxByParseTxHashVout(
			lastBreachRemedyTransaction.InputTxHex,
			lastBreachRemedyTransaction.InputAddress,
			lastBreachRemedyTransaction.InputAddressScriptPubKey,
			lastBreachRemedyTransaction.InputRedeemScript)
		if err != nil {
			log.Println(err)
			return errors.New(fmt.Sprintf(enum.Tips_rsmc_failToGetInput, "breachRemedyTransaction"))
		}

		if lastBreachRemedyTransaction.CurrState == dao.TxInfoState_CreateAndSign {
			return nil
		}

		signedBRTxid, signedBRHex, err := omnicore.OmniSignRawTransactionForUnsend(lastBreachRemedyTransaction.BrTxHex, inputs, lastTempAddressPrivateKey)
		if err != nil {
			return errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "breachRemedyTransaction"))
		}
		lastBreachRemedyTransaction.Txid = signedBRTxid
		lastBreachRemedyTransaction.BrTxHex = signedBRHex
		lastBreachRemedyTransaction.SignAt = time.Now()
		lastBreachRemedyTransaction.CurrState = dao.TxInfoState_CreateAndSign
		return tx.Update(lastBreachRemedyTransaction)
	}
	return nil
}

func checkBobRsmcData(rsmcHex, toAddress string, commitmentTransaction *dao.CommitmentTransaction) error {
	_, err := omnicore.VerifyOmniTxHex(rsmcHex, commitmentTransaction.PropertyId, commitmentTransaction.AmountToCounterparty, toAddress, true)
	if err != nil {
		return err
	}
	return nil
}
