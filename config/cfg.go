package config

import (
	"log"
	"strconv"
	"time"

	"github.com/go-ini/ini"
)

var (
	Cfg               *ini.File
	ServerPort        = 60020
	ReadTimeout       = 5 * time.Second
	WriteTimeout      = 10 * time.Second
	TrackerHost       = "localhost:60060"
	TrackerServerPort = 60060

	ChainNode_Type = "test"
	ChainNode_Host = "62.234.216.108:18332"
	ChainNode_User = "omniwallet"
	ChainNode_Pass = "cB3]iL2@eZ1?cB2?"
	//mainnet
	//	//ChainNode_Host = "62.234.188.160:8332"
	//	//ChainNode_User = "uprets"
	//	//ChainNode_Pass = "pass"

	//P2P
	P2P_hostIp     = "127.0.0.1"
	P2P_sourcePort = 4001
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	Cfg, err := ini.Load("config/conf.ini")
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
	ReadTimeout = time.Duration(section.Key("readTimeout").MustInt(5)) * time.Second
	WriteTimeout = time.Duration(section.Key("writeTimeout").MustInt(5)) * time.Second

	chainNode, err := Cfg.GetSection("chainNode")
	if err != nil {
		log.Println(err)
		return
	}
	ChainNode_Type = chainNode.Key("netType").String()
	ChainNode_Host = chainNode.Key("host").String()
	ChainNode_User = chainNode.Key("user").String()
	ChainNode_Pass = chainNode.Key("pass").String()
	if ChainNode_Type == "main" {
		chainNodeMain, err := Cfg.GetSection("chainNodeMain")
		if err != nil {
			log.Println(err)
			return
		}
		ChainNode_Host = chainNodeMain.Key("host").String()
		ChainNode_User = chainNodeMain.Key("user").String()
		ChainNode_Pass = chainNodeMain.Key("pass").String()
	}
	if len(ChainNode_Host) == 0 {
		log.Println("empty omnicore host")
		return
	}
	if len(ChainNode_User) == 0 {
		log.Println("empty omnicore account")
		return
	}
	if len(ChainNode_Pass) == 0 {
		log.Println("empty omnicore password")
		return
	}

	p2pNode, err := Cfg.GetSection("p2p")
	if err != nil {
		log.Println(err)
		return
	}
	P2P_hostIp = p2pNode.Key("hostIp").String()
	P2P_sourcePort = p2pNode.Key("sourcePort").MustInt()

	//tracker
	tracker, err := Cfg.GetSection("tracker")
	if err != nil {
		log.Println(err)
		return
	}
	if len(tracker.Key("hostIp").String()) == 0 {
		panic("empty tracker hostIp")
	}

	TrackerServerPort = tracker.Key("port").MustInt(60060)
	TrackerHost = tracker.Key("hostIp").MustString("localhost") + ":" + strconv.Itoa(TrackerServerPort)
}
