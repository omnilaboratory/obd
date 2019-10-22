package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"log"
	"sync"
	"time"
)

type htlcTxManager struct {
	operationFlag sync.Mutex
}

var HtlcTxService htlcTxManager

// query bob,and ask bob
func (service *htlcTxManager) AliceFindPathOfSingleHopAndSendToBob(msgData string, user bean.User) (data map[string]interface{}, bob string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcRequestCreate{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = db.Select(q.Eq("CreateBy", user.PeerId), q.Eq("CurrState", dao.NS_Finish), q.Eq("H", reqData.H)).First(rAndHInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	channelAliceInfos := getAllChannels(rAndHInfo.SenderPeerId)
	if len(channelAliceInfos) == 0 {
		return nil, "", errors.New("sender's channel not found")
	}
	//if has the channel direct
	for _, item := range channelAliceInfos {
		if item.PeerIdA == rAndHInfo.SenderPeerId && item.PeerIdB == rAndHInfo.RecipientPeerId {
			return nil, "", errors.New("has direct channel")
		}
		if item.PeerIdB == rAndHInfo.SenderPeerId && item.PeerIdA == rAndHInfo.RecipientPeerId {
			return nil, "", errors.New("has direct channel")
		}
	}

	channelCarlInfos := getAllChannels(rAndHInfo.RecipientPeerId)
	if len(channelCarlInfos) == 0 {
		return nil, "", errors.New("recipient's channel not found")
	}

	bob, aliceChannel, carlChannel := getTwoChannelOfSingleHop(*rAndHInfo, channelAliceInfos, channelCarlInfos)
	if tool.CheckIsString(&bob) == false {
		return nil, "", errors.New("no inter channel can use")
	}

	log.Println(aliceChannel)
	log.Println(carlChannel)
	log.Println(bob)

	// operate db
	htlcSingleHopPathInfo := &dao.HtlcSingleHopPathInfo{}
	htlcSingleHopPathInfo.FirstChannelId = aliceChannel.Id
	htlcSingleHopPathInfo.SecondChannelId = carlChannel.Id
	htlcSingleHopPathInfo.InterNodePeerId = bob
	htlcSingleHopPathInfo.HtlcCreateRandHInfoRequestHash = rAndHInfo.RequestHash
	htlcSingleHopPathInfo.CurrState = dao.NS_Create
	htlcSingleHopPathInfo.CreateBy = user.PeerId
	htlcSingleHopPathInfo.CreateAt = time.Now()
	err = db.Save(htlcSingleHopPathInfo)
	if err != nil {
		return nil, "", err
	}

	data = make(map[string]interface{})
	data["request_hash"] = rAndHInfo.RequestHash
	data["h"] = rAndHInfo.H
	return data, bob, nil
}

func (service *htlcTxManager) BobConfirmPath(msgData string, user bean.User) (data map[string]interface{}, senderPeerId string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	htlcSignRequestCreate := &bean.HtlcSignRequestCreate{}
	err = json.Unmarshal([]byte(msgData), htlcSignRequestCreate)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if htlcSignRequestCreate.Approval {
		if tool.CheckIsString(&htlcSignRequestCreate.ChannelAddressPrivateKey) == false {
			return nil, "", errors.New("channel_address_private_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.LastTempAddressPrivateKey) == false {
			return nil, "", errors.New("last_temp_address_private_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.CurrRsmcTempAddressPubKey) == false {
			return nil, "", errors.New("curr_rsmc_temp_address_pub_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.CurrRsmcTempAddressPrivateKey) == false {
			return nil, "", errors.New("curr_rsmc_temp_address_private_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.CurrHtlcTempAddressForHt1aPubKey) == false {
			return nil, "", errors.New("curr_htlc_temp_address_for_ht1a_pub_key is empty")
		}
		if tool.CheckIsString(&htlcSignRequestCreate.CurrHtlcTempAddressForHt1aPrivateKey) == false {
			return nil, "", errors.New("curr_htlc_temp_address_for_ht1a_private_key is empty")
		}
	}

	tx, err := db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	defer tx.Rollback()

	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = tx.Select(q.Eq("RequestHash", htlcSignRequestCreate.RequestHash)).First(rAndHInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	htlcSingleHopPathInfo := &dao.HtlcSingleHopPathInfo{}
	err = tx.Select(q.Eq("HtlcCreateRandHInfoRequestHash", htlcSignRequestCreate.RequestHash)).First(htlcSingleHopPathInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	htlcSingleHopPathInfo.CurrState = dao.NS_Refuse
	if htlcSignRequestCreate.Approval {
		htlcSingleHopPathInfo.CurrState = dao.NS_Finish

		htlcSingleHopPathInfo.BobCurrRsmcTempPubKey = htlcSignRequestCreate.CurrRsmcTempAddressPubKey
		htlcSingleHopPathInfo.BobCurrHtlcTempPubKey = htlcSignRequestCreate.CurrHtlcTempAddressPubKey

		//锁定两个通道
		aliceChannel := &dao.ChannelInfo{}
		err := tx.One("Id", htlcSingleHopPathInfo.FirstChannelId, aliceChannel)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}
		carlChannel := &dao.ChannelInfo{}
		err = tx.One("Id", htlcSingleHopPathInfo.SecondChannelId, carlChannel)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}

		aliceChannel.CurrState = dao.ChannelState_HtlcBegin
		err = tx.Update(aliceChannel)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}

		carlChannel.CurrState = dao.ChannelState_HtlcBegin
		err = tx.Update(carlChannel)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}

		if aliceChannel.PeerIdB == user.PeerId {
			tempAddrPrivateKeyMap[aliceChannel.PubKeyB] = htlcSignRequestCreate.ChannelAddressPrivateKey
		} else {
			tempAddrPrivateKeyMap[aliceChannel.PubKeyA] = htlcSignRequestCreate.ChannelAddressPrivateKey
		}
		bobLatestCommitmentTx, err := getLatestCommitmentTx(aliceChannel.ChannelId, user.PeerId)
		if err == nil {
			tempAddrPrivateKeyMap[bobLatestCommitmentTx.RSMCTempAddressPubKey] = htlcSignRequestCreate.LastTempAddressPrivateKey
		}
		tempAddrPrivateKeyMap[htlcSingleHopPathInfo.BobCurrRsmcTempPubKey] = htlcSignRequestCreate.CurrRsmcTempAddressPrivateKey
		tempAddrPrivateKeyMap[htlcSingleHopPathInfo.BobCurrHtlcTempPubKey] = htlcSignRequestCreate.CurrHtlcTempAddressPrivateKey
	}
	htlcSingleHopPathInfo.SignBy = user.PeerId
	htlcSingleHopPathInfo.SignAt = time.Now()
	err = tx.Update(htlcSingleHopPathInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	data = make(map[string]interface{})
	data["approval"] = htlcSignRequestCreate.Approval
	data["request_hash"] = htlcSignRequestCreate.RequestHash
	return data, rAndHInfo.SenderPeerId, nil
}

// -44
func (service *htlcTxManager) AliceOpenHtlcChannel(msgData string, user bean.User) (outData map[string]interface{}, targetUser string, err error) {
	if tool.CheckIsString(&msgData) == false {
		err = errors.New("empty json data")
		log.Println(err)
		return nil, "", err
	}
	htlcRequestOpen := &bean.HtlcRequestOpen{}
	err = json.Unmarshal([]byte(msgData), htlcRequestOpen)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&htlcRequestOpen.RequestHash) == false {
		err = errors.New("empty request_hash")
		log.Println(err)
		return nil, "", err
	}

	htlcSingleHopPathInfo := dao.HtlcSingleHopPathInfo{}
	err = db.Select(q.Eq("", htlcRequestOpen.RequestHash)).First(&htlcSingleHopPathInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	hAndRInfo := dao.HtlcRAndHInfo{}
	err = db.Select(q.Eq("RequestHash", htlcSingleHopPathInfo.HtlcCreateRandHInfoRequestHash)).First(&hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&htlcRequestOpen.ChannelAddressPrivateKey) == false {
		err = errors.New("channel_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.LastTempAddressPrivateKey) == false {
		err = errors.New("last_temp_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.CurrRsmcTempAddressPubKey) == false {
		err = errors.New("curr_rsmc_temp_address_pub_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.CurrRsmcTempAddressPrivateKey) == false {
		err = errors.New("curr_rsmc_temp_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.CurrHtlcTempAddressForHt1aPubKey) == false {
		err = errors.New("curr_htlc_temp_address_for_ht1a_pub_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&htlcRequestOpen.CurrHtlcTempAddressForHt1aPrivateKey) == false {
		err = errors.New("curr_htlc_temp_address_for_ht1a_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	//1、上一个交易必然是RSMC交易，所以需要结算上一个交易，为其创建BR交易
	//2、然后创建HTLC的commitment交易（Cna和Cnb），它有一个输入（三个btc的input），三个输出（rsmc，bob，htlc）
	//3、关于htlc的输出，也是把资金放到一个临时多签地址里面，这个资金在Alice(交易发起方)一方会创建一个锁定一天的交易（HT1a）
	//4、HT1a的构造，Cna的第三个输出作为输入，
	// 	其输出就是产生htlc里面的rsmc（为何要用这种呢？这个本身是alice自己的余额，所以提现是需要限制的，限制就是rsmc）
	// 	和CommitmentTx一样，要产生rsmc，就是要创建一个临时多签地址，所以又需要一组私钥
	// 	所以alice要创建BR，Cna，HT1a，HED1a,HTRD1a

	//launch database transaction, if anything goes wrong, roll back.
	tx, err := db.Begin(true)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback()

	channelInfo := dao.ChannelInfo{}
	err = tx.One("Id", htlcSingleHopPathInfo.FirstChannelId, &channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if user.PeerId == channelInfo.PeerIdA {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = htlcRequestOpen.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = htlcRequestOpen.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	}

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = tx.Select(q.Eq("ChannelId", htlcSingleHopPathInfo.FirstChannelId), q.Eq("CurrState", dao.FundingTransactionState_Accept)).OrderBy("CreateAt").Reverse().First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	//PeerIdA(概念中的Alice) 对上一次承诺交易的废弃
	err = htlcAliceAbortLastCommitmentTx(tx, channelInfo, user, *fundingTransaction, *htlcRequestOpen)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	//PeerIdB(概念中的Bob) 对上一次承诺交易的废弃
	err = htlcBobAbortLastCommitmentTx(tx, channelInfo, user, *fundingTransaction, *htlcRequestOpen)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	//开始创建htlc的承诺交易
	//Cna Alice这一方的交易
	commitmentTransactionOfA, err := service.htlcCreateAliceTxs(tx, channelInfo, user, *fundingTransaction, *htlcRequestOpen, htlcSingleHopPathInfo, hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(commitmentTransactionOfA)
	//Cnb Alice这一方的交易
	commitmentTransactionOfB, err := service.htlcCreateBobTxs(tx, channelInfo, user, *fundingTransaction, *htlcRequestOpen, htlcSingleHopPathInfo, hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(commitmentTransactionOfB)

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	return nil, "", nil
}

// 创建PeerIdA方的htlc的承诺交易，rsmc的Rd
// 这里要做一个判断，作为这次交易的发起者，
// 如果PeerIdA是发起者，在这Cna的逻辑中创建HT1a和HED1a
// 如果PeerIdB是发起者，那么在Cna中就应该创建HTLC Time Delivery 1b(HED1b) 和HTLC Execution  1a(HE1b)
func (service *htlcTxManager) htlcCreateAliceTxs(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, htlcRequestOpen bean.HtlcRequestOpen,
	htlcSingleHopPathInfo dao.HtlcSingleHopPathInfo, hAndRInfo dao.HtlcRAndHInfo) (*dao.CommitmentTransaction, error) {
	owner := channelInfo.PeerIdA
	bobIsInterNodeSoAliceSend2Bob := true
	if operator.PeerId == channelInfo.PeerIdB {
		bobIsInterNodeSoAliceSend2Bob = false
	}

	var lastCommitmentATx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastCommitmentATx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// create Cna tx
	commitmentTxInfo, err := htlcCreateCna(tx, channelInfo, operator, fundingTransaction, htlcRequestOpen, htlcSingleHopPathInfo, hAndRInfo, bobIsInterNodeSoAliceSend2Bob, lastCommitmentATx, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create rsmc RDna tx
	_, err = htlcCreateRDOfRsmc(
		tx, channelInfo, operator, fundingTransaction, htlcRequestOpen,
		htlcSingleHopPathInfo, bobIsInterNodeSoAliceSend2Bob, commitmentTxInfo, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// output2,给htlc创建的交易
	if bobIsInterNodeSoAliceSend2Bob { // 如是通道中的Alice转账给Bob，bob作为中间节点  创建HT1a
		// create ht1a
		htlcTimeoutTxA, err := createHtlcTimeoutTx(tx, owner, channelInfo, fundingTransaction, *commitmentTxInfo, htlcRequestOpen, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutTxA)
		// 继续创建htrd
		htrdTransaction, err := htlcCreateRD(tx, channelInfo, operator, fundingTransaction, htlcRequestOpen, htlcSingleHopPathInfo, bobIsInterNodeSoAliceSend2Bob, htlcTimeoutTxA, owner)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htrdTransaction)

		//创建hed1a  此交易要修改创建时机，等到bob拿到R的时候，再来创建，这个时候就需要广播交易（关闭通道），那么在很多情况下，其实是不用创建的
		hednA, err := htlcCreateExecutionDeliveryA(tx, channelInfo, fundingTransaction, *commitmentTxInfo, htlcRequestOpen, owner, hAndRInfo, "")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(hednA)

	} else {
		// 如果是bob转给alice，Alice作为中间商，作为当前通道的接收者
		// 这个时候，Cna产生的output2是锁定的bob的钱，在Alice这一方，Alice如果拿到了R，就去创建HE1a,得到的转账收益，
		// 而如果她后续要提现，就又要创建HERD1a，
		// 而bob的钱，需要在这里等待一天（因为一个hop）才能赎回：HTD1a，因为现在已经被锁定在了output2里面，等待Alice拿到R
		// 所以这个时候，只需要创建HTD1a,保证bob的钱能回去，因为不是bob自己广播，所以bob的钱的赎回，不需要走rsmc的流程，自己提现，才需要通过RSMC机制，完成去信任机制
		// create HTD1B for Alice
		htlcTimeoutDeliveryTxB, err := createHtlcTimeoutDeliveryTx(tx, owner, channelInfo.PeerIdA, channelInfo, fundingTransaction, *commitmentTxInfo, htlcRequestOpen, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutDeliveryTxB)
	}

	return commitmentTxInfo, nil
}

// 创建PeerIdA方的htlc的承诺交易，rsmc的Rd
// 这里要做一个判断，作为这次交易的发起者，
// 如果PeerIdA是发起者，在这Cna的逻辑中创建HT1a和HED1a
// 如果PeerIdB是发起者，那么在Cna中就应该创建HTLC Time Delivery 1b(HED1b) 和HTLC Execution  1a(HE1b)
func (service *htlcTxManager) htlcCreateBobTxs(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, htlcRequestOpen bean.HtlcRequestOpen,
	htlcSingleHopPathInfo dao.HtlcSingleHopPathInfo, hAndRInfo dao.HtlcRAndHInfo) (*dao.CommitmentTransaction, error) {
	owner := channelInfo.PeerIdB
	bobIsInterNodeSoAliceSend2Bob := true
	if operator.PeerId == channelInfo.PeerIdB {
		bobIsInterNodeSoAliceSend2Bob = false
	}

	var lastCommitmentBTx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastCommitmentBTx)
	if err != nil {
		lastCommitmentBTx = nil
	}
	// create Cnb tx
	commitmentTxInfo, err := htlcCreateCnb(tx, channelInfo, operator, fundingTransaction, htlcRequestOpen, htlcSingleHopPathInfo, hAndRInfo, bobIsInterNodeSoAliceSend2Bob, lastCommitmentBTx, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create rsmc RDnb tx
	_, err = htlcCreateRDOfRsmc(
		tx, channelInfo, operator, fundingTransaction, htlcRequestOpen,
		htlcSingleHopPathInfo, bobIsInterNodeSoAliceSend2Bob, commitmentTxInfo, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// htlc txs
	// output2,给htlc创建的交易
	if bobIsInterNodeSoAliceSend2Bob { // 如是通道中的Alice转账给Bob，bob作为中间节点  创建HT1a
		// create ht1a
		htlcTimeoutTxA, err := createHtlcTimeoutTx(tx, owner, channelInfo, fundingTransaction, *commitmentTxInfo, htlcRequestOpen, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutTxA)
		// 继续创建htrd
		htrdTransaction, err := htlcCreateRD(tx, channelInfo, operator, fundingTransaction, htlcRequestOpen, htlcSingleHopPathInfo, bobIsInterNodeSoAliceSend2Bob, htlcTimeoutTxA, owner)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htrdTransaction)

		//创建hed1a  此交易要修改创建时机，等到bob拿到R的时候，再来创建，这个时候就需要广播交易（关闭通道），那么在很多情况下，其实是不用创建的
		hednA, err := htlcCreateExecutionDeliveryA(tx, channelInfo, fundingTransaction, *commitmentTxInfo, htlcRequestOpen, owner, hAndRInfo, "")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(hednA)

	} else {
		// 如果是bob转给alice，Alice作为中间商，作为当前通道的接收者
		// 这个时候，Cna产生的output2是锁定的bob的钱，在Alice这一方，Alice如果拿到了R，就去创建HE1a,得到的转账收益，
		// 而如果她后续要提现，就又要创建HERD1a，
		// 而bob的钱，需要在这里等待一天（因为一个hop）才能赎回：HTD1a，因为现在已经被锁定在了output2里面，等待Alice拿到R
		// 所以这个时候，只需要创建HTD1a,保证bob的钱能回去，因为不是bob自己广播，所以bob的钱的赎回，不需要走rsmc的流程，自己提现，才需要通过RSMC机制，完成去信任机制
		// create HTD1B for Alice
		htlcTimeoutDeliveryTxB, err := createHtlcTimeoutDeliveryTx(tx, owner, channelInfo.PeerIdA, channelInfo, fundingTransaction, *commitmentTxInfo, htlcRequestOpen, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutDeliveryTxB)
	}

	return commitmentTxInfo, nil
}
