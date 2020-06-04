package config

import (
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	Init_node_chain_hash = "1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P"
	//Dust                 = 0.00000540
)

//database
const (
	DBname        = "obdserver.db"
	TrackerDbName = "trackerServer.db"
)

func GetHtlcFee() float64 {
	return 0.00001
}

// ins*150 + outs*34 + 10 + 80 = transaction size
// https://shimo.im/docs/5w9Fi1c9vm8yp1ly
//https://bitcoinfees.earn.com/api/v1/fees/recommended
func GetMinerFee() float64 {
	price := httpGetRecommendedMiner()
	if price == 0 {
		price = 6
	} else {
		price = price / 6
	}
	if price < 4 {
		price = 4
	}
	txSize := 150 + 68 + 90
	result, _ := decimal.NewFromFloat(float64(txSize) * price).Div(decimal.NewFromFloat(100000000)).Round(8).Float64()
	return result
}

var minerFeePricePerByte = 0.0
var successGetMinerFeePriceAt time.Time

func httpGetRecommendedMiner() (price float64) {
	if successGetMinerFeePriceAt.IsZero() == false {
		now := time.Now().Add(-6 * time.Hour)
		if now.Before(successGetMinerFeePriceAt) {
			return minerFeePricePerByte
		}
	}
	url := "https://bitcoinfees.earn.com/api/v1/fees/recommended"
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		successGetMinerFeePriceAt = time.Now()
		minerFeePricePerByte = gjson.Get(string(body), "hourFee").Float()
		return minerFeePricePerByte
	}
	return 0
}

func GetMinMinerFee(ins int) float64 {
	txSize := ins*150 + 68 + 90
	result, _ := decimal.NewFromFloat(float64(txSize) * 3.5).Div(decimal.NewFromFloat(100000000)).Round(8).Float64()
	return result
}

func GetOmniDustBtc() float64 {
	return 0.00000546
}
