package rpc

import (
	"encoding/hex"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"testing"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func TestClient_GetMiningInfo(t *testing.T) {

	client := NewClient()

	//s, err := client.OmniGetAllBalancesForAddress("n1zZKFcbmNCFG2zb3xTLsZ3CEk6uDZnX7z")
	//log.Println(s)
	//log.Println(err)
	//return

	hex := "0200000002ffba9dad856f0dcf400b884f1be3115b57ccb7ae7543d2a97994dd2c9f99b09e00000000d900473044022072ddc4f59a68056db70fde556d4cdae19e7764dfc1cbee3aafcb71c7e956025b0220305ac730a8e61d25948998dcf960a62859bb187af3108465104c859908240da80147304402202d74600caa77adc93329d9f05534b72453282978a5e820002095479c2daa696d02204d07cc5e20b7c377311fc6ca6b88d1fa875a85147afefb255d770983db5f67290147522103985e8880628058da7c49b0968e4e7d2819240b60255a1c9b5f2407a4056b5f54210348776fb7b9b8e0477059ea9a7baec286aeba71d9eda9935c85158b67ee2533af52aee8030000ffba9dad856f0dcf400b884f1be3115b57ccb7ae7543d2a97994dd2c9f99b09e02000000d90047304402201e130a551a0fb7cb9c037ae347e051bc4c0127ca731815e1fd717c4a3317097402203aaa3c2be6ce294a51d9fcfefecf3c27ebb3c37142f45202f4c52532c03ebcd701473044022065658728e48f5ab7ad05a4ad19de753d1d422129285f9da872b60334ee0efb760220396b6633ee59677454c9e8753370f58c2f0e37a715b226de9a5e5a5d0f7733c30147522103985e8880628058da7c49b0968e4e7d2819240b60255a1c9b5f2407a4056b5f54210348776fb7b9b8e0477059ea9a7baec286aeba71d9eda9935c85158b67ee2533af52aee8030000033c500000000000001976a9147a019f584f6a65d114d5f17264c9eb32f763d72c88ac0000000000000000166a146f6d6e69000000000000008900000000769cfd8022020000000000001976a9147a019f584f6a65d114d5f17264c9eb32f763d72c88ac00000000"
	result, err := client.DecodeRawTransaction(hex)
	log.Println(err)
	log.Println(result)
	//hex = "020000000173ecf2230c838c61d8fad3b446475bfe89fdccfa35e0121dd54fc6efb0f47f1d00000000d90047304402206a4c6607a9c454323f763052e065c7290559965612bed6e1d1e53256d72a38de02206e149f2ef912f9e41c0fa68abb25bc06bf9cf1b3549b73ed5bc945a8700107230147304402200b9c1e91881359d02c354cbab7d166e11ec9dc1397d4c624a743a946685e827802207c004e4bcad33843b90498855a307db574cc1c17199818a5b5f78845fa4feb430147522102c57b02d24356e1d31d34d2e3a09f7d68a4bdec6c0556595bb6391ce5d6d4fc6621030138baf7b8df30e1aa40ee42f349e9b0d4c92abb0ee37b2c9d61bf0df58f408252aeffffffff03344700000000000017a914bc86aa4fe56c2efda016bb1b0fa928372b8c51ab870000000000000000166a146f6d6e6900000000000000890000000017d784001c0200000000000017a914bc86aa4fe56c2efda016bb1b0fa928372b8c51ab8700000000"
	//result, err = client.DecodeRawTransaction(hex)
	//log.Println(err)
	//log.Println(result)
	//result, err := client.GetMiningInfo()
	//log.Println(err)
	//log.Println(result)
	//
	//result, err = client.ListUnspent("12A7mKppn4XsYBzPDGg8HY1L2zRS1uFeWS")
	//log.Println(err)
	//log.Println(result)
	//
	//result, err = client.ListUnspent("1FuiQiycRNxfWy5twwEbQbQkWyFUntgbCG")
	//log.Println(err)
	//log.Println(result)

	//client.send("importaddress", []interface{}{"12A7mKppn4XsYBzPDGg8HY1L2zRS1uFeWS", "", true})
	//log.Println(1)
	//client.send("importaddress", []interface{}{"1KoMjWRTRRZogAEZKYAhNKgtejzb4wGPPW", "", true})
	//log.Println(2)
	//client.send("importaddress", []interface{}{"mqnj5uu2jRwY5pe3Y8YyQqpJ6UKgEyqKuY", "", true})
	//log.Println(3)
	return

	//result, err = client.send("importaddress", []interface{}{"mre4gBmjKiBm8gwZmpCNcnnHiDY7TXr2wD", "", false})
	//log.Println(err)
	//log.Println(result)

	isValid, err := client.ValidateAddress("mfteg3UFwYQVRtYV6NXPaKyLCcmBwGuAXu")
	log.Println(isValid)
	log.Println(err)

	result, err = client.ListUnspent("mfteg3UFwYQVRtYV6NXPaKyLCcmBwGuAXu")
	log.Println(err)
	log.Println(result)

}

func TestDemo(t *testing.T) {
	name := "Yuki will you marry me ? Tetsu.Yuki will you marry me ? Tetsu."
	s := hex.EncodeToString([]byte(name))
	log.Println(s)
	//bytes, e := hex.DecodeString(s)

	bytes, e := hex.DecodeString("03c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee961")
	log.Println(e)
	log.Println(string(bytes))
}

func TestClient_GetBlockCount(t *testing.T) {
	client := NewClient()
	balance, err := client.GetBalanceByAddress("2N3vGUfBSNALGGxUo8gTYpVQAmLwjXomLhF")
	log.Println(err)
	log.Println(balance)
}

func TestClient_SendRawTransaction(t *testing.T) {
	client := NewClient()
	result, err := client.SendRawTransaction("0200000001788435d51b49b3e9786e9b5b00c76d1684f72ea537d20980fa254d53f00480aa00000000d900473044022056546f616136aa143eb43014ed9e1eccc47ddf70be9c6b28e876a689f983befa022004239aee7b8fa4175e1f7372bfebe3ff1c6689679f3502288cecf0f7fc52e25c014730440220682f20dedef38b05a38b35c8b27903c3fb5b3325dd54f614b6d20a1c0c7ac5a70220268c34a86b8de1dbee3a8dfc584bee4b8ed1591dc62cfaa113d36c69316098920147522103b4df7d3026a437f537dcc0a9e681ffdfb000c9f1223189adf18364588d46e55921036f8a9b88615bb30d9c2dcbf7ef869134f46bf70394c5cb1f440c68ee2f136aaa52aeffffffff033c1900000000000017a914f07c2b51b5774d534ec389937da9232e147b5fb8870000000000000000166a146f6d6e690000000000000079000000000bebc2001c0200000000000017a914f07c2b51b5774d534ec389937da9232e147b5fb88700000000")
	//result, err := client.GetAddressInfo("2MwKVXga7i82DgwmQ9nTPFSuAGP6pTkNQYr")
	log.Println(err)
	log.Println(result)
}

func TestClient_GetBalanceByAddress(t *testing.T) {
	client := NewClient()

	privkeys := []string{
		"cTBs2yp9DFeJhsJmg9ChFDuC694oiVjSakmU7s6CFr35dfhcko1V",
	}

	//srciptPubkey := "a91475138ee96bf42cec92a6815d4fd47b821fbdeceb87"
	outputItems := []TransactionOutputItem{
		{ToBitCoinAddress: "2Mx1x4dp19FUvHEoyM2Lt5toX4n22oaTXxo", Amount: 0.0001},
	}

	redeemScript := "52210389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0210303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e2852ae"

	txid, hex, err := client.BtcCreateAndSignRawTransaction("2N3vGUfBSNALGGxUo8gTYpVQAmLwjXomLhF", privkeys, outputItems, 0, 0, &redeemScript)
	log.Println(err)
	//log.Println(hex)
	log.Println(txid)

	privkeys = []string{
		"cUC9UsuybBiS7ZBFBhEFaeuhBXbPSm6yUBZVaMSD2DqS3aiBouvS",
	}

	fromBitCoinAddress := "2N3vGUfBSNALGGxUo8gTYpVQAmLwjXomLhF"
	result, err := client.ListUnspent(fromBitCoinAddress)

	array := gjson.Parse(result).Array()
	log.Println("listunspent", array)

	//out, _ := decimal.NewFromFloat(minerFee).Add(outTotalAmount).Float64()

	balance := 0.0
	var inputs []map[string]interface{}
	for _, item := range array {
		node := make(map[string]interface{})
		node["txid"] = item.Get("txid").String()
		node["vout"] = item.Get("vout").Int()
		node["redeemScript"] = redeemScript
		node["scriptPubKey"] = item.Get("scriptPubKey").String()
		inputs = append(inputs, node)
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Get("amount").Float())).Round(8).Float64()
	}
	log.Println("input list ", inputs)

	hex, err = client.SignRawTransactionWithKey(hex, privkeys, inputs, "NONE|ANYONECANPAY")
	parse := gjson.Parse(hex)
	//log.Println(parse)
	//log.Println(err)
	//log.Println(hex)
	result, err = client.DecodeRawTransaction(parse.Get("hex").String())
	//log.Println(result)
	log.Println(gjson.Get(result, "txid"))
	//result, err := client.SendRawTransaction(hex)
	//log.Println(err)
	//log.Println(result)
}
