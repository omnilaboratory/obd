package tool

import (
	"log"
	"testing"
)

func TestGenerateInitHashCode(t *testing.T) {
	log.Println(GenerateInitHashCode())
}

func TestSignMsgWithMd5(t *testing.T) {
	log.Println(SignMsgWithMd5([]byte("mrt1gppm")))
}
