package service

import (
	"github.com/gin-gonic/gin"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/omnilaboratory/obd/tracker/config"
	"github.com/omnilaboratory/obd/tracker/rpc"
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
func (manager *rpcManager) GetBalanceByAddress(context *gin.Context) {
	address := context.Query("address")
	var result string
	var msg string
	if tool.CheckIsAddress(address) == false {
		msg = "wrong address"
	}
	getbalance, err := rpc.NewClient().GetBalanceByAddress(address)
	if err != nil {
		msg = err.Error()
	}
	f, flag := getbalance.Float64()
	if flag == false {
		msg = "wrong balance"
	}
	result = tool.FloatToString(f, 8)
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) ImportAddress(context *gin.Context) {
	address := context.Query("address")
	if tool.CheckIsAddress(address) == false {
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
func (manager *rpcManager) OmniGetAllBalancesForAddress(context *gin.Context) {
	address := context.Query("address")
	if tool.CheckIsString(&address) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error Address",
		})
		return
	}
	result, err := rpc.NewClient().OmniGetAllBalancesForAddress(address)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "OmniGetAllBalancesForAddress",
		"data": result,
	})
}
func (manager *rpcManager) OmniGetBalancesForAddress(context *gin.Context) {
	address := context.Query("address")
	propertyId, _ := strconv.Atoi(context.Query("propertyId"))
	if tool.CheckIsString(&address) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error Address",
		})
		return
	}
	result, err := rpc.NewClient().OmniGetbalance(address, propertyId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "OmniGetAllBalancesForAddress",
		"data": result,
	})
}
func (manager *rpcManager) TestMemPoolAccept(context *gin.Context) {
	hex := context.Query("hex")
	result, err := rpc.NewClient().TestMemPoolAccept(hex)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "TestMemPoolAccept",
		"data": result,
	})
}
func (manager *rpcManager) SendRawTransaction(context *gin.Context) {
	hex := context.Query("hex")
	result, err := rpc.NewClient().SendRawTransaction(hex)
	msg := ""
	if err != nil {
		result = ""
		msg = err.Error()
	}

	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}
func (manager *rpcManager) OmniDecodeTransaction(context *gin.Context) {
	hex := context.Query("hex")
	result, err := rpc.NewClient().OmniDecodeTransaction(hex)
	msg := ""
	if err != nil {
		result = ""
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) OmniListTransactions(context *gin.Context) {
	address := context.Query("address")
	result, err := rpc.NewClient().OmniListTransactions(address, 100, 0)
	msg := ""
	if err != nil {
		result = ""
		msg = err.Error()
	}
	if result == "[]" {
		result = ""
		msg = "no tx"
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) GetNewAddress(context *gin.Context) {
	label := context.Query("label")
	result, err := rpc.NewClient().GetNewAddress(label)
	msg := ""
	if err != nil {
		result = ""
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) OmniSend(context *gin.Context) {
	fromAddress := context.Query("fromAddress")
	msg := ""
	if tool.CheckIsAddress(fromAddress) == false {
		msg = "wrong fromAddress"
	}

	toAddress := context.Query("toAddress")
	if tool.CheckIsAddress(toAddress) == false {
		msg = "wrong toAddress"
	}

	propertyIdStr := context.Query("propertyId")

	propertyId, err := strconv.Atoi(propertyIdStr)
	if err != nil {
		msg = "wrong propertyId"
	}
	amountStr := context.Query("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		msg = "wrong amount"
	}

	result, err := rpc.NewClient().OmniSend(fromAddress, toAddress, propertyId, amount)
	if err != nil {
		result = ""
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) OmniListProperties(context *gin.Context) {
	result, err := rpc.NewClient().OmniListProperties()
	msg := ""
	if err != nil {
		result = ""
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) OmniSendIssuanceFixed(context *gin.Context) {
	fromAddress := context.Query("fromAddress")
	msg := ""
	if tool.CheckIsAddress(fromAddress) == false {
		msg = "wrong fromAddress"
	}

	ecosystem, err := strconv.Atoi(context.Query("ecosystem"))
	if err != nil {
		msg = "wrong ecosystem"
	}
	divisibleType, err := strconv.Atoi(context.Query("divisibleType"))
	if err != nil {
		msg = "wrong divisibleType"
	}
	name := context.Query("name")
	data := context.Query("data")

	amount, err := strconv.ParseFloat(context.Query("amount"), 64)
	if err != nil {
		msg = "wrong amount"
	}

	result, err := rpc.NewClient().OmniSendIssuanceFixed(fromAddress, ecosystem, divisibleType, name, data, amount)
	if err != nil {
		result = ""
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}
func (manager *rpcManager) OmniSendIssuanceManaged(context *gin.Context) {
	fromAddress := context.Query("fromAddress")
	msg := ""
	if tool.CheckIsAddress(fromAddress) == false {
		msg = "wrong fromAddress"
	}

	ecosystem, err := strconv.Atoi(context.Query("ecosystem"))
	if err != nil {
		msg = "wrong ecosystem"
	}
	divisibleType, err := strconv.Atoi(context.Query("divisibleType"))
	if err != nil {
		msg = "wrong divisibleType"
	}
	name := context.Query("name")
	data := context.Query("data")

	result, err := rpc.NewClient().OmniSendIssuanceManaged(fromAddress, ecosystem, divisibleType, name, data)
	if err != nil {
		result = ""
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) OmniSendGrant(context *gin.Context) {
	fromAddress := context.Query("fromAddress")
	msg := ""
	if tool.CheckIsAddress(fromAddress) == false {
		msg = "wrong fromAddress"
	}

	propertyId, err := strconv.Atoi(context.Query("propertyId"))
	if err != nil {
		msg = "wrong propertyId"
	}
	amount, err := strconv.ParseFloat(context.Query("amount"), 64)
	if err != nil {
		msg = "wrong amount"
	}
	memo := context.Query("data")

	result, err := rpc.NewClient().OmniSendGrant(fromAddress, int64(propertyId), amount, memo)
	if err != nil {
		result = ""
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) OmniSendRevoke(context *gin.Context) {
	fromAddress := context.Query("fromAddress")
	msg := ""
	if tool.CheckIsAddress(fromAddress) == false {
		msg = "wrong fromAddress"
	}

	propertyId, err := strconv.Atoi(context.Query("propertyId"))
	if err != nil {
		msg = "wrong propertyId"
	}
	amount, err := strconv.ParseFloat(context.Query("amount"), 64)
	if err != nil {
		msg = "wrong amount"
	}
	memo := context.Query("data")

	result, err := rpc.NewClient().OmniSendRevoke(fromAddress, int64(propertyId), amount, memo)
	if err != nil {
		result = ""
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}
func (manager *rpcManager) BtcSignRawTransactionFromJson(context *gin.Context) {
	data := context.Query("data")
	result, err := rpc.NewClient().BtcSignRawTransactionFromJson(data)
	msg := ""
	if err != nil {
		result = ""
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
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
func (manager *rpcManager) OmniGettransaction(context *gin.Context) {
	txid := context.Query("txid")
	var err error
	var msg string
	var result string
	if tool.CheckIsString(&txid) == false {
		msg = "error txid"
	}
	result, err = rpc.NewClient().OmniGettransaction(txid)
	if err != nil {
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}
func (manager *rpcManager) OmniGetProperty(context *gin.Context) {
	propertyId := context.Query("propertyId")
	id, err := strconv.Atoi(propertyId)
	result := ""
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	result, err = rpc.NewClient().OmniGetProperty(int64(id))
	if err != nil {
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

var smartFeeSpanTime time.Time
var cacheFeeRate float64

func (manager *rpcManager) EstimateSmartFee(context *gin.Context) {
	if smartFeeSpanTime.IsZero() == false {
		if time.Now().Sub(smartFeeSpanTime).Minutes() > 10 {
			cacheFeeRate = 0
		}
	}
	if cacheFeeRate == 0 {
		fee := rpc.NewClient().EstimateSmartFee()
		cacheFeeRate = fee
		smartFeeSpanTime = time.Now()
	}

	context.JSON(http.StatusOK, gin.H{
		"msg":  "EstimateSmartFee",
		"data": cacheFeeRate,
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

func (manager *rpcManager) GetMiningInfo(context *gin.Context) {
	result, err := rpc.NewClient().GetMiningInfo()
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) GetNetworkInfo(context *gin.Context) {
	result, err := rpc.NewClient().GetNetworkInfo()
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"data": result,
	})
}

func (manager *rpcManager) GetChainNodeType(context *gin.Context) {

	context.JSON(http.StatusOK, gin.H{
		"msg":               "",
		"chainNodeType":     cfg.ChainNode_Type,
		"trackerP2pAddress": cfg.P2pLocalAddress,
	})
}
