package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"log"
	"time"
)

type commitmentTxManager struct{}

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
	if len(data.ChannelId) != 32 {
		return nil, nil, errors.New("wrong channel_id")
	}

	if data.Amount <= 0 {
		return nil, nil, errors.New("wrong payment amount")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId)).First(channelInfo)
	if err != nil {
		return nil, nil, err
	}

	var isAliceCreateTransfer = true
	if creator.PeerId == channelInfo.PeerIdB {
		isAliceCreateTransfer = false
	}

	var createSide = 0
	if isAliceCreateTransfer == false {
		createSide = 1
	}

	commitmentTxInfo := &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Eq("CreateSide", createSide), q.Or(q.Eq("PeerIdA", creator.PeerId), q.Eq("PeerIdB", creator.PeerId))).OrderBy("CreateAt").Reverse().First(commitmentTxInfo)
	if err != nil {
		return nil, nil, errors.New("not find the commitmentTxInfo")
	}
	balance := commitmentTxInfo.AmountM
	if balance < 0 {
		return nil, nil, errors.New("not enough balance")
	}

	targetUser = &channelInfo.PeerIdB
	if isAliceCreateTransfer == false {
		targetUser = &channelInfo.PeerIdA
	}

	if balance < data.Amount {
		return nil, nil, errors.New("not enough payment amount")
	}

	// if alice transfer to bob, alice is creator
	if isAliceCreateTransfer {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = data.ChannelAddressPrivateKey
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = data.ChannelAddressPrivateKey
	}
	//store the privateKey of last temp addr
	tempAddrPrivateKeyMap[commitmentTxInfo.PubKey2] = data.LastTempPrivateKey
	data.LastTempPrivateKey = ""

	return data, targetUser, err
}
func (service *commitmentTxManager) GetLatestCommitmentTxByChannelId(jsonData string, user *bean.User) (node *dao.CommitmentTxInfo, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}
	node = &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(node)
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
	node = &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(node)
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
	node = &dao.BreachRemedyTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}

func (service *commitmentTxManager) GetItemsByChannelId(jsonData string) (nodes []dao.CommitmentTxInfo, count *int, err error) {
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
		return nil, nil, err
	}

	nodes = []dao.CommitmentTxInfo{}
	tempCount, err := db.Select(q.Eq("ChannelId", chanId)).Count(&dao.CommitmentTxInfo{})
	if err != nil {
		return nil, nil, err
	}
	count = &tempCount
	err = db.Select(q.Eq("ChannelId", chanId)).OrderBy("CreateAt").Reverse().Skip(int(skip)).Limit(int(pageSize)).Find(&nodes)
	return nodes, count, err
}

func (service *commitmentTxManager) GetItemById(id int) (node *dao.CommitmentTxInfo, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitmentTxManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.CommitmentTxInfo{})
}

func (service *commitmentTxManager) Del(id int) (node *dao.CommitmentTxInfo, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}

	node = &dao.CommitmentTxInfo{}
	err = db.One("Id", id, node)
	if err != nil {
		return nil, err
	}
	err = db.DeleteStruct(node)
	return node, err
}

type commitTxSignedManager struct{}

var CommitTxSignedService commitTxSignedManager

func (service *commitTxSignedManager) CommitmentTxSign(jsonData string, signer *bean.User) (*dao.CommitmentTxInfo, *dao.CommitmentTxInfo, *string, error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, nil, nil, errors.New("empty json data")
	}

	data := &bean.CommitmentTxSigned{}
	err := json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(data.ChannelId) != 32 {
		return nil, nil, nil, errors.New("wrong ChannelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId)).First(channelInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	//确定是谁发起的转账发起方 谁是签名收款方 默认是alice发起转账，bob是签收方，如果签收方是alice 那么就是bob发起的转账请求
	var isAliceCreateTransfer = true
	if signer.PeerId == channelInfo.PeerIdA {
		isAliceCreateTransfer = false
	}

	creatorSide := 0
	targetUser := channelInfo.PeerIdA
	if isAliceCreateTransfer == false {
		creatorSide = 1
		targetUser = channelInfo.PeerIdB
	}
	log.Println(creatorSide)

	if data.Attitude == false {
		return nil, nil, &targetUser, errors.New("signer disagree transaction")
	}

	// if alice transfer to bob,bob is signer
	if isAliceCreateTransfer {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = data.ChannelAddressPrivateKey // data.ChannelAddressPrivateKey from signer
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = data.ChannelAddressPrivateKey
	}

	//启动事务
	tx, _ := db.Begin(true)
	defer tx.Rollback()

	// 得到充值交易
	var fundingTransaction = &dao.FundingTransaction{}
	err = tx.One("ChannelId", data.ChannelId, fundingTransaction)
	if err != nil {
		return nil, nil, &targetUser, err
	}
	//获取alice 处于10的  c
	commitmentATxInfo, err := createAliceSideTxs(tx, data, channelInfo, fundingTransaction, signer)
	if err != nil {
		return nil, nil, nil, err
	}
	commitmentBTxInfo, err := createBobSideTxs(tx, data, channelInfo, fundingTransaction, signer)
	if err != nil {
		return nil, nil, nil, err
	}
	tx.Commit()

	return commitmentATxInfo, commitmentBTxInfo, &targetUser, err
}

func createAliceSideTxs(tx storm.Node, data *bean.CommitmentTxSigned, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, signer *bean.User) (*dao.CommitmentTxInfo, error) {
	var creatorSide = 0
	var isAliceCreateTransfer = true
	if signer.PeerId == channelInfo.PeerIdA {
		isAliceCreateTransfer = false
	}

	var lastCommitmentATx = &dao.CommitmentTxInfo{}
	err := tx.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CreateSide", creatorSide), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Or(q.Eq("PeerIdA", signer.PeerId), q.Eq("PeerIdB", signer.PeerId))).OrderBy("CreateAt").Reverse().First(lastCommitmentATx)
	if err != nil {
		return nil, err
	}

	// create BR1a tx
	breachRemedyTransaction, err := createBRTx(creatorSide, channelInfo, lastCommitmentATx, signer)
	if err != nil {
		return nil, err
	}

	var brPrivKeys = make([]string, 0)
	var brPrivKey = tempAddrPrivateKeyMap[lastCommitmentATx.PubKey2]
	if tool.CheckIsString(&brPrivKey) {
		brPrivKeys = append(brPrivKeys, brPrivKey)
	}
	delete(tempAddrPrivateKeyMap, lastCommitmentATx.PubKey2)

	txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
		lastCommitmentATx.MultiAddress,
		brPrivKeys,
		[]rpc.TransactionInputItem{
			{breachRemedyTransaction.InputTxid, breachRemedyTransaction.InputVout, breachRemedyTransaction.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{breachRemedyTransaction.PubKeyB, breachRemedyTransaction.Amount},
		},
		0,
		0)
	if err != nil {
		return nil, err
	}
	breachRemedyTransaction.Txid = txid
	breachRemedyTransaction.TxHexFirstSign = hex
	breachRemedyTransaction.FirstSignAt = time.Now()
	breachRemedyTransaction.CurrState = dao.TxInfoState_OtherSign
	err = tx.Save(breachRemedyTransaction)

	lastRDTransaction := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CreateSide", creatorSide), q.Eq("CommitmentTxId", lastCommitmentATx.Id), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Or(q.Eq("PeerIdA", signer.PeerId), q.Eq("PeerIdB", signer.PeerId))).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
	if err != nil {
		return nil, err
	}

	lastCommitmentATx.CurrState = dao.TxInfoState_Abord
	lastRDTransaction.CurrState = dao.TxInfoState_Abord
	tx.Update(lastCommitmentATx)
	tx.Update(lastRDTransaction)

	var outputBean = commitmentOutputBean{}
	if isAliceCreateTransfer {
		outputBean.TempAddress = data.CurrTempPubKeyFromStarter
		//default alice transfer to bob ,then alice minus money
		outputBean.AmountM = lastCommitmentATx.AmountM - data.Amount
		outputBean.AmountB = lastCommitmentATx.AmountB + data.Amount
	} else {
		outputBean.TempAddress = data.CurrTempPubKey
		// if bob transfer to alice,then alice add money
		outputBean.AmountM = lastCommitmentATx.AmountM + data.Amount
		outputBean.AmountB = lastCommitmentATx.AmountB - data.Amount
	}
	outputBean.ToAddressB = channelInfo.PubKeyB

	// create C2a tx
	commitmentTxInfo, err := createCommitmentATx(creatorSide, channelInfo, fundingTransaction, outputBean, signer)
	if err != nil {
		return nil, err
	}

	var privkeys = make([]string, 0)
	var privKey = tempAddrPrivateKeyMap[channelInfo.PubKeyB]
	if tool.CheckIsString(&privKey) {
		privkeys = append(privkeys, privKey)
	}
	delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)

	txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
		channelInfo.ChannelPubKey,
		privkeys,
		[]rpc.TransactionInputItem{
			{commitmentTxInfo.InputTxid, commitmentTxInfo.InputVout, commitmentTxInfo.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{commitmentTxInfo.MultiAddress, commitmentTxInfo.AmountM},
			{commitmentTxInfo.PubKeyB, commitmentTxInfo.AmountB},
		},
		0,
		0)
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.LastCommitmentTxId = lastCommitmentATx.Id
	commitmentTxInfo.Txid = txid
	commitmentTxInfo.TxHexFirstSign = hex
	commitmentTxInfo.FirstSignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_OtherSign
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		return nil, err
	}

	// create RD2a tx
	rdTransaction, err := createRDaTx(creatorSide, channelInfo, commitmentTxInfo, channelInfo.PubKeyA, signer)
	if err != nil {
		return nil, err
	}

	txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
		commitmentTxInfo.MultiAddress,
		privkeys,
		[]rpc.TransactionInputItem{
			{rdTransaction.InputTxid, rdTransaction.InputVout, rdTransaction.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{rdTransaction.PubKeyA, rdTransaction.Amount},
		},
		0,
		rdTransaction.Sequnence)
	if err != nil {
		return nil, err
	}
	rdTransaction.Txid = txid
	rdTransaction.TxHexFirstSign = hex
	rdTransaction.FirstSignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_OtherSign
	err = tx.Save(rdTransaction)
	if err != nil {
		return nil, err
	}
	return commitmentTxInfo, err
}

func createBobSideTxs(tx storm.Node, data *bean.CommitmentTxSigned, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, signer *bean.User) (*dao.CommitmentTxInfo, error) {
	var createSide = 1
	var isAliceCreateTransfer = true
	if signer.PeerId == channelInfo.PeerIdA {
		isAliceCreateTransfer = false
	}

	var lastCommitmentATx = &dao.CommitmentTxInfo{}
	err := tx.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CreateSide", createSide), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Or(q.Eq("PeerIdA", signer.PeerId), q.Eq("PeerIdB", signer.PeerId))).OrderBy("CreateAt").Reverse().First(lastCommitmentATx)
	if err != nil {
		lastCommitmentATx = nil
	}

	//如果是单通道充值，bob第一次没有c rd
	if lastCommitmentATx != nil {

		// create BRa tx
		breachRemedyTransaction, err := createBRTx(createSide, channelInfo, lastCommitmentATx, signer)
		if err != nil {
			return nil, err
		}

		var brPrivKeys = make([]string, 0)
		if tool.CheckIsString(&data.LastTempPrivateKey) {
			brPrivKeys = append(brPrivKeys, data.LastTempPrivateKey)
		}

		txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
			lastCommitmentATx.MultiAddress,
			brPrivKeys,
			[]rpc.TransactionInputItem{
				{breachRemedyTransaction.InputTxid, breachRemedyTransaction.InputVout, breachRemedyTransaction.InputAmount},
			},
			[]rpc.TransactionOutputItem{
				{breachRemedyTransaction.PubKeyB, breachRemedyTransaction.Amount},
			},
			0,
			0)
		if err != nil {
			return nil, err
		}
		breachRemedyTransaction.Txid = txid
		breachRemedyTransaction.TxHexFirstSign = hex
		breachRemedyTransaction.FirstSignAt = time.Now()
		breachRemedyTransaction.CurrState = dao.TxInfoState_OtherSign
		err = tx.Save(breachRemedyTransaction)

		lastRDTransaction := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CreateSide", createSide), q.Eq("CommitmentTxId", lastCommitmentATx.Id), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Or(q.Eq("PeerIdA", signer.PeerId), q.Eq("PeerIdB", signer.PeerId))).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
		if err != nil {
			return nil, err
		}

		lastCommitmentATx.CurrState = dao.TxInfoState_Abord
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		tx.Update(lastCommitmentATx)
		tx.Update(lastRDTransaction)
	}

	var outputBean = commitmentOutputBean{}
	if isAliceCreateTransfer {
		outputBean.TempAddress = data.CurrTempPubKey
		//default alice transter to bob,then bob add money
		outputBean.AmountM = fundingTransaction.AmountA + data.Amount
		outputBean.AmountB = fundingTransaction.AmountB - data.Amount
		if lastCommitmentATx != nil {
			outputBean.AmountM = lastCommitmentATx.AmountM + data.Amount
			outputBean.AmountB = lastCommitmentATx.AmountB - data.Amount
		}
	} else {
		outputBean.TempAddress = data.CurrTempPubKeyFromStarter
		outputBean.AmountM = fundingTransaction.AmountA - data.Amount
		outputBean.AmountB = fundingTransaction.AmountB + data.Amount
		if lastCommitmentATx != nil {
			outputBean.AmountM = lastCommitmentATx.AmountM - data.Amount
			outputBean.AmountB = lastCommitmentATx.AmountB + data.Amount
		}
	}
	outputBean.ToAddressB = channelInfo.PubKeyA

	// create C2b tx
	commitmentTxInfo, err := createCommitmentATx(createSide, channelInfo, fundingTransaction, outputBean, signer)
	if err != nil {
		return nil, err
	}

	var privkeys = make([]string, 0)
	var privKey = tempAddrPrivateKeyMap[channelInfo.PubKeyA]
	if tool.CheckIsString(&privKey) {
		privkeys = append(privkeys, privKey)
	}
	delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)

	txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
		channelInfo.ChannelPubKey,
		privkeys,
		[]rpc.TransactionInputItem{
			{commitmentTxInfo.InputTxid, commitmentTxInfo.InputVout, commitmentTxInfo.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{commitmentTxInfo.MultiAddress, commitmentTxInfo.AmountM},
			{commitmentTxInfo.PubKeyB, commitmentTxInfo.AmountB},
		},
		0,
		0)
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.LastCommitmentTxId = lastCommitmentATx.Id
	commitmentTxInfo.Txid = txid
	commitmentTxInfo.TxHexFirstSign = hex
	commitmentTxInfo.FirstSignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_OtherSign
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		return nil, err
	}

	// create RDb tx
	rdTransaction, err := createRDaTx(createSide, channelInfo, commitmentTxInfo, channelInfo.PubKeyB, signer)
	if err != nil {
		return nil, err
	}

	txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
		commitmentTxInfo.MultiAddress,
		privkeys,
		[]rpc.TransactionInputItem{
			{rdTransaction.InputTxid, rdTransaction.InputVout, rdTransaction.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{rdTransaction.PubKeyA, rdTransaction.Amount},
		},
		0,
		rdTransaction.Sequnence)
	if err != nil {
		return nil, err
	}
	rdTransaction.Txid = txid
	rdTransaction.TxHexFirstSign = hex
	rdTransaction.FirstSignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_OtherSign
	err = tx.Save(rdTransaction)
	if err != nil {
		return nil, err
	}
	return commitmentTxInfo, err
}

func (service *commitTxSignedManager) GetItemsByChannelId(jsonData string) (nodes []dao.CommitmentTxInfo, count *int, err error) {
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
		return nil, nil, err
	}

	nodes = []dao.CommitmentTxInfo{}
	tempCount, err := db.Select(q.Eq("ChannelId", chanId)).Count(&dao.CommitmentTxInfo{})
	if err != nil {
		return nil, nil, err
	}
	count = &tempCount
	err = db.Select(q.Eq("ChannelId", chanId)).Skip(int(skip)).Limit(int(pageSize)).Find(&nodes)
	return nodes, count, err
}

func (service *commitTxSignedManager) GetItemById(id int) (node *dao.CommitmentTxInfo, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitTxSignedManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.CommitmentTxInfo{})
}
