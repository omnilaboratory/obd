package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/tracker/service"
	"github.com/satori/go.uuid"
	"net/http"
	"strings"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors())
	go service.ObdNodeManager.TrackerStart()
	router.GET("/ws", wsClientConnect)

	apiv1 := router.Group("/api/v1/")
	{
		apiv1.GET("GetHtlcCurrState", service.HtlcService.GetHtlcCurrState)
		apiv1.GET("getChannelState", service.ChannelService.GetChannelState)
		apiv1.GET("getUserState", service.NodeAccountService.GetUserState)
		apiv1.GET("getUserP2pNodeId", service.NodeAccountService.GetUserP2pNodeId)
		apiv1.GET("getNodeInfoByP2pAddress", service.NodeAccountService.GetNodeInfoByP2pAddress)
	}
	apiv2 := router.Group("/api/common/")
	{
		apiv2.GET("getObdNodes", service.NodeAccountService.GetAllObdNodes)
		apiv2.GET("getUsers", service.NodeAccountService.GetAllUsers)
		apiv2.GET("getChannels", service.ChannelService.GetChannels)
	}

	apiv3 := router.Group("/api/rpc/")
	{
		apiv3.GET("getChainNodeType", service.RpcService.GetChainNodeType)
		apiv3.GET("getBlockCount", service.RpcService.GetBlockCount)
		apiv3.GET("getOmniBalance", service.RpcService.GetOmniBalance)
		apiv3.GET("getBalanceByAddress", service.RpcService.GetBalanceByAddress)
		apiv3.GET("importAddress", service.RpcService.ImportAddress)
		apiv3.GET("listReceivedByAddress", service.RpcService.ListReceivedByAddress)
		apiv3.GET("getTransactionById", service.RpcService.GetTransactionById)
		apiv3.GET("omniGettransaction", service.RpcService.OmniGettransaction)
		apiv3.GET("omniGetProperty", service.RpcService.OmniGetProperty)
		apiv3.POST("createRawTransaction", service.RpcService.CreateRawTransaction)
		apiv3.GET("estimateSmartFee", service.RpcService.EstimateSmartFee)
		apiv3.GET("listUnspent", service.RpcService.ListUnspent)
		apiv3.GET("omniGetAllBalancesForAddress", service.RpcService.OmniGetAllBalancesForAddress)
		apiv3.GET("omniGetBalancesForAddress", service.RpcService.OmniGetBalancesForAddress)
		apiv3.GET("testMemPoolAccept", service.RpcService.TestMemPoolAccept)
		apiv3.GET("sendRawTransaction", service.RpcService.SendRawTransaction)
		apiv3.GET("omniDecodeTransaction", service.RpcService.OmniDecodeTransaction)
		apiv3.GET("omniListTransactions", service.RpcService.OmniListTransactions)
		apiv3.GET("getNewAddress", service.RpcService.GetNewAddress)
		apiv3.GET("omniSend", service.RpcService.OmniSend)
		apiv3.GET("omniListProperties", service.RpcService.OmniListProperties)
		apiv3.GET("omniSendIssuanceFixed", service.RpcService.OmniSendIssuanceFixed)
		apiv3.GET("omniSendIssuanceManaged", service.RpcService.OmniSendIssuanceManaged)
		apiv3.GET("omniSendGrant", service.RpcService.OmniSendGrant)
		apiv3.GET("omniSendRevoke", service.RpcService.OmniSendRevoke)
		apiv3.GET("btcSignRawTransactionFromJson", service.RpcService.BtcSignRawTransactionFromJson)
		apiv3.GET("getMiningInfo", service.RpcService.GetMiningInfo)
		apiv3.GET("getNetworkInfo", service.RpcService.GetNetworkInfo)
	}

	return router
}

//跨域
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method               //请求方法
		origin := c.Request.Header.Get("Origin") //请求头部
		var headerKeys []string                  // 声明请求头keys
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")                                       // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE") //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")                                                                                                                                                           // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")                                                                                                                                                  //  跨域请求是否需要带cookie信息 默认设置为true
			c.Set("content-type", "application/json")                                                                                                                                                              // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next() //  处理请求
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
	go newClient.Read()
}
