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

func TestDemo(t *testing.T) {
	name := "abc"
	s := hex.EncodeToString([]byte(name))
	log.Println(s)
	//bytes, e := hex.DecodeString(s)
	bytes, e := hex.DecodeString("6f6d6e6900000000800004ef0000000002faf080")
	log.Println(e)
	log.Println(string(bytes))
}

func TestClient_OmniSend(t *testing.T) {
	client := NewClient()

	channelAddr := "2N5EJHxmzRPSYxkijP95p4ZmvbJJ9Vk1hio"
	toChanelAddr := "2NCprPtFBuBdDUG7xsY7QvAMDWZwfaJvKxR"
	ChannelAddressRedeemScript := "522103b4df7d3026a437f537dcc0a9e681ffdfb000c9f1223189adf18364588d46e55921024e70d53f1e328c20d28950116d32e503c113be554af643ef68808ad83371737352ae"
	txid, hex, usedTxid, err := client.OmniCreateAndSignRawTransactionForCommitmentTx(
		channelAddr,
		[]string{
			"cVaSpCw5UaioTmsrxr3jL4bcn1wsdbbeQuUoQVW6P9pZxNhqa6wc",
			"cSV9Tmayp42czShgzkVQ8gJYBqnuZ27FdYho5QehUgQL83UJQ3rY",
		},
		toChanelAddr,
		121,
		2,
		0,
		0, &ChannelAddressRedeemScript)
	log.Println(err)
	log.Println(txid)
	log.Println(hex)
	log.Println(usedTxid)

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
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Get("amount").Float())).Float64()
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
