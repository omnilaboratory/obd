package service

import (
	"github.com/gin-gonic/gin"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/tidwall/gjson"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type rpcManager struct {
	mu sync.Mutex
}

var RpcService rpcManager

func (manager *rpcManager) GetOmniBalance(context *gin.Context) {
	reqData := &bean.GetOmniBalanceRequest{}
	reqData.Address = context.Query("address")
	if tool.CheckIsString(&reqData.Address) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error Address",
		})
		return
	}
	reqData.PropertyId, _ = strconv.Atoi(context.Query("propertyId"))
	if reqData.PropertyId == 0 {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error propertyId",
		})
		return
	}

	getbalance, err := rpc.NewClient().OmniGetbalance(reqData.Address, reqData.PropertyId)
	balance := 0.0
	if err == nil {
		balance = gjson.Get(getbalance, "balance").Float()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "OmniGetbalance",
		"data": balance,
	})
}

func (manager *rpcManager) ImportAddress(context *gin.Context) {
	address := context.Query("address")
	if tool.CheckIsString(&address) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error Address",
		})
		return
	}

	_, _ = rpc.NewClient().ValidateAddress(address)
	context.JSON(http.StatusOK, gin.H{
		"msg": "OmniGetbalance",
	})
}

type cacheListReceivedByAddressInfo struct {
	SpanTime time.Time
	Result   string
}

var cacheListReceivedByAddress map[string]cacheListReceivedByAddressInfo

func (manager *rpcManager) ListReceivedByAddress(context *gin.Context) {
	address := context.Query("address")
	if tool.CheckIsString(&address) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error Address",
		})
		return
	}

	if cacheListReceivedByAddress == nil {
		cacheListReceivedByAddress = make(map[string]cacheListReceivedByAddressInfo)
	}
	cacheInfo := cacheListReceivedByAddress[address]
	if cacheInfo.SpanTime.IsZero() == false {
		if time.Now().Sub(cacheInfo.SpanTime).Minutes() > 2 {
			if cacheInfo.Result != "" {
				cacheInfo.Result = ""
			}
		}
	}
	if cacheInfo.Result == "" {
		temp, err := rpc.NewClient().ListReceivedByAddress(address)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"msg": err.Error(),
			})
			return
		}
		cacheInfo.Result = temp
		cacheInfo.SpanTime = time.Now()
		cacheListReceivedByAddress[address] = cacheInfo
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "OmniGetbalance",
		"data": cacheInfo.Result,
	})
}
func (manager *rpcManager) ListUnspent(context *gin.Context) {
	address := context.Query("address")
	if tool.CheckIsString(&address) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error Address",
		})
		return
	}
	result, err := rpc.NewClient().ListUnspent(address)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "ListUnspent",
		"data": result,
	})
}

func (manager *rpcManager) GetTransactionById(context *gin.Context) {
	txid := context.Query("txid")
	if tool.CheckIsString(&txid) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error txid",
		})
		return
	}
	result, err := rpc.NewClient().GetTransactionById(txid)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "GetTransactionById",
		"data": result,
	})
}
func (manager *rpcManager) EstimateSmartFee(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"msg":  "EstimateSmartFee",
		"data": rpc.NewClient().EstimateSmartFee(),
	})
}
func (manager *rpcManager) CreateRawTransaction(context *gin.Context) {

	data := make(map[string]interface{})
	err := context.BindJSON(&data)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error data",
		})
		return
	}
	inputItems := data["inputs"].([]interface{})
	inputs := make([]map[string]interface{}, 0)
	for _, node := range inputItems {
		item := node.(map[string]interface{})
		inputs = append(inputs, item)
	}

	outputs := data["outputs"].(map[string]interface{})
	result, err := rpc.NewClient().CreateRawTransaction(inputs, outputs)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "CreateRawTransaction",
		"data": result,
	})
}

var blockSpanTime time.Time
var cacheBlockCount int

func (manager *rpcManager) GetBlockCount(context *gin.Context) {
	if blockSpanTime.IsZero() == false {
		if time.Now().Sub(blockSpanTime).Minutes() > 2 {
			cacheBlockCount = 0
		}
	}
	if cacheBlockCount == 0 {
		blockCount, _ := rpc.NewClient().GetBlockCount()
		cacheBlockCount = blockCount
		blockSpanTime = time.Now()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "blockCount",
		"data": cacheBlockCount,
	})
}
