package bean

import (
	"github.com/asdine/storm"
	"github.com/omnilaboratory/obd/bean/chainhash"
	"github.com/tyler-smith/go-bip32"
)

// OutPoint defines a bitcoin data type that is used to track previous transaction outputs.
type OutPoint struct {
	Hash  chainhash.Hash
	Index uint32
}

type User struct {
	P2PLocalAddress string    `json:"p2p_local_address"`
	P2PLocalPeerId  string    `json:"p2p_local_peer_id"`
	PeerId          string    `json:"peer_id"`
	Mnemonic        string    `json:"mnemonic"`
	State           UserState `json:"state"`
	IsAgent         bool      `json:"is_agent"`
	ChangeExtKey    *bip32.Key
	CurrAddrIndex   int       `json:"curr_addr_index"`
	Db              *storm.DB //db
}

type TransactionInputItem struct {
	Txid         string  `json:"txid"`
	ScriptPubKey string  `json:"scriptPubKey"`
	RedeemScript string  `json:"redeem_script"`
	Vout         uint32  `json:"vout"`
	Amount       float64 `json:"value"`
}

type TransactionOutputItem struct {
	ToBitCoinAddress string
	Amount           float64
}
