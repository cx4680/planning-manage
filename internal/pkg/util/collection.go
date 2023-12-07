package util

import (
	"reflect"
	"strconv"
)

func ListToMap(list interface{}, key string) map[string]interface{} {
	res := make(map[string]interface{})
	arr := ToSlice(list)
	for _, row := range arr {
		immutable := reflect.ValueOf(row).Elem()
		value := immutable.FieldByName(key)
		var val string
		if value.Type().String() == "int" || value.Type().String() == "int64" {
			val = strconv.FormatInt(immutable.FieldByName(key).Int(), 10)
		} else {
			val = immutable.FieldByName(key).String()
		}
		res[val] = row
	}
	return res
}

func ListToMaps(list interface{}, key string) map[string]interface{} {
	res := make(map[string][]interface{})
	arr := ToSlice(list)
	for _, row := range arr {
		immutable := reflect.ValueOf(row).Elem()
		value := immutable.FieldByName(key)
		var val string
		if value.Type().String() == "int" || value.Type().String() == "int64" {
			val = strconv.FormatInt(immutable.FieldByName(key).Int(), 10)
		} else {
			val = immutable.FieldByName(key).String()
		}
		list, contain := res[val]
		if contain {
			res[val] = append(list, row)
		} else {
			res[val] = []interface{}{row}
		}
	}
	return res
}

func ToSlice(arr interface{}) []interface{} {
	ret := make([]interface{}, 0)
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		ret = append(ret, arr)
		return ret
	}
	l := v.Len()
	for i := 0; i < l; i++ {
		ret = append(ret, v.Index(i).Interface())
	}
	return ret
}
