package service

import (
	"log"
	"testing"
)

func TestFundingManager_NextTemporaryChanID(t *testing.T) {
	tempId := FundingCreateService.NextTemporaryChanID()
	log.Println(tempId)
	log.Println(string(tempId[:]))
	tempId = FundingCreateService.NextTemporaryChanID()
	log.Println(tempId)
	log.Println(string(tempId[:]))
}
