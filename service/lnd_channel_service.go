package service

import (
	"github.com/satori/go.uuid"
	"lnd-server/modules"
)

type ChannelManager struct {
}

var Channel_Service = ChannelManager{}

// openChannel init data
func (c *ChannelManager) OpenChannel(data *modules.OpenChannelData) error {
	data.Chain_hash = modules.Init_node_chain_hash
	uuid_str, _ := uuid.NewV4()
	println(uuid_str.String())
	data.Temporary_channel_id = uuid_str.Bytes()
	return nil
}
