package omnicore

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

func TestCreateTransaction(t *testing.T) {
	transaction, err := CreateTransaction("5HusYj2b2x4nroApgfvaSfKYZhRbKFH41bVyPooymbC6KfgSXdD", "1KKKK6N21XKo48zWKuQKXdvSsCf95ibHFa", 91234, "81b4c832d70cb56ff957589752eb4125a4cab78a25a8fc52d6a09e5bd4404d48")
	if err != nil {
		fmt.Println(err)
		return
	}
	data, _ := json.Marshal(transaction)
	fmt.Println(string(data))
}

func TestCheckUnspent(t *testing.T) {
	Unspent1 := new(Unspent) //pointer
	Unspent1.TxID = "c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de"
	Unspent1.VOut = 0

	Unspent2 := new(Unspent)
	Unspent2.TxID = "ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3"
	Unspent2.VOut = 2

	var unspent_slice []*Unspent

	unspentlist := append(unspent_slice, Unspent1, Unspent2)

	json1, _ := json.Marshal(unspentlist)

	fmt.Println(string(json1))

	fmt.Println("now unmarshal..........")

	const jsonStream = `
        { "txid" : "c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de" , "vout" : 0} 
        { "txid" : "ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3" , "vout" : 2 } 
   		`

	decoder := json.NewDecoder(strings.NewReader(jsonStream))
	for {
		var m Unspent
		if err := decoder.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %d \n ", m.TxID, m.VOut)
	}

	fmt.Println("or in this way ..........")

	var jsonBlob = []byte(` [ 
        { "txid" : "c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de" , "vout" : 0 } , 
        { "txid" : "ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3" ,     "vout" : 2 } 
    ] `)

	var unspent_tx []Unspent
	err := json.Unmarshal(jsonBlob, &unspent_tx)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%s: %d \n ", unspent_tx[0].TxID, unspent_tx[0].VOut)
	fmt.Printf("%s: %d \n ", unspent_tx[1].TxID, unspent_tx[1].VOut)

}

func TestGetOmMarker(t *testing.T) {
	fmt.Printf("%s \n ", GetOmMarker())
}

func TestOmniCreatePayloadSimpleSend(t *testing.T) {
	var property_id uint32 = 100
	var amount uint64 = 10
	payload_byte_slice := OmniCreatePayloadSimpleSend(property_id, amount)
	str := string(payload_byte_slice)
	fmt.Printf(str)
}

func TestCreateTransactionBase(t *testing.T) {
	const jsonStream = `
		{ "txid" : "c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de" , "vout" : 0} 
		{ "txid" : "ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3" , "vout" : 2 } 
	   `

	tx, tx_hex, _ := CreateRawTransaction(jsonStream)
	fmt.Printf("The final base transaction hex")
	fmt.Printf("%s \n ", tx_hex)
	fmt.Println("Must be equal to: ")
	fmt.Println(TxToHex(tx))

}

// extend the opreturn data to base transaction
func TestCreateTransactionOpreturn(t *testing.T) {
	const jsonStream = `
		{ "txid" : "c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de" , "vout" : 0} 
		{ "txid" : "ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3" , "vout" : 2 } 
	   `

	// 2) Construct payload
	payload, payload_hex := Omni_createpayload_simplesend("2", "0.1", true)
	fmt.Println("expect: 00000000000000020000000000989680")

	fmt.Println(payload_hex)

	// 3) Construct transaction base
	tx, tx_hex, _ := CreateRawTransaction(jsonStream)
	fmt.Printf("The final base transaction hex")
	fmt.Printf("%s \n ", tx_hex)

	// 4) Attach payload output
	tx, _ = Omni_createrawtx_opreturn(tx, payload, payload_hex)

	fmt.Println("expect: 0100000002de95b97cf4c67ec01485fd698ec154a325ff69dd3e58435d7024bae7f69534c20000000000ffffffffb3b60aaa69b860c9bf31e742e3b37e75a2a553fd0bebf8aaf7da0e9bb07316ee0200000000ffffffff010000000000000000166a146f6d6e690000000000000002000000000098968000000000")
	fmt.Println("export: ")
	fmt.Println(TxToHex(tx))

	//5) Attach reference/receiver output
	receiver := "1Njbpr7EkLA1R8ag8bjRN7oks7nv5wUn3o"
	tx, _ = Omni_createrawtx_reference(tx, receiver, &chaincfg.MainNetParams)

	expected := "0100000002de95b97cf4c67ec01485fd698ec154a325ff69dd3e58435d7024bae7f69534c20000000000ffffffffb3b60aaa69b860c9bf31e742e3b37e75a2a553fd0bebf8aaf7da0e9bb07316ee0200000000ffffffff020000000000000000166a146f6d6e690000000000000002000000000098968022020000000000001976a914ee692ea81da1b12d3dd8f53fd504865c9d843f5288ac00000000"
	fmt.Println("export: ")
	exported := TxToHex(tx)
	fmt.Println(exported)

	if strings.Compare(expected, exported) == 1 {
		fmt.Println("succeed constructing opreturn raw transaction offline")
	}
}
