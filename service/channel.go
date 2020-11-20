package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"strings"
	"time"
)

type channelManager struct{}

var ChannelService = channelManager{}

// AliceOpenChannel init ChannelInfo
func (this *channelManager) AliceOpenChannel(msg bean.RequestMessage, user *bean.User) (openChannelInfo *bean.RequestOpenChannel, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New(enum.Tips_common_wrong + "msg.data")
	}

	reqData := &bean.SendChannelOpen{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		return nil, err
	}

	openChannelInfo = &bean.RequestOpenChannel{}
	openChannelInfo.FundingAddress, err = getAddressFromPubKey(reqData.FundingPubKey)
	if err != nil {
		return nil, err
	}

	openChannelInfo.ChainHash = config.Init_node_chain_hash
	openChannelInfo.TemporaryChannelId = bean.ChannelIdService.NextTemporaryChanID()
	openChannelInfo.FunderNodeAddress = P2PLocalPeerId
	openChannelInfo.FunderPeerId = user.PeerId
	openChannelInfo.FundingPubKey = reqData.FundingPubKey
	openChannelInfo.FunderAddressIndex = reqData.FunderAddressIndex
	openChannelInfo.IsPrivate = reqData.IsPrivate

	channelInfo := &dao.ChannelInfo{}
	channelInfo.RequestOpenChannel = *openChannelInfo
	channelInfo.PeerIdA = user.PeerId
	channelInfo.PeerIdB = msg.RecipientUserPeerId
	channelInfo.PubKeyA = reqData.FundingPubKey
	channelInfo.AddressA = openChannelInfo.FundingAddress
	channelInfo.CurrState = dao.ChannelState_Create
	channelInfo.CreateAt = time.Now()
	channelInfo.CreateBy = user.PeerId

	err = user.Db.Save(channelInfo)
	return openChannelInfo, err
}

// obd init ChannelInfo for Bob
func (this *channelManager) BeforeBobOpenChannelAtBobSide(msg string, user *bean.User) (err error) {
	if tool.CheckIsString(&msg) == false {
		return errors.New(enum.Tips_common_wrong + "msg")
	}

	aliceOpenChannelInfo := bean.RequestOpenChannel{}
	err = json.Unmarshal([]byte(msg), &aliceOpenChannelInfo)
	if err != nil {
		return err
	}

	channelInfo := &dao.ChannelInfo{}
	channelInfo.RequestOpenChannel = aliceOpenChannelInfo
	channelInfo.PeerIdA = aliceOpenChannelInfo.FunderPeerId
	channelInfo.PeerIdB = user.PeerId
	channelInfo.PubKeyA = aliceOpenChannelInfo.FundingPubKey
	channelInfo.AddressA = aliceOpenChannelInfo.FundingAddress
	channelInfo.CurrState = dao.ChannelState_Create
	channelInfo.CreateAt = time.Now()
	channelInfo.CreateBy = user.PeerId
	err = user.Db.Save(channelInfo)
	return err
}

func (this *channelManager) BobCheckChannelAddressExist(jsonData string, user *bean.User) (exist bool, err error) {
	reqData := &bean.SendSignOpenChannel{}
	err = json.Unmarshal([]byte(jsonData), &reqData)

	if err != nil {
		return false, err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		return false, errors.New(enum.Tips_common_wrong + " temporary_channel_id")
	}

	channelInfo := &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("PeerIdB", user.PeerId),
		q.Eq("CurrState", dao.ChannelState_Create)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return false, errors.New(enum.Tips_channel_notFoundChannelInCreate + reqData.TemporaryChannelId)
	}

	if channelInfo.PeerIdB != user.PeerId {
		return false, errors.New(enum.Tips_rsmc_notTargetUser)
	}

	channelInfo.PubKeyB = reqData.FundingPubKey
	multiSig, err := rpcClient.CreateMultiSig(2, []string{channelInfo.PubKeyA, channelInfo.PubKeyB})
	if err != nil {
		log.Println(err)
		return false, err
	}
	channelAddress := gjson.Get(multiSig, "address").String()

	existAddress := false
	result, err := rpcClient.ListReceivedByAddress(channelAddress)
	if err == nil {
		array := gjson.Parse(result).Array()
		if len(array) > 0 {
			existAddress = true
		}
	}
	count, _ := user.Db.Select(q.Eq("ChannelAddress", channelAddress)).Count(&dao.ChannelInfo{})
	if count > 0 {
		existAddress = true
	}
	return existAddress, nil
}

func (this *channelManager) BobAcceptChannel(msg bean.RequestMessage, user *bean.User) (channelInfo *dao.ChannelInfo, err error) {
	reqData := &bean.SendSignOpenChannel{}
	err = json.Unmarshal([]byte(msg.Data), &reqData)

	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&reqData.TemporaryChannelId) == false {
		return nil, errors.New(enum.Tips_common_wrong + "temporary_channel_id")
	}

	bobFundingAddress := ""
	if reqData.Approval {
		if tool.CheckIsString(&reqData.FundingPubKey) == false {
			return nil, errors.New(enum.Tips_common_wrong + "funding_pubkey")
		}

		bobFundingAddress, err = getAddressFromPubKey(reqData.FundingPubKey)
		if err != nil {
			return nil, err
		}
	}

	channelInfo = &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("TemporaryChannelId", reqData.TemporaryChannelId),
		q.Eq("PeerIdB", user.PeerId),
		q.Eq("CurrState", dao.ChannelState_Create)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, errors.New(enum.Tips_channel_notFoundChannelInCreate + reqData.TemporaryChannelId)
	}

	if channelInfo.PeerIdB != user.PeerId {
		return nil, errors.New(enum.Tips_channel_notThePeerIdB)
	}

	if channelInfo.PeerIdA != msg.RecipientUserPeerId {
		return nil, errors.New(enum.Tips_common_wrong + msg.RecipientUserPeerId)
	}

	if reqData.Approval {
		channelInfo.PubKeyB = reqData.FundingPubKey
		channelInfo.FundeeAddressIndex = reqData.FundeeAddressIndex
		channelInfo.AddressB = bobFundingAddress
		multiSig, err := rpcClient.CreateMultiSig(2, []string{channelInfo.PubKeyA, channelInfo.PubKeyB})
		if err != nil {
			log.Println(err)
			return nil, err
		}

		channelAddress := gjson.Get(multiSig, "address").String()

		existAddress := false
		result, err := rpcClient.ListReceivedByAddress(channelAddress)
		if err == nil {
			array := gjson.Parse(result).Array()
			if len(array) > 0 {
				existAddress = true
			}
		}

		count, _ := user.Db.Select(q.Eq("ChannelAddress", channelAddress)).Count(&dao.ChannelInfo{})
		if count > 0 {
			existAddress = true
		}
		if existAddress == false {
			channelInfo.ChannelAddress = gjson.Get(multiSig, "address").String()
			channelInfo.ChannelAddressRedeemScript = gjson.Get(multiSig, "redeemScript").String()

			addrInfoStr, err := rpcClient.GetAddressInfo(channelInfo.ChannelAddress)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			channelInfo.ChannelAddressScriptPubKey = gjson.Parse(addrInfoStr).Get("scriptPubKey").String()
			channelInfo.CurrState = dao.ChannelState_WaitFundAsset
		} else {
			return nil, errors.New(enum.Tips_channel_changePubkeyForChannel + reqData.FundingPubKey)
		}
	} else {
		channelInfo.CurrState = dao.ChannelState_OpenChannelRefuse
		channelInfo.RefuseReason = user.PeerId + " do not agree with it"
	}

	channelInfo.AcceptAt = time.Now()
	err = user.Db.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return channelInfo, err
}

//当bob操作完，发送信息到Alice所在的obd，obd处理先从bob得到发给alice的信息，然后再发给Alice的轻客户端
func (this *channelManager) AfterBobAcceptChannelAtAliceSide(jsonData string, user *bean.User) (outputData interface{}, err error) {
	bobChannelInfo := &dao.ChannelInfo{}
	err = json.Unmarshal([]byte(jsonData), &bobChannelInfo)
	if err != nil {
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("TemporaryChannelId", bobChannelInfo.TemporaryChannelId),
		q.Eq("PeerIdA", user.PeerId),
		q.Eq("PeerIdB", bobChannelInfo.PeerIdB),
		q.Eq("CurrState", dao.ChannelState_Create)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, errors.New(enum.Tips_channel_notFoundChannelInCreate + bobChannelInfo.TemporaryChannelId)
	}

	if bobChannelInfo.CurrState == dao.ChannelState_WaitFundAsset {
		channelInfo.PubKeyB = bobChannelInfo.PubKeyB
		channelInfo.AddressB = bobChannelInfo.AddressB
		channelInfo.ChannelAddress = bobChannelInfo.ChannelAddress
		channelInfo.ChannelAddressRedeemScript = bobChannelInfo.ChannelAddressRedeemScript
		channelInfo.ChannelAddressScriptPubKey = bobChannelInfo.ChannelAddressScriptPubKey
		channelInfo.CurrState = dao.ChannelState_WaitFundAsset
	} else {
		channelInfo.CurrState = dao.ChannelState_OpenChannelRefuse
		channelInfo.RefuseReason = bobChannelInfo.RefuseReason
	}
	channelInfo.AcceptAt = time.Now()
	err = user.Db.Update(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return channelInfo, err
}

type ChannelVO struct {
	TemporaryChannelId string           `json:"temporary_channel_id"`
	IsPrivate          bool             `json:"is_private"`
	ChannelId          string           `json:"channel_id"`
	ChannelAddress     string           `json:"channel_address"`
	PropertyId         int64            `json:"property_id"`
	CurrState          dao.ChannelState `json:"curr_state"`
	PeerIdA            string           `json:"peer_ida"`
	PeerIdB            string           `json:"peer_idb"`
	BtcFundingTimes    int              `json:"btc_funding_times"`
	BtcAmount          float64          `json:"btc_amount"`
	AssetAmount        float64          `json:"asset_amount"`
	BalanceA           float64          `json:"balance_a"`
	BalanceB           float64          `json:"balance_b"`
	BalanceHtlc        float64          `json:"balance_htlc"`
	CreateAt           time.Time        `json:"create_at"`
}

type pageVO struct {
	Data       interface{} `json:"data"`
	PageNum    int         `json:"pageNum"`
	PageSize   int         `json:"pageSize"`
	TotalCount int         `json:"totalCount"`
	TotalPage  int         `json:"totalPage"`
}

// AssetFundingAllItem
func (this *channelManager) AllItem(jsonData string, user bean.User) (data *pageVO, err error) {
	data = &pageVO{}
	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()
	pageIndex := gjson.Get(jsonData, "page_index").Int()
	if pageIndex <= 0 {
		pageIndex = 1
	}
	pageSize := gjson.Get(jsonData, "page_size").Int()
	if pageSize <= 0 {
		pageSize = 10
	}
	skip := (pageIndex - 1) * pageSize

	var infos []dao.ChannelInfo
	err = tx.Select(
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().Skip(int(skip)).Limit(int(pageSize)).
		Find(&infos)

	tempCount, err := tx.Select(
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		Count(&dao.ChannelInfo{})
	if err != nil {
		return nil, err
	}

	totalPage := int(tempCount) / int(pageSize)
	if int(tempCount)%int(pageSize) != 0 {
		totalPage += 1
	}

	data.TotalPage = totalPage
	data.TotalCount = tempCount
	data.PageNum = int(pageIndex)
	data.PageSize = int(pageSize)

	if infos != nil {
		items := make([]ChannelVO, 0)
		for _, info := range infos {
			item := ChannelVO{}
			item.TemporaryChannelId = info.TemporaryChannelId
			item.ChannelId = info.ChannelId
			item.ChannelAddress = info.ChannelAddress
			item.IsPrivate = info.IsPrivate
			item.CurrState = info.CurrState
			item.PropertyId = info.PropertyId
			item.AssetAmount = info.Amount
			item.BtcAmount = info.BtcAmount
			item.PeerIdA = info.PeerIdA
			item.PeerIdB = info.PeerIdB
			item.CreateAt = info.CreateAt
			item.BtcFundingTimes = 3
			if item.CurrState == dao.ChannelState_Create {
				result, err := rpcClient.ListReceivedByAddress(info.ChannelAddress)
				if err == nil {
					if len(gjson.Parse(result).Array()) > 0 {
						btcFundingTimes := len(gjson.Parse(result).Array()[0].Get("txids").Array())
						if btcFundingTimes > 3 {
							btcFundingTimes = 3
						}
						item.BtcFundingTimes = btcFundingTimes
					}
				}
			}

			if info.CurrState >= dao.ChannelState_CanUse {
				commitmentTxInfo, _ := getLatestCommitmentTxUseDbTx(tx, info.ChannelId, user.PeerId)
				if commitmentTxInfo.Id > 0 {
					item.BalanceA = commitmentTxInfo.AmountToRSMC
					item.BalanceB = commitmentTxInfo.AmountToCounterparty
					item.BalanceHtlc = commitmentTxInfo.AmountToHtlc
				}
			}
			items = append(items, item)
		}
		data.Data = items
	}
	_ = tx.Commit()
	return data, err
}

// AssetFundingTotalCount
func (this *channelManager) TotalCount(user bean.User) (count int, err error) {
	return user.Db.Select(
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		Count(&dao.ChannelInfo{})
}

// GetChannelByTemporaryChanId
func (this *channelManager) GetChannelByTemporaryChanId(jsonData string, user bean.User) (node *dao.ChannelInfo, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New(enum.Tips_common_empty + "temporary_channel_id")
	}
	node = &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("TemporaryChannelId", jsonData)).
		First(node)
	return node, err
}

// DelChannelByTemporaryChanId
func (this *channelManager) DelChannelByTemporaryChanId(jsonData string, user bean.User) (node *dao.ChannelInfo, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New(enum.Tips_common_empty + "temporary_channel_id")
	}
	node = &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("TemporaryChannelId", jsonData)).
		First(node)
	if tool.CheckIsString(&node.ChannelId) {
		return nil, errors.New(enum.Tips_channel_cannotDelChannel)
	}
	if err == nil {
		err = user.Db.DeleteStruct(node)
	}
	return node, err
}

func (this *channelManager) GetChannelInfoByChannelId(jsonData string, user bean.User) (info *dao.ChannelInfo, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New(enum.Tips_common_empty + "ChannelId")
	}

	info = &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("ChannelId", jsonData),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(info)
	return info, err
}

func (this *channelManager) GetChannelInfoById(jsonData string, user bean.User) (info *dao.ChannelInfo, err error) {
	id, err := strconv.Atoi(jsonData)
	if err != nil {
		return nil, err
	}
	info = &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("Id", id),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(info)
	return info, err
}

//请求关闭通道
func (this *channelManager) RequestCloseChannel(msg bean.RequestMessage, user *bean.User) (interface{}, error) {
	channelId, err := getChannelIdFromJson(msg.Data)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if channelInfo.CurrState != dao.ChannelState_CanUse && channelInfo.CurrState != dao.ChannelState_HtlcTx {
		return nil, errors.New("wrong channel state " + strconv.Itoa(int(channelInfo.CurrState)))
	}

	targetUser := channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	}

	if targetUser != msg.RecipientUserPeerId {
		return nil, errors.New("wrong targetUser " + msg.RecipientUserPeerId)
	}

	lastCommitmentTx, err := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if lastCommitmentTx.CurrState != dao.TxInfoState_Htlc_GetH && lastCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, errors.New("latest commitment tx state is wrong")
	}

	if channelInfo.CurrState == dao.ChannelState_HtlcTx {
		flag := httpGetHtlcStateFromTracker(lastCommitmentTx.HtlcRoutingPacket, lastCommitmentTx.HtlcH)
		if flag == 1 {
			return nil, errors.New("R is backward")
		}
	}

	closeChannel := &dao.CloseChannel{}
	closeChannel.ChannelId = channelId
	closeChannel.Owner = user.PeerId
	closeChannel.CurrState = 0
	_ = tx.Select(
		q.Eq("ChannelId", closeChannel.ChannelId),
		q.Eq("Owner", closeChannel.Owner),
		q.Eq("CurrState", closeChannel.CurrState)).
		Find(closeChannel)

	if closeChannel.Id == 0 {
		closeChannel.CreateAt = time.Now()
		dataBytes, _ := json.Marshal(closeChannel)
		closeChannel.RequestHex = tool.SignMsgWithSha256(dataBytes)
		err = tx.Save(closeChannel)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	_ = tx.Commit()

	toData := make(map[string]interface{})
	toData["channel_id"] = channelId
	toData["close_channel_hash"] = closeChannel.RequestHex
	return toData, nil
}

//关闭通道的请求到达对方节点obd
func (this *channelManager) BeforeBobSignCloseChannelAtBobSide(data string, user bean.User) (retData map[string]interface{}, err error) {
	var channelId = gjson.Get(data, "channel_id").String()
	var closeChannelHash = gjson.Get(data, "close_channel_hash").String()

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if channelInfo.CurrState != dao.ChannelState_CanUse && channelInfo.CurrState != dao.ChannelState_HtlcTx {
		return nil, errors.New("wrong channel state " + strconv.Itoa(int(channelInfo.CurrState)))
	}

	requestSenderUser := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		requestSenderUser = channelInfo.PeerIdB
	}

	closeChannel := &dao.CloseChannel{}
	closeChannel.ChannelId = channelId
	closeChannel.Owner = requestSenderUser
	closeChannel.CurrState = 0
	_ = tx.Select(
		q.Eq("ChannelId", closeChannel.ChannelId),
		q.Eq("Owner", requestSenderUser),
		q.Eq("CurrState", closeChannel.CurrState)).
		Find(closeChannel)

	if closeChannel.Id == 0 {
		closeChannel.CreateAt = time.Now()
		closeChannel.RequestHex = closeChannelHash
		err = tx.Save(closeChannel)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	_ = tx.Commit()

	retData = make(map[string]interface{})
	retData["channel_id"] = channelId
	retData["request_close_channel_hash"] = closeChannelHash
	return retData, nil
}

//对方签收是否关闭
func (this *channelManager) SignCloseChannel(msg bean.RequestMessage, user bean.User) (retData map[string]interface{}, err error) {

	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New("empty inputData")
	}

	reqData := &bean.CloseChannelSign{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.ChannelId) == false {
		err = errors.New("empty channel_id")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.RequestCloseChannelHash) == false {
		err = errors.New("empty request_close_channel_hash")
		log.Println(err)
		return nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", reqData.ChannelId)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if channelInfo.CurrState != dao.ChannelState_CanUse && channelInfo.CurrState != dao.ChannelState_HtlcTx {
		return nil, errors.New("wrong channel state " + strconv.Itoa(int(channelInfo.CurrState)))
	}

	requestSenderUser := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		requestSenderUser = channelInfo.PeerIdB
	}
	if requestSenderUser != msg.RecipientUserPeerId {
		return nil, errors.New("wrong RecipientUserPeerId")
	}

	closeChannelStarterData := &dao.CloseChannel{}
	err = tx.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", 0),
		q.Eq("RequestHex", reqData.RequestCloseChannelHash)).
		First(closeChannelStarterData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	closeChannelStarterData.Approval = reqData.Approval
	closeChannelStarterData.CurrState = 1
	_ = tx.Update(closeChannelStarterData)

	if reqData.Approval {
		channelInfo.CurrState = dao.ChannelState_Close
		channelInfo.CloseAt = time.Now()
		err = tx.Update(channelInfo)
		if err != nil {
			return nil, err
		}
	}
	_ = tx.Commit()

	retData = make(map[string]interface{})
	retData["channel_id"] = reqData.ChannelId
	retData["request_close_channel_hash"] = closeChannelStarterData.RequestHex
	retData["approval"] = reqData.Approval
	return retData, nil
}

//直接强制关闭通道
func (this *channelManager) ForceCloseChannel(msg bean.RequestMessage, user *bean.User) (interface{}, error) {
	channelId, err := getChannelIdFromJson(msg.Data)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&channelId) == false {
		err = errors.New(enum.Tips_common_wrong + "channel_id")
		log.Println(err)
		return nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	targetUser := user.PeerId
	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", targetUser),
			q.Eq("PeerIdB", targetUser))).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	latestCommitmentTx, err := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, targetUser)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if latestCommitmentTx.CurrState != dao.TxInfoState_Create &&
		latestCommitmentTx.CurrState != dao.TxInfoState_Htlc_GetH &&
		latestCommitmentTx.CurrState != dao.TxInfoState_Htlc_GetR &&
		latestCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, errors.New(enum.Tips_channel_wrongLatestCommitmentTxState)
	}

	// 当前是处于htlc的状态，且是获取到H
	if channelInfo.CurrState == dao.ChannelState_HtlcTx {
		_, err = this.CloseHtlcChannelSigned(tx, channelInfo, latestCommitmentTx, *user)
		if err != nil {
			return nil, err
		}
	} else {
		if latestCommitmentTx.CurrState == dao.TxInfoState_Create {
			err = tx.One("Id", latestCommitmentTx.LastCommitmentTxId, latestCommitmentTx)
			if err != nil {
				return nil, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
			}
		}

		if latestCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
			return nil, errors.New(enum.Tips_channel_LatestCommitmentTxNotInReadySendState)
		}

		//region 广播承诺交易 最近的rsmc的资产分配交易 因为是omni资产，承诺交易被拆分成了两个独立的交易
		if tool.CheckIsString(&latestCommitmentTx.RSMCTxHex) {
			commitmentTxid, err := rpcClient.SendRawTransaction(latestCommitmentTx.RSMCTxHex)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			log.Println(commitmentTxid)
		}
		if tool.CheckIsString(&latestCommitmentTx.ToCounterpartyTxHex) {
			commitmentTxidToBob, err := rpcClient.SendRawTransaction(latestCommitmentTx.ToCounterpartyTxHex)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			log.Println(commitmentTxidToBob)
		}
		//endregion

		//region 广播RD
		latestRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
		_ = tx.Select(
			q.Eq("ChannelId", channelInfo.ChannelId),
			q.Eq("CommitmentTxId", latestCommitmentTx.Id),
			q.Eq("Owner", targetUser)).
			OrderBy("CreateAt").Reverse().
			First(latestRevocableDeliveryTx)

		if latestRevocableDeliveryTx.Id > 0 {
			_, err = rpcClient.SendRawTransaction(latestRevocableDeliveryTx.TxHex)
			if err != nil {
				log.Println(err)
				msg := err.Error()
				//如果omnicore返回的信息里面包含了non-BIP68-final (code 64)， 则说明因为需要等待1000个区块高度，广播是对的
				if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
					return nil, err
				}
			}

			latestRevocableDeliveryTx.CurrState = dao.TxInfoState_SendHex
			latestRevocableDeliveryTx.SendAt = time.Now()
			err = tx.Update(latestRevocableDeliveryTx)
			if err != nil {
				return nil, err
			}

			err = addRDTxToWaitDB(latestRevocableDeliveryTx)
			if err != nil {
				return nil, err
			}
		}
		//endregion

		// region update state
		latestCommitmentTx.CurrState = dao.TxInfoState_SendHex
		latestCommitmentTx.SendAt = time.Now()
		err = tx.Update(latestCommitmentTx)
		if err != nil {
			return nil, err
		}
		//endregion
	}

	channelInfo.CurrState = dao.ChannelState_Close
	channelInfo.CloseAt = time.Now()
	err = tx.Update(channelInfo)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	//同步通道信息到tracker
	sendChannelStateToTracker(*channelInfo, *latestCommitmentTx)

	return channelInfo, nil

}

//请求方节点处理关闭通道的操作
func (this *channelManager) AfterBobSignCloseChannelAtAliceSide(jsonData string, user bean.User) (interface{}, error) {

	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New(enum.Tips_common_empty + "inputData")
	}
	reqData := &bean.CloseChannelSign{}
	err := json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id")
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&reqData.RequestCloseChannelHash) == false {
		err = errors.New("empty request_close_channel_hash")
		log.Println(err)
		return nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	closeChannelStarterData := &dao.CloseChannel{}
	err = tx.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", 0),
		q.Eq("RequestHex", reqData.RequestCloseChannelHash)).
		First(closeChannelStarterData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	targetUser := user.PeerId
	channelInfo := &dao.ChannelInfo{}
	err = tx.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Or(
			q.Eq("PeerIdA", targetUser),
			q.Eq("PeerIdB", targetUser))).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	closeChannelStarterData.Approval = reqData.Approval
	if reqData.Approval == false {
		_ = tx.Update(closeChannelStarterData)
		_ = tx.Commit()

		log.Println("disagree close channel")
		return nil, errors.New("disagree close channel")
	}

	latestCommitmentTx, err := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, targetUser)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if latestCommitmentTx.CurrState != dao.TxInfoState_Htlc_GetH && latestCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, errors.New("latest commitment tx state is wrong")
	}

	// 当前是处于htlc的状态，且是获取到H
	if channelInfo.CurrState == dao.ChannelState_HtlcTx {
		_, err = this.CloseHtlcChannelSigned(tx, channelInfo, latestCommitmentTx, user)
		if err != nil {
			return nil, err
		}
	} else {
		//region 广播承诺交易 最近的rsmc的资产分配交易 因为是omni资产，承诺交易被拆分成了两个独立的交易
		if tool.CheckIsString(&latestCommitmentTx.RSMCTxHex) {
			commitmentTxid, err := rpcClient.SendRawTransaction(latestCommitmentTx.RSMCTxHex)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			log.Println(commitmentTxid)
		}
		if tool.CheckIsString(&latestCommitmentTx.ToCounterpartyTxHex) {
			commitmentTxidToBob, err := rpcClient.SendRawTransaction(latestCommitmentTx.ToCounterpartyTxHex)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			log.Println(commitmentTxidToBob)
		}
		//endregion

		//region 广播RD
		latestRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(
			q.Eq("ChannelId", channelInfo.ChannelId),
			q.Eq("Owner", targetUser)).
			OrderBy("CreateAt").Reverse().
			First(latestRevocableDeliveryTx)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		_, err = rpcClient.SendRawTransaction(latestRevocableDeliveryTx.TxHex)
		if err != nil {
			log.Println(err)
			msg := err.Error()
			//如果omnicore返回的信息里面包含了non-BIP68-final (code 64)， 则说明因为需要等待1000个区块高度，广播是对的
			if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
				return nil, err
			}
		}
		//endregion

		// region update state
		latestCommitmentTx.CurrState = dao.TxInfoState_SendHex
		latestCommitmentTx.SendAt = time.Now()
		err = tx.Update(latestCommitmentTx)
		if err != nil {
			return nil, err
		}

		latestRevocableDeliveryTx.CurrState = dao.TxInfoState_SendHex
		latestRevocableDeliveryTx.SendAt = time.Now()
		err = tx.Update(latestRevocableDeliveryTx)
		if err != nil {
			return nil, err
		}

		err = addRDTxToWaitDB(latestRevocableDeliveryTx)
		if err != nil {
			return nil, err
		}
		//endregion
	}

	channelInfo.CurrState = dao.ChannelState_Close
	channelInfo.CloseAt = time.Now()
	err = tx.Update(channelInfo)
	if err != nil {
		return nil, err
	}

	closeChannelStarterData.CurrState = 1
	_ = tx.Update(closeChannelStarterData)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	//同步通道信息到tracker
	sendChannelStateToTracker(*channelInfo, *latestCommitmentTx)

	return channelInfo, nil
}

//  htlc  when getH close channel
func (this *channelManager) CloseHtlcChannelSigned(tx storm.Node, channelInfo *dao.ChannelInfo, latestCommitmentTx *dao.CommitmentTransaction, user bean.User) (outData interface{}, err error) {
	// 提现操作的发起者
	closeOpStarter := user.PeerId

	//region 广播主承诺交易 三笔
	if tool.CheckIsString(&latestCommitmentTx.RSMCTxHex) {
		commitmentTxid, err := rpcClient.SendRawTransaction(latestCommitmentTx.RSMCTxHex)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(commitmentTxid)
	}

	latestRsmcRD := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", latestCommitmentTx.Id),
		q.Eq("RDType", 0),
		q.Eq("Owner", closeOpStarter)).
		OrderBy("CreateAt").Reverse().
		First(latestRsmcRD)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = rpcClient.SendRawTransaction(latestRsmcRD.TxHex)
	if err != nil {
		log.Println(err)
		msg := err.Error()
		if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
			return nil, err
		}
	}

	if tool.CheckIsString(&latestCommitmentTx.ToCounterpartyTxHex) {
		commitmentTxidToBob, err := rpcClient.SendRawTransaction(latestCommitmentTx.ToCounterpartyTxHex)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(commitmentTxidToBob)
	}

	// htlc部分
	if tool.CheckIsString(&latestCommitmentTx.HtlcTxHex) {
		commitmentTxidToHtlc, err := rpcClient.SendRawTransaction(latestCommitmentTx.HtlcTxHex)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(commitmentTxidToHtlc)
	}
	// endregion

	// region htlc的相关交易广播

	// 提现人是这次htlc的转账发起者
	if latestCommitmentTx.HtlcSender == closeOpStarter {
		ht1a := &dao.HTLCTimeoutTxForAAndExecutionForB{}
		err = tx.Select(
			q.Eq("ChannelId", channelInfo.ChannelId),
			q.Eq("CommitmentTxId", latestCommitmentTx.Id),
			q.Eq("Owner", closeOpStarter),
			q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
			First(ht1a)
		if ht1a.Id > 0 {
			htrd := &dao.RevocableDeliveryTransaction{}
			err = tx.Select(
				q.Eq("CommitmentTxId", ht1a.Id),
				q.Eq("Owner", closeOpStarter),
				q.Eq("RDType", 1),
				q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
				First(htrd)
			if htrd.Id > 0 && tool.CheckIsString(&ht1a.RSMCTxHex) {
				//广播alice的ht1a
				_, err = rpcClient.SendRawTransaction(ht1a.RSMCTxHex)
				if err == nil { //如果已经超时 比如alice的3天超时，bob得到R后的交易的无等待锁定
					if tool.CheckIsString(&htrd.TxHex) {
						_, err = rpcClient.SendRawTransaction(htrd.TxHex)
						if err != nil {
							log.Println(err)
							msg := err.Error()
							if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
								return nil, err
							}
						}
						_ = addRDTxToWaitDB(htrd)
						ht1a.CurrState = dao.TxInfoState_SendHex
						ht1a.SendAt = time.Now()
						_ = tx.Update(ht1a)
					}
				} else {
					//如果是alice的（ht1a的锁定时间内的提现交易，就需要判断时候是正常的超时广播（含有non-BIP68-final (code 64)），如果不是，就返回
					log.Println(err)
					msg := err.Error()
					if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
						return nil, err
					}
					_ = addHT1aTxToWaitDB(ht1a, htrd)
				}
			}
		}
	} else {
		//如果是htlc的转账接收者
		htdnx := &dao.HTLCTimeoutDeliveryTxB{}
		err = tx.Select(
			q.Eq("CommitmentTxId", latestCommitmentTx.Id),
			q.Eq("CurrState", dao.TxInfoState_CreateAndSign),
			q.Eq("Owner", closeOpStarter)).
			First(htdnx)
		if htdnx.Id > 0 && tool.CheckIsString(&htdnx.TxHex) {
			_, err = rpcClient.SendRawTransaction(htdnx.TxHex)
			if err != nil {
				log.Println(err)
				msg := err.Error()
				if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
					return nil, err
				}
			}
			_ = addHTDnxTxToWaitDB(htdnx)
			htdnx.CurrState = dao.TxInfoState_SendHex
			htdnx.SendAt = time.Now()
			_ = tx.Update(htdnx)
		}
	}
	// endregion

	// region update obj state to db
	latestCommitmentTx.CurrState = dao.TxInfoState_SendHex
	latestCommitmentTx.SendAt = time.Now()
	err = tx.Update(latestCommitmentTx)
	if err != nil {
		return nil, err
	}

	latestRsmcRD.CurrState = dao.TxInfoState_SendHex
	latestRsmcRD.SendAt = time.Now()
	err = tx.Update(latestRsmcRD)
	if err != nil {
		return nil, err
	}

	err = addRDTxToWaitDB(latestRsmcRD)
	if err != nil {
		return nil, err
	}
	//endregion
	return channelInfo, nil
}
