package modules

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"math"
)

//
const (
	 
	//This value Init_node_chain_hash is the first Mastercoins were generated during the month of August 2013,
	//and its hash is the parameter [chain_hash:chain_hash] used in open_channel message. 
	 
	Init_node_chain_hash = "1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P"
	maxBtcPaymentMSat    = lnwire.MilliSatoshi(math.MaxUint32)
)

//database 
const (
	DBname     = "lndserver.db"
	Userbucket = "user"
)
