package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/config"
	"crypto/sha256"
	"github.com/satori/go.uuid"
)

type ChannelManager struct{}

var Channel_Service = ChannelManager{}

// openChannel init data
func (c *ChannelManager) OpenChannel(data *bean.OpenChannelInfo) error {
	data.Chain_hash = config.Init_node_chain_hash
	tempId, _ := c.getTemporayChaneelId()
	data.Temporary_channel_id = *tempId
	return nil
}

func (c *ChannelManager) getTemporayChaneelId() (tempId *chainhash.Hash, err error) {
	uuidStr, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	hash.Write([]byte(uuidStr.String()))
	sum := hash.Sum(nil)
	return chainhash.NewHash(sum)
}
