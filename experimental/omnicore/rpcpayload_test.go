package omnicore

import (
	"fmt"
	"testing"
)

func TestOmni_createpayload_simplesend(t *testing.T) {

	_, payload_hex := Omni_createpayload_simplesend("2", "0.1", true)
	fmt.Println("expect: 00000000000000020000000000989680")

	fmt.Println(payload_hex)

}
