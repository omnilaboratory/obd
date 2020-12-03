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
		log.Println(string(body))
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
		log.Println(string(body))
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
func HttpEstimateSmartFeeFromTracker() (result float64) {
	url := "http://" + config.TrackerHost + "/api/rpc/estimateSmartFee"
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
