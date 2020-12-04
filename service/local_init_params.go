package service

import (
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/rpc"
)

type commitmentTxOutputBean struct {
	AmountToRsmc               float64
	AmountToCounterparty       float64
	AmountToHtlc               float64
	RsmcTempPubKey             string
	HtlcTempPubKey             string
	OppositeSideChannelPubKey  string
	OppositeSideChannelAddress string
}

var P2PLocalPeerId string
var rpcClient *rpc.Client

var OnlineUserMap = make(map[string]*bean.User)
