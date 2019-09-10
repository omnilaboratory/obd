package rpc

import (
	"github.com/tidwall/gjson"
	"log"
	"testing"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func TestClient_GetBlockCount(t *testing.T) {
	client := NewClient()
	balance, err := client.GetBalanceByAddress("mtu1CPCHK1yfTCwiTquSKRHcBrW2mHmfJH")
	log.Println(err)
	log.Println(balance)
}

func TestClient_SendRawTransaction(t *testing.T) {
	client := NewClient()
	result, err := client.SendRawTransaction("0200000001cb4b3acd38449d9a309a0e6c84d930d4880e5aca614b2e0dd9da255bb3a9b8e40000000092004730440220097fe51de0fbaab4ebd274f79b8be13c3ce703ffa4b4deb69755278447a337bd02206e60213b7b6fe632cc779ffa6116f8fcb31508c621be94e6e899e6fa116059670100475221035df813872b751da7c26e35261a2e6d7667d5649d19671f15f341189da4291139210303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e2852aee803000001b8050100000000001976a91492c53581aa6f00960c4a1a50039c00ffdbe9e74a88ac00000000")
	log.Println(err)
	log.Println(result)
}

func TestClient_GetBalanceByAddress(t *testing.T) {
	client := NewClient()

	privkeys := []string{
		//"cTBs2yp9DFeJhsJmg9ChFDuC694oiVjSakmU7s6CFr35dfhcko1V",
		"cUC9UsuybBiS7ZBFBhEFaeuhBXbPSm6yUBZVaMSD2DqS3aiBouvS",
	}

	srciptPubkey := "a91475138ee96bf42cec92a6815d4fd47b821fbdeceb87"
	inputItems := []TransactionInputItem{
		{
			Txid:         "c75078491c82248717baf4d2cac74176f921ac22fadb3627c4aa542b211a8680",
			Vout:         1,
			ScriptPubKey: srciptPubkey,
			Amount:       0.0001,
		},
	}
	outputItems := []TransactionOutputItem{
		{ToBitCoinAddress: "2Mx1x4dp19FUvHEoyM2Lt5toX4n22oaTXxo", Amount: 0.0001},
	}

	redemScript := "52210389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0210303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e2852ae"

	txid, hex, err := client.BtcCreateAndSignRawTransactionForUnsendInputTx("2N3vGUfBSNALGGxUo8gTYpVQAmLwjXomLhF", privkeys, inputItems, outputItems, 0, 0, &redemScript)
	log.Println(err)
	log.Println(hex)
	log.Println(txid)
	privkeys = []string{
		"cTBs2yp9DFeJhsJmg9ChFDuC694oiVjSakmU7s6CFr35dfhcko1V",
	}

	//hex = "02000000014d58d4a73854b749d8188f9d25c96fa780e155f2dcbb9203d71f5be19b7b9e10010000006a47304402201e78cf4a6e8af625c2f0f5456435dab81f8d0275d0d68886a0e66a2f88ace4c302206f1e48462f24909de19e3da1b88e4fe767b4e6a012cdee95f2a33eb20679258801210389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0ffffffff02a08601000000000017a91475138ee96bf42cec92a6815d4fd47b821fbdeceb8721890300000000001976a91492c53581aa6f00960c4a1a50039c00ffdbe9e74a88ac00000000"
	//result2, err := client.SendRawTransaction(hex)
	//log.Println(err)
	//log.Println(result2)
	//result, err := client.SignRawTransactionWithKey(hex, privkeys, nil, "ALL")

	var inputs []map[string]interface{}
	for _, item := range inputItems {
		node := make(map[string]interface{})
		node["txid"] = item.Txid
		node["vout"] = item.Vout
		node["redeemScript"] = redemScript
		node["scriptPubKey"] = item.ScriptPubKey
		inputs = append(inputs, node)
	}

	result, err := client.SignRawTransactionWithKey(hex, privkeys, inputs, "ALL")
	log.Println(err)
	s, err := client.DecodeRawTransaction(gjson.Get(result, "hex").String())
	log.Println(s)
}
