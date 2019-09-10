package rpc

import (
	"log"
	"testing"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func TestClient_GetBlockCount(t *testing.T) {
	client := NewClient()
	balance, err := client.GetBalanceByAddress("2N3vGUfBSNALGGxUo8gTYpVQAmLwjXomLhF")
	log.Println(err)
	log.Println(balance)
}

func TestClient_SendRawTransaction(t *testing.T) {
	client := NewClient()
	result, err := client.SendRawTransaction("0200000001fcfbb9b399a3a4a413f126197f6ad3e6fd5067ea59fef7231f9efe3a51017caf01000000d90047304402204833933b22101d24f59375ef5302d5eecb97264ea19ffff96cdd2bbc690ab9140220106d8dbf73a9a1dc903db0f1c9f6f322e277d1b7d28bff75ac60b463a15469a90147304402201dfd74fcf77ff23cd2deba8218b92eb2c8d6cacd4c491bc6296d01444c8f9ad10220752b21d258299687c96e7c293296b7c248c45571b4f70965e3a49400876742dc014752210389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0210303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e2852aeffffffff03701101000000000017a9144cf2089b6e009b57ac5b8c908853a19d4b0e5ab187a03e0d000000000017a91475138ee96bf42cec92a6815d4fd47b821fbdeceb8730750000000000001976a914fd1dc7fb3727d29b43200480580049c3bf8a041b88ac00000000")
	log.Println(err)
	log.Println(result)
}

func TestClient_GetBalanceByAddress(t *testing.T) {
	client := NewClient()

	privkeys := []string{
		"cTBs2yp9DFeJhsJmg9ChFDuC694oiVjSakmU7s6CFr35dfhcko1V",
		"cUC9UsuybBiS7ZBFBhEFaeuhBXbPSm6yUBZVaMSD2DqS3aiBouvS",
	}

	//srciptPubkey := "a91475138ee96bf42cec92a6815d4fd47b821fbdeceb87"
	outputItems := []TransactionOutputItem{
		{ToBitCoinAddress: "2Mx1x4dp19FUvHEoyM2Lt5toX4n22oaTXxo", Amount: 0.0001},
	}

	redemScript := "52210389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0210303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e2852ae"

	txid, hex, err := client.BtcCreateAndSignRawTransaction("2N3vGUfBSNALGGxUo8gTYpVQAmLwjXomLhF", privkeys, outputItems, 0, 0, &redemScript)
	log.Println(err)
	log.Println(hex)
	log.Println(txid)

	result, err := client.SendRawTransaction(hex)
	log.Println(err)
	log.Println(result)
}
