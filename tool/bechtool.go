package tool

import "errors"

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
