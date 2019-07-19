package modules

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"math"
)

const (
	Init_node_chain_hash = "1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P"
	maxBtcPaymentMSat    = lnwire.MilliSatoshi(math.MaxUint32)
)

//数据库
const (
	DBname     = "lndserver.db"
	Userbucket = "user"
)
