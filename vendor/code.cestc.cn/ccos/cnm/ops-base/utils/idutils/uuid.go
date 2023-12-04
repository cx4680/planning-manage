package idutils

import (
	uuid "github.com/satori/go.uuid"
)

func GetUUID() string {
	var u2 = uuid.NewV4()
	return u2.String()
}
