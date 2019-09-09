package tool

import (
	"fmt"
	"testing"
)

func TestGetAddress(t *testing.T) {
	fmt.Println(SignMsg([]byte("03870f2aebd7ac762bf26de14bf4624781cd4e4ed3ca4ada16c883f1d7a492ec0a")))
}
