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

	hex := "0200000002a556f8850b3da1db5446d800e29df1ae48cdfaa16b80034b62fcbfc6f9ce59c400000000d9004730440220318226a4a1b2c1c5001539c2738f36ece6801c4f9a3681de2a3f9cc13564cb9c022030ac8dcedefdc9d969b8510bb0159ffab88c7718dd40119c9df19c92f0f031630147304402205dc31458000fe4acea94e57d37a8c9a7914a9b26a3b87cc42cad1099da6d3ce9022077575e3b7d795f3c2dfb31b8213f0b17eb5b259028fc8d310140fe35b41ad1db0147522102fcd183938b962b42d1a541d1d9444b9ff4ed199b000b4f24011b9f0dacc1c62a210238b21ae976c7059057b2b973fec5be8c7b37cea1b15aff52ecf309d5f527e76e52aee8030000a556f8850b3da1db5446d800e29df1ae48cdfaa16b80034b62fcbfc6f9ce59c402000000db00483045022100e195a7d6180e48f775a4f95ad22bb552de9af6a2460d777a0b21c439e6b2804f022062e3c350228ffbc7db5e316bd7324c1d45a09d2e3db917d4fc08187ae693abf501483045022100aea0d93e149fdd9c05ca7b883b39bd399a6c813bd5fdf6da3205f00ee345cace0220206069c722551e3e9629fa61861bae449ce0f91806b85ac8b86cc9a3467457210147522102fcd183938b962b42d1a541d1d9444b9ff4ed199b000b4f24011b9f0dacc1c62a210238b21ae976c7059057b2b973fec5be8c7b37cea1b15aff52ecf309d5f527e76e52aee80300000326140000000000001976a914a2bebc3bbc138a248296ad96e6aaf71d83f69c3688ac0000000000000000166a146f6d6e69000000000000008900000000000f424022020000000000001976a9146cbffc572e8842c31891ff88323c06a3c45dffef88ac00000000"
	result, err := client.DecodeRawTransaction(hex)
	log.Println(err)
	log.Println(result)
	return
	result, err = client.OmniDecodeTransaction(hex)
	log.Println(err)
	log.Println(result)
	//txId, err := client.SendRawTransaction(hex)
	//log.Println(err)
	//log.Println(txId)
	txId, err := client.OmniGetAllBalancesForAddress("mre4gBmjKiBm8gwZmpCNcnnHiDY7TXr2wD")
	log.Println(err)
	log.Println(txId)
	txId, err = client.OmniGetAllBalancesForAddress("mvqxxQskQkLVA9MNC5FFXPM1oY8uuMDBQa")
	log.Println(err)
	log.Println(txId)
	txId, err = client.OmniGettransaction("3eb735c4bdcfab26f9b8d44b31cac8d5f5de0165f15db7a997e4513c759050ff")
	//txId, err = client.OmniGettransaction("3eb735c4bdcfab26f9b8d44b31cac8d5f5de0165f15db7a997e4513c759050ff")
	log.Println(err)
	log.Println(txId)
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
