package service

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"log"
	"testing"
)

func TestMyTx(t *testing.T) {
	createMyTx()
}
func TestMyTxList(t *testing.T) {
	result, err := rpcClient.ListUnspent("mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM")
	log.Println(result)
	log.Println(err)
	tx, err := rpcClient.GetTransactionById("ab49456042bc098d16ed760623b70e45e2a457b529592a1542d4f640b758c924")
	log.Println(tx)
	log.Println(err)
}

func TestGenerateBTC(t *testing.T) {
	wifKey, pubKey, err := GenerateBTCTest()
	log.Println(wifKey)
	log.Println(pubKey)
	log.Println(err)
}

//第一步，构建一个带自定义问题的交易，这个交易要转账给一个新的地址 0.1btc， 这个新的地址要用这个钱，需要给出解答
//首先我得有个账号
//然后给这个账号转账，获得一个input
//然后把这个input转给另一个地址
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
func TestCreateTx32(t *testing.T) {
	temp := "c1"
	log.Println([]byte(temp))
}
func TestCreateTx4(t *testing.T) {

	result, err := rpcClient.DecodeScript("47304402201ac28bcd22c86dd011196bad8d25c388b542df2ae2261ae806ec332c1408d32e022051f7131e9c9fe9d2a08c4f28421ea66e3deec320afd89a957d684679473f0c2b012103a65e09459907e81b72e2840594586235af81873ade02bf7c8f4098c454cc57a8")
	log.Println(err)
	log.Println(result)

	decodeString, err := hex.DecodeString("6a146f6d6e6900000000000002c10000000005f5e100")
	log.Println(decodeString)
	log.Println(err)
	bytes, err := hexutil.Decode("0x047304402201ac28bcd22c86dd011196bad8d25c388b542df2ae2261ae806ec332c1408d32e022051f7131e9c9fe9d2a08c4f28421ea66e3deec320afd89a957d684679473f0c2b012103a65e09459907e81b72e2840594586235af81873ade02bf7c8f4098c454cc57a8")
	log.Println(err)
	log.Println(bytes)
	log.Println(string(bytes))
	s := hex.EncodeToString(bytes)
	log.Println(s)
}
