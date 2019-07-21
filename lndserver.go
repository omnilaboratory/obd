package main

import (
	"LightningOnOmni/routers"
	"LightningOnOmni/service"
	"fmt"
	"log"
	"net/http"
)

func main() {
	service.Global_params.Interval = 1000
	service.Global_params.MaximumClients = 1024 * 1024
	service.Global_params.PoolSize = 4 * 1024
	fmt.Println("Starting application...")
	go service.Global_manager.Start()
	log.Fatal(http.ListenAndServe(":60020", routers.InitRouter()))
}
