package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"time"
)

type commitTxManager struct{}

var CommitmentTxService commitTxManager

func (service *commitTxManager) CreateNewCommitmentTxRequest(jsonData string, creator *bean.User) (data *bean.CommitmentTx, targetUser *string, err error) {
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

	commitmentTxInfo := &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Or(q.Eq("PeerIdA", creator.PeerId), q.Eq("PeerIdB", creator.PeerId))).OrderBy("CreateAt").Reverse().First(commitmentTxInfo)
	if err != nil {
		return nil, nil, err
	}
	bob := commitmentTxInfo.PeerIdB
	var balance = commitmentTxInfo.AmountM
	if creator.PeerId == commitmentTxInfo.PeerIdB {
		balance = commitmentTxInfo.AmountB
		bob = commitmentTxInfo.PeerIdA
	}

	if balance < data.Amount {
		return nil, nil, errors.New("not enough payment amount")
	}
	return data, &bob, err
}
func (service *commitTxManager) GetNewestCommitmentTxByChannelId(jsonData string, user *bean.User) (node *dao.CommitmentTxInfo, err error) {
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

func (service *commitTxManager) GetNewestRDTxByChannelId(jsonData string, user *bean.User) (node *dao.RevocableDeliveryTransaction, err error) {
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
func (service *commitTxManager) GetNewestBRTxByChannelId(jsonData string, user *bean.User) (node *dao.BreachRemedyTransaction, err error) {
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

func (service *commitTxManager) GetItemsByChannelId(jsonData string) (nodes []dao.CommitmentTxInfo, count *int, err error) {
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

func (service *commitTxManager) GetItemById(id int) (node *dao.CommitmentTxInfo, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitTxManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.CommitmentTxInfo{})
}

func (service *commitTxManager) Del(id int) (node *dao.CommitmentTxInfo, err error) {
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

func (service *commitTxSignedManager) CommitmentTxSign(jsonData string, signer *bean.User) (commitmentTxInfo *dao.CommitmentTxInfo, targetUser *string, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, nil, errors.New("empty json data")
	}
	data := &bean.CommitmentTxSigned{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, nil, err
	}
	if len(data.ChannelId) != 32 {
		return nil, nil, errors.New("wrong ChannelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId)).First(channelInfo)
	if err != nil {
		return nil, nil, err
	}

	alice := channelInfo.PeerIdB
	if signer.PeerId == commitmentTxInfo.PeerIdB {
		alice = commitmentTxInfo.PeerIdA
	}

	if data.Attitude == false {
		return nil, &alice, errors.New("signer disagree transaction")
	}

	var fundingTransaction = &dao.FundingTransaction{}
	err = db.One("ChannelId", data.ChannelId, fundingTransaction)
	if err != nil {
		return nil, nil, err
	}

	lastCommitmentTx := &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.TxInfoState_OtherSign), q.Or(q.Eq("PeerIdA", signer.PeerId), q.Eq("PeerIdB", signer.PeerId))).OrderBy("CreateAt").Reverse().First(lastCommitmentTx)
	if err != nil {
		return nil, nil, err
	}

	tx, _ := db.Begin(true)
	defer tx.Rollback()

	// create BRa tx
	breachRemedyTransaction, err := createBRaTx(channelInfo, lastCommitmentTx, signer)
	if err != nil {
		return nil, nil, err
	}

	txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
		commitmentTxInfo.MultiAddress,
		[]string{
			data.ReceiverSignature,
		},
		[]rpc.TransactionInputItem{
			{breachRemedyTransaction.InputTxid, breachRemedyTransaction.InputVout, breachRemedyTransaction.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{breachRemedyTransaction.PubKeyB, breachRemedyTransaction.Amount},
		},
		0,
		0)
	if err != nil {
		return nil, nil, err
	}
	breachRemedyTransaction.Txid = txid
	breachRemedyTransaction.TxHexFirstSign = hex
	breachRemedyTransaction.FirstSignAt = time.Now()
	breachRemedyTransaction.CurrState = dao.TxInfoState_OtherSign
	err = tx.Save(breachRemedyTransaction)

	// create C2a tx
	commitmentTxInfo, err = createCommitmentATx(channelInfo, fundingTransaction, signer)
	if err != nil {
		return nil, nil, err
	}
	txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
		channelInfo.ChannelPubKey,
		[]string{
			data.ReceiverSignature,
		},
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
		return nil, nil, err
	}
	commitmentTxInfo.Txid = txid
	commitmentTxInfo.TxHexFirstSign = hex
	commitmentTxInfo.FirstSignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_OtherSign
	_ = tx.Save(commitmentTxInfo)

	// create RDa tx
	rdTransaction, err := createRDaTx(channelInfo, commitmentTxInfo, signer)
	if err != nil {
		return nil, nil, err
	}

	txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
		commitmentTxInfo.MultiAddress,
		[]string{
			data.ReceiverSignature,
		},
		[]rpc.TransactionInputItem{
			{rdTransaction.InputTxid, rdTransaction.InputVout, rdTransaction.InputAmount},
		},
		[]rpc.TransactionOutputItem{
			{rdTransaction.PubKeyA, rdTransaction.Amount},
		},
		0,
		rdTransaction.Sequnence)
	if err != nil {
		return nil, nil, err
	}
	rdTransaction.Txid = txid
	rdTransaction.TxHexFirstSign = hex
	rdTransaction.FirstSignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_OtherSign
	_ = tx.Save(rdTransaction)

	err = tx.Save(commitmentTxInfo)
	tx.Commit()

	return commitmentTxInfo, &alice, err
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
