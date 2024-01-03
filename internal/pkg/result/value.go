package result

import (
	"reflect"
)

func IsNil(data interface{}) bool {
	value := reflect.ValueOf(data)
	k := value.Kind()
	switch k {
	case reflect.Interface, reflect.Slice, reflect.Pointer, reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer:
		return value.IsNil()
		// case reflect.Array, reflect.Struct:
		// 	return data == nil
	}

	return data == nil
}
