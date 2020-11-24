package rpc

import (
	"encoding/hex"
	"log"
	"testing"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func TestClient_GetMiningInfo(t *testing.T) {

	client := NewClient()

	hex := "0200000001cfb3ed1a3e13dbd03ed0329ab03f3c86e7502b534748b39a52e1d3abe8d8e216010000006b483045022100840180244c88891b116fe469d597ed640224032389f608f9e36fce05b292450c022069ab1fb78c0634664bdd87643c43d1079596a2cd73ffa912560e2fbe9689d37801210238b21ae976c7059057b2b973fec5be8c7b37cea1b15aff52ecf309d5f527e76effffffff0382910b00000000001976a914a2bebc3bbc138a248296ad96e6aaf71d83f69c3688ac0000000000000000166a146f6d6e690000000000000089000000001dcd6500220200000000000017a914e771a39d6fde8e0b1cbcba5384f35d14b0160f608700000000"
	result, err := client.DecodeRawTransaction(hex)
	log.Println(err)
	log.Println(result)

	accept, err := client.TestMemPoolAccept(hex)
	log.Println(accept)

	//transaction, err := client.SendRawTransaction(hex)
	//log.Println(transaction)

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
}
