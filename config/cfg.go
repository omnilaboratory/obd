package config

import (
	"log"
	"time"

	"github.com/go-ini/ini"
)

var (
	Cfg          *ini.File
	ServerPort   = 60020
	GrpcPort     = 60021
	ReadTimeout  = 5 * time.Second
	WriteTimeout = 10 * time.Second

	//0.13 omnicore
	ChainNode_Host = "62.234.216.108:18334"
	//0.18 omnicore
	ChainNode_Type = "test"
	//ChainNode_Host = "62.234.216.108:18334"
	ChainNode_User = "omniwallet"
	ChainNode_Pass = "cB3]iL2@eZ1?cB2?"

	//P2P
	P2P_sourcePort = 3001
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
	GrpcPort = section.Key("grpcPort").MustInt(60021)
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

	p2pNode, err := Cfg.GetSection("p2p")
	if err != nil {
		log.Println(err)
		return
	}
	P2P_sourcePort = p2pNode.Key("sourcePort").MustInt()
}
