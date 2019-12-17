package service

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

func TestGenerateBTC(t *testing.T) {
	wifKey, pubKey, err := GenerateBTCTest()
	log.Println(wifKey)
	log.Println(pubKey)
	log.Println(err)
}

func TestCreateTx1(t *testing.T) {
	createTx()
}
func TestCreateTx(t *testing.T) {
	buildRawTx()
}
func TestCreateTx2(t *testing.T) {
	transaction, err := CreateTransaction("5HusYj2b2x4nroApgfvaSfKYZhRbKFH41bVyPooymbC6KfgSXdD", "1KKKK6N21XKo48zWKuQKXdvSsCf95ibHFa", 91234, "81b4c832d70cb56ff957589752eb4125a4cab78a25a8fc52d6a09e5bd4404d48")
	if err != nil {
		fmt.Println(err)
		return
	}
	data, _ := json.Marshal(transaction)
	fmt.Println(string(data))
}
func TestCreateTx3(t *testing.T) {
	createTx3()
}
