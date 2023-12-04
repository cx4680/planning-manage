package spliterror

import (
	"strings"
)

func SplitError(str string) string {
	return strings.Split(str, ".")[len(strings.Split(str, "."))-1]
}
