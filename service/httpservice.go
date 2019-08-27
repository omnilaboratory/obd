package service

import (
	"LightningOnOmni/bean"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type httpService struct{}

var HttpService httpService

func (s httpService) TestBd(context *gin.Context) {
	node, err := FundingTransactionService.CreateFundingTx("")
	if err != nil {
		context.JSON(http.StatusOK, gin.H{
			"msg":  "userInfo",
			"data": err.Error(),
		})
		return
	}
	bytes, _ := json.Marshal(node)
	context.JSON(http.StatusOK, gin.H{
		"msg":  "test CreateFunding",
		"data": string(bytes),
	})
}

func (s httpService) UserInfo(context *gin.Context) {
	user, err := UserService.UserInfo(context.Query("email"))
	if err != nil {
		context.JSON(http.StatusOK, gin.H{
			"msg":  "userInfo",
			"data": err.Error(),
		})
		return
	}
	bytes, _ := json.Marshal(user)
	context.JSON(http.StatusOK, gin.H{
		"msg":  "userInfo",
		"data": string(bytes),
	})
}

func (s httpService) UserLogin(context *gin.Context) {
	user := bean.User{}
	user.PeerId = context.Query("email")
	UserService.UserLogin(&user)
	bytes, _ := json.Marshal(user)
	context.JSON(http.StatusOK, gin.H{
		"msg":  "userLogin",
		"data": string(bytes),
	})
}
func (s httpService) UserLogout(context *gin.Context) {
	user := bean.User{}
	user.PeerId = context.Query("email")
	logout := UserService.UserLogout(&user)
	if logout != nil {
		context.JSON(http.StatusOK, gin.H{
			"msg":  "userLogout",
			"data": logout.Error(),
		})
	} else {
		bytes, _ := json.Marshal(user)
		context.JSON(http.StatusOK, gin.H{
			"msg":  "userLogout",
			"data": string(bytes),
		})
	}

}

func (s httpService) GetNodeData(context *gin.Context) {
	nodeService := NodeService{}
	id, _ := strconv.Atoi(context.Query("id"))
	data, _ := nodeService.Get(id)
	bytes, _ := json.Marshal(data)

	context.JSON(http.StatusOK, gin.H{
		"msg":  "getNodeData",
		"data": string(bytes),
	})

}
