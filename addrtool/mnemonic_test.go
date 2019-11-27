package addrtool

import (
	"encoding/json"
	"log"
	"testing"
)

func Test_Demo1(t *testing.T) {
	mnemonic := "clown attack stem effort file shield lunch lion topple alcohol lemon salt suspect imitate mimic tiger original achieve either coyote demand neither creek alpha"
	wallet, _ := CreateWallet(mnemonic, 0)
	bytes, _ := json.Marshal(wallet)
	log.Println(string(bytes))

}

func Test_Demo2(t *testing.T) {

}
