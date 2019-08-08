package main

import (
	"LightningOnOmni/routers"
	"log"
	"net/http"
	"time"
)

func main() {
	routersInit := routers.InitRouter()
	server := &http.Server{
		Addr:           ":60020",
		Handler:        routersInit,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(server.ListenAndServe())
}
