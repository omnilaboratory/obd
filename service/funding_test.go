package service

import (
	"log"
	"testing"
)

func TestFundingManager_NextTemporaryChanID(t *testing.T) {
	for i := 0; i < 10; i++ {
		tempId := FundingCreateService.NextTemporaryChanID()
		log.Println(tempId)
		log.Println(string(tempId[:]))
	}
}

func TestFundingManager_DeleteItem(t *testing.T) {

}
