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
	hex1 := "02000000015c2f029b1bd32721bfe39a4424e53837404e2295bef8f782a3c6ed02b1585ebd010000006a473044022047630a53d390ea30de7f042390186a1536f6e8aa99114a342813dd3d56e0ad9c022021b72906b8c3a0c6de41a4e0a64a2a14343a2f48de40fd0e23db01172370bd6b012102241c98bb5c64f34bd414f05e722924a6946569f817190bbe9752f1c0416b20dcffffffff02502400000000000017a9147297c8404e9dc908ce10fb1db3660736dd45f7ef8702c90b00000000001976a914928f34815d1a8f54afe239ad68391fcddb505a6588ac00000000"
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
