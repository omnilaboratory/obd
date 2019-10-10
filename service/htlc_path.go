package service

import (
	"LightningOnOmni/dao"
	"github.com/asdine/storm/q"
	"log"
)

type pathManager struct{}

var PathService = pathManager{}

//查找出与自己相关所有通道
func (p *pathManager) GetPath(fromPeerId, toPeerId string, amount float64) (targetNode *dao.ChannelInfo, nodes []dao.ChannelInfo, err error) {
	var tempANodes []dao.ChannelInfo
	err = db.Select(q.Eq("PeerIdA", fromPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempANodes)
	if err != nil {
		log.Println(err)
	}

	for _, node := range tempANodes {
		commitmentNode := &dao.CommitmentTransaction{}
		err = db.Select(q.Eq("ChannelId", node.ChannelId), q.Eq("Owner", node.PeerIdB)).OrderBy("CreateAt").Reverse().First(commitmentNode)
		if err != nil {
			continue
		}

		if commitmentNode.AmountToRSMC < amount {
			continue
		}

		if node.PeerIdB == toPeerId {
			return &node, nil, nil
		}

		nodes = append(nodes, node)
	}

	var tempBNodes []dao.ChannelInfo
	err = db.Select(q.Eq("PeerIdB", fromPeerId), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&tempBNodes)
	if err != nil {
		log.Println(err)
	}

	for _, node := range tempBNodes {
		commitmentNode := &dao.CommitmentTransaction{}
		err = db.Select(q.Eq("ChannelId", node.ChannelId), q.Eq("Owner", node.PeerIdA)).OrderBy("CreateAt").Reverse().First(commitmentNode)
		if err != nil {
			continue
		}
		if commitmentNode.AmountToRSMC < amount {
			continue
		}
		if node.PeerIdA == toPeerId {
			return &node, nil, nil
		}
		nodes = append(nodes, node)
	}

	nodes = append(tempANodes, tempBNodes...)
	return nil, nodes, nil
}
