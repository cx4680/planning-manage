package utils

import (
	"math/rand"
	"strings"
)

var numberChars = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}

func RandomNumberString(lenNum int) string {
	str := strings.Builder{}
	length := len(numberChars)
	for i := 0; i < lenNum; i++ {
		l := numberChars[rand.Intn(length)]
		str.WriteString(l)
	}
	return str.String()
}
