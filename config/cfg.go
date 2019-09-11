package config

import (
	"log"
	"time"

	"github.com/go-ini/ini"
)

var (
	Cfg          *ini.File
	ServerPort   int           = 60020
	GrpcPort     int           = 60021
	ReadTimeout  time.Duration = 5 * time.Second
	WriteTimeout time.Duration = 10 * time.Second

	//0.13 omnicore
	//ChainNode_Host string = "62.234.216.108:18332"
	//0.18 omnicore
	ChainNode_Host string = "62.234.216.108:18334"
	ChainNode_User string = "omniwallet"
	ChainNode_Pass string = "cB3]iL2@eZ1?cB2?"

	Schedule_Delay10min time.Duration = 10 * time.Minute
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
	ChainNode_Host = chainNode.Key("host").String()
	ChainNode_User = chainNode.Key("user").String()
	ChainNode_Pass = chainNode.Key("pass").String()

	schedule, err := Cfg.GetSection("schedule")
	if err != nil {
		log.Println(err)
		return
	}
	Schedule_Delay10min = time.Duration(schedule.Key("delay1").MustInt(5)) * time.Second
}
