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
	requestData := &bean.HtlcRequestOpen{}
	err = json.Unmarshal([]byte(msgData), requestData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&requestData.RequestHash) == false {
		err = errors.New("empty request_hash")
		log.Println(err)
		return nil, "", err
	}

	htlcSingleHopPathInfo := dao.HtlcSingleHopPathInfo{}
	err = db.Select(q.Eq("", requestData.RequestHash)).First(&htlcSingleHopPathInfo)
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

	if tool.CheckIsString(&requestData.ChannelAddressPrivateKey) == false {
		err = errors.New("channel_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&requestData.LastTempAddressPrivateKey) == false {
		err = errors.New("last_temp_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&requestData.CurrRsmcTempAddressPubKey) == false {
		err = errors.New("curr_rsmc_temp_address_pub_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&requestData.CurrRsmcTempAddressPrivateKey) == false {
		err = errors.New("curr_rsmc_temp_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&requestData.CurrHtlcTempAddressForHt1aPubKey) == false {
		err = errors.New("curr_htlc_temp_address_for_ht1a_pub_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&requestData.CurrHtlcTempAddressForHt1aPrivateKey) == false {
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
	dbTx, err := db.Begin(true)
	if err != nil {
		return nil, "", err
	}
	defer dbTx.Rollback()

	channelInfo := dao.ChannelInfo{}
	err = dbTx.One("Id", htlcSingleHopPathInfo.FirstChannelId, &channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if user.PeerId == channelInfo.PeerIdA {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = requestData.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = requestData.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	}

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = dbTx.Select(q.Eq("ChannelId", htlcSingleHopPathInfo.FirstChannelId), q.Eq("CurrState", dao.FundingTransactionState_Accept)).OrderBy("CreateAt").Reverse().First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	//PeerIdA(概念中的Alice) 对上一次承诺交易的废弃
	err = htlcAliceAbortLastCommitmentTx(dbTx, channelInfo, user, *fundingTransaction, *requestData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	//PeerIdB(概念中的Bob) 对上一次承诺交易的废弃
	err = htlcBobAbortLastCommitmentTx(dbTx, channelInfo, user, *fundingTransaction, *requestData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	//开始创建htlc的承诺交易
	//Cna Alice这一方的交易
	commitmentTransactionOfA, err := service.htlcCreateAliceSideTxs(dbTx, channelInfo, user, *fundingTransaction, *requestData, htlcSingleHopPathInfo, hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(commitmentTransactionOfA)

	//Cnb Bob 那一方的交易
	commitmentTransactionOfB, err := service.htlcCreateBobSideTxs(dbTx, channelInfo, user, *fundingTransaction, *requestData, htlcSingleHopPathInfo, hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(commitmentTransactionOfB)

	err = dbTx.Commit()
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
func (service *htlcTxManager) htlcCreateAliceSideTxs(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
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
		htlcTimeoutTxA, err := createHtlcTimeoutTxForAliceSide(tx, owner, channelInfo, fundingTransaction, *commitmentTxInfo, htlcRequestOpen, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutTxA)
		// 继续创建htrd
		htrdTransaction, err := htlcCreateRD(tx, channelInfo, operator, fundingTransaction, htlcRequestOpen, bobIsInterNodeSoAliceSend2Bob, htlcTimeoutTxA, owner)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htrdTransaction)
	} else {
		// 如果是bob转给alice，Alice作为中间商，作为当前通道的接收者
		// 这个时候，Cna产生的output2是锁定的bob的钱，在Alice这一方，Alice如果拿到了R，就去创建HE1a,得到的转账收益，
		// 而如果她后续要提现，就又要创建HERD1a，
		// 而bob的钱，需要在这里等待一天（因为一个hop）才能赎回：HTD1a，因为现在已经被锁定在了output2里面，等待Alice拿到R
		// 所以这个时候，只需要创建HTD1a,保证bob的钱能回去，因为不是bob自己广播，所以bob的钱的赎回，不需要走rsmc的流程，自己提现，才需要通过RSMC机制，完成去信任机制
		// create HTD for Alice
		htlcTimeoutDeliveryTxB, err := createHtlcTimeoutDeliveryTx(tx, owner, channelInfo.AddressA, 6*24, channelInfo, fundingTransaction, *commitmentTxInfo, htlcRequestOpen, operator)
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
func (service *htlcTxManager) htlcCreateBobSideTxs(dbTx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen,
	htlcSingleHopPathInfo dao.HtlcSingleHopPathInfo, hAndRInfo dao.HtlcRAndHInfo) (*dao.CommitmentTransaction, error) {

	owner := channelInfo.PeerIdB
	bobIsInterNodeSoAliceSend2Bob := true
	if operator.PeerId == channelInfo.PeerIdB {
		bobIsInterNodeSoAliceSend2Bob = false
	}

	var lastCommitmentBTx = &dao.CommitmentTransaction{}
	err := dbTx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastCommitmentBTx)
	if err != nil {
		lastCommitmentBTx = nil
	}
	// create Cnb dbTx
	commitmentTxInfo, err := htlcCreateCnb(dbTx, channelInfo, operator, fundingTransaction, requestData, htlcSingleHopPathInfo, hAndRInfo, bobIsInterNodeSoAliceSend2Bob, lastCommitmentBTx, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create rsmc RDnb dbTx
	_, err = htlcCreateRDOfRsmc(
		dbTx, channelInfo, operator, fundingTransaction, requestData,
		htlcSingleHopPathInfo, bobIsInterNodeSoAliceSend2Bob, commitmentTxInfo, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// htlc txs
	// output2,给htlc创建的交易，如何处理output2里面的钱
	if bobIsInterNodeSoAliceSend2Bob {
		// 如是通道中的Alice转账给Bob，bob作为中间节点  创建HTD1b ，Alice的钱在超时的情况下，可以返回到Alice账号
		// 当前操作的请求者是Alice
		// create HTD1b
		htlcTimeoutDeliveryTxB, err := createHtlcTimeoutDeliveryTx(dbTx, channelInfo.PeerIdA, channelInfo.AddressA, 6*24, channelInfo, fundingTransaction, *commitmentTxInfo, requestData, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutDeliveryTxB)
		// 有了R后，创建给bob的收款交易(HE1b )及其后续交易（herd1b, Rsmc，因为在Bob方，他想提现，就必须走Rsmc的保护体制）
	} else {
		// 如果是Bob给Alice转账，现在是bob的钱被锁定在在output2里面，Bob需要在超时的时候，拿回自己的钱，或者Alice得到R的时候，生成HED1A
		// 创建bob的超时赎回的交易，当前操作的请求者是bob
		// create ht1a
		htlcTimeoutTxB, err := createHtlcTimeoutTxForBobSide(dbTx, owner, channelInfo, fundingTransaction, *commitmentTxInfo, requestData, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutTxB)

		// 继续创建htrd
		htrdTransaction, err := htlcCreateRD(dbTx, channelInfo, operator, fundingTransaction, requestData, bobIsInterNodeSoAliceSend2Bob, htlcTimeoutTxB, owner)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htrdTransaction)
	}

	return commitmentTxInfo, nil
}
