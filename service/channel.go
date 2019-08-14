package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/config"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"crypto/sha256"
	"github.com/satori/go.uuid"
	"log"
)

type ChannelManager struct{}

var ChannelService = ChannelManager{}

// openChannel init data
func (c *ChannelManager) OpenChannel(data *bean.OpenChannelInfo) error {
	client := rpc.NewClient()
	address, err := client.GetNewAddress("serverBob")
	if err != nil {
		log.Println(err)
		return err
	}

	multiAddr, err := client.CreateMultiSig(2, []string{string(data.FundingPubKey), address})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(multiAddr)

	var node = &dao.OpenChannelInfo{}

	data.ChainHash = config.Init_node_chain_hash
	sum, tempId, _ := c.getTemporaryChannelId()
	data.TemporaryChannelId = *tempId
	node.TemporaryChannelIdStr = sum

	node.OpenChannelInfo = *data

	db, err := dao.DB_Manager.GetDB()
	if err != nil {
		log.Println(err)
		return err
	}
	return db.Save(node)
}

func (c *ChannelManager) getTemporaryChannelId() (result string, tempId *chainhash.Hash, err error) {
	uuidStr, err := uuid.NewV4()
	if err != nil {
		return "", nil, err
	}
	hash := sha256.New()
	hash.Write([]byte(uuidStr.String()))
	sum := hash.Sum(nil)
	hashes, err := chainhash.NewHash(sum)
	return string(sum), hashes, err
}
