package service

import (
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"errors"
	"github.com/asdine/storm/q"
	"log"
)

//
type TreeNode struct {
	parentNode     *TreeNode
	level          int
	currNodePeerId string
	amount         float64
	isRoot         bool
	isTarget       bool
	channel        *dao.ChannelInfo
	children       []*TreeNode
	childMap       map[string]*TreeNode
}

type PathNode struct {
	ParentNode     *PathNode            `json:"parent_node"`
	Level          int                  `json:"level"`
	CurrNodePeerId string               `json:"curr_node_peer_id"`
	IsRoot         bool                 `json:"is_root"`
	IsTarget       bool                 `json:"is_target"`
	PeerMap        map[string]*PathNode `json:"child_map"`
}

type PathBranchInfo struct {
	Peer2Peer string           `json:"peer_2_peer"`
	Amount    float64          `json:"amount"`
	Channel   *dao.ChannelInfo `json:"channel"`
}

type pathManager struct{}

var PathService = pathManager{}

func (p *pathManager) CreateDemoChannelNetwork(realSenderPeerId, realReceiverPeerId string, amount float64, currNode *PathNode, tree *PathNode, nodeMap map[string]*PathNode, branchMap map[string]*PathBranchInfo) {
	if nodeMap == nil {
		nodeMap = make(map[string]*PathNode)
	}
	if currNode == nil {
		currNode = tree
		nodeMap[realReceiverPeerId] = currNode
	}

	amount += float64(currNode.Level) * 1.0
	var nodes []dao.DemoChannelInfo
	err := db.Select(q.Eq("PeerIdB", realReceiverPeerId)).Find(&nodes)
	if err == nil {
		for _, item := range nodes {
			if item.AmountA > amount {
				interSender := item.PeerIdA
				peer2Peer := item.PeerIdA + "2" + item.PeerIdB
				if branchMap[peer2Peer] == nil {
					branchInfo := PathBranchInfo{
						Peer2Peer: peer2Peer,
						Amount:    item.AmountA,
					}
					branchMap[peer2Peer] = &branchInfo
				} else {
					continue
				}

				if nodeMap[interSender] == nil {
					newNode := PathNode{
						ParentNode:     currNode,
						Level:          currNode.Level + 1,
						CurrNodePeerId: interSender,
						IsRoot:         false,
						PeerMap:        make(map[string]*PathNode),
					}
					nodeMap[interSender] = &newNode
					currNode.PeerMap[interSender] = &newNode

					if interSender == realSenderPeerId {
						newNode.IsTarget = true
						return
					} else {
						p.CreateDemoChannelNetwork(realSenderPeerId, interSender, amount, currNode, tree, nodeMap, branchMap)
					}
				}
			}
		}
	}
	nodes = make([]dao.DemoChannelInfo, 0)
	err = db.Select(q.Eq("PeerIdA", realReceiverPeerId)).Find(&nodes)
	if err == nil {
		for _, item := range nodes {
			if item.AmountB > amount {
				interSender := item.PeerIdB
				peer2Peer := item.PeerIdB + "2" + item.PeerIdA
				if branchMap[peer2Peer] == nil {
					branchInfo := PathBranchInfo{
						Peer2Peer: peer2Peer,
						Amount:    item.AmountB,
					}
					branchMap[peer2Peer] = &branchInfo
				} else {
					continue
				}
				if nodeMap[interSender] == nil {
					newNode := PathNode{
						ParentNode:     currNode,
						Level:          currNode.Level + 1,
						CurrNodePeerId: interSender,
						IsRoot:         false,
						PeerMap:        make(map[string]*PathNode),
					}
					nodeMap[interSender] = &newNode
					currNode.PeerMap[interSender] = &newNode

					if interSender == realSenderPeerId {
						newNode.IsTarget = true
						return
					} else {
						p.CreateDemoChannelNetwork(realSenderPeerId, interSender, amount, currNode, tree, nodeMap, branchMap)
					}
				}
			}
		}
	}
}

func (p *pathManager) findDemoPath(realSenderPeerId, interSenderPeerId, realReceiverPeerId string, tree *PathNode, nodeMap map[string]*PathNode, branchMap map[string]*PathBranchInfo, path []*PathNode) (error, []*PathNode) {
	if nodeMap[realSenderPeerId] == nil {
		return errors.New("not found"), nil
	}

	if path == nil {
		path = make([]*PathNode, 0)
		path = append(path, nodeMap[realSenderPeerId])
	}

	currNode := nodeMap[interSenderPeerId]
	childPeers := currNode.PeerMap
	for key, node := range childPeers {
		if key == realReceiverPeerId {
			path = append(path, node)
			return nil, path
		} else {
			interSenderPeerId = key

		}
	}

	return nil, path
}

func (p *pathManager) CreateTreeFromReceiver(realSenderPeerId, receiverPeerId string, amount float64, currNode *PathNode, tree *PathNode, nodeMap map[string]*PathNode, branchMap map[string]*PathBranchInfo) {
	if nodeMap == nil {
		nodeMap = make(map[string]*PathNode)
	}
	if tree == nil {
		return
	}

	if currNode == nil {
		currNode = tree
		nodeMap[receiverPeerId] = currNode
	}

	//PeerIdA 作为sender PeerIdB 作为receiver
	var tempPeerAChannels []dao.ChannelInfo
	err := db.Select(q.Eq("PeerIdB", receiverPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempPeerAChannels)
	if err != nil {
		log.Println(err)
	}
	//看看查到的直接通道里面有没有目标对象
	if tempPeerAChannels != nil && len(tempPeerAChannels) > 0 {
		//如果没有找到目标对象，就找中间商，中间商的下个通道的钱是否足够
		for _, channel := range tempPeerAChannels {
			interSenderPeerId := channel.PeerIdA
			commitmentTxInfo, err := getLatestCommitmentTx(channel.ChannelId, interSenderPeerId)
			if err != nil || commitmentTxInfo.AmountToRSMC < amount {
				continue
			}
			peer2Peer := channel.PeerIdA + "2" + channel.PeerIdB
			if branchMap[peer2Peer] == nil {
				branchInfo := PathBranchInfo{}
				branchInfo.Peer2Peer = peer2Peer
				branchInfo.Channel = &channel
				branchInfo.Amount = commitmentTxInfo.AmountToRSMC
				branchMap[peer2Peer] = &branchInfo
			} else {
				continue
			}

			if nodeMap[interSenderPeerId] == nil {
				p.dealNodeFromReceiver(channel, commitmentTxInfo, realSenderPeerId, interSenderPeerId, amount, currNode, tree, nodeMap, branchMap)
			}
		}
	}

	//PeerIdB 作为sender PeerIdA 作为receiver
	var tempBNodes []dao.ChannelInfo
	err = db.Select(q.Eq("PeerIdA", receiverPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempBNodes)
	if err != nil {
		log.Println(err)
	}

	//看看查到的直接通道里面有没有目标对象
	if tempBNodes != nil && len(tempBNodes) > 0 {
		//如果没有找到目标对象，就找中间商，中间商的下个通道的钱是否足够
		for _, channel := range tempBNodes {
			interSenderPeerId := channel.PeerIdB
			commitmentTxInfo, err := getLatestCommitmentTx(channel.ChannelId, interSenderPeerId)
			if err != nil || commitmentTxInfo.AmountToRSMC < amount {
				continue
			}
			peer2Peer := interSenderPeerId + "2" + channel.PeerIdA
			if branchMap[peer2Peer] == nil {
				branchInfo := PathBranchInfo{}
				branchInfo.Peer2Peer = peer2Peer
				branchInfo.Channel = &channel
				branchInfo.Amount = commitmentTxInfo.AmountToRSMC
				branchMap[peer2Peer] = &branchInfo
			} else {
				continue
			}
			if nodeMap[interSenderPeerId] == nil {
				p.dealNodeFromReceiver(channel, commitmentTxInfo, realSenderPeerId, interSenderPeerId, amount, currNode, tree, nodeMap, branchMap)
			}
		}
	}
}

func (p *pathManager) dealNodeFromReceiver(currChannel dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTransaction, realSenderPeerId, interSenderPeerId string, needReceiveAmount float64,
	parentNode *PathNode, tree *PathNode, nodeMap map[string]*PathNode, branchMap map[string]*PathBranchInfo) {

	if commitmentTxInfo.AmountToRSMC > needReceiveAmount {
		needReceiveAmount += tool.GetHtlcFee() * float64(parentNode.Level)
		newNode := &PathNode{
			ParentNode:     parentNode,
			CurrNodePeerId: interSenderPeerId,
			Level:          parentNode.Level + 1,
			IsRoot:         false,
			PeerMap:        make(map[string]*PathNode, 0),
		}
		parentNode.PeerMap[interSenderPeerId] = newNode
		nodeMap[interSenderPeerId] = newNode
		if interSenderPeerId == realSenderPeerId {
			newNode.IsTarget = true
			return
		} else {
			//查找bob的通道满足余额大于需要额度的列表
			if findNextLevelValidAlices(currChannel, interSenderPeerId, needReceiveAmount, nodeMap) {
				p.CreateTreeFromReceiver(realSenderPeerId, interSenderPeerId, needReceiveAmount, newNode, tree, nodeMap, branchMap)
			}
		}
	}
}

func findNextLevelValidAlices(currChannel dao.ChannelInfo, interSenderPeerId string, amount float64, nodeMap map[string]*PathNode) bool {
	var nextTempANodes []dao.ChannelInfo
	err := db.Select(q.Not(q.Eq("Id", currChannel.Id)), q.Eq("PeerIdB", interSenderPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&nextTempANodes)
	if err == nil {
		for _, nextNode := range nextTempANodes {
			tempCurrSender := nextNode.PeerIdA
			if nodeMap[tempCurrSender] != nil {
				continue
			}
			commitmentTxInfo, err := getLatestCommitmentTx(nextNode.ChannelId, tempCurrSender)
			if err == nil {
				if commitmentTxInfo.AmountToRSMC > amount {
					return true
				}
			}
		}
	}
	var nextTempBNodes []dao.ChannelInfo
	err = db.Select(q.Not(q.Eq("Id", currChannel.Id)), q.Eq("PeerIdA", interSenderPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&nextTempBNodes)
	if err == nil {
		for _, nextNode := range nextTempBNodes {
			tempCurrSender := nextNode.PeerIdB
			if nodeMap[tempCurrSender] != nil {
				continue
			}
			commitmentTxInfo, err := getLatestCommitmentTx(nextNode.ChannelId, tempCurrSender)
			if err == nil {
				if commitmentTxInfo.AmountToRSMC > amount {
					return true
				}
			}
		}
	}
	return false
}

func (p *pathManager) CreateTree(senderPeerId, receiverPeerId string, amount float64, currNode *TreeNode, tree *TreeNode, nodeMap map[string]TreeNode) (err error) {
	if nodeMap == nil {
		nodeMap = make(map[string]TreeNode)
	}
	var tempBNodes []dao.ChannelInfo
	err = db.Select(q.Eq("PeerIdA", senderPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempBNodes)
	if err != nil {
		log.Println(err)
	}
	//看看查到的直接通道里面有没有目标对象
	if tempBNodes != nil && len(tempBNodes) > 0 {
		//如果没有找到目标对象，就找中间商，中间商的下个通道的钱是否足够
		for _, node := range tempBNodes {
			return p.dealNode(node, senderPeerId, node.PeerIdB, receiverPeerId, amount, currNode, tree, nodeMap)
		}
	}

	var tempANodes []dao.ChannelInfo
	err = db.Select(q.Eq("PeerIdB", senderPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempANodes)
	if err != nil {
		log.Println(err)
	}

	//看看查到的直接通道里面有没有目标对象
	if tempANodes != nil && len(tempANodes) > 0 {
		//如果没有找到目标对象，就找中间商，中间商的下个通道的钱是否足够
		for _, node := range tempANodes {
			return p.dealNode(node, senderPeerId, node.PeerIdA, receiverPeerId, amount, currNode, tree, nodeMap)
		}
	}
	return errors.New("not found")
}

func (p *pathManager) dealNode(node dao.ChannelInfo, fromPeerId, targetUser, toPeerId string, amount float64, currNode *TreeNode, tree *TreeNode, nodeMap map[string]TreeNode) error {
	commitmentNode := &dao.CommitmentTransaction{}
	err := db.Select(q.Eq("ChannelId", node.ChannelId), q.Eq("Owner", fromPeerId)).OrderBy("CreateAt").Reverse().First(commitmentNode)
	if err == nil && commitmentNode.AmountToRSMC > amount {
		if targetUser == toPeerId {
			childNode := &TreeNode{
				parentNode:     currNode,
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
					currNodePeerId: targetUser,
					isTarget:       false,
					amount:         commitmentNode.AmountToRSMC,
					channel:        &node,
					children:       make([]*TreeNode, 0),
				}
				if tree == nil {
					rootNode := &TreeNode{
						parentNode:     nil,
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
				return p.CreateTree(targetUser, toPeerId, amount, childNode, tree, nodeMap)
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
