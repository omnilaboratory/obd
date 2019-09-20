package service

import (
	"LightningOnOmni/dao"
	"github.com/asdine/storm/q"
	"log"
	"testing"
	"time"
)

func TestChannelManager_AliceOpenChannel(t *testing.T) {
	hex := "02000000013ec76f18e3fb709436c5ee4f66c95baa7cb533c6ac975c96bb79ce0e42a7428f00000000d90047304402202c47c570fae9c4dd2e233bbeaf6baa30071b29f0b1c0fd18709163bca154c324022041539e44cc7f5796da8dde64f7a3914e005187595487182ce43250776074cb380147304402201040640a7fb04b1a4a1da22e95e2e864ba7a2efdfd08b8fb1064e5a58cba5ccf02204550dabf44d1af93b64bfcdc2a87dd720f58a3b0f000d6adb7b790068254df4001475221034434a59d648b5a7585182ef71dbc3ecc44236e5fa028b4c55c6adb76fd473ca1210373869bdf9667964b51d9a93ab92a20cd524e21be4d5f4690f51330e827379bbe52aeffffffff033c1900000000000017a91427f4ab8c95d6f4b945f88da4b0a72a3d6031c29d870000000000000000166a146f6d6e690000000000000079000000003b9aca001c0200000000000017a91427f4ab8c95d6f4b945f88da4b0a72a3d6031c29d8700000000"
	inputs, err := getRdInputsFromCommitmentTx(hex, "2MvtVSk3K2Kqn851kqJnvhKC73xdSM5Yhww", "522103ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d210373869bdf9667964b51d9a93ab92a20cd524e21be4d5f4690f51330e827379bbe52ae")
	log.Println(err)
	log.Println(inputs)
}

func TestDecodeTx(t *testing.T) {
	result, err := rpcClient.OmniDecodeTransaction("02000000017840e4bd98a8d022d3ca359239127922eb7329edd21e206637173af427e5e57a010000006a47304402202d5c989ab0fdb94adee355a12f417f1e17f856bb790c68d1e96a666e4ed4309502201d61614d482ec9b64bfdf5741eb30348cab512d156103c52eac3df1c087eac600121034434a59d648b5a7585182ef71dbc3ecc44236e5fa028b4c55c6adb76fd473ca1ffffffff03960b0f00000000001976a9140505ea289a01ba42c259f6608b79c3738c69aacd88ac0000000000000000166a146f6d6e690000000000000079000000003b9aca001c0200000000000017a914dff7bd260fc3ebb602f94ac21347b69e20a4847c8700000000")
	log.Println(err)
	log.Println(result)
	result, err = rpcClient.DecodeRawTransaction("02000000017840e4bd98a8d022d3ca359239127922eb7329edd21e206637173af427e5e57a010000006a47304402202d5c989ab0fdb94adee355a12f417f1e17f856bb790c68d1e96a666e4ed4309502201d61614d482ec9b64bfdf5741eb30348cab512d156103c52eac3df1c087eac600121034434a59d648b5a7585182ef71dbc3ecc44236e5fa028b4c55c6adb76fd473ca1ffffffff03960b0f00000000001976a9140505ea289a01ba42c259f6608b79c3738c69aacd88ac0000000000000000166a146f6d6e690000000000000079000000003b9aca001c0200000000000017a914dff7bd260fc3ebb602f94ac21347b69e20a4847c8700000000")
	log.Println(err)
	log.Println(result)
}

func TestGetBalance(t *testing.T) {

	address := "2N9cf5TUd12ChU32ZyV58WJAF9GFmVUuHxK"
	balance, _ := rpcClient.OmniGetAllBalancesForAddress(address)
	log.Println(balance)
	balance1, _ := rpcClient.GetBalanceByAddress(address)
	log.Println(balance1)
	result, _ := rpcClient.ListUnspent(address)
	log.Println(result)
}

func TestSend(t *testing.T) {
	//result, err := rpcClient.SendRawTransaction("020000000258921046d2381049ceb5c83d97a4aabe93d8d4e06c32177d586e9ceed6bc2153000000006a4730440220274d9a08463987d772b3daa66cdfd337426ccb69fce5b3f8645cfbd6df35ba6f02201cf910b02bb03105cdefe78e67f30fa39d9ee6be0c90dcb6d29870a1018aee2c0121034434a59d648b5a7585182ef71dbc3ecc44236e5fa028b4c55c6adb76fd473ca1fffffffff9c98f5bccd1059b26521eb677d95d044e95bafa05d1c922a68b4234ca8f014a020000006a473044022038afbde1a2b60f528df1d292becf938e7567f82c5d8c903607839c7ab164d72d02205ae92cf56746d34d009db9b418039fc5bb41b726d48a27cebd511e90478ad4f30121034434a59d648b5a7585182ef71dbc3ecc44236e5fa028b4c55c6adb76fd473ca1ffffffff02102700000000000017a914dff7bd260fc3ebb602f94ac21347b69e20a4847c876a190f00000000001976a9140505ea289a01ba42c259f6608b79c3738c69aacd88ac00000000")
	//log.Println(err)
	//log.Println(result)
	result, err := rpcClient.SendRawTransaction("02000000017840e4bd98a8d022d3ca359239127922eb7329edd21e206637173af427e5e57a010000006a47304402202d5c989ab0fdb94adee355a12f417f1e17f856bb790c68d1e96a666e4ed4309502201d61614d482ec9b64bfdf5741eb30348cab512d156103c52eac3df1c087eac600121034434a59d648b5a7585182ef71dbc3ecc44236e5fa028b4c55c6adb76fd473ca1ffffffff03960b0f00000000001976a9140505ea289a01ba42c259f6608b79c3738c69aacd88ac0000000000000000166a146f6d6e690000000000000079000000003b9aca001c0200000000000017a914dff7bd260fc3ebb602f94ac21347b69e20a4847c8700000000")
	log.Println(err)
	log.Println(result)
}

func TestCreateAddress(t *testing.T) {
	address, _ := rpcClient.GetNewAddress("newAddress")
	log.Println(address)
	result, _ := rpcClient.DumpPrivKey(address)
	log.Println(result)
	rpcClient.ValidateAddress(address)
}

func TestChannelManager_Test(t *testing.T) {

	result, err := rpcClient.OmniDecodeTransaction("010000000163af14ce6d477e1c793507e32a5b7696288fa89705c0d02a3f66beb3c5b8afee0100000000ffffffff02ac020000000000004751210261ea979f6a06f9dafe00fb1263ea0aca959875a7073556a088cdfadcd494b3752102a3fd0a8a067e06941e066f78d930bfc47746f097fcd3f7ab27db8ddf37168b6b52ae22020000000000001976a914946cb2e08075bcbaf157e47bcb67eb2b2339d24288ac00000000")

	//result, err := rpcClient.OmniListProperties()
	log.Println(err)
	log.Println(result)
}

func TestTask(t *testing.T) {
	log.Println("aaa")
	node := &dao.RDTxWaitingSend{}
	node.TransactionHex = "111"
	node.IsEnable = true
	node.CreateAt = time.Now()
	db.Save(node)

	var nodes []dao.RDTxWaitingSend
	err := db.Select().Find(&nodes)
	if err != nil {
		return
	}

	for _, item := range nodes {
		item.IsEnable = false
		item.TransactionHex = "33333"
		item.FinishAt = time.Now()
		err := db.Update(&item)
		log.Println(err)
		db.UpdateField(&item, "IsEnable", false)
	}
	var nodes2 []dao.RDTxWaitingSend

	db.Select(q.Eq("IsEnable", true)).Find(&nodes2)
	if err != nil {
		return
	}

}
