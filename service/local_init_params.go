package service

import (
	"github.com/omnilaboratory/obd/bean"
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

var OnlineUserMap = make(map[string]*bean.User)
