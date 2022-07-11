package service

import (
	"encoding/json"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tyler-smith/go-bip39"
	"log"
	"testing"
)

func Test_Demo3(t *testing.T) {

}
func Test_Demo2(t *testing.T) {
	mnemonic := "coyote antenna senior reward diesel vault into used veteran model throw relief"
	userId := tool.SignMsgWithSha256([]byte(mnemonic))

	changeExtKey, _ := HDWalletService.CreateChangeExtKey(mnemonic)

	user := &bean.User{}
	user.CurrAddrIndex = 0
	user.ChangeExtKey = changeExtKey
	user.Mnemonic = mnemonic
	user.PeerId = userId
	for i := 0; i < 4; i++ {
		wallet, _ := HDWalletService.GetAddressByIndex(user, uint32(i))
		bytes, _ := json.Marshal(wallet)
		log.Println(string(bytes))
	}

}

func Test_Demo1(t *testing.T) {
	mnemonic := "unfold tortoise zoo hand sausage project boring corn test same elevator mansion bargain coffee brick tilt forum purpose hundred embody weapon ripple when narrow"
	seed := bip39.NewSeed(mnemonic, "")
	masterKey, _ := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	// This gives the path: m/44H
	acc44H, _ := masterKey.Child(hdkeychain.HardenedKeyStart + 44)
	// This gives the path: m/44H/60H
	acc44H60H, _ := acc44H.Child(hdkeychain.HardenedKeyStart + 1)
	// This gives the path: m/44H/60H/0H
	acc44H60H0H, _ := acc44H60H.Child(hdkeychain.HardenedKeyStart + 0)
	// This gives the path: m/44H/60H/0H/0
	acc44H60H0H0, _ := acc44H60H0H.Child(0)
	// This gives the path: m/44H/60H/0H/0/0
	acc44H60H0H00, _ := acc44H60H0H0.Child(0)
	addr, _ := acc44H60H0H00.Address(&chaincfg.TestNet3Params)
	log.Println(addr)
	btcecPrivKey, _ := acc44H60H0H00.ECPrivKey()

	privateKey := btcecPrivKey.ToECDSA()

	log.Println(privateKey)
}
