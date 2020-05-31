package tracker

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/tracker/service"
	"github.com/satori/go.uuid"
	"net/http"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors())
	go service.ObdNodeManager.TrackerStart()
	router.GET("/ws", wsClientConnect)

	apiv1 := router.Group("/api/v1")
	{
		apiv1.GET("/getHtlcTxState", service.HtlcService.GetHtlcCurrState)
		apiv1.GET("/getChannelState", service.ChannelService.GetChannelState)
		apiv1.GET("/getUserState", service.NodeAccountService.GetUserState)
	}
	apiv2 := router.Group("/api/common")
	{
		apiv2.GET("/getObdNodes", service.NodeAccountService.GetAllObdNodes)
		apiv2.GET("/getUsers", service.NodeAccountService.GetAllUsers)
		apiv2.GET("/getChannels", service.ChannelService.GetChannels)
	}

	return router
}

//跨域
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}

func wsClientConnect(c *gin.Context) {
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	uuidStr := uuid.NewV4()
	newClient := &service.ObdNode{
		Id:          uuidStr.String(),
		Socket:      conn,
		SendChannel: make(chan []byte)}
	service.ObdNodeManager.Connected <- newClient
	go newClient.Write()
	newClient.Read()
}
