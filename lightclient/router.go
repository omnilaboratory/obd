package lightclient

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

func InitRouter(conn *grpc.ClientConn) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	go globalWsClientManager.Start()
	router.GET("/ws", wsClientConnect)

	return router

}

func handlerRpcReq(c *gin.Context) {
	bytes, _ := c.GetRawData()
	log.Println(string(bytes))
	parse := gjson.Parse(string(bytes))
	log.Println(parse)
	c.JSON(http.StatusOK, gin.H{
		"result": parse.Raw,
	})
}

func wsClientConnect(c *gin.Context) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if error != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}

	uuidStr := uuid.NewV4()
	client := &Client{
		Id:          uuidStr.String(),
		Socket:      conn,
		SendChannel: make(chan []byte)}

	globalWsClientManager.Connected <- client
	go client.Write()
	client.Read()
}
