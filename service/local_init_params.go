package service

import (
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/rpc"
)

type commitmentOutputBean struct {
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
var TrackerChan chan []byte

//for store the privateKey
var tempAddrPrivateKeyMap = make(map[string]string)

var OnlineUserMap = make(map[string]*bean.User)
