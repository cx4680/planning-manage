package util

import "strings"

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
