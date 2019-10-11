package service

import (
	"LightningOnOmni/dao"
	"errors"
	"github.com/asdine/storm/q"
	"log"
)

//
type TreeNode struct {
	parentNode     *TreeNode
	level          int
	parentPeerId   string
	currNodePeerId string
	amount         float64
	isTarget       bool
	channel        *dao.ChannelInfo
	children       []*TreeNode
}

type pathManager struct{}

var PathService = pathManager{}

func (p *pathManager) CreateTree(fromPeerId, toPeerId string, amount float64, currNode *TreeNode, tree *TreeNode) (err error) {
	var tempBNodes []dao.ChannelInfo
	err = db.Select(q.Eq("PeerIdA", fromPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempBNodes)
	if err != nil {
		log.Println(err)
	}
	//看看查到的直接通道里面有没有目标对象
	if tempBNodes != nil && len(tempBNodes) > 0 {
		//如果没有找到目标对象，就找中间商，中间商的下个通道的钱是否足够
		for _, node := range tempBNodes {
			return p.dealNode(node, fromPeerId, node.PeerIdB, toPeerId, amount, currNode, tree)
		}
	}

	var tempANodes []dao.ChannelInfo
	err = db.Select(q.Eq("PeerIdB", fromPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempANodes)
	if err != nil {
		log.Println(err)
	}

	//看看查到的直接通道里面有没有目标对象
	if tempANodes != nil && len(tempANodes) > 0 {
		//如果没有找到目标对象，就找中间商，中间商的下个通道的钱是否足够
		for _, node := range tempANodes {
			return p.dealNode(node, fromPeerId, node.PeerIdA, toPeerId, amount, currNode, tree)
		}
	}
	return errors.New("not found")
}

func (p *pathManager) dealNode(node dao.ChannelInfo, fromPeerId, targetUser, toPeerId string, amount float64, currNode *TreeNode, tree *TreeNode) error {
	commitmentNode := &dao.CommitmentTransaction{}
	err := db.Select(q.Eq("ChannelId", node.ChannelId), q.Eq("Owner", fromPeerId)).OrderBy("CreateAt").Reverse().First(commitmentNode)
	if err == nil && commitmentNode.AmountToRSMC > amount {
		if targetUser == toPeerId {
			childNode := &TreeNode{
				parentNode:     currNode,
				parentPeerId:   fromPeerId,
				currNodePeerId: toPeerId,
				isTarget:       true,
				amount:         commitmentNode.AmountToRSMC,
				channel:        &node,
				children:       make([]*TreeNode, 0),
			}
			if tree == nil {
				tree = childNode
				currNode = childNode
			} else {
				currNode.children = append(currNode.children, childNode)
			}
			return nil
		} else {
			//查找bob的通道满足余额大于需要额度的列表
			if findNextLevelValidBobs(fromPeerId, targetUser, amount, node, tree) {
				childNode := &TreeNode{
					parentNode:     currNode,
					parentPeerId:   fromPeerId,
					currNodePeerId: targetUser,
					isTarget:       false,
					amount:         commitmentNode.AmountToRSMC,
					channel:        &node,
					children:       make([]*TreeNode, 0),
				}
				if tree == nil {
					rootNode := &TreeNode{
						parentNode:     nil,
						parentPeerId:   nil,
						currNodePeerId: fromPeerId,
						isTarget:       false,
						amount:         0,
						channel:        nil,
						children:       make([]*TreeNode, 0),
					}
					tree = rootNode
					currNode = rootNode
				}
				currNode.children = append(currNode.children, childNode)
				return p.CreateTree(targetUser, toPeerId, amount, childNode, tree)
			}
		}
	}
	return errors.New("not found")
}

func findNextLevelValidBobs(fromPeerId string, targetPeerId string, amount float64, currChannel dao.ChannelInfo, tree *TreeNode) bool {
	var nextTempBNodes []dao.ChannelInfo
	err := db.Select(q.Not(q.Eq("PeerIdB", fromPeerId)), q.Eq("PeerIdA", targetPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&nextTempBNodes)
	if err == nil {
		for _, nextNode := range nextTempBNodes {
			//1、排除已经存在的节点

			//2、获取余额
			commitmentNode := &dao.CommitmentTransaction{}
			err = db.Select(q.Eq("ChannelId", nextNode.ChannelId), q.Eq("Owner", targetPeerId)).OrderBy("CreateAt").Reverse().First(commitmentNode)
			if err == nil {
				if commitmentNode.AmountToRSMC > amount {
					return true
				}
			}
		}
	}
	//当bob是PeerIdB的时候
	var nextTempANodes []dao.ChannelInfo
	err = db.Select(q.Not(q.Eq("PeerIdA", fromPeerId)), q.Eq("PeerIdB", targetPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&nextTempANodes)
	if err == nil {
		for _, nextNode := range nextTempANodes {
			//获取余额
			commitmentNode := &dao.CommitmentTransaction{}
			err = db.Select(q.Eq("ChannelId", nextNode.ChannelId), q.Eq("Owner", targetPeerId)).OrderBy("CreateAt").Reverse().First(commitmentNode)
			if err == nil {
				if commitmentNode.AmountToRSMC > amount {
					return true
				}
			}
		}
	}
	return false
}

//查找出与自己相关所有通道
func (p *pathManager) GetPath(fromPeerId, toPeerId string, amount float64, path []dao.ChannelInfo) (targetNode *dao.ChannelInfo, nodes []dao.ChannelInfo, err error) {
	if nodes == nil {
		log.Println("make nodes array")
		nodes = make([]dao.ChannelInfo, 0)
	}
	if path == nil {
		log.Println("make path array")
		nodes = make([]dao.ChannelInfo, 0)
	}
	// if parentPeerId is PeerIdA in channel
	// then get the PeerIdB
	var tempBNodes []dao.ChannelInfo
	err = db.Select(q.Eq("PeerIdA", fromPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempBNodes)
	if err != nil {
		log.Println(err)
	}

	//如果找到目标对象
	if tempBNodes != nil && len(tempBNodes) > 0 {
		for _, node := range tempBNodes {
			if node.PeerIdB == toPeerId {
				path = append(path, node)
				return &node, nil, nil
			}
		}

		//如果没有找到目标对象，就找中间商，中间商的下个通道的钱是否足够
		for _, node := range tempBNodes {
			//查找bob的通道满足余额大于需要额度的列表
			getValidBobs(fromPeerId, amount, node, nodes, path)
		}
	}

	var tempANodes []dao.ChannelInfo
	err = db.Select(q.Eq("PeerIdB", fromPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempANodes)
	if err != nil {
		log.Println(err)
	}

	if tempANodes != nil && len(tempANodes) > 0 {
		for _, node := range tempANodes {
			if node.PeerIdA == toPeerId {
				path = append(path, node)
				return &node, nil, nil
			}
		}
		for _, node := range tempANodes {
			getValidBobs(fromPeerId, amount, node, nodes, path)
		}
	}

	for _, nextChannnel := range nodes {
		nextFromPeerId := nextChannnel.PeerIdB
		if nextFromPeerId == fromPeerId {
			nextFromPeerId = nextChannnel.PeerIdA
		}
		return p.GetPath(nextFromPeerId, toPeerId, amount, path)
	}
	return nil, nil, errors.New("not found path")
	//return nil, nil
}

func getValidBobs(fromPeerId string, amount float64, currChannel dao.ChannelInfo, nodes []dao.ChannelInfo, path []dao.ChannelInfo) {
	currPeerBob := currChannel.PeerIdB
	var nextTempBNodes []dao.ChannelInfo
	err := db.Select(q.Not(q.Eq("PeerIdB", fromPeerId)), q.Eq("PeerIdA", currPeerBob), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&nextTempBNodes)
	if err == nil {
		for _, nextNode := range nextTempBNodes {
			//获取余额
			commitmentNode := &dao.CommitmentTransaction{}
			err = db.Select(q.Eq("ChannelId", nextNode.ChannelId), q.Eq("Owner", currPeerBob)).OrderBy("CreateAt").Reverse().First(commitmentNode)
			if err == nil {
				if commitmentNode.AmountToRSMC > amount {
					nodes = append(nodes, currChannel)
				}
			}
		}
	}
	//当bob是PeerIdB的时候
	var nextTempANodes []dao.ChannelInfo
	err = db.Select(q.Not(q.Eq("PeerIdA", fromPeerId)), q.Eq("PeerIdB", currPeerBob), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&nextTempANodes)
	if err == nil {
		for _, nextNode := range nextTempANodes {
			//获取余额
			commitmentNode := &dao.CommitmentTransaction{}
			err = db.Select(q.Eq("ChannelId", nextNode.ChannelId), q.Eq("Owner", currPeerBob)).OrderBy("CreateAt").Reverse().First(commitmentNode)
			if err == nil {
				if commitmentNode.AmountToRSMC > amount {
					nodes = append(nodes, currChannel)
				}
			}
		}
	}
}
