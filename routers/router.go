package routers

import (
	"LightningOnOmni/service"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"net/http"
)

func InitRouter() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/ws", clientConnect)
	m.HandleFunc("/test", test).Methods("GET")
	return m
}

func clientConnect(res http.ResponseWriter, req *http.Request) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if error != nil {
		http.NotFound(res, req)
		return
	}
	uuid_str, _ := uuid.NewV4()
	client := &service.Client{
		Id:           uuid_str.String(),
		Socket:       conn,
		Send_channel: make(chan []byte)}

	service.Global_manager.Register <- client
	go client.Write()
	client.Read()
}

func test(writer http.ResponseWriter, request *http.Request) {
	bytes, err := json.Marshal(&service.User{Id: "1", Email: "123@qq.com"})
	if err != nil {
		fmt.Fprintf(writer, "wrong data")
		return
	}
	fmt.Fprintf(writer, string(bytes))
}
