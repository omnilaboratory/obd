package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"log"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

type commitmentTxManager struct {
	operationFlag sync.Mutex
}

var CommitmentTxService commitmentTxManager

func (service *commitmentTxManager) CreateNewCommitmentTxRequest(jsonData string, creator *bean.User) (data *bean.CommitmentTx, targetUser *string, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, nil, errors.New("empty json data")
	}
	data = &bean.CommitmentTx{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, nil, err
	}
	if bean.ChannelIdService.IsEmpty(data.ChannelId) {
		return nil, nil, errors.New("wrong channel_id")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		return nil, nil, err
	}

	err = checkBtcFundFinish(channelInfo.ChannelAddress)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	lastCommitmentTxInfo := &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.TxInfoState_CreateAndSign), q.Eq("Owner", creator.PeerId)).OrderBy("CreateAt").Reverse().First(lastCommitmentTxInfo)
	if err != nil {
		return nil, nil, errors.New("not find the lastCommitmentTxInfo")
	}
	balance := lastCommitmentTxInfo.AmountToRSMC
	if balance < 0 {
		return nil, nil, errors.New("not enough balance")
	}

	if balance < data.Amount {
		return nil, nil, errors.New("not enough payment amount")
	}

	if tool.CheckIsString(&data.ChannelAddressPrivateKey) == false {
		return nil, nil, errors.New("wrong ChannelAddressPrivateKey")
	}

	if tool.CheckIsString(&data.LastTempAddressPrivateKey) == false {
		return nil, nil, errors.New("wrong LastTempAddressPrivateKey")
	}

	if _, err := getAddressFromPubKey(data.CurrTempAddressPubKey); err != nil {
		return nil, nil, errors.New("wrong CurrTempAddressPubKey")
	}

	if tool.CheckIsString(&data.CurrTempAddressPrivateKey) == false {
		return nil, nil, errors.New("wrong CurrTempAddressPrivateKey")
	}

	if data.Amount <= 0 {
		return nil, nil, errors.New("wrong payment amount")
	}

	isAliceCreateTransfer := true
	targetUser = &channelInfo.PeerIdB
	if creator.PeerId == channelInfo.PeerIdB {
		isAliceCreateTransfer = false
		targetUser = &channelInfo.PeerIdA
	}

	//store the privateKey of last temp addr
	// if alice transfer to bob, alice is the creator
	if isAliceCreateTransfer {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = data.ChannelAddressPrivateKey
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = data.ChannelAddressPrivateKey
	}

	tempAddrPrivateKeyMap[lastCommitmentTxInfo.RSMCTempAddressPubKey] = data.LastTempAddressPrivateKey
	tempAddrPrivateKeyMap[data.CurrTempAddressPubKey] = data.CurrTempAddressPrivateKey
	data.ChannelAddressPrivateKey = ""
	data.LastTempAddressPrivateKey = ""
	data.CurrTempAddressPrivateKey = ""

	data.RequestCommitmentHash = lastCommitmentTxInfo.CurrHash

	// store the request data for -352
	var tempInfo = &dao.CommitmentTxRequestInfo{}
	_ = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("UserId", creator.PeerId), q.Eq("IsEnable", true)).First(tempInfo)
	tempInfo.CommitmentTx = *data
	tempInfo.LastTempAddressPubKey = lastCommitmentTxInfo.RSMCTempAddressPubKey
	if tempInfo.Id == 0 {
		tempInfo.ChannelId = data.ChannelId
		tempInfo.UserId = creator.PeerId
		tempInfo.CreateAt = time.Now()
		tempInfo.IsEnable = true
		err = db.Save(tempInfo)
	} else {
		err = db.Update(tempInfo)
	}
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	return data, targetUser, err
}

type commitmentTxSignedManager struct {
	operationFlag sync.Mutex
}

var CommitmentTxSignedService commitmentTxSignedManager

func (service *commitmentTxSignedManager) CommitmentTxSign(jsonData string, signer *bean.User) (*dao.CommitmentTransaction, *dao.CommitmentTransaction, *string, error) {
	if tool.CheckIsString(&jsonData) == false {
		err := errors.New("empty json data")
		log.Println(err)
		return nil, nil, nil, err
	}

	data := &bean.CommitmentTxSigned{}
	err := json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, err
	}

	if tool.CheckIsString(&data.RequestCommitmentHash) == false {
		err = errors.New("wrong RequestCommitmentHash")
		log.Println(err)
		return nil, nil, nil, err
	}

	if bean.ChannelIdService.IsEmpty(data.ChannelId) {
		err = errors.New("wrong ChannelId")
		log.Println(err)
		return nil, nil, nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, err
	}

	err = checkBtcFundFinish(channelInfo.ChannelAddress)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, err
	}

	//Make sure who creates the transaction, who will sign the transaction.
	//The default creator is Alice, and Bob is the signer.
	//While if ALice is the signer, then Bob creates the transaction.
	targetUser := channelInfo.PeerIdA
	if signer.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	} else {
		targetUser = channelInfo.PeerIdB
	}

	if data.Approval == false {
		return nil, nil, &targetUser, errors.New("signer disagree transaction")
	}

	service.operationFlag.Lock()
	defer service.operationFlag.Unlock()

	var dataFromCreator = &dao.CommitmentTxRequestInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("UserId", targetUser), q.Eq("IsEnable", true)).OrderBy("CreateAt").Reverse().First(dataFromCreator)
	if err != nil {
		log.Println(err)
		return nil, nil, &targetUser, err
	}

	if dataFromCreator.RequestCommitmentHash != data.RequestCommitmentHash {
		err = errors.New("error RequestCommitmentHash")
		log.Println(err)
		return nil, nil, nil, err
	}

	var requestCommitmentTx = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrHash", data.RequestCommitmentHash), q.Eq("Owner", targetUser), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(requestCommitmentTx)
	if err != nil {
		err = errors.New("not found the requested commitment tx")
		log.Println(err)
		return nil, nil, nil, err
	}

	dealCount, err := db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("LastHash", data.RequestCommitmentHash), q.Eq("Owner", targetUser)).Count(&dao.CommitmentTransaction{})
	if err != nil {
		log.Println(err)
		return nil, nil, &targetUser, err
	}
	if dealCount > 0 {
		err = errors.New("the request commitment tx is invalid")
		log.Println(err)
		return nil, nil, nil, err
	}

	//for c rd br
	if tool.CheckIsString(&data.ChannelAddressPrivateKey) == false {
		err = errors.New("fail to get the signer's channel address private key")
		log.Println(err)
		return nil, nil, nil, err
	}
	if signer.PeerId == channelInfo.PeerIdB {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = data.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = data.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	}

	//for c rd
	if _, err := getAddressFromPubKey(data.CurrTempAddressPubKey); err != nil {
		err = errors.New("fail to get the signer's curr temp address pub key")
		log.Println(err)
		return nil, nil, nil, err
	}
	//for c rd
	if tool.CheckIsString(&data.CurrTempAddressPrivateKey) == false {
		err = errors.New("fail to get the signer's curr temp address private key")
		log.Println(err)
		return nil, nil, nil, err
	}

	//check the starter's private key
	// for c rd br
	creatorChannelAddressPrivateKey := ""
	if signer.PeerId == channelInfo.PeerIdB {
		creatorChannelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyA]
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	} else {
		creatorChannelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyB]
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	}
	if tool.CheckIsString(&creatorChannelAddressPrivateKey) == false {
		err = errors.New("fail to get the starer's channel private key")
		log.Println(err)
		return nil, nil, nil, err
	}

	// for c rd
	creatorCurrTempAddressPrivateKey := tempAddrPrivateKeyMap[dataFromCreator.CurrTempAddressPubKey]
	if tool.CheckIsString(&creatorCurrTempAddressPrivateKey) == false {
		err = errors.New("fail to get the starer's curr temp address private key")
		log.Println(err)
		return nil, nil, nil, err
	}
	defer delete(tempAddrPrivateKeyMap, requestCommitmentTx.RSMCMultiAddressScriptPubKey)

	//for br
	creatorLastTempAddressPrivateKey := tempAddrPrivateKeyMap[dataFromCreator.LastTempAddressPubKey]
	if tool.CheckIsString(&creatorLastTempAddressPrivateKey) == false {
		err = errors.New("fail to get the starter's last temp address  private key")
		log.Println(err)
		return nil, nil, nil, err
	}
	defer delete(tempAddrPrivateKeyMap, dataFromCreator.LastTempAddressPubKey)

	//launch database transaction, if anything goes wrong, roll back.
	tx, err := db.Begin(true)
	if err != nil {
		return nil, nil, nil, err
	}
	defer tx.Rollback()

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = tx.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.FundingTransactionState_Accept)).OrderBy("CreateAt").Reverse().First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, nil, &targetUser, err
	}
	commitmentATxInfo, err := createAliceSideTxs(tx, data, *dataFromCreator, channelInfo, fundingTransaction, signer)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, err
	}

	commitmentBTxInfo, err := createBobSideTxs(tx, data, *dataFromCreator, channelInfo, fundingTransaction, signer)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, err
	}

	dataFromCreator.IsEnable = false
	err = tx.Update(dataFromCreator)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, err
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, nil, nil, err
	}

	return commitmentATxInfo, commitmentBTxInfo, &targetUser, err
}

func createAliceSideTxs(tx storm.Node, signData *bean.CommitmentTxSigned, dataFromCreator dao.CommitmentTxRequestInfo, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, signer *bean.User) (*dao.CommitmentTransaction, error) {
	owner := channelInfo.PeerIdA

	var isAliceSendToBob = true
	if signer.PeerId == channelInfo.PeerIdA {
		isAliceSendToBob = false
	}

	var lastCommitmentTx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", signData.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastCommitmentTx)
	if err != nil {
		lastCommitmentTx = nil
	}

	if lastCommitmentTx != nil {

		count, _ := tx.Select(q.Eq("CommitmentTxId", lastCommitmentTx.Id)).Count(&dao.BreachRemedyTransaction{})
		if count > 0 {
			err = errors.New("already exist BreachRemedyTransaction ")
			return nil, err
		}

		// create BRa tx  for bob ，let the lastCommitmentTx abort,
		breachRemedyTransaction, err := createBRTx(channelInfo.PeerIdB, channelInfo, lastCommitmentTx, signer)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if breachRemedyTransaction.Amount > 0 {
			lastTempAddressPrivateKey := ""
			if isAliceSendToBob {
				lastTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCreator.LastTempAddressPubKey]
			} else {
				lastTempAddressPrivateKey = signData.LastTempAddressPrivateKey
			}
			if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
				err = errors.New("fail to get the lastTempAddressPrivateKey")
				log.Println(err)
				return nil, err
			}

			inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTx.RSMCTxHash, lastCommitmentTx.RSMCMultiAddress, lastCommitmentTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
				lastCommitmentTx.RSMCMultiAddress,
				[]string{
					lastTempAddressPrivateKey,
					tempAddrPrivateKeyMap[channelInfo.PubKeyB],
				},
				inputs,
				channelInfo.AddressB,
				fundingTransaction.FunderAddress,
				fundingTransaction.PropertyId,
				breachRemedyTransaction.Amount,
				0,
				0,
				&lastCommitmentTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			breachRemedyTransaction.Txid = txid
			breachRemedyTransaction.TransactionSignHex = hex
			breachRemedyTransaction.SignAt = time.Now()
			breachRemedyTransaction.CurrState = dao.TxInfoState_CreateAndSign
			err = tx.Save(breachRemedyTransaction)
		}

		lastRDTransaction := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(q.Eq("ChannelId", signData.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentTx.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		lastCommitmentTx.CurrState = dao.TxInfoState_Abord
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentTx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = tx.Update(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	// create Cna tx
	var outputBean = commitmentOutputBean{}
	if isAliceSendToBob {
		outputBean.RsmcTempPubKey = dataFromCreator.CurrTempAddressPubKey
		//default alice transfer to bob ,then alice minus money
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentTx != nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	} else {
		outputBean.RsmcTempPubKey = signData.CurrTempAddressPubKey
		// if bob transfer to alice,then alice add money
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentTx != nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	}
	outputBean.OppositeSideChannelAddress = channelInfo.AddressB
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB

	commitmentTxInfo, err := createCommitmentTx(owner, channelInfo, fundingTransaction, outputBean, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Rsmc

	usedTxidTemp := ""
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTx(
			channelInfo.ChannelAddress,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
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

	//create to Bob tx
	if commitmentTxInfo.AmountToOther > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTxToBob(
			channelInfo.ChannelAddress,
			usedTxidTemp,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			channelInfo.AddressB,
			fundingTransaction.FunderAddress,
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
	if lastCommitmentTx != nil {
		commitmentTxInfo.LastCommitmentTxId = lastCommitmentTx.Id
	}

	commitmentTxInfo.SignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	if lastCommitmentTx != nil {
		commitmentTxInfo.LastHash = lastCommitmentTx.CurrHash
	}
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

	// create RDna tx
	rdTransaction, err := createRDTx(owner, channelInfo, commitmentTxInfo, channelInfo.AddressA, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := ""
	if isAliceSendToBob {
		currTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCreator.CurrTempAddressPubKey]
	} else {
		currTempAddressPrivateKey = signData.CurrTempAddressPrivateKey
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.RSMCTxHash, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
		commitmentTxInfo.RSMCMultiAddress,
		[]string{
			currTempAddressPrivateKey,
			tempAddrPrivateKeyMap[channelInfo.PubKeyB],
		},
		inputs,
		rdTransaction.OutputAddress,
		fundingTransaction.FunderAddress,
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
		log.Println(err)
		return nil, err
	}
	return commitmentTxInfo, err
}

func createBobSideTxs(tx storm.Node, signData *bean.CommitmentTxSigned, dataFromCreator dao.CommitmentTxRequestInfo, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, signer *bean.User) (*dao.CommitmentTransaction, error) {
	owner := channelInfo.PeerIdB
	var isAliceSendToBob = true
	if signer.PeerId == channelInfo.PeerIdA {
		isAliceSendToBob = false
	}

	var lastCommitmentTx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", signData.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign), q.Or(q.Eq("PeerIdA", signer.PeerId), q.Eq("PeerIdB", signer.PeerId))).OrderBy("CreateAt").Reverse().First(lastCommitmentTx)
	if err != nil {
		lastCommitmentTx = nil
	}
	//In unilataral funding mode, only Alice is required to fund the channel.
	//So during funding procedure, on Bob side, he has no commitment transaction and revockable delivery transaction.
	if lastCommitmentTx != nil {

		count, _ := tx.Select(q.Eq("CommitmentTxId", lastCommitmentTx.Id)).Count(&dao.BreachRemedyTransaction{})
		if count > 0 {
			err = errors.New("already exist BreachRemedyTransaction ")
			return nil, err
		}

		// create BRb tx for alice
		breachRemedyTransaction, err := createBRTx(channelInfo.PeerIdA, channelInfo, lastCommitmentTx, signer)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if breachRemedyTransaction.Amount > 0 {
			lastTempAddressPrivateKey := ""
			if isAliceSendToBob {
				lastTempAddressPrivateKey = signData.LastTempAddressPrivateKey
			} else {
				lastTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCreator.LastTempAddressPubKey]
			}

			if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
				err = errors.New("fail to get the lastTempAddressPrivateKey")
				log.Println(err)
				return nil, err
			}

			inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTx.RSMCTxHash, lastCommitmentTx.RSMCMultiAddress, lastCommitmentTx.RSMCMultiAddressScriptPubKey)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
				lastCommitmentTx.RSMCMultiAddress,
				[]string{
					lastTempAddressPrivateKey,
					tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				},
				inputs,
				channelInfo.AddressA,
				fundingTransaction.FunderAddress,
				fundingTransaction.PropertyId,
				breachRemedyTransaction.Amount,
				0,
				0,
				&lastCommitmentTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			breachRemedyTransaction.Txid = txid
			breachRemedyTransaction.TransactionSignHex = hex
			breachRemedyTransaction.SignAt = time.Now()
			breachRemedyTransaction.CurrState = dao.TxInfoState_CreateAndSign
			err = tx.Save(breachRemedyTransaction)
		}

		lastRDTransaction := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(q.Eq("ChannelId", signData.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentTx.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		lastCommitmentTx.CurrState = dao.TxInfoState_Abord
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentTx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = tx.Update(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	// create Cnb tx
	var outputBean = commitmentOutputBean{}
	if isAliceSendToBob {
		outputBean.RsmcTempPubKey = signData.CurrTempAddressPubKey
		//by default, alice transfers money to bob,then bob's balance increases.
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentTx != nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	} else {
		outputBean.RsmcTempPubKey = dataFromCreator.CurrTempAddressPubKey
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentTx != nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	}
	outputBean.OppositeSideChannelAddress = channelInfo.AddressA
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA

	commitmentTxInfo, err := createCommitmentTx(owner, channelInfo, fundingTransaction, outputBean, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Rsmc

	usedTxidTemp := ""
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTx(
			channelInfo.ChannelAddress,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
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

	//create to alice tx
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTxToBob(
			channelInfo.ChannelAddress,
			usedTxidTemp,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			channelInfo.AddressA,
			fundingTransaction.FunderAddress,
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

	if lastCommitmentTx != nil {
		commitmentTxInfo.LastCommitmentTxId = lastCommitmentTx.Id
	}
	commitmentTxInfo.SignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	commitmentTxInfo.CurrHash = ""
	commitmentTxInfo.LastHash = ""
	if lastCommitmentTx != nil {
		commitmentTxInfo.LastHash = lastCommitmentTx.CurrHash
	}
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

	// create RDb tx
	rdTransaction, err := createRDTx(owner, channelInfo, commitmentTxInfo, channelInfo.AddressB, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := ""
	if isAliceSendToBob {
		currTempAddressPrivateKey = signData.CurrTempAddressPrivateKey
	} else {
		currTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCreator.CurrTempAddressPubKey]
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.RSMCTxHash, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
		commitmentTxInfo.RSMCMultiAddress,
		[]string{
			tempAddrPrivateKeyMap[channelInfo.PubKeyA],
			currTempAddressPrivateKey,
		},
		inputs,
		rdTransaction.OutputAddress,
		fundingTransaction.FunderAddress,
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
		log.Println(err)
		return nil, err
	}
	return commitmentTxInfo, err
}
