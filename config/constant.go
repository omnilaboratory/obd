package config

const (
	Init_node_chain_hash = "1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P"
	//Dust                 = 0.00000540
)

//database
const (
	DBname        = "obdserver.db"
	TrackerDbName = "trackerServer.db"
)

func GetHtlcFee() float64 {
	return 0.00001
}

// ins*150 + outs*34 + 10 + 80 = transaction size
// https://shimo.im/docs/5w9Fi1c9vm8yp1ly
func GetMinerFee() float64 {
	return 0.00003
}

func GetOmniDustBtc() float64 {
	return 0.0000054
}
