package config

import (
	"flag"
	"log"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/go-ini/ini"
)

var (
	//Cfg               *ini.File
	configPath   = flag.String("configPath", "config/conf.ini", "Config file path")
	ServerPort   = 60020
	ReadTimeout  = 60 * time.Second
	WriteTimeout = 60 * time.Second

	HtlcFeeRate = 0.0001
	HtlcMaxFee  = 0.01

	TrackerHost = "localhost:60060"

	ChainNodeType = "test"
	//P2P
	P2P_hostIp     = "127.0.0.1"
	P2P_sourcePort = 4001
	BootstrapPeers addrList
)

func parseHostname(hostname string) string {
	P2pHostIps, err := net.LookupIP(hostname)
	if err != nil {
		panic("Can't parse hostname")
	}

	return P2pHostIps[0].String()
}

func Init() {
	testing.Init()
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//Cfg, err := ini.Load("config/conf.ini")
	Cfg, err := ini.Load(*configPath)
	if err != nil {
		log.Println(err)
		return
	}
	section, err := Cfg.GetSection("server")
	if err != nil {
		log.Println(err)
		return
	}
	ServerPort = section.Key("port").MustInt(60020)
	ReadTimeout = time.Duration(section.Key("readTimeout").MustInt(60)) * time.Second
	WriteTimeout = time.Duration(section.Key("writeTimeout").MustInt(60)) * time.Second

	htlcNode, err := Cfg.GetSection("htlc")
	if err != nil {
		log.Println(err)
		return
	}
	HtlcFeeRate = htlcNode.Key("feeRate").MustFloat64(0.0001)
	HtlcMaxFee = htlcNode.Key("maxFee").MustFloat64(0.01)

	p2pNode, err := Cfg.GetSection("p2p")
	if err != nil {
		log.Println(err)
		return
	}

	P2P_hostIp = parseHostname(p2pNode.Key("hostIp").String())
	P2P_sourcePort = p2pNode.Key("sourcePort").MustInt()

	//tracker
	tracker, err := Cfg.GetSection("tracker")
	if err != nil {
		log.Println(err)
		return
	}

	if len(tracker.Key("host").String()) == 0 {
		panic("empty tracker host")
	}

	RawTrackerHost := tracker.Key("host").MustString("localhost:60060")
	RawHostIP := strings.Split(RawTrackerHost, ":")
	ParseHostname := parseHostname(RawHostIP[0])
	TrackerHost = ParseHostname + ":" + RawHostIP[1]

}
