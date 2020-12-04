package conn

import (
	"errors"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HttpGetBlockCountFromTracker() (flag int) {
	url := "http://" + config.TrackerHost + "/api/rpc/getBlockCount"
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return int(gjson.Get(string(body), "data").Int())
	}
	return 0
}

func HttpGetOmniBalanceFromTracker(address string, propertyId int) (balance float64) {
	url := "http://" + config.TrackerHost + "/api/rpc/getOmniBalance?address=" + address + "&propertyId=" + strconv.Itoa(propertyId)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return gjson.Get(string(body), "data").Float()
	}
	return 0
}
func HttpListReceivedByAddressFromTracker(address string) (result string) {
	if tool.CheckIsAddress(address) == false {
		return ""
	}
	url := "http://" + config.TrackerHost + "/api/rpc/listReceivedByAddress?address=" + address
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return gjson.Get(string(body), "data").Str
	}
	return ""
}
func HttpImportAddressFromTracker(address string) (result string) {
	if tool.CheckIsAddress(address) == false {
		return ""
	}
	url := "http://" + config.TrackerHost + "/api/rpc/importAddress?address=" + address
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return gjson.Get(string(body), "data").Str
	}
	return ""
}

func HttpGetTransactionByIdFromTracker(txid string) (result string) {
	url := "http://" + config.TrackerHost + "/api/rpc/getTransactionById?txid=" + txid
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return gjson.Get(string(body), "data").Str
	}
	return ""
}
func HttpListUnspentFromTracker(address string) (result string) {
	if tool.CheckIsAddress(address) == false {
		return ""
	}
	url := "http://" + config.TrackerHost + "/api/rpc/listUnspent?address=" + address
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return gjson.Get(string(body), "data").Str
	}
	return ""
}

var smartFeeSpanTime time.Time
var cacheFeeRate float64

func HttpEstimateSmartFeeFromTracker() (result float64) {
	if smartFeeSpanTime.IsZero() == false {
		if time.Now().Sub(smartFeeSpanTime).Minutes() > 10 {
			cacheFeeRate = 0
		}
	}
	if cacheFeeRate == 0 {
		url := "http://" + config.TrackerHost + "/api/rpc/estimateSmartFee"
		log.Println(url)
		resp, err := http.Get(url)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			cacheFeeRate = gjson.Get(string(body), "data").Float()
			smartFeeSpanTime = time.Now()
		}
	}
	return cacheFeeRate
}

func HttpCreateRawTransactionFromTracker(data string) (result string) {
	url := "http://" + config.TrackerHost + "/api/rpc/createRawTransaction"
	log.Println(url)
	request, _ := http.NewRequest("POST", url, strings.NewReader(data))
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return gjson.Get(string(body), "data").Str
	}
	return ""
}

func HttpOmniGetAllBalancesForAddressFromTracker(address string) (result string) {
	if tool.CheckIsAddress(address) == false {
		return ""
	}
	url := "http://" + config.TrackerHost + "/api/rpc/omniGetAllBalancesForAddress?address=" + address
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return gjson.Get(string(body), "data").Str
	}
	return ""
}
func HttpOmniGetBalancesForAddressFromTracker(address string, propertyId int) (result string) {
	if tool.CheckIsAddress(address) == false {
		return ""
	}
	url := "http://" + config.TrackerHost + "/api/rpc/omniGetBalancesForAddress?address=" + address + "&propertyId=" + strconv.Itoa(propertyId)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return gjson.Get(string(body), "data").Str
	}
	return ""
}

func HttpTestMemPoolAcceptFromTracker(hex string) (result string) {
	if tool.CheckIsString(&hex) == false {
		return ""
	}
	url := "http://" + config.TrackerHost + "/api/rpc/testMemPoolAccept?hex=" + hex
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return gjson.Get(string(body), "data").Str
	}
	return ""
}
func HttpSendRawTransactionFromTracker(hex string) (result string, err error) {
	if tool.CheckIsString(&hex) == false {
		return "", errors.New("error hex")
	}
	url := "http://" + config.TrackerHost + "/api/rpc/sendRawTransaction?hex=" + hex
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		err = nil
		if gjson.Get(string(body), "data").Str == "" {
			err = errors.New(gjson.Get(string(body), "msg").Str)
		}
		return gjson.Get(string(body), "data").Str, err
	}
	return "", errors.New("error hex")
}
func HttpOmniDecodeTransactionFromTracker(hex string) (result string, err error) {
	if tool.CheckIsString(&hex) == false {
		return "", errors.New("error hex")
	}
	url := "http://" + config.TrackerHost + "/api/rpc/omniDecodeTransaction?hex=" + hex
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		err = nil
		if gjson.Get(string(body), "data").Str == "" {
			err = errors.New(gjson.Get(string(body), "msg").Str)
		}
		return gjson.Get(string(body), "data").Str, err
	}
	return "", errors.New("error hex")
}
func HttpOmniListTransactionsFromTracker(address string) (result string, err error) {
	if tool.CheckIsAddress(address) == false {
		return "", errors.New("error address")
	}
	url := "http://" + config.TrackerHost + "/api/rpc/OmniListTransactions?address=" + address
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		err = nil
		if gjson.Get(string(body), "data").Str == "" {
			err = errors.New(gjson.Get(string(body), "msg").Str)
		}
		return gjson.Get(string(body), "data").Str, err
	}
	return "", errors.New("error result")
}

func HttpOmniGetPropertyFromTracker(propertyId int64) (result string, err error) {
	if propertyId < 1 {
		return "", errors.New("error propertyId")
	}
	url := "http://" + config.TrackerHost + "/api/rpc/omniGetProperty?propertyId=" + strconv.Itoa(int(propertyId))
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		err = nil
		if gjson.Get(string(body), "data").Str == "" {
			err = errors.New(gjson.Get(string(body), "msg").Str)
		}
		return gjson.Get(string(body), "data").Str, err
	}
	return "", errors.New("error result")
}
func HttpOmniGettransactionFromTracker(txid string) (result string, err error) {
	if tool.CheckIsString(&txid) == false && len(txid) != 64 {
		return "", errors.New("wrong txid")
	}
	url := "http://" + config.TrackerHost + "/api/rpc/omniGettransaction?txid=" + txid
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		err = nil
		if gjson.Get(string(body), "data").Str == "" {
			err = errors.New(gjson.Get(string(body), "msg").Str)
		}
		return gjson.Get(string(body), "data").Str, err
	}
	return "", errors.New("error result")
}
