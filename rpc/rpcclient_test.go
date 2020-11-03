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

	hex := "020000000196655c03ea41f34cbd461cbca281a74af50837020c0cd99c9270ffd37ac64d3b0100000000ffffffff03bec81000000000001976a9147a019f584f6a65d114d5f17264c9eb32f763d72c88ac0000000000000000166a146f6d6e690000000000000089000000000bebc2001c0200000000000017a9149dcd7d0d1fc9be6b52259f424ff00e7ff49c3a6e8700000000"
	//hex = "0200000002922eb906431e0920ad1b40d1bbc877cc10a9e366d18f1a0eefe035ec877f49c800000000930000483045022100b10b4e4290426b72a40e0cdd70bded855331fbe4d4f6b7e4449b7dd019a0814d022017280b6e50bc9ae5b2b5c370593ea75ac58d34efed7e51fecfb3528a62af2daa0147522102212af74a0c82d640fdf32ac045a04b4f0d35bbb2245a6ab27babdfe7a2dfcbb021027a6f5270e8fb897f70b35f15bfb648deabbc5b92af2c2d023aaecea5fef0c6d952aee8030000922eb906431e0920ad1b40d1bbc877cc10a9e366d18f1a0eefe035ec877f49c802000000930000483045022100d617d51fd957fb3ad9a414cfe11ef0a20f3e0f38e491da09372ddde66cc62c9d0220650e127059386ca9395c4d362176cd2c94377bc6004f0793e0a93ff464e8bd4f0147522102212af74a0c82d640fdf32ac045a04b4f0d35bbb2245a6ab27babdfe7a2dfcbb021027a6f5270e8fb897f70b35f15bfb648deabbc5b92af2c2d023aaecea5fef0c6d952aee80300000326140000000000001976a9140b060fcf9c573c6290c1e5a9de315b8eda2c3a5388ac0000000000000000166a146f6d6e6900000000000000890000000005f5e10022020000000000001976a9140b060fcf9c573c6290c1e5a9de315b8eda2c3a5388ac00000000"
	//result, err := client.TestMemPoolAccept(hex)
	//log.Println(err)
	//log.Println(result)
	//result, err = client.DecodeRawTransaction(hex)
	//log.Println(err)
	//log.Println(result)
	//pass, err := client.CheckMultiSign(hex, 1)
	//log.Println(err)
	//log.Println(pass)

	hex = "02000000017fd7d7ad18460885d0e1172c959ce5e5d05433119ecc49f99b71e55e13d85cbc00000000da0047304402204fb0654362250e00b3f16d3923cccf7f9c281784a6bc8bb99048fd4322abe71702203aac14edb8fcc303ee350f8e7ad12eec6c0671f56b88ea17ee7e18b501e1103b01483045022100a5199b14c6acba19f95372d2802d4de48a0d360120a95f5090406da2aac9eca202200e40dc24b0ade852261b023c0a5f0154d58521a7915791ecd000224d4fb6821f01475221020cbd54b5d0cd602a161beb99fe1a8f2ed9aaaf66eec054d26a617d7df67e9ae12103d30073909bd072c070ac7c9b2d75084db079f32929fd00010c20d976389dffdf52aeffffffff03181b00000000000017a9143ed93405eed1961b16b1f22ea33928fe0ddb2e2f870000000000000000166a146f6d6e6900000000000000890000000005f5e1001c0200000000000017a9143ed93405eed1961b16b1f22ea33928fe0ddb2e2f8700000000"
	result, err := client.TestMemPoolAccept(hex)
	log.Println(err)
	log.Println(result)
	result, err = client.DecodeRawTransaction(hex)
	log.Println(err)
	log.Println(result)
	pass, err := client.CheckMultiSign(false, hex, 1)
	log.Println(err)
	log.Println(pass)
	//
	//hex ="02000000029268cc2b0431d103cbeeb1ed9a54bc11acc579ebfc6f359067abee201b5297aa00000000d900473044022049098355519704d8127eb110f361d5db162f27cbbb572987a2f210df6207035102206e4a0d34e72887532ad1fafbe5c86da35a7297b610af26b0b33962b01a9d86c30147304402200eff836e6ce9be3ab77089cae34240ffbc63bd0dbbc8460ceba03bf78a98086402204122f462aef2b83a82c9802b256d2d4274ae20fc06c4f4042269442dd800028401475221035ae96b9763347ea51c673b976c57e5033908a1f02c9719aba57d5ec21413b903210360917f53381f2b05bb3ec299b6bf7e7446a5c6ed287d65cfc6e38858c380017252aee80300009268cc2b0431d103cbeeb1ed9a54bc11acc579ebfc6f359067abee201b5297aa02000000d900473044022016b104dc322a9aeb689181169cecf77ed8a126a9a4d28b2a61125953cb4f7d7102206b960be4389e19270a5ac600b4bbe5093388e981ab4cd2fb39c47af6c5f9a92d01473044022064bc9bc94f11bfb64bc7546c788ef36e68dfa03c52cf2d575adc8fb69eeaf85202205f2488bd372426c56e0e8319525d4771b894234fc862c4a8845a50863eb0279e01475221035ae96b9763347ea51c673b976c57e5033908a1f02c9719aba57d5ec21413b903210360917f53381f2b05bb3ec299b6bf7e7446a5c6ed287d65cfc6e38858c380017252aee80300000336500000000000001976a9147a019f584f6a65d114d5f17264c9eb32f763d72c88ac0000000000000000166a146f6d6e690000000000000089000000000bebc20022020000000000001976a9147a019f584f6a65d114d5f17264c9eb32f763d72c88ac00000000"
	//result, err = client.TestMemPoolAccept(hex)
	//log.Println(err)
	//log.Println(result)
	//result, err = client.DecodeRawTransaction(hex)
	//log.Println(err)
	//log.Println(result)
	//pass, err = client.CheckMultiSign(hex, 2)
	//log.Println(err)
	//log.Println(pass)
	//
	//result, err = client.OmniDecodeTransaction(hex)
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
