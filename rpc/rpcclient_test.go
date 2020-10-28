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
	result, _ := client.ListReceivedByAddress("2N3GWPBvcyjXoXaXjWetKTzUXndGQCkgMna")

	//balance, err := client.OmniGetbalance("n4ExCKJY11hCu3xKkjDFLE1ZG4awjYsk3E", 137)
	//log.Println(balance)

	//result := client.EstimateSmartFee()
	//result, err := client.DecodeRawTransaction("02000000018451f5b707c6f1a1eb0b3f3e9eb6d736b3a8e15c70d01e3cc136783449d13174000000006a473044022079bcf8b5a3daf409990ce5fb89b3b734f421b90b24f8262d4920f5c92630d92b02200200dbdd98029dff9bb02cbff638a446b4b5b7597a55a9b03f4e62756b1566370121022496530640e1c4c7a2a9a5878439b19eebb2fbce2fabf7a9d847402860d10813ffffffff02c43500000000000017a914129965444fa07857158de1178ace3aa3afc82efd87b00a0a00000000001976a91428272468efa366443623045f833f53e7d63aa4d388ac00000000")
	log.Println(result)

	//result, err := client.TestMemPoolAccept("0200000001367d440085ab142c1f5687c6b9376ef48af3041cb33f83516b415cc5c01e6149000000006b4830450221008dc68b00366e4b187c99cbe5c21912413aaf87c9a17aeb524e0cf644d836497d02204e431c4956812ea7d03dbca25129069f758464188f23d4dcb593ab70d3e62d58012102c57b02d24356e1d31d34d2e3a09f7d68a4bdec6c0556595bb6391ce5d6d4fc66ffffffff02409c00000000000017a9141d171331ff123a8d86037821c29fcdac760722f587e2161200000000001976a9147a019f584f6a65d114d5f17264c9eb32f763d72c88ac00000000")
	//log.Println(err)
	//log.Println(result)
	//
	//result, err = client.DecodeRawTransaction("0200000001174ed5cbf3fa9127d24fafa2e59c76ad9ab715d1b68aef106d6313f453e4f145000000006b483045022100df2fb3d0b4d016b9127f1fcb80ad0b5b81dab31bb3668bad7612b9aa417f6c9b022013adf05bd272629e720019bb1f775f8c59c0440075383dff59c4f18aa324deff0121020cbd54b5d0cd602a161beb99fe1a8f2ed9aaaf66eec054d26a617d7df67e9ae1ffffffff02202400000000000017a914000566661c7cb62897b0b0aed64731df1474f67b8776731700000000001976a9140b060fcf9c573c6290c1e5a9de315b8eda2c3a5388ac00000000")
	//log.Println(err)
	//log.Println(result)

	result, err := client.OmniDecodeTransaction("020000000196655c03ea41f34cbd461cbca281a74af50837020c0cd99c9270ffd37ac64d3b000000009200473044022064412d4f83733248d666033c85d83e2a1fc231bf15f92f1242a16cd270fc4c0d02203c6f9a78a7fcc5bca7d0d7b7bbd65a519cd3a2410e1d5df5bbc54143e27f6122010047522102c57b02d24356e1d31d34d2e3a09f7d68a4bdec6c0556595bb6391ce5d6d4fc66210360917f53381f2b05bb3ec299b6bf7e7446a5c6ed287d65cfc6e38858c380017252aeffffffff03307500000000000017a914f6542333fe0cef450702f4a982b39ad473206e9e870000000000000000166a146f6d6e690000000000000089000000000bebc2001c0200000000000017a914f6542333fe0cef450702f4a982b39ad473206e9e8700000000")
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
