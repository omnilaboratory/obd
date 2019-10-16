package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"log"
	"strconv"
	"sync"
	"time"
)

type htlcHMessageManager struct {
	operationFlag sync.Mutex
}

var HtlcHMessageService htlcHMessageManager

// DealHtlcRequest
//
// Process type -40: Alice start a request to transfer to Carol.
func (service *htlcHMessageManager) DealHtlcRequest(jsonData string, creator *bean.User) (data *bean.HtlcHRespond, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty json data")
	}
	htlcHRequest := &bean.HtlcHRequest{}
	err = json.Unmarshal([]byte(jsonData), htlcHRequest)
	if err != nil {
		return nil, err
	}

	// Check out if the HTLC can be launched.
	err = checkIfHtlcCanBeLaunched(creator, htlcHRequest)
	if err != nil {  // CAN NOT launch HTLC.
		return nil, err
	}

	// Record the request data to database.
	tx, err := db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()
	createRandHInfo := &dao.HtlcCreateRandHInfo{}
	createRandHInfo.SenderPeerId = creator.PeerId
	createRandHInfo.RecipientPeerId = htlcHRequest.RecipientPeerId
	createRandHInfo.PropertyId = htlcHRequest.PropertyId
	createRandHInfo.Amount = htlcHRequest.Amount
	createRandHInfo.CurrState = dao.NS_Create
	createRandHInfo.CreateAt = time.Now()
	createRandHInfo.CreateBy = creator.PeerId
	err = tx.Save(createRandHInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	bytes, err := json.Marshal(createRandHInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	createRandHInfo.RequestHash = msgHash
	err = tx.Update(createRandHInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	data = &bean.HtlcHRespond{}
	data.PropertyId = htlcHRequest.PropertyId
	data.Amount = htlcHRequest.Amount
	data.RequestHash = msgHash

	return data, nil
}

func (service *htlcHMessageManager) DealHtlcResponse(jsonData string, user *bean.User) (data interface{}, senderPeerId *string, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, nil, errors.New("empty json data")
	}
	htlcHRespond := &bean.HtlcHRespond{}
	err = json.Unmarshal([]byte(jsonData), htlcHRespond)
	if err != nil {
		return nil, nil, err
	}

	createRandHInfo := &dao.HtlcCreateRandHInfo{}
	err = db.Select(q.Eq("RequestHash", htlcHRespond.RequestHash), q.Eq("CurrState", dao.NS_Create)).First(createRandHInfo)
	if err != nil {
		return nil, nil, err
	}

	createRandHInfo.CurrState = dao.NS_Refuse
	if htlcHRespond.Approval {
		s, _ := tool.RandBytes(32)
		temp := append([]byte(createRandHInfo.RequestHash), s...)
		r := tool.SignMsgWithSha256(temp)
		h := tool.SignMsgWithSha256([]byte(r))

		createRandHInfo.R = r
		createRandHInfo.H = h
		createRandHInfo.CurrState = dao.NS_Finish
	}

	createRandHInfo.SignAt = time.Now()
	createRandHInfo.SignBy = user.PeerId
	err = db.Update(createRandHInfo)
	if err != nil {
		return nil, &createRandHInfo.SenderPeerId, err
	}

	responseData := make(map[string]interface{})
	responseData["id"] = createRandHInfo.Id
	responseData["request_hash"] = htlcHRespond.RequestHash
	responseData["approval"] = htlcHRespond.Approval
	if htlcHRespond.Approval {
		responseData["h"] = createRandHInfo.H
	}
	return responseData, &createRandHInfo.SenderPeerId, nil
}

func (service *htlcHMessageManager) GetHtlcCreatedRandHInfoList(user *bean.User) (data interface{}, err error) {
	var createRandHInfoList []dao.HtlcCreateRandHInfo
	err = db.Select(q.Eq("CreateBy", user.PeerId)).Find(&createRandHInfoList)
	if err != nil {
		return nil, err
	}
	for _, item := range createRandHInfoList {
		item.R = ""
	}
	return createRandHInfoList, nil
}

func (service *htlcHMessageManager) GetHtlcCreatedRandHInfoItemById(msgData string, user *bean.User) (data interface{}, err error) {
	id, err := strconv.Atoi(msgData)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	var createRandHInfo dao.HtlcCreateRandHInfo
	err = db.Select(q.Eq("Id", id), q.Eq("CreateBy", user.PeerId)).First(&createRandHInfo)
	if err != nil {
		return nil, err
	}
	createRandHInfo.R = ""
	return createRandHInfo, nil
}

func (service *htlcHMessageManager) GetHtlcSignedRandHInfoList(user *bean.User) (data interface{}, err error) {
	var createRandHInfoList []dao.HtlcCreateRandHInfo
	err = db.Select(q.Eq("RecipientPeerId", user.PeerId), q.Eq("SignBy", user.PeerId)).Find(&createRandHInfoList)
	if err != nil {
		return nil, err
	}
	return createRandHInfoList, nil
}

func (service *htlcHMessageManager) GetHtlcSignedRandHInfoItem(msgData string, user *bean.User) (data interface{}, err error) {
	id, err := strconv.Atoi(msgData)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	var createRandHInfo dao.HtlcCreateRandHInfo
	err = db.Select(q.Eq("Id", id), q.Eq("SignBy", user.PeerId)).First(&createRandHInfo)
	if err != nil {
		return nil, err
	}
	return createRandHInfo, nil
}

// checkIfHtlcCanBeLaunched
//
// Check out if the HTLC can be launched.
// There is four case that CAN NOT launch HTLC.
//
// Case 1. There is a direct channel between Alice and Carol.
//
// Case 2. There is NOT a middleman, CAN NOT launch HTLC.
//
// Case 3. The middleman is NOT online, CAN NOT launch HTLC.
//
// Case 4. There is a online middleman, but NOT enough balance in channel.
func checkIfHtlcCanBeLaunched(creator *bean.User, htlcHRequest *bean.HtlcHRequest) error {

	// Case 1. There is a direct channel between Alice and Carol.
	// Get all channels of Alice.
	channelsOfAlice := getAllChannels(creator.PeerId)
	if len(channelsOfAlice) == 0 {
		return errors.New("The sender has no channel.")
	}
	
	// Check out if there is a direct channel between Alice and Carol.
	// If Yes, NO need to launch HTLC.
	for _, item := range channelsOfAlice {
		if item.PeerIdA == creator.PeerId && item.PeerIdB == htlcHRequest.RecipientPeerId {
			return errors.New("There is a direct channel between Alice and Carol.")
		}

		if item.PeerIdB == creator.PeerId && item.PeerIdA == htlcHRequest.RecipientPeerId {
			return errors.New("There is a direct channel between Alice and Carol.")
		}
	}

	// Case 2. There is NOT a middleman, CAN NOT launch HTLC.
	// Get all channels of Carol.
	channelsOfCarol := getAllChannels(htlcHRequest.RecipientPeerId)
	if len(channelsOfCarol) == 0 {
		return errors.New("The recipient has no channel.")
	}

	var hasMiddleman = false
	arrMiddleman := make([]string, 0)  // Save the middlemen.
	
	// Looking for a middleman.
	for _, itemAlice := range channelsOfAlice {
		if itemAlice.PeerIdA == creator.PeerId { // PeerIdA is Alice.
			for _, itemCarol := range channelsOfCarol {
				// PeerIdB of Alice's channel maybe a middleman.
				if itemAlice.PeerIdB == itemCarol.PeerIdA || itemAlice.PeerIdB == itemCarol.PeerIdB {
					arrMiddleman = append(arrMiddleman, itemAlice.PeerIdB)
					hasMiddleman = true
					break
				}
			}
		}
		
		if itemAlice.PeerIdB == creator.PeerId { // PeerIdB is Alice.
			for _, itemCarol := range channelsOfCarol {
				// PeerIdA of Alice's channel maybe a middleman.
				if itemAlice.PeerIdA == itemCarol.PeerIdA || itemAlice.PeerIdA == itemCarol.PeerIdB {
					arrMiddleman = append(arrMiddleman, itemAlice.PeerIdA)
					hasMiddleman = true
					break
				}
			}
		}
	}

	// There is NOT a middleman.
	if hasMiddleman == false {
		return errors.New("There is NOT a middleman, CAN NOT launch HTLC.")
	}
	
	// Case 3. The middleman is NOT online, CAN NOT launch HTLC.
	var allMiddlemanNotOnline = true
	arrOnlineMiddleman := make([]string, 0)  // Save the online middlemen.
	
	// Find all online middlemen.
	for _, item := range arrMiddleman {
		if FindUserIsOnline(item) == nil { // The middleman is online.
			// Save the middleman for using in future.
			arrOnlineMiddleman    = append(arrOnlineMiddleman, item)
			allMiddlemanNotOnline = false
		}
	}
	
	// The middleman is NOT online.
	if allMiddlemanNotOnline == true {
		return errors.New("The middleman is NOT online, CAN NOT launch HTLC.")
	}

	// Case 4. There is a online middleman, but NOT enough balance in channel.
	// Save the qualified middleman.
	arrQualifiedMiddleman := make([]string, 0)

	// Looking for a qualified middleman.
	for _, middleman := range arrOnlineMiddleman {

		var channelAlice, channelCarol *dao.ChannelInfo
		var commitmentTxAlice, commitmentTxCarol *dao.CommitmentTransaction
		var err error

		// Get channel between Alice and Middleman. creator.PeerId is Alice
		channelAlice, err = GetChannelInfoByTwoPeerID(creator.PeerId, middleman)
		if err != nil {
			return err
		}

		// Get channel between Carol and Middleman. htlcHRequest.RecipientPeerId is Carol
		channelCarol, err = GetChannelInfoByTwoPeerID(htlcHRequest.RecipientPeerId, middleman)
		if err != nil {
			return err
		}

		// Get Alice's balance in channel between Alice and Middleman.
		commitmentTxAlice, err = getLatestCommitmentTx(channelAlice.ChannelId, creator.PeerId)
		if err != nil {
			return err
		}

		// Get Middleman's balance in channel between Carol and Middleman.
		commitmentTxCarol, err = getLatestCommitmentTx(channelCarol.ChannelId, middleman)
		if err != nil {
			return err
		}

		// If there is enough balance that Alice transfer to Middleman and 
		// Middleman transfer to Carol, then record the Middleman.
		if commitmentTxAlice.AmountToRSMC >= (htlcHRequest.Amount + tool.GetHtlcFee()) && 
		   commitmentTxCarol.AmountToRSMC >= (htlcHRequest.Amount + tool.GetHtlcFee()) {

			arrQualifiedMiddleman = append(arrQualifiedMiddleman, middleman)
		}
	}

	// There is NOT a qualified middleman.
	if len(arrQualifiedMiddleman) == 0 {
		return errors.New("NOT enough balance in channels, CAN NOT launch HTLC.")
	}

	return nil
}

// GetChannelInfoByTwoPeerID
//
// Get a channel info by two peer ID.
func GetChannelInfoByTwoPeerID(peerIdA string, peerIdB string) (channelInfo *dao.ChannelInfo, err error) {
	channelInfo = &dao.ChannelInfo{}
	err = db.Select(q.Eq("CurrState", dao.ChannelState_Accept), 
		q.Or(q.And(q.Eq("PeerIdA", peerIdA), q.Eq("PeerIdB", peerIdB))), 
		q.And(q.Eq("PeerIdA", peerIdB), q.Eq("PeerIdB", peerIdA))).First(channelInfo)

	return channelInfo, err
}