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
	"time"
)

type htlcHMessageManager struct {
	operationFlag sync.Mutex
}

var HtlcHMessageService htlcHMessageManager

// DealHtlcRequest
//
// Process type -40: Alice start a request to transfer to Carol.
func (service *htlcHMessageManager) DealHtlcRequest(jsonData string,
	creator *bean.User) (data *bean.HtlcHRespond, err error) {

	//------------
	// ** We will launch a HTLC transfer for testing purpose. **
	// It tests Alice transfer money to Carol through Bob (middleman).
	//
	// a) There IS NOT a direct channel between Alice and Carol.
	// b) There is a direct channel between Alice and Bob.
	// c) There is a direct channel between Bob and Carol.
	//------------

	// [jsonData] is content inputed from [Alice] websocket client.
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty json data")
	}

	// Get [HtlcHRequest] struct object from [jsonData].
	htlcHRequest := &bean.HtlcHRequest{}
	err = json.Unmarshal([]byte(jsonData), htlcHRequest)
	if err != nil {
		return nil, err
	}

	// Check out if the HTLC can be launched.
	err = checkIfHtlcCanBeLaunched(creator, htlcHRequest)
	if err != nil { // CAN NOT launch HTLC.
		return nil, err
	}

	// Record the request data to database.
	tx, err := db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// If there is an error, roll back the database.
	defer tx.Rollback()

	// [HtlcRAndHInfo] struct object save the base data about HTLC flow.
	rAndHInfo := &dao.HtlcRAndHInfo{}
	rAndHInfo.SenderPeerId = creator.PeerId
	rAndHInfo.RecipientPeerId = htlcHRequest.RecipientPeerId
	rAndHInfo.PropertyId = htlcHRequest.PropertyId
	rAndHInfo.Amount = htlcHRequest.Amount
	rAndHInfo.CurrState = dao.NS_Create
	rAndHInfo.CreateAt = time.Now()
	rAndHInfo.CreateBy = creator.PeerId

	// Cache data. DO NOT write to database actually.
	err = tx.Save(rAndHInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Generate a request hash for HTLC Request every time.
	// The [RequestHash] is used to uniquely identify each HTLC request.
	bytes, err := json.Marshal(rAndHInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	rAndHInfo.RequestHash = msgHash

	// Update the cache data. DO NOT write to database actually.
	err = tx.Update(rAndHInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit transaction. Write to database actually.
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Generate response message.
	// If no error, the response data is displayed in websocket client of Carol.
	// Otherwise, it is displayed in websocket client of Alice.
	data = &bean.HtlcHRespond{}
	data.PropertyId = htlcHRequest.PropertyId
	data.Amount = htlcHRequest.Amount
	data.RequestHash = msgHash

	return data, nil
}

// DealHtlcResponse
//
// Process type -41: Carol response to transfer H to Alice.
//  * H is <hash_of_preimage_R>
func (service *htlcHMessageManager) DealHtlcResponse(jsonData string,
	user *bean.User) (data interface{}, senderPeerId *string, err error) {

	// [jsonData] is content inputed from [Carol] websocket client.
	if tool.CheckIsString(&jsonData) == false {
		return nil, nil, errors.New("empty json data")
	}

	htlcHRespond := &bean.HtlcHRespond{}
	err = json.Unmarshal([]byte(jsonData), htlcHRespond)
	if err != nil {
		return nil, nil, err
	}

	// [HtlcRAndHInfo] has saved to database in [Type -40].
	// So, get the object from database now.
	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = db.Select(q.Eq("RequestHash", htlcHRespond.RequestHash), q.Eq("CurrState", dao.NS_Create)).First(rAndHInfo)
	if err != nil {
		return nil, nil, err
	}

	rAndHInfo.CurrState = dao.NS_Refuse

	// Carol approved the request from Alice.
	if htlcHRespond.Approval {

		// Generate the R and H.
		// For temp solution currently, the R and H save to database.
		//  * R is <preimage_R>
		//  * H is <hash_of_preimage_R>

		s, _ := tool.RandBytes(32)
		temp := append([]byte(rAndHInfo.RequestHash), s...)
		temp = append(temp, user.PeerId...)

		r := tool.SignMsgWithRipemd160(temp)
		h := tool.SignMsgWithSha256([]byte(r))

		rAndHInfo.R = r
		rAndHInfo.H = h
		rAndHInfo.CurrState = dao.NS_Finish
	}

	rAndHInfo.SignAt = time.Now()
	rAndHInfo.SignBy = user.PeerId
	err = db.Update(rAndHInfo)
	if err != nil {
		return nil, &rAndHInfo.SenderPeerId, err
	}

	// Generate response message.
	// If no error, the response data is displayed in websocket client of Alice.
	// Otherwise, it is displayed in websocket client of Carol.
	responseData := make(map[string]interface{})
	responseData["id"] = rAndHInfo.Id
	responseData["request_hash"] = htlcHRespond.RequestHash
	responseData["approval"] = htlcHRespond.Approval
	if htlcHRespond.Approval {
		responseData["h"] = rAndHInfo.H
	}
	return responseData, &rAndHInfo.SenderPeerId, nil
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
	arrMiddleman := make([]string, 0) // Save the middlemen.

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
	arrOnlineMiddleman := make([]string, 0) // Save the online middlemen.

	// Find all online middlemen.
	for _, item := range arrMiddleman {
		if FindUserIsOnline(item) == nil { // The middleman is online.
			// Save the middleman for using in future.
			arrOnlineMiddleman = append(arrOnlineMiddleman, item)
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
		if commitmentTxAlice.AmountToRSMC >= (htlcHRequest.Amount+tool.GetHtlcFee()) &&
			commitmentTxCarol.AmountToRSMC >= (htlcHRequest.Amount+tool.GetHtlcFee()) {

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
		q.Or(q.And(q.Eq("PeerIdA", peerIdA), q.Eq("PeerIdB", peerIdB)),
			q.And(q.Eq("PeerIdA", peerIdB), q.Eq("PeerIdB", peerIdA)))).First(channelInfo)

	return channelInfo, err
}
