package service

import (
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/config"
	"crypto/sha256"
	"github.com/satori/go.uuid"
)

//type = -32
type OpenChannelInfo struct {
	Chain_hash                    chainhash.ChainHash `json:"chain_hash"`
	Temporary_channel_id          chainhash.Hash      `json:"temporary_channel_id"`
	funding_satoshis              uint64              `json:"funding_satoshis"`
	push_msat                     uint64              `json:"push_msat"`
	dust_limit_satoshis           uint64              `json:"dust_limit_satoshis"`
	max_htlc_value_in_flight_msat uint64              `json:"max_htlc_value_in_flight_msat"`
	channel_reserve_satoshis      uint64              `json:"channel_reserve_satoshis"`
	htlc_minimum_msat             uint64              `json:"htlc_minimum_msat"`
	feerate_per_kw                uint32              `json:"feerate_per_kw"`
	to_self_delay                 uint16              `json:"to_self_delay"`
	max_accepted_htlcs            uint16              `json:"max_accepted_htlcs"`
	funding_pubkey                chainhash.Point     `json:"funding_pubkey"`
	revocation_basepoint          chainhash.Point     `json:"revocation_basepoint"`
	payment_basepoint             chainhash.Point     `json:"payment_basepoint"`
	delayed_payment_basepoint     chainhash.Point     `json:"delayed_payment_basepoint"`
	htlc_basepoint                chainhash.Point     `json:"htlc_basepoint"`
}

//type = -33
type AcceptChannelInfo OpenChannelInfo

//type: -38 (close_channel)
type Close_channel struct {
	channel_id   chainhash.Hash
	len          uint16
	scriptpubkey []byte
	signature    chainhash.Signauture
}

type ChannelManager struct {
}

var Channel_Service = ChannelManager{}

// openChannel init data
func (c *ChannelManager) OpenChannel(data *OpenChannelInfo) error {
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
