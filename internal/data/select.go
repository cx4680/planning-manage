package data

import (
	"reflect"

	"gorm.io/gorm/schema"
)

const (
	gormTag = "gorm"
	column  = "COLUMN"
)

func SelectColumn(source any) []string {
	value := reflect.ValueOf(source)
	k := value.Kind()
	if reflect.Struct != k {
		panic("invalid source data type")
	}

	sourceFields := reflect.VisibleFields(reflect.TypeOf(source))
	var columns []string
	for _, field := range sourceFields {
		tag := field.Tag.Get(gormTag)
		for k, v := range schema.ParseTagSetting(tag, ";") {
			if k == column {
				columns = append(columns, v)
			}
		}
	}

	return columns
}
