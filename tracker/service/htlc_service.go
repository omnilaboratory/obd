package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/gin-gonic/gin"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/omnilaboratory/obd/tracker/dao"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type graphEdge struct {
	ParentNodeIndex int
	PathPeerIds     string
	ChannelIds      string
	PathIndexArr    []int
	ChannelId       string
	Level           uint16
	CurrNodePeerId  string
	IsRoot          bool
	IsTarget        bool
}

type htlcManager struct {
	mu       sync.Mutex
	openList []*graphEdge
}

var HtlcService htlcManager

func (manager *htlcManager) getPath(obdClient *ObdNode, msgData string) (path interface{}, err error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if tool.CheckIsString(&msgData) == false {
		return "", errors.New("wrong inputData")
	}
	pathRequest := &bean.HtlcPathRequest{}
	err = json.Unmarshal([]byte(msgData), pathRequest)
	if err != nil {
		return "", err
	}

	if tool.CheckIsString(&pathRequest.RealPayerPeerId) == false {
		return "", errors.New("wrong realPayerPeerId")
	}
	if tool.CheckIsString(&pathRequest.PayeePeerId) == false {
		return "", errors.New("wrong PayeePeerId")
	}
	if pathRequest.Amount < config.GetOmniDustBtc() {
		return "", errors.New("wrong amount")
	}

	manager.createChannelNetwork(pathRequest.PayerObdNodeId, pathRequest.RealPayerPeerId, pathRequest.PayeePeerId, pathRequest.PropertyId, pathRequest.Amount, nil, true)
	resultIndex := -1
	minLength := 99
	for index, node := range manager.openList {
		if node.IsTarget {
			log.Println(node.ChannelIds)
			tempLength := len(strings.Split(node.ChannelIds, ","))
			if tempLength < minLength {
				resultIndex = index
				minLength = tempLength
			}
		}
	}
	retNode := make(map[string]interface{})
	retNode["senderPeerId"] = pathRequest.RealPayerPeerId
	retNode["h"] = pathRequest.H
	retNode["amount"] = pathRequest.Amount
	retNode["path"] = ""
	if resultIndex != -1 {
		splitArr := strings.Split(manager.openList[resultIndex].ChannelIds, ",")
		path := ""
		for i := len(splitArr) - 1; i > -1; i-- {
			path += splitArr[i] + ","
		}
		path = strings.TrimSuffix(path, ",")
		retNode["path"] = path
	}
	return retNode, nil
}

func (manager *htlcManager) updateHtlcInfo(obdClient *ObdNode, msgData string) (err error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if tool.CheckIsString(&msgData) == false {
		return errors.New("wrong inputData")
	}
	reqData := &bean.UpdateHtlcTxStateRequest{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		return err
	}

	if tool.CheckIsString(&reqData.Path) == false {
		return errors.New("path")
	}
	if tool.CheckIsString(&reqData.H) == false {
		return errors.New("h")
	}
	if tool.CheckIsString(&reqData.CurrChannelId) == false {
		return errors.New("currChannelId")
	}

	htlcTxInfo := &dao.HtlcTxInfo{}
	_ = db.Select(q.Eq("Path", reqData.Path), q.Eq("H", reqData.H)).First(htlcTxInfo)
	if htlcTxInfo.Id == 0 {
		htlcTxInfo.Path = reqData.Path
		htlcTxInfo.H = reqData.H
		htlcTxInfo.DirectionFlag = bean.HtlcTxState_PayMoney
		htlcTxInfo.CurrChannelId = reqData.CurrChannelId
		htlcTxInfo.CreateAt = time.Now()
		_ = db.Save(htlcTxInfo)
	} else {
		if tool.CheckIsString(&htlcTxInfo.R) == false && tool.CheckIsString(&reqData.R) {
			htlcTxInfo.R = reqData.R
		}
		htlcTxInfo.DirectionFlag = reqData.DirectionFlag
		htlcTxInfo.CurrChannelId = reqData.CurrChannelId
		_ = db.Update(htlcTxInfo)
	}
	return nil
}

func (manager *htlcManager) GetHtlcCurrState(context *gin.Context) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	reqData := &bean.GetHtlcTxStateRequest{}
	reqData.Path = context.Query("path")
	if tool.CheckIsString(&reqData.Path) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error path",
		})
		return
	}
	reqData.H = context.Query("h")
	if tool.CheckIsString(&reqData.H) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error h",
		})
		return
	}

	htlcTxInfo := &dao.HtlcTxInfo{}
	_ = db.Select(q.Eq("Path", reqData.Path), q.Eq("H", reqData.H)).First(htlcTxInfo)
	retData := make(map[string]interface{})
	retData["flag"] = 0
	if htlcTxInfo.Id > 0 {
		retData["flag"] = htlcTxInfo.DirectionFlag
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "htlcState",
		"data": retData,
	})
}

func (manager *htlcManager) createChannelNetwork(payerObdNodeId, realPayerPeerId, currPayeePeerId string, propertyId int64, amount float64, currNode *graphEdge, isBegin bool) {
	if isBegin {
		manager.openList = make([]*graphEdge, 0)
		realPayeeEdge := &graphEdge{
			ParentNodeIndex: -1,
			CurrNodePeerId:  currPayeePeerId,
			PathPeerIds:     "",
			ChannelIds:      "",
			PathIndexArr:    make([]int, 0),
			Level:           0,
			IsRoot:          true,
			IsTarget:        false,
		}
		manager.openList = append(manager.openList, realPayeeEdge)
	}

	if currNode == nil {
		currNode = manager.openList[0]
	}
	//如果当前用户已经在备选的用户列表中出现了，就return
	if strings.Contains(currNode.PathPeerIds, currNode.CurrNodePeerId) {
		return
	}

	if currNode.Level > 0 {
		amount += config.GetHtlcFee()
	}

	currNodeIndex := 0
	for index, tempNode := range manager.openList {
		if tempNode == currNode {
			currNodeIndex = index
			break
		}
	}

	pathPeerIds := currNode.PathPeerIds + "," + currNode.CurrNodePeerId
	pathIndexArr := currNode.PathIndexArr
	if currNode.PathPeerIds == "" {
		pathPeerIds = currNode.CurrNodePeerId
		pathIndexArr = make([]int, 0)
	}
	pathIndexArr = append(pathIndexArr, currNodeIndex)
	newEdge := graphEdge{
		ParentNodeIndex: currNodeIndex,
		PathPeerIds:     pathPeerIds,
		PathIndexArr:    pathIndexArr,
		Level:           currNode.Level + 1,
		IsRoot:          false,
	}

	var nodes []dao.ChannelInfo
	err := errors.New("no channel")
	if isBegin {
		err = db.Select(
			q.Eq("PropertyId", propertyId),
			q.Eq("CurrState", 20),
			q.Or(
				q.Eq("PeerIdB", currPayeePeerId),
				q.Eq("PeerIdA", currPayeePeerId)),
			q.Or(
				q.Eq("ObdNodeIdA", payerObdNodeId),
				q.Eq("ObdNodeIdB", payerObdNodeId))).
			Find(&nodes)
	} else {
		err = db.Select(
			q.Eq("PropertyId", propertyId),
			q.Eq("CurrState", 20),
			q.Or(
				q.Eq("PeerIdB", currPayeePeerId),
				q.Eq("PeerIdA", currPayeePeerId))).
			Find(&nodes)
	}

	if err == nil {
		for _, item := range nodes {
			interSender := item.PeerIdA
			leftAmount := item.AmountA
			if item.PeerIdA == currPayeePeerId {
				interSender = item.PeerIdB
				leftAmount = item.AmountB
			}

			if leftAmount >= amount {
				if _, ok := userOfOnlineMap[interSender]; ok == false {
					continue
				}
				channelIds := item.ChannelId
				if tool.CheckIsString(&currNode.ChannelIds) {
					channelIds = currNode.ChannelIds + "," + item.ChannelId
				}

				newEdge.ChannelIds = channelIds
				newEdge.ChannelId = item.ChannelId
				newEdge.CurrNodePeerId = interSender
				newEdge.IsTarget = false
				manager.openList = append(manager.openList, &newEdge)

				if interSender == realPayerPeerId {
					newEdge.IsTarget = true
					newEdge.PathPeerIds += "," + newEdge.CurrNodePeerId
				} else {
					if newEdge.Level < 6 {
						manager.createChannelNetwork("", realPayerPeerId, interSender, propertyId, amount, &newEdge, false)
					}
				}
			}
		}
	}
}
