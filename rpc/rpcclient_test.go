package rpc

import (
	"LightningOnOmni/bean/chainhash"
	"crypto/sha256"
	"github.com/tidwall/gjson"
	"log"
	"testing"

	"github.com/satori/go.uuid"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func TestClient_GetBlockCount(t *testing.T) {

	uuid_str, err := uuid.NewV4()
	if err != nil {
		log.Println(err)
	}
	hash := sha256.New()
	hash.Write([]byte(uuid_str.String()))
	sum := hash.Sum(nil)
	tempId, err := chainhash.NewHash(sum)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(tempId[:]))

	client := NewClient()
	//hex := "02000000000101d5df58bb45caac351120a9d228e580a7ffa9f928d78925a3ccbd746f256cb0fe0000000023220020b92c9c68bdc00c5f838e64fc436b994546c6e27911881a3384d0c54f4acc91f10a0000000160decb1d0000000017a914c893cec2a62aca000b8cf963a9530f24a4a4e23987040047304402207183958394599d0023995852d89d8cfd57687feb3638f92082d7fcc911e78ff4022003f50ae3514ca0387d9dae533398526a8c92e847e7f241f98af762eaad234c0901473044022079c25009db65d6e82bd1fb5271164ab41701359fe44e42d9ac93e40bc3026b460220143c20659211159cfec5e11ac95c49a3488baac43d57de1e6c735e7708e4ddb901475221035792effbf392a1f59ebc1a6df0eec31ca83f434706d07b5797f8cb7df8c27def210345f04d044f19006bd60620b86354296ad8a526027c00117b0f4324b0d59db1a152ae00000000"
	//result, _ := client.Omni_getbalance("n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn",31)
	//result, err := client.Omni_gettransaction("1f35a5a0d166fef6b2e24b321efa1108a01e6ea1205397bc96613fb73b902d18")
	//result, err := client.Omni_listtransactions()
	var keys []string
	_, result, err := client.BtcCreateAndSignRawTransaction("n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn", keys, []TransactionOutputItem{{ToBitCoinAddress: "n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA", Amount: 0.0001}}, 0.00001, 0)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(result)
}

func TestClient_DecodeRawTransaction(t *testing.T) {
	client := NewClient()
	//hex := "0100000001be898a63ce2f0bf3297189fb9103aa63f0eafd74a6cb1974f2bbd59b1ed13140520000006a47304402206b5eb7fdcc6df35ee0cd5186b6d9e6f29bcaa8a85af1f9a4726c976ee746f00b0220764d3d9f7c86d61afd773ac31f63256fe8b766819253f55f837451b7c93cad3d01210343c3b527ebf385b15cedfa7e9b840b32353482bfc29ddc931a155ff41db0123effffffff02306f0100000000001976a914d8b7fc6790003813df46aacd8bdc84d1f672147b88ac10270000000000001976a914fd1dc7fb3727d29b43200480580049c3bf8a041b88ac00000000"
	hex := "0100000001946f7d2d3b7f61e87c20a96fd01f9c6ef50fc534588c9d5dd73d0c7241deca36010000006a47304402204107627dcece427aa8c565e76c9d18af455833fd66cbbc8fbbf96bd68f828f10022043c84bcbf09d543b18b66745611280d2bee1ed398e4aab2d5f0fcffd6d99e56b01210343c3b527ebf385b15cedfa7e9b840b32353482bfc29ddc931a155ff41db0123effffffff0210270000000000001976a914fd1dc7fb3727d29b43200480580049c3bf8a041b88ac38440100000000001976a914d8b7fc6790003813df46aacd8bdc84d1f672147b88ac00000000"
	result, err := client.SendRawTransaction(hex)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(result)
}

func TestClient_Validateaddress(t *testing.T) {
	client := NewClient()
	//r, err := client.Omni_getbalance(" n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA", 121)
	r, err := client.ValidateAddress(" n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA")
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(r)
}
func TestClient_OmniRawTransaction(t *testing.T) {
	client := NewClient()
	txid, err := client.OmniRawTransaction("n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn", nil, "n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA", 121, 7, 0.00001, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(txid)
}

func TestClient_GetTransactionById(t *testing.T) {
	client := NewClient()
	//result, err := client.GetTransactionById("434b1d74135ec0bf01c0d086792afdcee8c9440ad0aa10dc1882a901ca2b71e4")
	result, err := client.CreateMultiSig(2, []string{"mvzP1iEEJRZgJoZoCbBPDzwv3ZULPauug3", "mzs5iXkCzmygs2q1bqQBKcsTm5HvGEpjyK"})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(result)
}

func TestClient_GetMiningInfo(t *testing.T) {

	client := NewClient()
	//sendrawtransaction [020000000115ab410718fb8d12bae999cf1d7f611872c36a8b3ade1a9bcf7bdc4f3218a28a00000000d9004730440220657eb3190656a0f4607fe231f5b5082f319c8a25ed81c0693e72a8add97a1373022024a35d89fab11da0a84f0f15dc86f7b0c366c723729b1602d02b6cfdb90a74f501473044022072b62d9633a61115419d9de541a47ced6a68b5eeed084c429e7c29651aafe2fa02203cab30acbd4dae7744a51af6ccc78fe88d6ec9299192b618681897be8b922d7e0147522103ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d210303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e2852aee803000001e87a0100000000001976a91492c53581aa6f00960c4a1a50039c00ffdbe9e74a88ac00000000]
	result, err := client.SendRawTransaction("020000000115ab410718fb8d12bae999cf1d7f611872c36a8b3ade1a9bcf7bdc4f3218a28a00000000d9004730440220657eb3190656a0f4607fe231f5b5082f319c8a25ed81c0693e72a8add97a1373022024a35d89fab11da0a84f0f15dc86f7b0c366c723729b1602d02b6cfdb90a74f501473044022072b62d9633a61115419d9de541a47ced6a68b5eeed084c429e7c29651aafe2fa02203cab30acbd4dae7744a51af6ccc78fe88d6ec9299192b618681897be8b922d7e0147522103ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d210303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e2852aee803000001e87a0100000000001976a91492c53581aa6f00960c4a1a50039c00ffdbe9e74a88ac00000000")
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(result)

	return
	address, err := client.GetNewAddress("")
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(address)

	isvalid, err := client.ValidateAddress("n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn")
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(isvalid)
	isvalid, err = client.ValidateAddress("n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA")
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(isvalid)

	isvilid, err := client.CreateMultiSig(2, []string{"0343c3b527ebf385b15cedfa7e9b840b32353482bfc29ddc931a155ff41db0123e", "0303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e28"})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(isvilid)
}

func TestClient_ValidateAddress(t *testing.T) {
	client := NewClient()
	balance, _ := client.DecodeRawTransaction("0200000001d6a01bc64b3d6c33524fb1dd546f20efa8c3ae72012e9b28e5b656a978760c3d0000000092004730440220055fe35ee8b800e4b8d3246df6a8d36cf5d2552531b35e80e6118ee62714c13f02206e0b42ab7e813013c46c576477351568b62d081eb0e43e949ce41d1e4f863b9701004752210389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0210303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e2852aeffffffff01e87a01000000000017a9143455ed8bf8c87aea1654e34a451e2d6aa9ba13628700000000")
	log.Println(balance)
	balance, _ = client.DecodeRawTransaction("0200000001d6a01bc64b3d6c33524fb1dd546f20efa8c3ae72012e9b28e5b656a978760c3d00000000d9004730440220642c72da3a733907ae08391d4c824dd495f477c239ff44065938ee28b7af8c5502206e696986cb91c1003197649b88f5e0e185be9921053ffa4b923203d9dd509a1c014730440220055fe35ee8b800e4b8d3246df6a8d36cf5d2552531b35e80e6118ee62714c13f02206e0b42ab7e813013c46c576477351568b62d081eb0e43e949ce41d1e4f863b97014752210389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0210303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e2852aeffffffff01e87a01000000000017a9143455ed8bf8c87aea1654e34a451e2d6aa9ba13628700000000")
	log.Println(balance)
}

func TestClient_AddMultiSigAddress(t *testing.T) {
	client := NewClient()
	result, err := client.AddMultiSigAddress(2, []string{"myjgni5hmXbRo9KsStMRK5gVbXeRZs5LLi", "n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA"})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(result)
}

func TestClient_SignMessage(t *testing.T) {
	client := NewClient()
	address := "mn4ja6fttFsyRLTxegeAftopBB8Twr6iTa"
	privKey := "cTAJCTpGNTGNfEhjez6JZBSB2wjgtuJepLdo2tpD4iy1X4Zw7q9c"
	//cTAJCTpGNTGNfEhjez6JZBSB2wjgtuJepLdo2tpD4iy1X4Zw7q9c
	msg := "{\n\"peer_id\":\"254698748@qq.com\"\n}"
	signature, err := client.SignMessageWithPrivKey(privKey, msg)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(signature)

	// ILXbebexMFhhrc7CwoMNHgag6XcO7Du1YD5vcSleq5WZFcFCz6x1Fk07CN3I1bm0AJRpfzk7PGk8ZxSAoC1KtBg=
	result, err := client.VerifyMessage(address, signature, msg)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(result)

	result, err = client.DecodeScript(signature)
	if err != nil {
		log.Println(err)
		return
	}
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
