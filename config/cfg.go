package config

import (
	"github.com/go-ini/ini"
	"log"
	"time"
)

var (
	Cfg          *ini.File
	ServerPort   int           = 60020
	ReadTimeout  time.Duration = 5 * time.Second
	WriteTimeout time.Duration = 10 * time.Second

	//Chainnode_Host string = "62.234.216.108:18332"
	Chainnode_Host string = "62.234.216.108:18334"
	Chainnode_User string = "omniwallet"
	Chainnode_Pass string = "cB3]iL2@eZ1?cB2?"

	Schedule_Delay1 time.Duration = 5 * time.Second
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
	Chainnode_Host = chainNode.Key("host").String()
	Chainnode_User = chainNode.Key("user").String()
	Chainnode_Pass = chainNode.Key("pass").String()

	schedule, err := Cfg.GetSection("schedule")
	if err != nil {
		log.Println(err)
		return
	}
	Schedule_Delay1 = time.Duration(schedule.Key("delay1").MustInt(5)) * time.Second
}
