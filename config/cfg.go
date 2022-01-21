package config

import (
	"flag"
	"github.com/mitchellh/go-homedir"
	"log"
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

	GrpcServerPort = 50051

	HtlcFeeRate = 0.0001
	HtlcMaxFee  = 0.01

	TrackerHost = "127.0.0.1:60060"

	ChainNodeType = "regtest"
	//P2P
	P2P_hostIp     = "127.0.0.1"
	P2P_port       = 4001
	BootstrapPeers addrList

	Init_node_chain_hash = "1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P"

	DataDirectory     = ""
	dataDirectoryName = ".obd"
)

func Init() {
	testing.Init()
	//flag.Parse()
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
	GrpcServerPort = section.Key("grpc_server_port").MustInt(50051)
	ReadTimeout = time.Duration(section.Key("readTimeout").MustInt(60)) * time.Second
	WriteTimeout = time.Duration(section.Key("writeTimeout").MustInt(60)) * time.Second

	specifiedDataDirectory, err := section.GetKey("dataDirectory")
	if err == nil {
		DataDirectory = specifiedDataDirectory.Value()
	} else {
		homeDirectory, err := homedir.Dir()
		if err != nil {
			panic(err.Error())
		}
		DataDirectory = homeDirectory + "/" + dataDirectoryName
	}

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
	P2P_hostIp = p2pNode.Key("hostIp").String()
	P2P_port = p2pNode.Key("port").MustInt()

	//tracker
	tracker, err := Cfg.GetSection("tracker")
	if err != nil {
		log.Println(err)
		return
	}
	TrackerHost = tracker.Key("host").MustString("localhost:60060")
}
