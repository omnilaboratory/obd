package rpc

import (
	"log"
	"testing"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func TestClient_OmniSend(t *testing.T) {
	client := NewClient()

	channelAddr := "2NFuD1Z4Cox1CXcRNDByXA17AYdwqrMzPoM"
	toChanelAddr := "2NFAnukYdESsw8JHBBaXBbdvfqQyi6zQkhG"
	ChannelAddressRedeemScript := "5221034434a59d648b5a7585182ef71dbc3ecc44236e5fa028b4c55c6adb76fd473ca1210373869bdf9667964b51d9a93ab92a20cd524e21be4d5f4690f51330e827379bbe52ae"
	txid, hex, usedTxid, err := client.OmniCreateAndSignRawTransactionForCommitmentTx(
		channelAddr,
		[]string{
			"cSBd2TDKagHMrgXSGatqvjVSbeUhTY2ejuFQjCiKJqNCt7xQvEbr",
			"cTrhCmWHV8B1P3npZuqFMxwhTUJj19ZD2uhAdkC2xb5NhtkTwXXn",
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
	result, err := client.SendRawTransaction("0200000001429c7af3e450dbb9b22ab2faaf4984b5a1bb63e6b846963c251fbc82fc3378fe010000006a47304402202af680fd43242be263154e167fc5318708a63714c1219b8018bf9d7dbb445dc90220157e2c2b6117897d019e219574ac60fa43306a60436b355b86eded07e4a4d1070121034434a59d648b5a7585182ef71dbc3ecc44236e5fa028b4c55c6adb76fd473ca1ffffffff02102700000000000017a914f881c1debb0c6e6b785516364721e277613f8684877ede0d00000000001976a9140505ea289a01ba42c259f6608b79c3738c69aacd88ac00000000")
	//result, err := client.GetAddressInfo("2MwKVXga7i82DgwmQ9nTPFSuAGP6pTkNQYr")
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
