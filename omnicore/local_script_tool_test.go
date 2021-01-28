package omnicore

import (
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/tool"
	"log"
	"testing"
)

func TestCreateMultiSig(t *testing.T) {
	s, i, i2 := CreateMultiSigAddr("02c57b02d24356e1d31d34d2e3a09f7d68a4bdec6c0556595bb6391ce5d6d4fc66", "032dedba91b8ed7fb32dec1e2270bd451dee3521d1d9f53059a05830b4aa0d635b", tool.GetCoreNet())
	//s, i, i2 := CreateMultiSigAddr("02c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c8", "032dedba91b8ed7fb32dec1e2270bd451dee3521d1d9f53059a05830b4aa0d635b", tool.GetCoreNet())
	log.Println(s)
	log.Println(i)
	log.Println(i2)
}

func TestVerifyOmniTxHex(t *testing.T) {
	log.Println("")
	hex1 := "0200000002638ca1bb822a0b783eae6091d2c88c68638ebe4a5b8f7e573fa98b13bc472bba00000000db00483045022100cbab1007964d4057a492fc105c93101f4ef90fb7bf7014676f81690b852a4ae702205c73c0e3fb6ce73a2b93927c073e0a7a76a2ffb078ef2f4fbfeaec58a9e6f33c01483045022100b35b41f40a4de2253ec5a1e3adc3954467f6c76de7e80cb8abb6820249fc924202201a42b85b341224b3f66f8f95baf942b1e96def420733cded0dec66f5672f7f19014752210212b76306d005679fd6f04da94d49af586ee53667d9b8b5c5bb5a79251e6f6c0d2103c384b8d9c65edea28ce205537bb58dc0096bc618e9e553455e1db1f36cc2564252aee8030000638ca1bb822a0b783eae6091d2c88c68638ebe4a5b8f7e573fa98b13bc472bba02000000d90047304402200bd55db16b31880f868b7a2d0263fe2590c282efb2d9f1d3bc9f0e159ef482bf02202574ef71f5c51d2e1c6bd30e6d800213c6b7fae3829b857906d45b927ba1770d0147304402205f5c4bb549dbc36907a0098f4e4a3604f9f9c7605ba680daac6c550d40c8e295022077285cf8c1bd786c48f601c3a55d9dfcf510743cc1a65900814bf0a54462a6d9014752210212b76306d005679fd6f04da94d49af586ee53667d9b8b5c5bb5a79251e6f6c0d2103c384b8d9c65edea28ce205537bb58dc0096bc618e9e553455e1db1f36cc2564252aee8030000034a140000000000001976a914a88134d1b26b1a2e756c4cbaec707d172f00091588ac0000000000000000166a146f6d6e69000000000000008900000000000f424022020000000000001976a91457bf5109ba28554cf727734844fcb2b934f4998f88ac00000000"
	transaction := DecodeRawTransaction(hex1, tool.GetCoreNet())
	log.Println(transaction)
	log.Println("omnicore 签名结果")

}

func TestSign(t *testing.T) {
	redeemhex := "0200000002acbd057ae190cd8fdad4c989fc8216cd9137814620eaf48bc0ff919888e534f30000000000e8030000acbd057ae190cd8fdad4c989fc8216cd9137814620eaf48bc0ff919888e534f30200000000e8030000034a140000000000001976a914928f34815d1a8f54afe239ad68391fcddb505a6588ac0000000000000000166a146f6d6e6900000000000000890000000005f5e10022020000000000001976a914928f34815d1a8f54afe239ad68391fcddb505a6588ac00000000"
	inputs := []bean.RawTxInputItem{}

	item := bean.RawTxInputItem{}
	item.ScriptPubKey = "a9143833fc9817cadba3088022c6cc3687fdda33558687"
	redeemScript := "522103af0e670036b6365494a3ca0ed1bccbfd810f71ac3a119903d514af79c17b33a02102a488048de367beb56aff7768c34d976c5b59c37c5faf009f6ae5a469f0c9e6e452ae"
	item.RedeemScript = redeemScript
	inputs = append(inputs, item)

	item = bean.RawTxInputItem{}
	item.ScriptPubKey = "a9143833fc9817cadba3088022c6cc3687fdda33558687"
	redeemScript = "522103af0e670036b6365494a3ca0ed1bccbfd810f71ac3a119903d514af79c17b33a02102a488048de367beb56aff7768c34d976c5b59c37c5faf009f6ae5a469f0c9e6e452ae"
	item.RedeemScript = redeemScript
	inputs = append(inputs, item)

	privkey := "cPsWdLTpT21gPkYDGjUEitMkkphmPJw3YPXi67pcmLkeLb5FXBjc"
	sign, err := SignRawHex(inputs, redeemhex, privkey, 2)
	log.Println(err)
	transaction := DecodeRawTransaction(sign, tool.GetCoreNet())
	log.Println("第一次 完成", sign)
	log.Println("第一次 完成", transaction)

	redeemhex = sign
	privkey = "cQS12CUD8byKopTV5GQ7RLeNAdL2efdYKTBxwYWdtP18recNuBft"
	sign, err = SignRawHex(inputs, redeemhex, privkey, 1)
	transaction = DecodeRawTransaction(sign, tool.GetCoreNet())
	log.Println("第二次 完成", sign)
	log.Println("第二次 完成", transaction)
}
