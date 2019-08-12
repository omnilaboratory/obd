package config

import (
	"github.com/go-ini/ini"
	"log"
	"time"
)

var (
	Cfg          *ini.File
	ServerPort   int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
)

func init() {
	Cfg, err := ini.Load("config/conf.ini")
	if err != nil {
		log.Println(err)
	}
	section, err := Cfg.GetSection("server")
	if err != nil {
		log.Println(err)
	}
	ServerPort = section.Key("port").MustInt(60020)
	ReadTimeout = time.Duration(section.Key("readTimeout").MustInt(5)) * time.Second
	WriteTimeout = time.Duration(section.Key("writeTimeout").MustInt(5)) * time.Second
}
