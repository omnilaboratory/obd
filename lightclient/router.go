package lightclient

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
	"github.com/satori/go.uuid"
	"github.com/unrolled/secure"
	"log"
	"net/http"
	"strconv"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	//router.Use(TlsHandler())

	go GlobalWsClientManager.Start()
	bean.CurrObdNodeInfo.WebsocketLink = "ws://" + config.P2P_hostIp + ":" + strconv.Itoa(config.ServerPort) + "/ws" + config.ChainNodeType
	router.GET("/ws"+config.ChainNodeType, wsClientConnect)

	return router
}

func TlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     config.P2P_hostIp + ":" + strconv.Itoa(config.ServerPort),
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}

func wsClientConnect(c *gin.Context) {

	wsConn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		http.NotFound(c.Writer, c.Request)
		return
	}

	client := &Client{
		Id:          uuid.NewV4().String(),
		Socket:      wsConn,
		SendChannel: make(chan []byte)}

	session := c.GetHeader("session")
	if session == tool.GetGRpcSession() {
		client.IsGRpcRequest = true
	}

	go client.Write()
	go client.Read()
	GlobalWsClientManager.Connected <- client
}
