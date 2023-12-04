package datetime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCurrent(t *testing.T) {
	tm := CurrentUnixMilli()
	assert.Equal(t, UnixMilliToTime(tm).Year(), time.Now().Year())

	milli := UnixMilliToTime(tm)
	assert.Equal(t, milli.Year(), time.Now().Year())
	assert.Equal(t, milli.Month(), time.Now().Month())
	assert.Equal(t, milli.Day(), time.Now().Day())
}
