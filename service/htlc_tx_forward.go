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

type htlcForwardTxManager struct {
	operationFlag sync.Mutex
}

const singleHopPerHopDuration = 6 * 24

// htlc 正向交易
var HtlcForwardTxService htlcForwardTxManager

// -42 find inter node and send msg to inter node
func (service *htlcForwardTxManager) AliceFindPathOfSingleHopAndSendToBob(msgData string, user bean.User) (data map[string]interface{}, bob string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcRequestFindPathAndSendH{}
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

	// operate db
	channelCount := 2
	htlcSingleHopPathInfo := &dao.HtlcSingleHopPathInfo{}
	htlcSingleHopPathInfo.ChannelIdArr = make([]int, channelCount)
	htlcSingleHopPathInfo.ChannelIdArr[0] = aliceChannel.Id
	htlcSingleHopPathInfo.ChannelIdArr[1] = carlChannel.Id
	htlcSingleHopPathInfo.InterNodePeerId = bob
	htlcSingleHopPathInfo.HAndRInfoRequestHash = rAndHInfo.RequestHash
	htlcSingleHopPathInfo.CurrState = dao.SingleHopPathInfoState_Created
	htlcSingleHopPathInfo.TotalStep = len(htlcSingleHopPathInfo.ChannelIdArr) * 2
	htlcSingleHopPathInfo.CurrStep = 0
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

// -43 send H to next node
func (service *htlcForwardTxManager) SendH(msgData string, user bean.User) (data map[string]interface{}, targetUserId string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcSendH{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	htlcSingleHopPathInfo := &dao.HtlcSingleHopPathInfo{}
	err = db.Select(q.Eq("HAndRInfoRequestHash", reqData.HAndRInfoRequestHash)).First(htlcSingleHopPathInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	rAndHInfo := dao.HtlcRAndHInfo{}
	err = db.Select(q.Eq("RequestHash", reqData.HAndRInfoRequestHash)).First(&rAndHInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	currChannelIndex := htlcSingleHopPathInfo.CurrStep
	if currChannelIndex < -1 || currChannelIndex > len(htlcSingleHopPathInfo.ChannelIdArr) {
		return nil, "", errors.New("err channel id")
	}
	carlChannel := &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("Id", htlcSingleHopPathInfo.ChannelIdArr[currChannelIndex]),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).First(carlChannel)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	targetUserId = carlChannel.PeerIdB
	if user.PeerId == carlChannel.PeerIdB {
		targetUserId = carlChannel.PeerIdA
	}
	data = make(map[string]interface{})
	data["request_hash"] = htlcSingleHopPathInfo.HAndRInfoRequestHash
	data["h"] = rAndHInfo.H
	return data, targetUserId, nil
}

// -44  下一个节点回复请求，如果答应，就把那些密钥信息传递过来，临时放到pathinfo里面
func (service *htlcForwardTxManager) SignGetH(msgData string, user bean.User) (data map[string]interface{}, targetUser string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	requestData := &bean.HtlcSignGetH{}
	err = json.Unmarshal([]byte(msgData), requestData)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	// region check input data
	if requestData.Approval {
		if tool.CheckIsString(&requestData.ChannelAddressPrivateKey) == false {
			return nil, "", errors.New("channel_address_private_key is empty")
		}
		if tool.CheckIsString(&requestData.LastTempAddressPrivateKey) == false {
			return nil, "", errors.New("last_temp_address_private_key is empty")
		}
		if tool.CheckIsString(&requestData.CurrRsmcTempAddressPubKey) == false {
			return nil, "", errors.New("curr_rsmc_temp_address_pub_key is empty")
		}
		if tool.CheckIsString(&requestData.CurrRsmcTempAddressPrivateKey) == false {
			return nil, "", errors.New("curr_rsmc_temp_address_private_key is empty")
		}
		if tool.CheckIsString(&requestData.CurrHtlcTempAddressForHt1aPubKey) == false {
			return nil, "", errors.New("curr_htlc_temp_address_for_ht1a_pub_key is empty")
		}
		if tool.CheckIsString(&requestData.CurrHtlcTempAddressForHt1aPrivateKey) == false {
			return nil, "", errors.New("curr_htlc_temp_address_for_ht1a_private_key is empty")
		}
	}
	// endregion

	tx, err := db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	defer tx.Rollback()

	// region query db data

	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = tx.Select(q.Eq("RequestHash", requestData.RequestHash)).First(rAndHInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	htlcSingleHopPathInfo := &dao.HtlcSingleHopPathInfo{}
	err = tx.Select(q.Eq("HAndRInfoRequestHash", requestData.RequestHash)).First(htlcSingleHopPathInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if htlcSingleHopPathInfo.CurrStep > 2 {
		return nil, "", errors.New("error step")
	}

	if requestData.Approval == false && htlcSingleHopPathInfo.CurrStep == 1 {
		err = errors.New("the receiver can not refuse")
		log.Println(err)
		return nil, "", err
	}
	if requestData.Approval == false {
		htlcSingleHopPathInfo.CurrState = dao.SingleHopPathInfoState_RefusedByInterNode
	} else {
		htlcSingleHopPathInfo.CurrState = dao.SingleHopPathInfoState_StepBegin
	}

	//endregion

	// region temp store data
	if requestData.Approval {
		//锁定两个通道
		if htlcSingleHopPathInfo.CurrStep == 0 {
			for _, id := range htlcSingleHopPathInfo.ChannelIdArr {
				channelInfo := &dao.ChannelInfo{}
				err := tx.One("Id", id, channelInfo)
				if err != nil {
					log.Println(err.Error())
					return nil, "", err
				}
				channelInfo.CurrState = dao.ChannelState_HtlcBegin
				err = tx.Update(channelInfo)
				if err != nil {
					log.Println(err.Error())
					return nil, "", err
				}
			}
		}

		currChannelIndex := htlcSingleHopPathInfo.CurrStep
		if currChannelIndex < -1 || currChannelIndex > len(htlcSingleHopPathInfo.ChannelIdArr) {
			return nil, "", errors.New("err channel id")
		}

		currChannel := &dao.ChannelInfo{}
		err := tx.One("Id", htlcSingleHopPathInfo.ChannelIdArr[currChannelIndex], currChannel)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}

		if currChannel.PeerIdB == user.PeerId {
			tempAddrPrivateKeyMap[currChannel.PubKeyB] = requestData.ChannelAddressPrivateKey
		} else {
			tempAddrPrivateKeyMap[currChannel.PubKeyA] = requestData.ChannelAddressPrivateKey
		}
		bobLatestCommitmentTx, err := getLatestCommitmentTx(currChannel.ChannelId, user.PeerId)
		if err == nil {
			tempAddrPrivateKeyMap[bobLatestCommitmentTx.RSMCTempAddressPubKey] = requestData.LastTempAddressPrivateKey
		}
		tempAddrPrivateKeyMap[htlcSingleHopPathInfo.BobCurrRsmcTempPubKey] = requestData.CurrRsmcTempAddressPrivateKey
		tempAddrPrivateKeyMap[htlcSingleHopPathInfo.BobCurrHtlcTempPubKey] = requestData.CurrHtlcTempAddressPrivateKey

		htlcSingleHopPathInfo.BobCurrRsmcTempPubKey = requestData.CurrRsmcTempAddressPubKey
		htlcSingleHopPathInfo.BobCurrHtlcTempPubKey = requestData.CurrHtlcTempAddressPubKey
		htlcSingleHopPathInfo.BobCurrHtlcTempForHt1bPubKey = requestData.CurrHtlcTempAddressForHt1aPubKey
	}

	// endregion

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
	data["approval"] = requestData.Approval
	data["request_hash"] = requestData.RequestHash
	return data, rAndHInfo.SenderPeerId, nil
}

// -45
func (service *htlcForwardTxManager) SenderBeginCreateHtlcCommitmentTx(msgData string, user bean.User) (outData map[string]interface{}, targetUser string, err error) {
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
	err = db.Select(q.Eq("HAndRInfoRequestHash", requestData.RequestHash)).First(&htlcSingleHopPathInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if htlcSingleHopPathInfo.CurrStep > 2 {
		return nil, "", errors.New("error step")
	}

	htlcSingleHopPathInfo.CurrStep += 1

	hAndRInfo := dao.HtlcRAndHInfo{}
	err = db.Select(q.Eq("RequestHash", htlcSingleHopPathInfo.HAndRInfoRequestHash)).First(&hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// region check input private key
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
	// endregion

	//1、上一个交易必然是RSMC交易，所以需要结算上一个交易，为其创建BR交易
	//2、然后创建HTLC的commitment交易（Cna和Cnb），它有一个输入（三个btc的input），三个输出（rsmc，bob，htlc）
	//3、关于htlc的输出，也是把资金放到一个临时多签地址里面，这个资金在Alice(交易发起方)一方会创建一个锁定一天的交易（HT1a）
	//4、HT1a的构造: Cna的第三个输出作为输入，
	// 	其输出就是产生htlc里面的rsmc（为何要用这种呢？这个本身是alice自己的余额，所以提现是需要限制的，限制就是rsmc）
	// 	和CommitmentTx一样，要产生rsmc，就是要创建一个临时多签地址，所以又需要一组私钥(Alice的临时地址，bob的通道地址)
	// 	所以Alice这一方要创建上个交易的BR，新的C2a，Rd,HT1a，HTRD1a

	//launch database transaction, if anything goes wrong, roll back.
	dbTx, err := db.Begin(true)
	if err != nil {
		return nil, "", err
	}
	defer dbTx.Rollback()

	// region prepare the data
	currChannelIndex := htlcSingleHopPathInfo.CurrStep - 1
	if currChannelIndex < -1 || currChannelIndex > len(htlcSingleHopPathInfo.ChannelIdArr) {
		return nil, "", errors.New("err channel id")
	}

	channelInfo := dao.ChannelInfo{}
	err = dbTx.One("Id", htlcSingleHopPathInfo.ChannelIdArr[currChannelIndex], &channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	//当前操作者是Alice Alice转账给Bob
	if user.PeerId == channelInfo.PeerIdA {
		targetUser = channelInfo.PeerIdB
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = requestData.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	} else { //当前操作者是Bob Bob转账给Alice
		targetUser = channelInfo.PeerIdA
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = requestData.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	}

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = dbTx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CurrState", dao.FundingTransactionState_Accept)).OrderBy("CreateAt").Reverse().First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// 创建上个交易的BR  begin
	//PeerIdA(概念中的Alice) 对上一次承诺交易的废弃
	err = htlcAliceAbortLastRsmcCommitmentTx(dbTx, channelInfo, user, *fundingTransaction, *requestData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	//PeerIdB(概念中的Bob) 对上一次承诺交易的废弃
	err = htlcBobAbortLastRsmcCommitmentTx(dbTx, channelInfo, user, *fundingTransaction, *requestData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	// 创建上个交易的BR  end

	//region 创建htlc相关的交易: Cna（1+3: Cna + toRsmc + toOther + toHtlc）+ 2(ht1a+htrd1a); cnb(1+3: Cnb + toRsmc + toOther + toHtlc) + 1(htd1b)

	///Cna Alice方(channel.PeerIdA)的交易
	commitmentTransactionOfA, err := service.htlcCreateAliceSideTxs(dbTx, channelInfo, user, *fundingTransaction, *requestData, htlcSingleHopPathInfo, hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(commitmentTransactionOfA)

	///Cnb Bob方(channel.PeerIdB)的交易
	commitmentTransactionOfB, err := service.htlcCreateBobSideTxs(dbTx, channelInfo, user, *fundingTransaction, *requestData, htlcSingleHopPathInfo, hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(commitmentTransactionOfB)

	//endregion

	htlcSingleHopPathInfo.CurrState = dao.SingleHopPathInfoState_StepFinish
	err = dbTx.Update(&htlcSingleHopPathInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	err = dbTx.Commit()
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	data := make(map[string]interface{})
	data["commitmentTransactionOfA"] = commitmentTransactionOfA
	data["commitmentTransactionOfB"] = commitmentTransactionOfB
	return data, targetUser, nil
}

// 创建Alice方的htlc的承诺交易，rsmc的Rd
// 这里要做一个判断，作为这次交易的发起者，
// 如果PeerIdA是发起者，在这Cna的逻辑中创建HT1a和HED1a
// 如果PeerIdB是发起者，那么在Cna中就应该创建HTLC Time Delivery 1b(HED1b) 和HTLC Execution  1a(HE1b)
func (service *htlcForwardTxManager) htlcCreateAliceSideTxs(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen,
	pathInfo dao.HtlcSingleHopPathInfo, hAndRInfo dao.HtlcRAndHInfo) (*dao.CommitmentTransaction, error) {

	owner := channelInfo.PeerIdA

	bobIsInterNodeSoAliceSend2Bob := true
	if operator.PeerId == channelInfo.PeerIdB {
		bobIsInterNodeSoAliceSend2Bob = false
	}

	var lastCommitmentATx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner)).OrderBy("CreateAt").Reverse().First(lastCommitmentATx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create Cna tx
	commitmentTxInfo, err := htlcCreateCna(tx, channelInfo, operator, fundingTransaction, requestData, pathInfo, hAndRInfo, bobIsInterNodeSoAliceSend2Bob, lastCommitmentATx, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create rsmc RDna tx
	rdTx, err := htlcCreateRDOfRsmc(
		tx, channelInfo, operator, fundingTransaction, requestData,
		pathInfo, bobIsInterNodeSoAliceSend2Bob, commitmentTxInfo, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(rdTx)
	timeout := (pathInfo.TotalStep/2 - pathInfo.CurrStep + 1) * singleHopPerHopDuration
	// output2,htlc的后续交易
	if bobIsInterNodeSoAliceSend2Bob { // 如是通道中的Alice转账给Bob，bob作为中间节点  创建HT1a
		// create ht1a
		htlcTimeoutTxA, err := createHtlcTimeoutTxForAliceSide(tx, owner, channelInfo, fundingTransaction, *commitmentTxInfo, requestData, timeout, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutTxA)
		// 继续创建htrd
		htrdTransaction, err := createHtlcRD(tx, channelInfo, operator, fundingTransaction, requestData, bobIsInterNodeSoAliceSend2Bob, htlcTimeoutTxA, owner)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htrdTransaction)

		//HEDa的构建放到得到R的时候   当bob得到R的时候，构建给bob的HED交易
		//htlcCreateExecutionDelivery

	} else { // bob is sender, 如果alice得到R，构建Htlc Execution交易以及接下来的HERD交易，如果超时，bob的钱就应该超时赎回HTDelivery
		// 如果是bob转给alice，Alice作为中间商，作为当前通道的接收者
		// create HTD for bob  锁定了bob的钱，超时了，就应该给bob赎回
		htlcTimeoutDeliveryTxB, err := createHtlcTimeoutDeliveryTx(tx, channelInfo.PeerIdB, channelInfo.AddressB, timeout, channelInfo, fundingTransaction, *commitmentTxInfo, requestData, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutDeliveryTxB)

		// Alice拿到了R，就可以用R去构建HTLC Execution 交易，因为是Alice这边的交易，那么就需要用RSMC来限制提现操作 即构建HE和HERD

	}

	return commitmentTxInfo, nil
}

// 创建PeerIdA方的htlc的承诺交易，rsmc的Rd
// 这里要做一个判断，作为这次交易的发起者，
// 如果PeerIdA是发起者，在这Cna的逻辑中创建HT1a和HED1a
// 如果PeerIdB是发起者，那么在Cna中就应该创建HTLC Time Delivery 1b(HED1b) 和HTLC Execution  1a(HE1b)
func (service *htlcForwardTxManager) htlcCreateBobSideTxs(dbTx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen,
	pathInfo dao.HtlcSingleHopPathInfo, hAndRInfo dao.HtlcRAndHInfo) (*dao.CommitmentTransaction, error) {

	owner := channelInfo.PeerIdB
	bobIsInterNodeSoAliceSend2Bob := true
	if operator.PeerId == channelInfo.PeerIdB {
		bobIsInterNodeSoAliceSend2Bob = false
	}

	var lastCommitmentBTx = &dao.CommitmentTransaction{}
	err := dbTx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner)).OrderBy("CreateAt").Reverse().First(lastCommitmentBTx)
	if err != nil {
		lastCommitmentBTx = nil
	}
	// create Cnb dbTx
	commitmentTxInfo, err := htlcCreateCnb(dbTx, channelInfo, operator, fundingTransaction, requestData, pathInfo, hAndRInfo, bobIsInterNodeSoAliceSend2Bob, lastCommitmentBTx, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create rsmc RDnb dbTx
	_, err = htlcCreateRDOfRsmc(
		dbTx, channelInfo, operator, fundingTransaction, requestData,
		pathInfo, bobIsInterNodeSoAliceSend2Bob, commitmentTxInfo, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	timeout := (pathInfo.TotalStep/2 - pathInfo.CurrStep + 1) * singleHopPerHopDuration
	// htlc txs
	// output2,给htlc创建的交易，如何处理output2里面的钱
	if bobIsInterNodeSoAliceSend2Bob {
		// 如是通道中的Alice转账给Bob，bob作为中间节点  创建HTD1b ，Alice的钱在超时的情况下，可以返回到Alice账号
		// 当前操作的请求者是Alice
		// create HTD1b 当超时的情况，Alice赎回自己的钱的交易
		htlcTimeoutDeliveryTxB, err := createHtlcTimeoutDeliveryTx(dbTx, channelInfo.PeerIdA, channelInfo.AddressA, timeout, channelInfo, fundingTransaction, *commitmentTxInfo, requestData, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutDeliveryTxB)

		// Htlc Execution 如果bob拿到了R，就构建bob的HE交易 然后后续的HERD

	} else {
		// create ht  for bob, bob超时赎回自己的钱钱
		htlcTimeoutTxB, err := createHtlcTimeoutTxForBobSide(dbTx, owner, channelInfo, fundingTransaction, *commitmentTxInfo, requestData, timeout, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutTxB)

		// 继续创建htrd
		htrdTransaction, err := createHtlcRD(dbTx, channelInfo, operator, fundingTransaction, requestData, bobIsInterNodeSoAliceSend2Bob, htlcTimeoutTxB, owner)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htrdTransaction)

		//HEDa 如果Alice拿到了R，就需要构建HTLC Execution Delivery 交易，把钱给Alice，这个交易的拥有着是Alice
	}
	return commitmentTxInfo, nil
}
