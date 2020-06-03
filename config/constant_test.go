package config

import (
	"log"
	"testing"
)

func TestGetMinerFee(t *testing.T) {
	fee := GetMinerFee()
	log.Println(fee)
}
