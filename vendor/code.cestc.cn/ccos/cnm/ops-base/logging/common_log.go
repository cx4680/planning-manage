package logging

import (
	"strings"
)

var genLog *Logger = New()
var crashLog *Logger = New()

var (
	DayRotate  = "day"
	HourRotate = "hour"
)

type CommonLogConfig struct {
	PathPreFix      string
	Rotate          string
	GenLogLevel     string
	BalanceLogLevel string
}

func isHourRotate(rotate string) bool {

	if rotate == HourRotate {
		return true
	}
	return false
}

func getNewPathName(clc CommonLogConfig, name string) string {

	hasSuf := strings.HasSuffix(clc.PathPreFix, "/")
	if hasSuf {
	} else {
		name = "/" + name
	}
	return clc.PathPreFix + name
}

func setCommonRotate(clc CommonLogConfig) {
	isHour := isHourRotate(clc.Rotate)
	if isHour {
		genLog.SetRotateByHour()
		crashLog.SetRotateByHour()
	} else {
		genLog.SetRotateByDay()
		crashLog.SetRotateByDay()
	}
}

func setCommonLogLevel(clc CommonLogConfig) {
	genLog.SetLevelByString(clc.GenLogLevel)
	crashLog.SetLevelByString("debug")
}

func setCommonOutput(clc CommonLogConfig) {
	genLog.SetOutputByName(getNewPathName(clc, "gen"), OtherLevelGen)
	crashLog.SetOutputByName(getNewPathName(clc, "crash"), OtherLevelCrash)
}

func setCommonTimer(clc CommonLogConfig) {

	genLog.SetPrintLevel(false)
	crashLog.SetPrintLevel(false)

	genLog.SetTimeFmt()
	crashLog.SetTimeFmt()
}

func InitCommonLog(clc CommonLogConfig) string {

	if len(clc.PathPreFix) == 0 {
		return ""
	}
	setCommonRotate(clc)
	if _config.WithFile {
		setCommonOutput(clc)
	}
	setCommonLogLevel(clc)
	setCommonTimer(clc)
	return ""
}

func GenLog(v ...interface{}) {
	genLog.Debug(v...)
}

func GenLogf(format string, v ...interface{}) {
	genLog.Debugf(format, v...)
}

func CrashLog(v ...interface{}) {
	crashLog.Debug(v...)
}

func CrashLogf(format string, v ...interface{}) {
	crashLog.Debugf(format, v...)
}
