package main

import (
	"LightningOnOmni/config"
	"LightningOnOmni/routers"
	"log"
	"net/http"
	"strconv"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	routersInit := routers.InitRouter()
	addr := ":" + strconv.Itoa(config.ServerPort)
	server := &http.Server{
		Addr:           addr,
		Handler:        routersInit,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(server.ListenAndServe())
}
