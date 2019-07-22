package routers

import (
	"LightningOnOmni/service"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"net/http"
)

func InitRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	go service.Global_manager.Start()
	router.GET("/ws", ClientConnect)

	apiv1 := router.Group("/api/v1")
	{
		apiv1.GET("/tags", func(context *gin.Context) {
			context.JSON(http.StatusOK, gin.H{
				"msg": "test",
			})
		})
	}
	return router
}

func ClientConnect(c *gin.Context) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if error != nil {
		http.NotFound(c.Writer, c.Request)
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
