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

	data.ChainHash = config.Init_node_chain_hash
	tempId, _ := c.getTemporaryChannelId()
	data.TemporaryChannelId = *tempId

	db, err := dao.DB_Manager.GetDB()
	if err != nil {
		log.Println(err)
		return err
	}
	return db.Save(data)
}

func (c *ChannelManager) getTemporaryChannelId() (tempId *chainhash.Hash, err error) {
	uuidStr, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	hash.Write([]byte(uuidStr.String()))
	sum := hash.Sum(nil)
	return chainhash.NewHash(sum)
}
