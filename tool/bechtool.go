package tool

import (
	"errors"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/omnilaboratory/obd/bean"
	"strconv"
	"strings"
	"time"
)

func GetMsgLengthFromInt(num int) (code string, err error) {
	if num < 0 || num > 931 {
		return "", errors.New("wrong num")
	}
	firstNum := num / 32
	secondNum := num % 32
	code = convertNumToCode(firstNum)
	code += convertNumToCode(secondNum)
	return code, nil
}

var codes = [][]string{
	{"q", "p", "z", "r", "y", "9", "x", "8"},
	{"g", "f", "2", "t", "v", "d", "w", "0"},
	{"s", "3", "j", "n", "5", "4", "k", "h"},
	{"c", "e", "6", "m", "u", "a", "7", "l"},
}

func convertNumToCode(num int) string {
	if num < 0 && num > 31 {
		return ""
	}
	rows := num / 8
	col := num % 8
	return codes[rows][col]
}

func getCodeIndex(code string) (row, col int, err error) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			if codes[i][j] == code {
				return i, j, nil
			}
		}
	}
	return 0, 0, errors.New("not found")
}

func ConvertNumToString(num int, result *string) {
	currValue := num % 32
	code := convertNumToCode(currValue)
	*result = code + *result
	num = num / 32
	if num == 0 {
		return
	} else {
		ConvertNumToString(num, result)
	}
}
func ConvertBechStringToNum(str string) (result int64, err error) {
	totalLength := len(str)
	result = 0
	for index, item := range str {
		currIndex := totalLength - index - 1
		pow := math.BigPow(32, int64(currIndex)).Int64()
		row, col, _ := getCodeIndex(string(item))
		itemValue := int64(int64(row*8+col) * pow)
		result += itemValue
	}
	return result, nil
}

func DecodeInvoiceObjFromCodes(encode string) (invoice bean.HtlcRequestInvoice, err error) {
	source := encode

	invoice = bean.HtlcRequestInvoice{}
	if strings.HasPrefix(encode, "obbc") {
		invoice.NetType = "obbc"
	}
	if strings.HasPrefix(encode, "obtb") {
		invoice.NetType = "obtb"
	}
	if strings.HasPrefix(encode, "obcrt") {
		invoice.NetType = "obcrt"
	}
	if len(invoice.NetType) == 0 {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, invoice.NetType)
	amoutEndIndex := strings.Index(encode, "s1p")
	amountStr := encode[0:amoutEndIndex]
	amount, err := strconv.Atoi(amountStr)
	invoice.Amount = float64(amount / 100000000)
	amountStr = amountStr + "s1"
	encode = strings.TrimPrefix(encode, amountStr)
	//propertyId
	prefix := encode[0:1]
	if prefix != "p" {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, prefix)
	lengthStr := encode[0:2]
	length, err := ConvertBechStringToNum(lengthStr)
	if err != nil {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, lengthStr)
	itemStr := encode[0:length]
	propertyId, err := ConvertBechStringToNum(itemStr)
	if err != nil {
		return invoice, errors.New("error encode")
	}
	invoice.PropertyId = propertyId
	encode = strings.TrimPrefix(encode, itemStr)

	//nodePeerId
	prefix = encode[0:1]
	if prefix != "n" {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, prefix)
	lengthStr = encode[0:2]
	length, err = ConvertBechStringToNum(lengthStr)
	if err != nil {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, lengthStr)
	itemStr = encode[0:length]
	invoice.RecipientNodePeerId = itemStr
	encode = strings.TrimPrefix(encode, itemStr)

	//userPeerId
	prefix = encode[0:1]
	if prefix != "u" {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, prefix)
	lengthStr = encode[0:2]
	length, err = ConvertBechStringToNum(lengthStr)
	if err != nil {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, lengthStr)
	itemStr = encode[0:length]
	invoice.RecipientUserPeerId = itemStr
	encode = strings.TrimPrefix(encode, itemStr)

	//H
	prefix = encode[0:1]
	if prefix != "h" {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, prefix)
	lengthStr = encode[0:2]
	length, err = ConvertBechStringToNum(lengthStr)
	if err != nil {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, lengthStr)
	itemStr = encode[0:length]
	invoice.H = itemStr
	encode = strings.TrimPrefix(encode, itemStr)

	//ExpiryTime
	prefix = encode[0:1]
	if prefix != "x" {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, prefix)
	lengthStr = encode[0:2]
	length, err = ConvertBechStringToNum(lengthStr)
	if err != nil {
		return invoice, errors.New("error encode")
	}
	encode = strings.TrimPrefix(encode, lengthStr)
	itemStr = encode[0:length]
	expiryTime, err := ConvertBechStringToNum(itemStr)
	if err != nil {
		return invoice, errors.New("error encode")
	}
	date := time.Unix(expiryTime, 0)
	invoice.ExpiryTime = bean.JsonDate(date)
	encode = strings.TrimPrefix(encode, itemStr)

	//Description
	prefix = encode[0:1]
	if prefix == "d" {
		encode = strings.TrimPrefix(encode, prefix)
		lengthStr = encode[0:2]
		length, err = ConvertBechStringToNum(lengthStr)
		if err != nil {
			return invoice, errors.New("error encode")
		}
		encode = strings.TrimPrefix(encode, lengthStr)
		itemStr = encode[0:length]
		if err != nil {
			return invoice, errors.New("error encode")
		}
		invoice.Description = itemStr
		encode = strings.TrimPrefix(encode, itemStr)
	}
	checkSum, err := ConvertBechStringToNum(encode)
	if err != nil {
		return invoice, errors.New("error encode")
	}
	source = strings.TrimSuffix(source, encode)

	bytes := []byte(source)
	sum := 0
	for _, item := range bytes {
		sum += int(item)
	}
	if int64(sum) != checkSum {
		return invoice, errors.New("error encode")
	}

	return invoice, nil
}
