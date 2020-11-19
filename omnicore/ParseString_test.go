package omnicore

import (
	"fmt"
	"testing"
)

func TestSubstring(t *testing.T) {
	str := "i dont know how to get substring	"
	strRightOfDecimal := str[0 : len(str)-1]
	fmt.Println(len(str))
	fmt.Println(strRightOfDecimal)

}

func TestStrToInt64(t *testing.T) {

	// divisible
	// 0 decimal
	str := "1000"
	number := StrToInt64(str, true)
	fmt.Printf("%v \n", number)

	// divisible
	// < 8 decimal
	str = "1000.0001"
	number = StrToInt64(str, true)
	fmt.Printf("%v \n", number)

	//  < 8 decimal
	str = "1000.000001"
	number = StrToInt64(str, true)
	fmt.Printf("%v \n", number)

	//  = 8 decimal
	str = "1000.00000001"
	number = StrToInt64(str, true)
	fmt.Printf("%v \n", number)

	//  > 8 decimal
	str = "1000.000000001"
	number = StrToInt64(str, true)
	fmt.Printf("%v \n", number)

	// indivisible
	// has decimal
	fmt.Println("Test indivisible")

	str = "1000.0001"
	number = StrToInt64(str, false)
	fmt.Printf("%v \n", number)

	str = "100000000000"
	number = StrToInt64(str, false)
	fmt.Printf("%v \n", number)

	// divisible
	// has decimal
	fmt.Println("Test divisible: 0.0001")
	fmt.Println("Expect: 10000")

	str = "0.0001"
	number = StrToInt64(str, true)
	fmt.Printf("%v \n", number)

	// negtive
	fmt.Println("Test negtive")

	str = "-1000.0001"
	number = StrToInt64(str, false)
	fmt.Printf("%v \n", number)

}
