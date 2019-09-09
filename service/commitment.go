package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"log"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
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

	if tool.CheckIsString(&data.ChannelAddressPrivateKey) == false {
		return nil, nil, errors.New("wrong ChannelAddressPrivateKey")
	}

	if data.Amount <= 0 {
		return nil, nil, errors.New("wrong payment amount")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		return nil, nil, err
	}

	isAliceCreateTransfer := true
	targetUser = &channelInfo.PeerIdB
	if creator.PeerId == channelInfo.PeerIdB {
		isAliceCreateTransfer = false
		targetUser = &channelInfo.PeerIdA
	}

	lastCommitmentTxInfo := &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Eq("Owner", creator.PeerId)).OrderBy("CreateAt").Reverse().First(lastCommitmentTxInfo)
	if err != nil {
		return nil, nil, errors.New("not find the lastCommitmentTxInfo")
	}
	balance := lastCommitmentTxInfo.AmountM
	if balance < 0 {
		return nil, nil, errors.New("not enough balance")
	}

	if balance < data.Amount {
		return nil, nil, errors.New("not enough payment amount")
	}

	if tool.CheckIsString(&data.LastTempPrivateKey) == false {
		return nil, nil, errors.New("wrong LastTempPrivateKey")
	}

	// if alice transfer to bob, alice is the creator
	if isAliceCreateTransfer {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = data.ChannelAddressPrivateKey
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = data.ChannelAddressPrivateKey
	}
	//store the privateKey of last temp addr
	tempAddrPrivateKeyMap[lastCommitmentTxInfo.TempPubKey] = data.LastTempPrivateKey
	data.LastTempPrivateKey = ""
	data.ChannelAddressPrivateKey = ""
	data.RequestCommitmentHash = lastCommitmentTxInfo.CurrHash

	// store the request data for -354
	var tempInfo = &dao.CommitmentTxRequestInfo{}
	_ = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("UserId", creator.PeerId), q.Eq("IsEnable", true)).First(tempInfo)
	tempInfo.CommitmentTx = *data
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

func (service *commitmentTxManager) GetLatestCommitmentTxByChannelId(jsonData string, user *bean.User) (node *dao.CommitmentTransaction, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty jsonData")
	}
	var reqData = &bean.GetBalanceRequest{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return nil, err
	}

	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		return nil, errors.New("wrong channelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("Owner", user.PeerId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}

func (service *commitmentTxManager) GetLatestRDTxByChannelId(jsonData string, user *bean.User) (node *dao.RevocableDeliveryTransaction, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("Owner", user.PeerId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}
func (service *commitmentTxManager) GetLatestBRTxByChannelId(jsonData string, user *bean.User) (node *dao.BreachRemedyTransaction, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.BreachRemedyTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}

func (service *commitmentTxManager) GetItemsByChannelId(jsonData string, user *bean.User) (nodes []dao.CommitmentTransaction, count *int, err error) {
	var chanId bean.ChannelID

	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		return nil, nil, err
	}

	pageIndex := gjson.Get(jsonData, "page_index").Int()
	if pageIndex <= 0 {
		pageIndex = 1
	}
	pageSize := gjson.Get(jsonData, "page_size").Int()
	if pageSize <= 0 {
		pageSize = 10
	}
	skip := (pageIndex - 1) * pageSize

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, nil, err
	}

	nodes = []dao.CommitmentTransaction{}
	tempCount, err := db.Select(q.Eq("ChannelId", chanId)).Count(&dao.CommitmentTransaction{})
	if err != nil {
		return nil, nil, err
	}
	count = &tempCount
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().Skip(int(skip)).Limit(int(pageSize)).Find(&nodes)
	return nodes, count, err
}

func (service *commitmentTxManager) GetItemById(id int) (node *dao.CommitmentTransaction, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitmentTxManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.CommitmentTransaction{})
}

func (service *commitmentTxManager) Del(id int) (node *dao.CommitmentTransaction, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}

	node = &dao.CommitmentTransaction{}
	err = db.One("Id", id, node)
	if err != nil {
		return nil, err
	}
	err = db.DeleteStruct(node)
	return node, err
}

type commitmentTxSignedManager struct {
	operationFlag sync.Mutex
}

var CommitmentTxSignedService commitmentTxSignedManager

func (service *commitmentTxSignedManager) CommitmentTxSign(jsonData string, signer *bean.User) (*dao.CommitmentTransaction, *dao.CommitmentTransaction, *string, error) {
	if tool.CheckIsString(&jsonData) == false {
		log.Println("empty json data")
		return nil, nil, nil, errors.New("empty json data")
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

	//确定是谁发起的转账发起方 谁是签名收款方 默认是alice发起转账，bob是签收方，如果签收方是alice 那么就是bob发起的转账请求
	//Make sure who creates the transaction, who will sign the transaction.
	//The default creator is Alice, and Bob is the signer.
	//While if ALice is the signer, then Bob creates the transaction.
	targetUser := channelInfo.PeerIdA
	if signer.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	} else {
		targetUser = channelInfo.PeerIdB
	}

	if data.Attitude == false {
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

	var requestCommitmentATx = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrHash", data.RequestCommitmentHash), q.Eq("Owner", targetUser), q.Eq("CurrState", dao.TxInfoState_OtherSign)).OrderBy("CreateAt").Reverse().First(requestCommitmentATx)
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

	// if alice transfer to bob,bob is the signer
	if signer.PeerId == channelInfo.PeerIdB {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = data.ChannelAddressPrivateKey // data.ChannelAddressPrivateKey from signer
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = data.ChannelAddressPrivateKey
	}

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

	var isAliceCreateTransfer = true
	if signer.PeerId == channelInfo.PeerIdA {
		isAliceCreateTransfer = false
	}

	var lastCommitmentATx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", signData.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_OtherSign)).OrderBy("CreateAt").Reverse().First(lastCommitmentATx)
	if err != nil {
		lastCommitmentATx = nil
	}

	if lastCommitmentATx != nil {
		// create BRa tx  for bob ，let the lastCommitmentTx abort,
		breachRemedyTransaction, err := createBRTx(channelInfo.PeerIdB, channelInfo, lastCommitmentATx, signer)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		var brPrivKeys = make([]string, 0)
		var brPrivKey = tempAddrPrivateKeyMap[lastCommitmentATx.TempPubKey]
		if tool.CheckIsString(&brPrivKey) {
			brPrivKeys = append(brPrivKeys, brPrivKey)
		}
		delete(tempAddrPrivateKeyMap, lastCommitmentATx.TempPubKey)

		txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
			lastCommitmentATx.MultiAddress,
			brPrivKeys,
			[]rpc.TransactionInputItem{
				{breachRemedyTransaction.InputTxid, breachRemedyTransaction.InputVout, breachRemedyTransaction.InputAmount},
			},
			[]rpc.TransactionOutputItem{
				{channelInfo.AddressB, breachRemedyTransaction.Amount},
			},
			0,
			0)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		breachRemedyTransaction.Txid = txid
		breachRemedyTransaction.TxHexFirstSign = hex
		breachRemedyTransaction.FirstSignAt = time.Now()
		breachRemedyTransaction.CurrState = dao.TxInfoState_OtherSign
		err = tx.Save(breachRemedyTransaction)

		lastRDTransaction := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(q.Eq("ChannelId", signData.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentATx.Id), q.Eq("CurrState", dao.TxInfoState_OtherSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		lastCommitmentATx.CurrState = dao.TxInfoState_Abord
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentATx)
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
	if isAliceCreateTransfer {
		outputBean.TempPubKey = dataFromCreator.CurrTempPubKey
		//default alice transfer to bob ,then alice minus money
		outputBean.AmountM, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountB, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentATx != nil {
			outputBean.AmountM, _ = decimal.NewFromFloat(lastCommitmentATx.AmountM).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountB, _ = decimal.NewFromFloat(lastCommitmentATx.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	} else {
		outputBean.TempPubKey = signData.CurrTempPubKey
		// if bob transfer to alice,then alice add money
		outputBean.AmountM, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountB, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentATx != nil {
			outputBean.AmountM, _ = decimal.NewFromFloat(lastCommitmentATx.AmountM).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountB, _ = decimal.NewFromFloat(lastCommitmentATx.AmountB).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	}
	outputBean.ToAddress = channelInfo.AddressB
	outputBean.ToPubKey = channelInfo.PubKeyB

	commitmentTxInfo, err := createCommitmentTx(owner, channelInfo, fundingTransaction, outputBean, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var privkeys = make([]string, 0)
	var privKey = tempAddrPrivateKeyMap[channelInfo.PubKeyB] //bob sign first
	if tool.CheckIsString(&privKey) {
		privkeys = append(privkeys, privKey)
	}
	delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)

	txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
		channelInfo.ChannelAddress,
		privkeys,
		[]rpc.TransactionInputItem{
			{commitmentTxInfo.InputTxid, commitmentTxInfo.InputVout, commitmentTxInfo.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{commitmentTxInfo.MultiAddress, commitmentTxInfo.AmountM},
			{outputBean.ToAddress, commitmentTxInfo.AmountB},
		},
		0,
		0)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.LastCommitmentTxId = lastCommitmentATx.Id
	commitmentTxInfo.Txid = txid
	commitmentTxInfo.TxHexFirstSign = hex
	commitmentTxInfo.FirstSignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_OtherSign
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	if lastCommitmentATx != nil {
		commitmentTxInfo.LastHash = lastCommitmentATx.CurrHash
	}
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsg(bytes)
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

	txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
		commitmentTxInfo.MultiAddress,
		privkeys,
		[]rpc.TransactionInputItem{
			{rdTransaction.InputTxid, rdTransaction.InputVout, rdTransaction.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{rdTransaction.OutputAddress, rdTransaction.Amount},
		},
		0,
		rdTransaction.Sequence)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rdTransaction.Txid = txid
	rdTransaction.TxHexFirstSign = hex
	rdTransaction.FirstSignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_OtherSign
	err = tx.Save(rdTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return commitmentTxInfo, err
}

func createBobSideTxs(tx storm.Node, data *bean.CommitmentTxSigned, dataFromCreator dao.CommitmentTxRequestInfo, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, signer *bean.User) (*dao.CommitmentTransaction, error) {
	owner := channelInfo.PeerIdB
	var isAliceCreateTransfer = true
	if signer.PeerId == channelInfo.PeerIdA {
		isAliceCreateTransfer = false
	}

	var lastCommitmentBTx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Or(q.Eq("PeerIdA", signer.PeerId), q.Eq("PeerIdB", signer.PeerId))).OrderBy("CreateAt").Reverse().First(lastCommitmentBTx)
	if err != nil {
		lastCommitmentBTx = nil
	}

	//In unilataral funding mode, only Alice is required to fund the channel.
	//So during funding procedure, on Bob side, he has no commitment transaction and revockable delivery transaction.
	if lastCommitmentBTx != nil {

		// create BRb tx for alice
		breachRemedyTransaction, err := createBRTx(channelInfo.PeerIdA, channelInfo, lastCommitmentBTx, signer)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		var brPrivKeys = make([]string, 0)
		if tool.CheckIsString(&data.LastTempPrivateKey) {
			brPrivKeys = append(brPrivKeys, data.LastTempPrivateKey)
		}

		txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
			lastCommitmentBTx.MultiAddress,
			brPrivKeys,
			[]rpc.TransactionInputItem{
				{breachRemedyTransaction.InputTxid, breachRemedyTransaction.InputVout, breachRemedyTransaction.InputAmount},
			},
			[]rpc.TransactionOutputItem{
				{channelInfo.AddressB, breachRemedyTransaction.Amount},
			},
			0,
			0)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		breachRemedyTransaction.Txid = txid
		breachRemedyTransaction.TxHexFirstSign = hex
		breachRemedyTransaction.FirstSignAt = time.Now()
		breachRemedyTransaction.CurrState = dao.TxInfoState_OtherSign
		err = tx.Save(breachRemedyTransaction)

		lastRDTransaction := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentBTx.Id), q.Eq("CurrState", dao.TxInfoState_OtherSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		lastCommitmentBTx.CurrState = dao.TxInfoState_Abord
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentBTx)
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
	if isAliceCreateTransfer {
		outputBean.TempPubKey = data.CurrTempPubKey
		//by default, alice transters money to bob,then bob's balance increases.
		outputBean.AmountM, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountB, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentBTx != nil {
			outputBean.AmountM, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountM).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountB, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountB).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	} else {
		outputBean.TempPubKey = dataFromCreator.CurrTempPubKey
		outputBean.AmountM, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountB, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentBTx != nil {
			outputBean.AmountM, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountM).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountB, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	}
	outputBean.ToAddress = channelInfo.AddressA
	outputBean.ToPubKey = channelInfo.PubKeyA

	commitmentTxInfo, err := createCommitmentTx(owner, channelInfo, fundingTransaction, outputBean, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var privkeys = make([]string, 0)
	var privKey = tempAddrPrivateKeyMap[channelInfo.PubKeyA]
	if tool.CheckIsString(&privKey) {
		privkeys = append(privkeys, privKey)
	}
	delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)

	txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
		channelInfo.ChannelAddress,
		privkeys,
		[]rpc.TransactionInputItem{
			{commitmentTxInfo.InputTxid, commitmentTxInfo.InputVout, commitmentTxInfo.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{commitmentTxInfo.MultiAddress, commitmentTxInfo.AmountM},
			{outputBean.ToAddress, commitmentTxInfo.AmountB},
		},
		0,
		0)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if lastCommitmentBTx != nil {
		commitmentTxInfo.LastCommitmentTxId = lastCommitmentBTx.Id
	}
	commitmentTxInfo.Txid = txid
	commitmentTxInfo.TxHexFirstSign = hex
	commitmentTxInfo.FirstSignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_OtherSign
	commitmentTxInfo.CurrHash = ""
	commitmentTxInfo.LastHash = ""
	if lastCommitmentBTx != nil {
		commitmentTxInfo.LastHash = lastCommitmentBTx.CurrHash
	}
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsg(bytes)
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

	txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionForUnsendInputTx(
		commitmentTxInfo.MultiAddress,
		privkeys,
		[]rpc.TransactionInputItem{
			{rdTransaction.InputTxid, rdTransaction.InputVout, rdTransaction.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{rdTransaction.OutputAddress, rdTransaction.Amount},
		},
		0,
		rdTransaction.Sequence)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rdTransaction.Txid = txid
	rdTransaction.TxHexFirstSign = hex
	rdTransaction.FirstSignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_OtherSign
	err = tx.Save(rdTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return commitmentTxInfo, err
}

func (service *commitmentTxSignedManager) GetItemsByChannelId(jsonData string) (nodes []dao.CommitmentTransaction, count *int, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()

	if len(array) != 32 {
		return nil, nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	pageIndex := gjson.Get(jsonData, "page_index").Int()
	if pageIndex <= 0 {
		pageIndex = 1
	}
	pageSize := gjson.Get(jsonData, "page_size").Int()
	if pageSize <= 0 {
		pageSize = 10
	}
	skip := (pageIndex - 1) * pageSize

	db, err := dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	nodes = []dao.CommitmentTransaction{}
	tempCount, err := db.Select(q.Eq("ChannelId", chanId)).Count(&dao.CommitmentTransaction{})
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	count = &tempCount
	err = db.Select(q.Eq("ChannelId", chanId)).Skip(int(skip)).Limit(int(pageSize)).Find(&nodes)
	return nodes, count, err
}

func (service *commitmentTxSignedManager) GetItemById(id int) (node *dao.CommitmentTransaction, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	node = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitmentTxSignedManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return db.Count(&dao.CommitmentTransaction{})
}
