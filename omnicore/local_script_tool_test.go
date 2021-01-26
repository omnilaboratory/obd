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
	hex1 := "0200000002c31d63e0c64f2ba4475aa71e7f57ebf6f52d532006c4450dccb57584a08a98b700000000da0047304402203d869fbe045650a5096a7349fa1171cc801e9ee7376eea94b0a26a59cd333bc502201890262aefb64abd4aa2db9b3275102806c9a5b21faa09ea0f2f4c78b863365401483045022100e3087d5cf084f01a751cb6375c2c6322bf5342968708fe4a05bb31e9596653a6022042f5dcd71b4f10073f64ce04499c8fac3ae75747acd8ce65cdf01dd6caf220e50147522103c384b8d9c65edea28ce205537bb58dc0096bc618e9e553455e1db1f36cc2564221022efbf96ac09bdbec9b39a16ad896aa5a79638fe8f1fdf36d1c5ada382efe7c1552aee8030000c31d63e0c64f2ba4475aa71e7f57ebf6f52d532006c4450dccb57584a08a98b702000000da00483045022100d706a3065e7c6e779bd474bb624d1a1e11ed76f2fb898411ed200d96bb274d1302206d0baf5c9d91f32d74b84a1c7d4a24f58305855a6376e3e008a46cffd671f93301473044022068b9b1ff0f6f66d84bbb56bebafebd2faf9af3744043de9ecc1b3e00d5159bea02201e73be1dcc42a11cab1334a71a6ba3c4e1fd14decc51e90b5f404467a83ad76e0147522103c384b8d9c65edea28ce205537bb58dc0096bc618e9e553455e1db1f36cc2564221022efbf96ac09bdbec9b39a16ad896aa5a79638fe8f1fdf36d1c5ada382efe7c1552aee8030000034a140000000000001976a914a88134d1b26b1a2e756c4cbaec707d172f00091588ac0000000000000000166a146f6d6e6900000000000000890000000005f5e10022020000000000001976a914a88134d1b26b1a2e756c4cbaec707d172f00091588ac00000000"
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
