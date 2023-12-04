package data

import (
	"net/url"
	"strings"

	"code.cestc.cn/zhangzhi/planning-manage/internal/api/constant"
)

func SplitCommaString(str string) []string {
	if str != "" {
		return strings.Split(str, constant.Comma)
	}

	return nil
}

func StringDecode(str string) string {
	if str != "" {
		unescape, err := url.QueryUnescape(str)
		if err != nil {
			return str
		}
		return unescape
	}
	return str
}

func StringArrayDecode(strs []string) []string {
	var array []string
	for _, str := range strs {
		unescape, err := url.QueryUnescape(str)
		if err != nil {
			array = append(array, str)
			continue
		}
		array = append(array, unescape)
	}
	return array
}
