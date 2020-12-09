package rpc

import (
	"log"
	"testing"
)

func TestClient_OmniListTransactions(t *testing.T) {
	client := NewClient()
	result, err := client.OmniListTransactions("2NAHWeCsgJGcg8q7shQDnvRbrHj2CrnqaXZ", 100, 0)
	log.Println(err)
	log.Println(result)

	s, err := client.GetTransactionById("629dd599be02b63233652bf1c6fbee0a1ccab2a8b9bc59776b5cfb45146a7d68")
	log.Println(s)
	log.Println(err)
}
