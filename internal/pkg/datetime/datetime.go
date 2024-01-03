package datetime

import "time"

const (
	FullTimeFmt = "2006-01-02 15:04:05"
	ZoneTimeFmt = "2006-01-02T15:04:05Z"
	DayTimeFmt  = "2006-01-02"
)

func CurrentUnixMilli() int64 {
	return time.Now().UnixMilli()
}

func UnixMilliToTime(msec int64) time.Time {
	return time.UnixMilli(msec)
}

func OffsetDayUnixMilli(day int) int64 {
	return time.Now().AddDate(0, 0, day).UnixMilli()
}

func GetNow() time.Time {
	return time.Now()
}

func GetNowStr() string {
	return time.Now().Format(FullTimeFmt)
}

func TimeToStr(t time.Time, fmt string) string {
	return t.Format(fmt)
}

func StrToTime(fmt, str string) (t time.Time) {
	t, _ = time.Parse(fmt, str)
	return
}
