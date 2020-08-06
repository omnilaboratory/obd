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

	//txid, err := client.SendRawTransaction("0200000002cdec954b67e1b9dd3b4ee907c15bd8a94d3dd311ebf3c07a53c689cf22892af000000000d90047304402204a2e4cb6e226e2e367cd9c4c64e1b9ec6587a9e3618c2ac9344d34f16533aa7202204604e2d4a529d32981ddb0def384b9950bf0c1ae21cc3f346df31bdab1f2dc3a0147304402204da912ce4a6b7c10d568ff6f772b8841a77bd6fb61f19961746189cc9fc4de46022030068690af2a706972e71b478ec39378daea9e9db9abd8b38ec5e96b5637bcc601475221035ae96b9763347ea51c673b976c57e5033908a1f02c9719aba57d5ec21413b9032103d51e1c595bc45357aeab25c5c0ff3eee60aa0a0671823c541cd2a83f4358280052aeffffffffcdec954b67e1b9dd3b4ee907c15bd8a94d3dd311ebf3c07a53c689cf22892af002000000d9004730440220696688fe0c95172b6c36f16e3f2c93dcac1361a1da10cc40e0708fdde62c0389022012307b1afc6ca410ccd4c0a4886093c0ea60ae72c8ba082d32579153a14399fd01473044022005d5977a6257dab4ae46f6d84bb44e9464fd3706ab32dfee09a56cce5374847b0220242ef21b26838e5a7b801d69de377ee668d98ab4b7965860e063ecbc479185a301475221035ae96b9763347ea51c673b976c57e5033908a1f02c9719aba57d5ec21413b9032103d51e1c595bc45357aeab25c5c0ff3eee60aa0a0671823c541cd2a83f4358280052aeffffffff0336500000000000001976a9147a019f584f6a65d114d5f17264c9eb32f763d72c88ac0000000000000000166a146f6d6e690000000000000089000000000bdc7fc022020000000000001976a9147c5416c655561b2eadbcc7ca32625456854040bb88ac00000000")
	//log.Println(err)
	//log.Println(txid)

	hex := "0200000002cdec954b67e1b9dd3b4ee907c15bd8a94d3dd311ebf3c07a53c689cf22892af000000000d90047304402204a2e4cb6e226e2e367cd9c4c64e1b9ec6587a9e3618c2ac9344d34f16533aa7202204604e2d4a529d32981ddb0def384b9950bf0c1ae21cc3f346df31bdab1f2dc3a0147304402204da912ce4a6b7c10d568ff6f772b8841a77bd6fb61f19961746189cc9fc4de46022030068690af2a706972e71b478ec39378daea9e9db9abd8b38ec5e96b5637bcc601475221035ae96b9763347ea51c673b976c57e5033908a1f02c9719aba57d5ec21413b9032103d51e1c595bc45357aeab25c5c0ff3eee60aa0a0671823c541cd2a83f4358280052aeffffffffcdec954b67e1b9dd3b4ee907c15bd8a94d3dd311ebf3c07a53c689cf22892af002000000d9004730440220696688fe0c95172b6c36f16e3f2c93dcac1361a1da10cc40e0708fdde62c0389022012307b1afc6ca410ccd4c0a4886093c0ea60ae72c8ba082d32579153a14399fd01473044022005d5977a6257dab4ae46f6d84bb44e9464fd3706ab32dfee09a56cce5374847b0220242ef21b26838e5a7b801d69de377ee668d98ab4b7965860e063ecbc479185a301475221035ae96b9763347ea51c673b976c57e5033908a1f02c9719aba57d5ec21413b9032103d51e1c595bc45357aeab25c5c0ff3eee60aa0a0671823c541cd2a83f4358280052aeffffffff0336500000000000001976a9147a019f584f6a65d114d5f17264c9eb32f763d72c88ac0000000000000000166a146f6d6e690000000000000089000000000bdc7fc022020000000000001976a9147c5416c655561b2eadbcc7ca32625456854040bb88ac00000000"
	result, err := client.OmniDecodeTransaction(hex)
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

	//isValid, err := client.ValidateAddress("mfteg3UFwYQVRtYV6NXPaKyLCcmBwGuAXu")
	//log.Println(isValid)
	//log.Println(err)
	//
	//result, err = client.ListUnspent("mfteg3UFwYQVRtYV6NXPaKyLCcmBwGuAXu")
	//log.Println(err)
	//log.Println(result)

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
