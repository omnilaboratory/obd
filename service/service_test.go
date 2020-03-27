package service

import (
	"github.com/asdine/storm/q"
	"log"
	"obd/dao"
	"testing"
	"time"
)

func TestDemoChannelTreeData(t *testing.T) {
	PathService.CreateDemoChannelNetwork("A", "F", 6, nil, true)

	for index, node := range PathService.openList {
		log.Println(index, node)
	}
	log.Println()

	for index, node := range PathService.openList {
		if node.IsTarget {
			log.Println("findPath:", index, node)
		}
	}

	//for key, node := range branchMap {
	//	log.Println(key, node)
	//}

}

func TestDelDemoChannelInfoData(t *testing.T) {
}

func TestDelDemoChannelInfoOne(t *testing.T) {
	err := db.Delete("DemoChannelInfo", 4)
	log.Println(err)
	err = db.Delete("DemoChannelInfo", 2)
	log.Println(err)
}

func TestGetDemoChannelInfoData(t *testing.T) {
	var nodes []dao.DemoChannelInfo
	db.All(&nodes)
	for _, value := range nodes {
		log.Println(value)
	}
}

func TestCreateDemoChannelInfoDataSingle(t *testing.T) {
	node := &dao.DemoChannelInfo{
		PeerIdA: "D",
		AmountA: 10,
		PeerIdB: "F",
		AmountB: 18,
	}
	db.Save(node)
}

func TestCreateDemoChannelInfoData(t *testing.T) {
	node := &dao.DemoChannelInfo{
		PeerIdA: "A",
		AmountA: 18,
		PeerIdB: "B",
		AmountB: 0,
	}
	db.Save(node)
	node = &dao.DemoChannelInfo{
		PeerIdA: "B",
		AmountA: 16,
		PeerIdB: "C",
		AmountB: 0,
	}
	db.Save(node)
	node = &dao.DemoChannelInfo{
		PeerIdA: "C",
		AmountA: 14,
		PeerIdB: "D",
		AmountB: 0,
	}
	db.Save(node)
	node = &dao.DemoChannelInfo{
		PeerIdA: "D",
		AmountA: 12,
		PeerIdB: "E",
		AmountB: 0,
	}
	db.Save(node)
	node = &dao.DemoChannelInfo{
		PeerIdA: "E",
		AmountA: 10,
		PeerIdB: "F",
		AmountB: 0,
	}
	db.Save(node)
	node = &dao.DemoChannelInfo{
		PeerIdA: "F",
		AmountA: 8,
		PeerIdB: "G",
		AmountB: 0,
	}
	db.Save(node)
	node = &dao.DemoChannelInfo{
		PeerIdA: "G",
		AmountA: 6,
		PeerIdB: "H",
		AmountB: 0,
	}
	db.Save(node)

}

func TestPathManager_GetPath(t *testing.T) {
	//PathService.GetPath("alice", "carol", 5, nil, true)
	//miniPathLength := 7
	//var miniPathNode *PathNode
	//
	//for _, node := range PathService.openList {
	//	if node.IsTarget {
	//		if int(node.Level) < miniPathLength {
	//			miniPathLength = int(node.Level)
	//			miniPathNode = node
	//		}
	//	}
	//}
	//if miniPathNode != nil {
	//	log.Println(miniPathNode)
	//	channelCount := miniPathNode.Level
	//	channelIdArr := make([]int, 0)
	//	for i := 1; i < int(channelCount); i++ {
	//		channelIdArr = append(channelIdArr, PathService.openList[miniPathNode.PathIdArr[i]].ChannelId)
	//	}
	//	channelIdArr = append(channelIdArr, miniPathNode.ChannelId)
	//	log.Println(channelIdArr)
	//} else {
	//	log.Println("no path find")
	//}
}

func TestChannelManager_AliceOpenChannel(t *testing.T) {
	hex := "02000000013ec76f18e3fb709436c5ee4f66c95baa7cb533c6ac975c96bb79ce0e42a7428f00000000d90047304402202c47c570fae9c4dd2e233bbeaf6baa30071b29f0b1c0fd18709163bca154c324022041539e44cc7f5796da8dde64f7a3914e005187595487182ce43250776074cb380147304402201040640a7fb04b1a4a1da22e95e2e864ba7a2efdfd08b8fb1064e5a58cba5ccf02204550dabf44d1af93b64bfcdc2a87dd720f58a3b0f000d6adb7b790068254df4001475221034434a59d648b5a7585182ef71dbc3ecc44236e5fa028b4c55c6adb76fd473ca1210373869bdf9667964b51d9a93ab92a20cd524e21be4d5f4690f51330e827379bbe52aeffffffff033c1900000000000017a91427f4ab8c95d6f4b945f88da4b0a72a3d6031c29d870000000000000000166a146f6d6e690000000000000079000000003b9aca001c0200000000000017a91427f4ab8c95d6f4b945f88da4b0a72a3d6031c29d8700000000"
	inputs, err := getInputsForNextTxByParseTxHashVout(hex, "2MvtVSk3K2Kqn851kqJnvhKC73xdSM5Yhww", "522103ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d210373869bdf9667964b51d9a93ab92a20cd524e21be4d5f4690f51330e827379bbe52ae")
	log.Println(err)
	log.Println(inputs)
}

func TestDecodeTx(t *testing.T) {
	hex := "02000000012ecd34ce812f36a876d6f5b3ab2ccb3478eea69e6af4a337fb1941ae8b8a62510000000092004730440220514b3ed6d636c69b2c936f9a57ecc248f00618c46e61bee5e8408192c4a25570022045d1c1b191a9b6c4ee56a51129baf9d79a26852d5ccbe572d8fcd961b360e8c8010047522103f1603966fc3986d7681a7bf7a1e6b8b44c6009939c28da21f065c1b991aeff12210216847047b926a1ff88e97fb0ebed8d0482c69521e9f8bc499c06b108a4972b8252aeffffffff033c1900000000000017a9140ff6b304e80589566854573a3c528ee0cb7dfbe4870000000000000000166a146f6d6e69000000000000007900000000773594001c0200000000000017a9140ff6b304e80589566854573a3c528ee0cb7dfbe48700000000"
	result, err := rpcClient.OmniDecodeTransaction(hex)
	log.Println(err)
	log.Println(result)
	result, err = rpcClient.DecodeRawTransaction(hex)
	log.Println(err)
	log.Println(result)
}

func TestGetBalance(t *testing.T) {

	address := "n362wgEWVqbymxYjVkkq7jLQjQdeW93Ncc"
	//address := "n362wgEWVqbymxYjVkkq7jLQjQdeW93Ncc"
	s, _ := rpcClient.OmniGetbalance(address, 121)
	log.Println(s)
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
	result, err := rpcClient.SendRawTransaction("02000000027e3a2539885f04f17e39feeccffdf5a1a65fb433427a6c669328f26885eee09d00000000d9004730440220641dc362f2897836f3d0fe3a12c97c5ea06d267e79bce44431110b549c641cea02206a775100fae0e5ea9af18e0219abfcbd7fd7887fdbc765df23df9106855d95130147304402203f6b4db66fa80e7d7045ef41f6f7064123e3d21e10356e029852ecc08c72074e02205e7665bf00ba6bf9e7840f705d328a53f7671e75ae7952f47b767df1f25f1a0d0147522102d9c077bd5bb97186313d74ca534584b3d290709880bc6945ab5c46799ad857302102e38f7e2ad2cb64c1822640b3f60b9edbe3a6f80a1e17f307e34a2526f4d20b1152aee80300007e3a2539885f04f17e39feeccffdf5a1a65fb433427a6c669328f26885eee09d02000000d900473044022077a3617ad4327b10bbbef0a4f9779b5d4f278299dae626cc94ee0d8f7ab96607022071820ce9f735948c4bb4d47971e958acdd30d32123b94242e7925263bb459d890147304402207e4f7b5eff63bb94464d4f551115a358672fdf29035c2a375007593f014ba854022077a175fe019aa1f0b8f848c59305e2cdc89ed8b26b5dc937d1ceaaef38e178f20147522102d9c077bd5bb97186313d74ca534584b3d290709880bc6945ab5c46799ad857302102e38f7e2ad2cb64c1822640b3f60b9edbe3a6f80a1e17f307e34a2526f4d20b1152aee803000003620b0000000000001976a914ec9c3fabfa57c7862ba594b70988a7b4f485744188ac0000000000000000166a146f6d6e6900000000000000790000000029b9270022020000000000001976a914ec9c3fabfa57c7862ba594b70988a7b4f485744188ac00000000")
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

	hexC := "02000000012d6e0d747667b159ef24d731a3f1f6bdc14d83c241087dc898aad85cc4a14f4400000000d900473044022065deb2b5c9ed9a41383c8300a81a95cf416bdb16616f35a12a99a38b14d02e98022054af1192e8e6c9a9aeaa5595a4659f93bd1011d8723fe57d448d5c3793ff26240147304402206d628f7405999bbc21a8cfe1a530b6969d96f7b648302e419f22440b79412759022044a30a9aa8bacb3371fc5952eabf50265115b12d36511062cc1089268e17c4ae0147522103b4df7d3026a437f537dcc0a9e681ffdfb000c9f1223189adf18364588d46e5592103c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee96152aeffffffff033c1900000000000017a914faafd20558ca1529b96fb1cdd40e4ef1915ed1e4870000000000000000166a146f6d6e690000000000000079000000001dcd65001c0200000000000017a914faafd20558ca1529b96fb1cdd40e4ef1915ed1e48700000000"
	result, err := rpcClient.DecodeRawTransaction(hexC)
	log.Println(err)
	log.Println(result)
	//hexRD := "0200000002952b9f91f48d53637685c88791ea4ea046c637fdca988662f322f3bbab88cf8600000000d90047304402204a33088ade500d6d0a40051231b4223186809b855cd8e534a0f4ccda2b90a8e302200657eea4dbdf52573402156d30125b6048e75e5d4376af16a262187aa99a0d6d01473044022042cea8d1389fb2c844f0e59ca9992943af962c973f424ffafef89f415934225202206ee20104936fbdb0fbc25a4bf6f6629b4b89a672e6ea4cb1070e50679920f6450147522102b8302d22a50fd84f34d528ff98998a6959bc7fb8f45b5f3fb44e23101aa5d8f22103647e81480a71989ee5c3391763d6aac445bb104f0dce688002c18a3bba6ed42b52aee8030000952b9f91f48d53637685c88791ea4ea046c637fdca988662f322f3bbab88cf8602000000d90047304402206dd99daaa88f50c3403bcba536e50f8d3288e2ca5e9fdcd97ba8d3bf9520384a0220158f12b5c43fec31c51ca7597d9dabaade62eec5798c7197db163691177865be0147304402205fe693c8ea0dcf48692fc51dec95913abde3c64084fb47bfade8291ffd9a4bef02202d571fe14754347c9153582ea6f744581fd5a30d147bbc63b2aa2704f89a29ee0147522102b8302d22a50fd84f34d528ff98998a6959bc7fb8f45b5f3fb44e23101aa5d8f22103647e81480a71989ee5c3391763d6aac445bb104f0dce688002c18a3bba6ed42b52aee803000003620b0000000000001976a914ec9c3fabfa57c7862ba594b70988a7b4f485744188ac0000000000000000166a146f6d6e6900000000000000790000000011e1a30022020000000000001976a914ec9c3fabfa57c7862ba594b70988a7b4f485744188ac00000000"
	hexRD := "02000000023cbcb3fe0254dd2a4ac81172e1de5520310fa0cea134ec827e6dbee153b10d7000000000d900473044022028c6c2f4de0baba040904b266b548020ee64f266aef4e00fa18f209ffecdd23e02207c96c5986a15a99b5f66c5e9d80b0e32140fc14d72dee23782ac47a6258da3d40147304402203edbc9b302eb3296090727472a35430da57cce8f2c8a2b10da69638e211df9360220117203aaa265440d5b61faed5926e92c5230753bc2348c528f553ab802d32fc60147522103ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d2103c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee96152aee80300003cbcb3fe0254dd2a4ac81172e1de5520310fa0cea134ec827e6dbee153b10d7002000000d900473044022029bb9797f6e4030120e614ed1153f8649f3aef6cc7c94fbd1a97c095ba55d882022013575c341f7cc69269407ac027097b2e9997631e6bad6df6508d60649fa1c4ff014730440220078e235854f316419ddd7801bc8737817bc51b25a8977709d0eae3c0d88eb63d02205af3887519aafe49a4acd833566f00f01b95d4fed363cc46df5eb0f5dc28a2270147522103ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d2103c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee96152aee803000003620b0000000000001976a914ec9c3fabfa57c7862ba594b70988a7b4f485744188ac0000000000000000166a146f6d6e690000000000000079000000001dcd650022020000000000001976a914ec9c3fabfa57c7862ba594b70988a7b4f485744188ac00000000"
	result, err = rpcClient.DecodeRawTransaction(hexRD)
	log.Println(err)
	log.Println(result)
	hexBR := "02000000023cbcb3fe0254dd2a4ac81172e1de5520310fa0cea134ec827e6dbee153b10d7000000000d9004730440220051b1a5d236d41efcd5b40613c6f0621b5870b61cb6cf0e7fcf1eb11ca25210902203280ac5f7f195e2a2f411bbf7c8ce09b64097af1d3361c19985916109edf363a0147304402200d1e967ee3105f2e5fcd2ecf81244b37c9e93362c2d55506a9dadcd9282f2c7602202fa88cd04286f2fea5336b03fb4238b06f955810384dff09d1858560456ee5090147522103ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d2103c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee96152aeffffffff3cbcb3fe0254dd2a4ac81172e1de5520310fa0cea134ec827e6dbee153b10d7002000000d90047304402201df6b76e21bcae0ed0883c06e7f46e1f4e930fe6eae85394ecfb5b29e0e1e0bc022008272bd938edf2a0bfcd10446431dfcc3452bcbb2fef0cc436aeb91bf44732f001473044022029988dfeca46b9a4592922953e56bfbc43d84cedc2ba21c1d4206c190f8735fc02200479e5f45ad6ddbb0ff5b8dceb1491379ca5b1580ef1e80401656d60e422e62a0147522103ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d2103c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee96152aeffffffff03620b0000000000001976a914ec9c3fabfa57c7862ba594b70988a7b4f485744188ac0000000000000000166a146f6d6e690000000000000079000000001dcd650022020000000000001976a914744846d33d79479478c2858c008ad93f77c2259d88ac00000000"
	result, err = rpcClient.DecodeRawTransaction(hexBR)
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
