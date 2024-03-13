package util

import (
	"strconv"
	"strings"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
)

func IsEmpty(str string) bool {
	if str == "" || len(str) <= 0 {
		return true
	}
	return false
}

func IsNotEmpty(str string) bool {
	return !IsEmpty(str)
}

func IsBlank(str string) bool {
	str = strings.Trim(str, " ")
	return IsEmpty(str)
}

func IsNotBlank(str string) bool {
	return !IsBlank(str)
}

func SplitString(str string, split string) []string {
	if str != "" {
		var stringList []string
		splitStrings := strings.Split(str, split)
		for _, splitString := range splitStrings {
			splitString = strings.TrimSpace(splitString)
			if splitString != "" {
				stringList = append(stringList, splitString)
			}
		}
		return stringList
	}
	return nil
}

func HandleRangeStr(rangeStr string) (bool, []int) {
	var rangeIntegers []int
	if rangeStr != "" {
		rangeCommaStrings := strings.Split(rangeStr, constant.Comma)
		for _, rangeCommaString := range rangeCommaStrings {
			rangeCommaString = strings.TrimSpace(rangeCommaString)
			if strings.Contains(rangeCommaString, constant.Hyphen) {
				rangeHyphens := strings.Split(rangeCommaString, constant.Hyphen)
				if len(rangeHyphens) != 2 {
					return true, nil
				}
				startStr := strings.TrimSpace(rangeHyphens[0])
				endStr := strings.TrimSpace(rangeHyphens[1])
				start, err := strconv.Atoi(startStr)
				if err != nil {
					return true, nil
				}
				end, err := strconv.Atoi(endStr)
				if err != nil {
					return true, nil
				}
				if start >= end {
					return true, nil
				}
				for i := start; i <= end; i++ {
					rangeIntegers = append(rangeIntegers, i)
				}
			} else {
				rangeComma, err := strconv.Atoi(rangeCommaString)
				if err != nil {
					return true, nil
				}
				rangeIntegers = append(rangeIntegers, rangeComma)
			}
		}
	}
	return false, rangeIntegers
}
