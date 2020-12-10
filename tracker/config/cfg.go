package cfg

import (
	"flag"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/go-ini/ini"
)

var (
	//Cfg               *ini.File
	configPath        = flag.String("trackerConfigPath", "tracker/config/conf.ini", "Config file path")
	ReadTimeout       = 60 * time.Second
	WriteTimeout      = 60 * time.Second
	TrackerServerPort = 60060

	ChainNode_Type = "test"
	ChainNode_Host = "62.234.216.108:18332"
	ChainNode_User = "omniwallet"
	ChainNode_Pass = "cB3]iL2@eZ1?cB2?"
)

func init() {
	testing.Init()
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	Cfg, err := ini.Load(*configPath)
	if err != nil {
		if strings.Contains(err.Error(), "open tracker/config/conf.ini") {
			Cfg, err = ini.Load("config/conf.ini")
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	section, err := Cfg.GetSection("server")
	if err != nil {
		log.Println(err)
		return
	}
	TrackerServerPort = section.Key("port").MustInt(60060)

	chainNode, err := Cfg.GetSection("chainNode")
	if err != nil {
		log.Println(err)
		return
	}
	ChainNode_Host = chainNode.Key("host").String()
	ChainNode_User = chainNode.Key("user").String()
	ChainNode_Pass = chainNode.Key("pass").String()
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
}
