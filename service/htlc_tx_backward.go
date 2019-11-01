package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"log"
	"sync"
)

type htlcBackwardTxManager struct {
	operationFlag sync.Mutex
}

// HTLC Reverse pass the R (Preimage R)
var HtlcBackwardTxService htlcBackwardTxManager

// SendRToPreviousNode
//
// Process type -46: Send R to Previous Node (middleman).
//  * R is <Preimage_R>
func (service *htlcBackwardTxManager) SendRToPreviousNode(msgData string,
	user bean.User) (data map[string]interface{}, previousNode string, err error) {

	// region Parse data inputed from websocket client of sender.
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcSendR{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	// endregion

	// region Check data inputed from websocket client of sender.
	if tool.CheckIsString(&reqData.RequestHash) == false {
		err = errors.New("empty request_hash")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("channel_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.CurrHtlcTempAddressForCnbPrivateKey) == false {
		err = errors.New("curr_htlc_temp_address_for_cnb_private_key is empty")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.CurrHtlcTempAddressForHE1bPubKey) == false {
		err = errors.New("curr_htlc_temp_address_for_he1b_pub_key is empty")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.CurrHtlcTempAddressForHE1bPrivateKey) == false {
		err = errors.New("curr_htlc_temp_address_for_he1b_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	// endregion

	// region Check out if the input R is correct.
	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = db.Select(
		q.Eq("RequestHash", reqData.RequestHash),
		q.Eq("R", reqData.R), // R from websocket client of sender
		q.Eq("CurrState", dao.NS_Finish)).First(rAndHInfo)

	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	// endregion

	// region Get peerId of previous node.
	htlcSingleHopPathInfo := dao.HtlcSingleHopPathInfo{}
	err = db.Select(q.Eq("HAndRInfoRequestHash",
		reqData.RequestHash)).First(&htlcSingleHopPathInfo)

	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// Currently solution is Alice to Bob to Carol.
	if htlcSingleHopPathInfo.CurrStep < 2 {
		return nil, "", errors.New("The transfer H has not completed yet.")
	} else if htlcSingleHopPathInfo.CurrStep > 3 {
		return nil, "", errors.New("The transfer R has completed.")
	}

	// If CurrStep = 2, that indicate the transfer H has completed.
	currChannelIndex := htlcSingleHopPathInfo.TotalStep - htlcSingleHopPathInfo.CurrStep - 1
	if currChannelIndex < -1 || currChannelIndex > len(htlcSingleHopPathInfo.ChannelIdArr) {
		return nil, "", errors.New("err channel id")
	}

	currChannel := &dao.ChannelInfo{}
	err = db.One("Id", htlcSingleHopPathInfo.ChannelIdArr[currChannelIndex], currChannel)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if currChannel.PeerIdA != user.PeerId && currChannel.PeerIdB != user.PeerId {
		return nil, "", errors.New("error user.")
	}

	if user.PeerId == currChannel.PeerIdA {
		previousNode = currChannel.PeerIdB
	} else {
		previousNode = currChannel.PeerIdA
	}

	// Transfer H or R increase step.
	htlcSingleHopPathInfo.CurrStep += 1
	// endregion

	// region Save private key to memory.
	//
	// PeerIdA of the channel is the sender of transfer R.
	// ChannelAddressPrivateKey is a private key of the sender.
	//
	// Example: When Bob transfer R to Alice, Bob is the sender.
	// Alice create HED1a need the private key of Bob for sign that.
	if user.PeerId == currChannel.PeerIdA {
		tempAddrPrivateKeyMap[currChannel.PubKeyA] = reqData.ChannelAddressPrivateKey
	} else { // PeerIdB is the sender of transfer R.
		tempAddrPrivateKeyMap[currChannel.PubKeyB] = reqData.ChannelAddressPrivateKey
	}

	tempAddrPrivateKeyMap[reqData.CurrHtlcTempAddressForHE1bPubKey] = reqData.CurrHtlcTempAddressForHE1bPrivateKey

	// Save pubkey to database.
	htlcSingleHopPathInfo.BobCurrHtlcTempForHt1bPubKey = reqData.CurrHtlcTempAddressForHE1bPubKey
	err = db.Update(htlcSingleHopPathInfo)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	// endregion

	// Generate response message.
	// If no error, the response data is displayed in websocket client of previous node.
	// Otherwise, it is displayed in websocket client of myself.
	responseData := make(map[string]interface{})
	responseData["id"] = rAndHInfo.Id
	responseData["request_hash"] = rAndHInfo.RequestHash
	responseData["r"] = rAndHInfo.R

	return responseData, previousNode, nil
}

// CheckRAndCreateCTxs
//
// Process type -47: Middleman node Check out if R is correct 
// and create commitment transactions.
//  * R is <Preimage_R>
func (service *htlcBackwardTxManager) CheckRAndCreateCTxs(msgData string, user bean.User) (
	data map[string]interface{}, targetUser string, err error) {

	// region Parse data inputed from websocket client of middleman node.
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcCheckRAndCreateTx{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	// endregion

	// region Check data inputed from websocket client of  middleman node.
	if tool.CheckIsString(&reqData.RequestHash) == false {
		err = errors.New("empty request_hash")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("channel_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.CurrHtlcTempAddressForCnaPrivateKey) == false {
		err = errors.New("curr_htlc_temp_address_for_cna_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	// endregion

	// region Check out if the request hash is correct.
	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = db.Select(
		q.Eq("RequestHash", reqData.RequestHash),
		q.Eq("CurrState", dao.NS_Finish)).First(rAndHInfo)

	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	// endregion
	
	// region Create HED1a commitment transaction. 
	// launch database transaction, if anything goes wrong, roll back.
	dbTx, err := db.Begin(true)
	if err != nil {
		return nil, "", err
	}
	defer dbTx.Rollback()

	// prepare the data
	htlcSingleHopPathInfo := dao.HtlcSingleHopPathInfo{}
	err = db.Select(q.Eq("HAndRInfoRequestHash",
		reqData.RequestHash)).First(&htlcSingleHopPathInfo)

	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	
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

	// Save private key to memory.
	//
	// PeerIdA of the channel is the creator of commitment transaction.
	// ChannelAddressPrivateKey is a private key of the creator.
	//
	// Example: When Bob transfer R to Alice has completed,
	// Alice begin create HED1a, HE1b, HERD1b. So, Alice is the creator.
	// Create HE1b, HERD1b need private key of Alice to sign that.
	if user.PeerId == channelInfo.PeerIdA {
		targetUser = channelInfo.PeerIdB
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = reqData.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	} else {  // PeerIdB is the creator of commitment transaction.
		targetUser = channelInfo.PeerIdA
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = reqData.ChannelAddressPrivateKey
		defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	}

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = dbTx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId), 
		q.Eq("CurrState", dao.FundingTransactionState_Accept)).
		OrderBy("CreateAt").Reverse().First(fundingTransaction)
		
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	// endregion



	// data = make(map[string]interface{})
	// data["approval"] = requestData.Approval
	// data["request_hash"] = requestData.RequestHash
	// return data, rAndHInfo.SenderPeerId, nil

	// 只有每个通道的转账发送方才能去创建关于R的交易
	// 如果转账发送方是PeerIdA（Alice），也就是Alice转账给bob，那就创建HED1a:Alice这边直接给钱bob，不需要时间锁定;HE1b,HERD1b，Bob因为是收款方，他自己的钱需要RSMC的方式锁在通道
	// 如果转账发送方是PeerIdB（Bob），也就是Bob转账给Alice，那就创建HE1a,HERD1a：Alice作为收款方，她得到的钱就需要RSMC锁定;HED1b：bob是发送方，他这边给Alice的钱是不需要锁定
	// 锁定的，就是自己的钱，就是给自己设定限制，给对方承诺

	return nil, "", nil
}
