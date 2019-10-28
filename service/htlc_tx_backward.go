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

type htlcBackwardTxManager struct {
	operationFlag sync.Mutex
}

// HTLC Reverse pass the R (Preimage R)
var HtlcBackwardTxService htlcBackwardTxManager

// CarolSendRToBob
//
// Process type -46: Carol send R to Bob (middleman).
//  * R is <Preimage_R> 
func (service *htlcBackwardTxManager) CarolSendRToBob(msgData string, 
	user bean.User) (data map[string]interface{}, bob string, err error) {

	// region Parse data inputed from [Carol] websocket client.
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
	
	// region Check data inputed from websocket client of Carol.
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

	// Get Bob peerId.
	htlcSingleHopPathInfo := dao.HtlcSingleHopPathInfo{}
	err = db.Select(q.Eq("", reqData.RequestHash)).First(&htlcSingleHopPathInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	bob = htlcSingleHopPathInfo.InterNodePeerId

	// Check R
	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = db.Select(
		q.Eq("RequestHash", reqData.RequestHash), 
		q.Eq("R", reqData.R), // R from websocket client of Carol
		q.Eq("CreateBy", user.PeerId), // Carol
		q.Eq("CurrState", dao.NS_Finish)).First(rAndHInfo)

	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	// Generate response message.
	// If no error, the response data is displayed in websocket client of Bob.
	// Otherwise, it is displayed in websocket client of Carol.
	responseData := make(map[string]interface{})
	responseData["id"] = rAndHInfo.Id
	responseData["request_hash"] = rAndHInfo.RequestHash
	responseData["r"] = rAndHInfo.R

	return responseData, bob, nil
}

// SignGetR
//
// Process type -47: Bob (middleman) check out if R is correct.
//  * R is <Preimage_R> 
func (service *htlcBackwardTxManager) SignGetR(msgData string, user bean.User) (
	data map[string]interface{}, targetUser string, err error) {

	// if tool.CheckIsString(&msgData) == false {
	// 	return nil, "", errors.New("empty json data")
	// }

	
	// data = make(map[string]interface{})
	// data["approval"] = requestData.Approval
	// data["request_hash"] = requestData.RequestHash
	// return data, rAndHInfo.SenderPeerId, nil
}
