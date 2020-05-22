package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/omnilaboratory/obd/tracker/dao"
	"log"
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

type channelManager struct {
	mu       sync.Mutex
	openList []*graphEdge
}

var channelService channelManager

func (manager *channelManager) updateChannelInfo(obdClient *ObdNode, msgData string) (err error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if tool.CheckIsString(&msgData) == false {
		return errors.New("wrong inputData")
	}
	channelInfos := make([]bean.ChannelInfoRequest, 0)
	err = json.Unmarshal([]byte(msgData), &channelInfos)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, item := range channelInfos {
		if tool.CheckIsString(&item.ChannelId) == false ||
			tool.CheckIsString(&item.PeerIdA) == false ||
			tool.CheckIsString(&item.PeerIdB) == false {
			continue
		}
		channelInfo := &dao.ChannelInfo{}
		_ = db.Select(q.Eq("ChannelId", item.ChannelId)).First(channelInfo)
		if channelInfo.Id == 0 {
			channelInfo.ChannelId = item.ChannelId
			channelInfo.PropertyId = item.PropertyId
			channelInfo.CurrState = item.CurrState
			channelInfo.PeerIdA = item.PeerIdA
			channelInfo.PeerIdB = item.PeerIdB
			channelInfo.AmountA = item.AmountA
			channelInfo.AmountB = item.AmountB
			if item.IsAlice {
				channelInfo.ObdNodeIdA = obdClient.Id
			} else {
				channelInfo.ObdNodeIdB = obdClient.Id
			}
			channelInfo.CreateAt = time.Now()
			_ = db.Save(channelInfo)
		} else {
			channelInfo.PropertyId = item.PropertyId
			channelInfo.CurrState = item.CurrState
			channelInfo.PeerIdA = item.PeerIdA
			channelInfo.PeerIdB = item.PeerIdB
			channelInfo.AmountA = item.AmountA
			channelInfo.AmountB = item.AmountB
			if item.IsAlice {
				channelInfo.ObdNodeIdA = obdClient.Id
			} else {
				channelInfo.ObdNodeIdB = obdClient.Id
			}
			_ = db.Update(channelInfo)
		}
	}
	return err
}
func (manager *channelManager) getPath(obdClient *ObdNode, msgData string) (path interface{}, err error) {
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

	manager.createChannelNetwork(pathRequest.RealPayerPeerId, pathRequest.PayeePeerId, pathRequest.PropertyId, pathRequest.Amount, nil, true)
	resultIndex := -1
	minLength := 99
	for index, node := range manager.openList {
		if node.IsTarget {
			log.Println(node.ChannelIds)
			resultIndex = index
			tempLength := len(strings.Split(node.ChannelIds, ","))
			if tempLength < minLength {
				minLength = tempLength
			}
		}
	}
	retNode := make(map[string]interface{})
	retNode["senderPeerId"] = pathRequest.RealPayerPeerId
	retNode["path"] = ""
	if resultIndex != -1 {
		splitArr := strings.Split(manager.openList[resultIndex].ChannelIds, ",")
		path := ""
		for i := len(splitArr) - 1; i > -1; i-- {
			path = splitArr[i] + ","
		}
		path = strings.TrimSuffix(path, ",")
		retNode["path"] = path
	}
	return retNode, nil
}

func (manager *channelManager) createChannelNetwork(realPayerPeerId, currPayeePeerId string, propertyId int64, amount float64, currNode *graphEdge, isBegin bool) {
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
	err := db.Select(
		q.Eq("PropertyId", propertyId),
		q.Or(
			q.Eq("PeerIdB", currPayeePeerId),
			q.Eq("PeerIdA", currPayeePeerId))).
		Find(&nodes)
	if err == nil {
		for _, item := range nodes {
			interSender := item.PeerIdA
			leftAmount := item.AmountA
			if item.PeerIdA == currPayeePeerId {
				interSender = item.PeerIdB
				leftAmount = item.AmountB
			}

			if leftAmount >= amount {

				//if _, ok := userOfOnlineMap[interSender]; ok == false {
				//	continue
				//}

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
						manager.createChannelNetwork(realPayerPeerId, interSender, propertyId, amount, &newEdge, false)
					}
				}
			}
		}
	}
}
