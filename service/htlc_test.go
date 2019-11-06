package service

import (
	"log"
	"testing"
)

func TestCreateNewAddress(t *testing.T) {
	address, _ := rpcClient.GetNewAddress("newAddress")
	log.Println(address)
	result, _ := rpcClient.DumpPrivKey(address)
	log.Println(result)
	rpcClient.ValidateAddress(address)
}

func TestGetBalanceByAddress(t *testing.T) {
	address := "2NFhMhDJT9TsnBCG6L2amH3eDXxgwW6EJh7"
	balance, err := rpcClient.GetBalanceByAddress(address)
	log.Println(err)
	log.Println(balance)
	result, err := rpcClient.OmniGetbalance(address, 121)
	log.Println(err)
	log.Println(result)
}

func TestUpdateData(t *testing.T) {
}
