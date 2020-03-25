package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"log"
	"obd/bean"
	"obd/config"
	"obd/dao"
	"obd/tool"
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
func (service *htlcForwardTxManager) AliceFindPathAndSendToBob(msgData string, user bean.User) (data map[string]interface{}, bob string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcRequestFindPathAndSendH{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.H) == false {
		return nil, "", errors.New("wrong h")
	}
	if tool.CheckIsString(&reqData.RecipientPeerId) == false {
		return nil, "", errors.New("wrong recipient_peer_id")
	}

	if err := FindUserIsOnline(reqData.RecipientPeerId); err != nil {
		return nil, "", err
	}

	if reqData.PropertyId < 0 {
		return nil, "", errors.New("wrong property_id")
	}
	if reqData.Amount < config.Dust {
		return nil, "", errors.New("wrong amount")
	}

	//find  path
	channelAliceInfos := getAllChannels(user.PeerId)
	if len(channelAliceInfos) == 0 {
		return nil, "", errors.New("sender's currChannel not found")
	}
	//if has the currChannel direct
	var directChannel dao.ChannelInfo
	for _, item := range channelAliceInfos {
		flag := false
		if item.PeerIdA == user.PeerId && item.PeerIdB == reqData.RecipientPeerId {
			flag = true
		}
		if item.PeerIdB == user.PeerId && item.PeerIdA == reqData.RecipientPeerId {
			flag = true
		}
		if flag {
			commitmentTxInfo, err := getLatestCommitmentTx(item.ChannelId, user.PeerId)
			if err == nil {
				if commitmentTxInfo.PropertyId == reqData.PropertyId &&
					commitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign &&
					commitmentTxInfo.AmountToRSMC >= reqData.Amount {
					directChannel = item
					break
				}
			}
		}
	}

	channelIdArr := make([]int, 0)
	if &directChannel != nil {
		log.Println("has direct currChannel")
		channelIdArr = append(channelIdArr, directChannel.Id)
	} else {
		channelCarlInfos := getAllChannels(reqData.RecipientPeerId)
		if len(channelCarlInfos) == 0 {
			return nil, "", errors.New("recipient's currChannel not found")
		}

		//find the path from transaction creator to the receiver
		//寻路 从发起方到收款方，支持直连和借道跳转
		PathService.GetPath(user.PeerId, reqData.RecipientPeerId, reqData.PropertyId, reqData.Amount, nil, true)
		miniPathLength := 7
		var miniPathNode *PathNode

		for _, node := range PathService.openList {
			if node.IsTarget {
				if int(node.Level) < miniPathLength {
					miniPathLength = int(node.Level)
					miniPathNode = node
				}
			}
		}
		if miniPathNode == nil {
			return nil, "", errors.New("no inter currChannel can use")
		}
		channelCount := miniPathNode.Level
		channelIdArr = append(channelIdArr, miniPathNode.ChannelId)
		for i := int(channelCount) - 1; i > 0; i-- {
			channelIdArr = append(channelIdArr, PathService.openList[miniPathNode.PathIdArr[i]].ChannelId)
		}
	}

	currChannel := &dao.ChannelInfo{}
	err = db.Select(q.Eq("Id", channelIdArr[0])).First(currChannel)
	if err != nil {
		err = errors.New("fail to find currChannel")
		log.Println(err)
		return nil, "", err
	}
	bob = currChannel.PeerIdB
	if currChannel.PeerIdB == user.PeerId {
		bob = currChannel.PeerIdA
	}

	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = db.Select(
		q.Eq("H", reqData.H),
		q.Eq("SenderPeerId", user.PeerId),
		q.Eq("RecipientPeerId", reqData.RecipientPeerId),
		q.Eq("PropertyId", reqData.PropertyId),
		q.Eq("CurrState", dao.NS_Create),
		q.Eq("Amount", reqData.Amount)).
		First(rAndHInfo)

	//first send the request ?
	//是第一次发送请求？
	if err != nil {
		rAndHInfo.SenderPeerId = user.PeerId
		rAndHInfo.RecipientPeerId = reqData.RecipientPeerId
		rAndHInfo.PropertyId = reqData.PropertyId
		rAndHInfo.Amount = reqData.Amount
		rAndHInfo.Memo = reqData.Memo
		rAndHInfo.H = reqData.H
		rAndHInfo.CurrState = dao.NS_Create
		rAndHInfo.CreateAt = time.Now()
		rAndHInfo.CreateBy = user.PeerId
		bytes, err := json.Marshal(rAndHInfo)
		rAndHInfo.RequestHash = tool.SignMsgWithSha256(bytes)

		err = db.Save(rAndHInfo)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}

		currBlockHeight, err := rpcClient.GetBlockCount()
		if err != nil {
			return nil, "", errors.New("fail to get blockHeight ,please try again later")
		}

		// operate db
		pathInfo := &dao.HtlcPathInfo{}
		pathInfo.ChannelIdArr = channelIdArr
		pathInfo.HAndRInfoRequestHash = rAndHInfo.RequestHash
		pathInfo.H = rAndHInfo.H
		pathInfo.CurrState = dao.HtlcPathInfoState_Created
		pathInfo.BeginBlockHeight = currBlockHeight
		pathInfo.TotalStep = len(channelIdArr) * 2
		pathInfo.CurrStep = 0
		pathInfo.CreateBy = user.PeerId
		pathInfo.CreateAt = time.Now()
		err = db.Save(pathInfo)
		if err != nil {
			return nil, "", err
		}
	}

	msgHash := MessageService.saveMsg(user.PeerId, bob, rAndHInfo.RequestHash)
	if tool.CheckIsString(&msgHash) == false {
		return nil, "", errors.New("fail to save msgHash")
	}

	data = make(map[string]interface{})
	data["request_hash"] = msgHash
	data["h"] = rAndHInfo.H
	data["channelId"] = currChannel.ChannelId
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

	if tool.CheckIsString(&reqData.H) == false {
		return nil, "", errors.New("wrong h")
	}
	if tool.CheckIsString(&reqData.HAndRInfoRequestHash) == false {
		return nil, "", errors.New("wrong h_and_r_info_request_hash")
	}

	message, err := MessageService.getMsg(reqData.HAndRInfoRequestHash)
	if err != nil {
		return nil, "", errors.New("wrong request_hash")
	}
	if message.Receiver != user.PeerId {
		return nil, "", errors.New("you are not the operator")
	}
	reqData.HAndRInfoRequestHash = message.Data

	pathInfo := &dao.HtlcPathInfo{}
	err = db.Select(q.Eq("HAndRInfoRequestHash", reqData.HAndRInfoRequestHash)).First(pathInfo)
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

	currChannelIndex := pathInfo.CurrStep
	if currChannelIndex < 0 || currChannelIndex >= len(pathInfo.ChannelIdArr) {
		return nil, "", errors.New("err channel id")
	}
	currChannel := &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("Id", pathInfo.ChannelIdArr[currChannelIndex]),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).First(currChannel)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	targetUserId = currChannel.PeerIdB
	if user.PeerId == currChannel.PeerIdB {
		targetUserId = currChannel.PeerIdA
	}

	msgHash := MessageService.saveMsg(user.PeerId, targetUserId, pathInfo.HAndRInfoRequestHash)
	if tool.CheckIsString(&msgHash) == false {
		return nil, "", errors.New("fail to save msgHash")
	}

	data = make(map[string]interface{})
	data["request_hash"] = msgHash
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

	if tool.CheckIsString(&requestData.RequestHash) == false {
		return nil, "", errors.New("request_hash is empty")
	}

	message, err := MessageService.getMsg(requestData.RequestHash)
	if err != nil {
		return nil, "", errors.New("wrong request_hash")
	}
	if message.Receiver != user.PeerId {
		return nil, "", errors.New("you are not the operator")
	}
	requestData.RequestHash = message.Data

	// region check input data
	if requestData.Approval {
		if tool.CheckIsString(&requestData.ChannelAddressPrivateKey) == false {
			return nil, "", errors.New("channel_address_private_key is empty")
		}
		if tool.CheckIsString(&requestData.CurrRsmcTempAddressPubKey) == false {
			return nil, "", errors.New("curr_rsmc_temp_address_pub_key is empty")
		}
		if tool.CheckIsString(&requestData.CurrRsmcTempAddressPrivateKey) == false {
			return nil, "", errors.New("curr_rsmc_temp_address_private_key is empty")
		}
		if tool.CheckIsString(&requestData.CurrHtlcTempAddressPubKey) == false {
			return nil, "", errors.New("curr_htlc_temp_address_pub_key is empty")
		}
		if tool.CheckIsString(&requestData.CurrHtlcTempAddressPrivateKey) == false {
			return nil, "", errors.New("curr_htlc_temp_address_private_key is empty")
		}
		//if tool.CheckIsString(&requestData.CurrHtlcTempAddressHe1bOfHPubKey) == false {
		//	return nil, "", errors.New("curr_htlc_temp_address_he1b_ofh_pub_key is empty")
		//}
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

	pathInfo := &dao.HtlcPathInfo{}
	err = tx.Select(q.Eq("HAndRInfoRequestHash", requestData.RequestHash)).First(pathInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if pathInfo.CurrStep > int(pathInfo.TotalStep/2) {
		return nil, "", errors.New("error step")
	}

	if requestData.Approval == false && pathInfo.CurrStep == int(pathInfo.TotalStep/2) {
		err = errors.New("the receiver can not refuse")
		log.Println(err)
		return nil, "", err
	}

	//endregion

	currChannelIndex := pathInfo.CurrStep
	if currChannelIndex < 0 || currChannelIndex >= len(pathInfo.ChannelIdArr) {
		return nil, "", errors.New("err channel id")
	}
	currChannel := &dao.ChannelInfo{}
	err = tx.One("Id", pathInfo.ChannelIdArr[currChannelIndex], currChannel)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	//判断是否是正确的操作者
	if user.PeerId != currChannel.PeerIdA && user.PeerId != currChannel.PeerIdB {
		err = errors.New("you are not the operator")
		log.Println(err.Error())
		return nil, "", err
	}

	// region temp store data
	if requestData.Approval {
		//锁定所有借道通道
		if pathInfo.CurrStep == 0 && pathInfo.CurrState == dao.HtlcPathInfoState_Created {
			for _, id := range pathInfo.ChannelIdArr {
				channelInfo := &dao.ChannelInfo{}
				err := tx.One("Id", id, channelInfo)
				if err != nil {
					log.Println(err.Error())
					return nil, "", err
				}
				if channelInfo.CurrState != dao.ChannelState_CanUse {
					return nil, "", errors.New("channel is busy :" + channelInfo.ChannelId)
				}
				channelInfo.CurrState = dao.ChannelState_HtlcTx
				err = tx.Update(channelInfo)
				if err != nil {
					log.Println(err.Error())
					return nil, "", err
				}
			}
		}

		currNodePubKey := currChannel.PubKeyA
		if currChannel.PeerIdB == user.PeerId {
			currNodePubKey = currChannel.PubKeyB
		}
		_, err := tool.GetPubKeyFromWifAndCheck(requestData.ChannelAddressPrivateKey, currNodePubKey)
		if err != nil {
			return nil, "", errors.New("ChannelAddressPrivateKey is wrong")
		}
		tempAddrPrivateKeyMap[currNodePubKey] = requestData.ChannelAddressPrivateKey

		bobLatestCommitmentTx, err := getLatestCommitmentTx(currChannel.ChannelId, user.PeerId)
		if err == nil {
			if tool.CheckIsString(&requestData.LastTempAddressPrivateKey) == false {
				return nil, "", errors.New("last_temp_address_private_key is empty")
			}
			_, err := tool.GetPubKeyFromWifAndCheck(requestData.LastTempAddressPrivateKey, bobLatestCommitmentTx.RSMCTempAddressPubKey)
			if err != nil {
				return nil, "", errors.New("LastTempAddressPrivateKey is wrong")
			}
			tempAddrPrivateKeyMap[bobLatestCommitmentTx.RSMCTempAddressPubKey] = requestData.LastTempAddressPrivateKey
		}
		pathInfo.CurrRsmcTempPubKey = requestData.CurrRsmcTempAddressPubKey
		_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrRsmcTempAddressPrivateKey, requestData.CurrRsmcTempAddressPubKey)
		if err != nil {
			return nil, "", errors.New("CurrRsmcTempAddressPubKey is wrong")
		}

		pathInfo.CurrHtlcTempPubKey = requestData.CurrHtlcTempAddressPubKey
		_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrHtlcTempAddressPrivateKey, requestData.CurrHtlcTempAddressPubKey)
		if err != nil {
			return nil, "", errors.New("CurrHtlcTempAddressPubKey is wrong")
		}
		//pathInfo.CurrHtlcTempForHe1bOfHPubKey = requestData.CurrHtlcTempAddressHe1bOfHPubKey

		tempAddrPrivateKeyMap[pathInfo.CurrRsmcTempPubKey] = requestData.CurrRsmcTempAddressPrivateKey
		tempAddrPrivateKeyMap[pathInfo.CurrHtlcTempPubKey] = requestData.CurrHtlcTempAddressPrivateKey
		pathInfo.CurrState = dao.HtlcPathInfoState_Forward
	} else {
		pathInfo.CurrState = dao.HtlcPathInfoState_RefusedByInterNode
	}
	// endregion

	err = tx.Update(pathInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if currChannel.PeerIdB == user.PeerId {
		targetUser = currChannel.PeerIdA
	} else {
		targetUser = currChannel.PeerIdB
	}
	msgHash := MessageService.saveMsg(user.PeerId, targetUser, requestData.RequestHash)
	if tool.CheckIsString(&msgHash) == false {
		return nil, "", errors.New("fail to save msgHash")
	}

	data = make(map[string]interface{})
	data["channelId"] = currChannel.ChannelId
	data["sender"] = targetUser
	data["approval"] = requestData.Approval
	data["request_hash"] = msgHash
	return data, targetUser, nil
}

// -45 开始创建此次借道的交易
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

	message, err := MessageService.getMsg(requestData.RequestHash)
	if err != nil {
		return nil, "", errors.New("wrong request_hash")
	}
	if message.Receiver != user.PeerId {
		return nil, "", errors.New("you are not the operator")
	}
	requestData.RequestHash = message.Data

	pathInfo := dao.HtlcPathInfo{}
	err = db.Select(q.Eq("HAndRInfoRequestHash", requestData.RequestHash)).First(&pathInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if pathInfo.CurrStep > int(pathInfo.TotalStep/2) {
		return nil, "", errors.New("error step")
	}

	pathInfo.CurrStep += 1

	hAndRInfo := dao.HtlcRAndHInfo{}
	err = db.Select(q.Eq("RequestHash", pathInfo.HAndRInfoRequestHash)).First(&hAndRInfo)
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
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrRsmcTempAddressPrivateKey, requestData.CurrRsmcTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("CurrRsmcTempAddressPrivateKey is wrong")
	}

	if tool.CheckIsString(&requestData.CurrHtlcTempAddressPrivateKey) == false {
		err = errors.New("CurrHtlcTempAddressPrivateKey is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&requestData.CurrHtlcTempAddressPubKey) == false {
		err = errors.New("CurrHtlcTempAddressPubKey is empty")
		log.Println(err)
		return nil, "", err
	}
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrHtlcTempAddressPrivateKey, requestData.CurrHtlcTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("CurrHtlcTempAddressPrivateKey is wrong")
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
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.CurrHtlcTempAddressForHt1aPrivateKey, requestData.CurrHtlcTempAddressForHt1aPubKey)
	if err != nil {
		return nil, "", errors.New("CurrHtlcTempAddressForHt1aPrivateKey is wrong")
	}

	//if tool.CheckIsString(&requestData.CurrHtlcTempAddressForHed1aOfHPubKey) == false {
	//	err = errors.New("curr_htlc_temp_address_for_hed1a_ofh_pub_key is empty")
	//	log.Println(err)
	//	return nil, "", err
	//}
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
	currChannelIndex := pathInfo.CurrStep - 1
	if currChannelIndex < 0 || currChannelIndex >= len(pathInfo.ChannelIdArr) {
		return nil, "", errors.New("err channel id")
	}

	currChannelInfo := dao.ChannelInfo{}
	err = dbTx.One("Id", pathInfo.ChannelIdArr[currChannelIndex], &currChannelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	service.operationFlag.Lock()
	defer service.operationFlag.Unlock()

	var lastCommitmentTxOfSender = &dao.CommitmentTransaction{}
	err = db.Select(
		q.Eq("ChannelId", currChannelInfo.ChannelId),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").
		Reverse().
		First(lastCommitmentTxOfSender)
	if err == nil {
		if lastCommitmentTxOfSender != nil &&
			lastCommitmentTxOfSender.TxType == dao.CommitmentTransactionType_Htlc {
			return nil, "", errors.New("already created")
		}
	}

	currNodePubKey := ""
	otherSideChannelPubKey := ""
	//当前操作者是Alice Alice转账给Bob
	if user.PeerId == currChannelInfo.PeerIdA {
		targetUser = currChannelInfo.PeerIdB
		currNodePubKey = currChannelInfo.PubKeyA
		otherSideChannelPubKey = currChannelInfo.PubKeyB
	} else { //当前操作者是Bob Bob转账给Alice
		targetUser = currChannelInfo.PeerIdA
		currNodePubKey = currChannelInfo.PubKeyB
		otherSideChannelPubKey = currChannelInfo.PubKeyA
	}

	otherSideChannelPrivateKey := tempAddrPrivateKeyMap[otherSideChannelPubKey]
	if tool.CheckIsString(&otherSideChannelPrivateKey) == false {
		return nil, targetUser, errors.New("sender private key is miss,send 44 again")
	}

	tempAddrPrivateKeyMap[currNodePubKey] = requestData.ChannelAddressPrivateKey
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.ChannelAddressPrivateKey, currNodePubKey)
	if err != nil {
		return nil, "", errors.New("ChannelAddressPrivateKey is wrong")
	}
	defer delete(tempAddrPrivateKeyMap, currNodePubKey)
	_, err = tool.GetPubKeyFromWifAndCheck(requestData.LastTempAddressPrivateKey, lastCommitmentTxOfSender.RSMCTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("LastTempAddressPrivateKey is wrong")
	}

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = dbTx.Select(
		q.Eq("ChannelId", currChannelInfo.ChannelId),
		q.Eq("CurrState", dao.FundingTransactionState_Accept)).
		OrderBy("CreateAt").Reverse().
		First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// 创建上个交易的BR  begin
	//PeerIdA(概念中的Alice) 对上一次承诺交易的废弃
	err = htlcAliceAbortLastRsmcCommitmentTx(dbTx, currChannelInfo, user, *fundingTransaction, *requestData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	//PeerIdB(概念中的Bob) 对上一次承诺交易的废弃
	err = htlcBobAbortLastRsmcCommitmentTx(dbTx, currChannelInfo, user, *fundingTransaction, *requestData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	// 创建上个交易的BR  end

	//region 创建htlc相关的交易: Cna（1+3: Cna + toRsmc + toOther + toHtlc）+ 2(ht1a+htrd1a); cnb(1+3: Cnb + toRsmc + toOther + toHtlc) + 1(htd1b)

	///Cna Alice方(channel.PeerIdA)的交易
	commitmentTransactionOfA, err := service.htlcCreateAliceSideTxs(dbTx, currChannelInfo, user, *fundingTransaction, *requestData, pathInfo, hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(commitmentTransactionOfA)

	///Cnb Bob方(channel.PeerIdB)的交易
	commitmentTransactionOfB, err := service.htlcCreateBobSideTxs(dbTx, currChannelInfo, user, *fundingTransaction, *requestData, pathInfo, hAndRInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(commitmentTransactionOfB)

	//endregion

	pathInfo.CurrState = dao.HtlcPathInfoState_Backward
	err = dbTx.Update(&pathInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	err = dbTx.Commit()
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	msgHash := MessageService.saveMsg(user.PeerId, targetUser, hAndRInfo.RequestHash)
	data := make(map[string]interface{})
	data["h"] = hAndRInfo.H
	data["request_hash"] = msgHash
	return data, targetUser, nil
}

// 创建Alice方的htlc的承诺交易，rsmc的Rd
// 这里要做一个判断，作为这次交易的发起者，
// 如果PeerIdA是发起者，在这Cna的逻辑中创建HT1a和HED1a
// 如果PeerIdB是发起者，那么在Cna中就应该创建HTLC Time Delivery 1b(HED1b) 和HTLC Execution  1a(HE1b)
func (service *htlcForwardTxManager) htlcCreateAliceSideTxs(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen,
	pathInfo dao.HtlcPathInfo, hAndRInfo dao.HtlcRAndHInfo) (*dao.CommitmentTransaction, error) {

	owner := channelInfo.PeerIdA

	aliceIsSender := true
	if operator.PeerId == channelInfo.PeerIdB {
		aliceIsSender = false
	}

	//当前的承诺交易已经作废，因为上一步已经为他创建了各种BR
	var lastCommitmentATx = &dao.CommitmentTransaction{}
	err := tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("Owner", owner)).
		OrderBy("CreateAt").Reverse().
		First(lastCommitmentATx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create Cna tx
	commitmentTxInfo, err := htlcCreateCna(tx, channelInfo, operator, fundingTransaction, requestData, pathInfo, hAndRInfo, aliceIsSender, lastCommitmentATx, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create rsmc RDna tx
	rdTx, err := htlcCreateRDOfRsmc(
		tx, channelInfo, operator, fundingTransaction, requestData,
		pathInfo, aliceIsSender, commitmentTxInfo, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(rdTx)
	timeout := (pathInfo.TotalStep/2 - pathInfo.CurrStep + 1) * singleHopPerHopDuration
	// output2,htlc的后续交易
	if aliceIsSender { // 如是通道中的Alice转账给Bob，bob作为中间节点  创建HT1a
		// create ht1a
		htlcTimeoutTxA, err := createHtlcTimeoutTxForAliceSide(tx, owner, channelInfo, fundingTransaction, *commitmentTxInfo, requestData, timeout, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutTxA)
		// 继续创建htrd
		htrdTransaction, err := createHtlcRD(tx, channelInfo, operator, fundingTransaction, requestData, aliceIsSender, htlcTimeoutTxA, owner)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htrdTransaction)

	} else {
		// bob is sender, 如果alice得到R，构建Htlc Execution交易以及接下来的HERD交易，如果超时，bob的钱就应该超时赎回HTDelivery
		// 如果是bob转给alice，Alice作为中间商，作为当前通道的接收者
		// create HTD for bob  锁定了bob的钱，超时了，就应该给bob赎回
		privateKeys := make([]string, 0)
		privateKeys = append(privateKeys, requestData.CurrHtlcTempAddressPrivateKey)
		privateKeys = append(privateKeys, tempAddrPrivateKeyMap[channelInfo.PubKeyB])
		htlcTimeoutDeliveryTxB, err := createHtlcTimeoutDeliveryTx(tx, channelInfo.PeerIdB, channelInfo.AddressB, timeout, channelInfo, fundingTransaction, *commitmentTxInfo, privateKeys, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutDeliveryTxB)
	}
	// 当前创建的是alice这一方的交易，给alice的钱就需要rsmc的再锁定，给bob的钱就直接给出去
	// 情况1：HEDa的构建 创建一个三签地址，锁住支付给bob资金
	// 情况2：当前操作人是bob，而Alice 作为中间商，资金支付的接收者，需要把bob的tohtlc的钱用三签地址锁住 也即创建he1a
	hedna, err := htlcCreateExecutionDeliveryOfHForAlice(tx, aliceIsSender, pathInfo, owner, channelInfo, fundingTransaction.PropertyId, *commitmentTxInfo, requestData, operator, hAndRInfo.H)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(hedna)

	return commitmentTxInfo, nil
}

// 创建PeerIdA方的htlc的承诺交易，rsmc的Rd
// 这里要做一个判断，作为这次交易的发起者，
// 如果PeerIdA是发起者，在这Cna的逻辑中创建HT1a和HED1a
// 如果PeerIdB是发起者，那么在Cna中就应该创建HTLC Time Delivery 1b(HED1b) 和HTLC Execution  1a(HE1b)
func (service *htlcForwardTxManager) htlcCreateBobSideTxs(dbTx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen,
	pathInfo dao.HtlcPathInfo, hAndRInfo dao.HtlcRAndHInfo) (*dao.CommitmentTransaction, error) {

	owner := channelInfo.PeerIdB
	aliceIsSender := true
	if operator.PeerId == channelInfo.PeerIdB {
		aliceIsSender = false
	}

	//当前的承诺交易已经作废，因为上一步已经为他创建了各种BR
	var lastCommitmentBTx = &dao.CommitmentTransaction{}
	err := dbTx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("Owner", owner),
		q.Or(
			q.Eq("PeerIdA", owner),
			q.Eq("PeerIdB", owner))).
		OrderBy("CreateAt").Reverse().
		First(lastCommitmentBTx)

	if err != nil {
		lastCommitmentBTx = nil
	}

	// create Cnb dbTx
	commitmentTxInfo, err := htlcCreateCnb(dbTx, channelInfo, operator, fundingTransaction, requestData, pathInfo, hAndRInfo, aliceIsSender, lastCommitmentBTx, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create rsmc RDnb dbTx
	_, err = htlcCreateRDOfRsmc(
		dbTx, channelInfo, operator, fundingTransaction, requestData,
		pathInfo, aliceIsSender, commitmentTxInfo, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	timeout := (pathInfo.TotalStep/2 - pathInfo.CurrStep + 1) * singleHopPerHopDuration
	// htlc txs
	// output2,给htlc创建的交易，如何处理output2里面的钱
	if aliceIsSender {
		// 如是通道中的Alice转账给Bob，bob作为中间节点  创建HTD1b ，Alice的钱在超时的情况下，可以返回到Alice账号
		// 当前操作的请求者是Alice
		// create HTD1b 当超时的情况，Alice赎回自己的钱的交易
		privateKeys := make([]string, 0)
		privateKeys = append(privateKeys, requestData.CurrHtlcTempAddressPrivateKey)
		privateKeys = append(privateKeys, tempAddrPrivateKeyMap[channelInfo.PubKeyA])
		htlcTimeoutDeliveryTxB, err := createHtlcTimeoutDeliveryTx(dbTx, channelInfo.PeerIdA, channelInfo.AddressA, timeout, channelInfo, fundingTransaction, *commitmentTxInfo, privateKeys, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutDeliveryTxB)

	} else {
		// create ht1b  for bob, bob超时赎回自己的钱钱
		htlcTimeoutTxB, err := createHtlcTimeoutTxForBobSide(dbTx, owner, channelInfo, fundingTransaction, *commitmentTxInfo, requestData, timeout, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutTxB)

		// 继续创建htrd
		htrdTransaction, err := createHtlcRD(dbTx, channelInfo, operator, fundingTransaction, requestData, aliceIsSender, htlcTimeoutTxB, owner)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htrdTransaction)
	}

	// 当前创建的是Bob这一方的交易，给bob的钱就需要rsmc的再锁定，给alice的钱就直接给出去
	// 情况1：aliceIsSender = true 当前操作人是alice，而bob 作为中间商，资金支付的接收者，需要把to htlc的钱用三签地址锁住 也即创建he1b 因为
	// 情况2：aliceIsSender = false  当前操作人是bob ,alice是中间商，Alice是资金接收者，bob这边就应该创建Hed1b
	hena, err := htlcCreateExecutionDeliveryOfHForBob(dbTx, aliceIsSender, owner, pathInfo, channelInfo, fundingTransaction.PropertyId, *commitmentTxInfo, requestData, operator, hAndRInfo.H)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(hena)

	return commitmentTxInfo, nil
}
