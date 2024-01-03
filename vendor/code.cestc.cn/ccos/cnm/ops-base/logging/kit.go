package logging

import (
	"path/filepath"
)

var DefaultKit Kit

type Kit interface {
	// Access log
	A() *Logger
	// SQL log
	S() *Logger
	// business log
	R() *Logger
}

type kit struct {
	a, s, r, o *Logger
}

func NewKit(a, s, r *Logger) Kit {
	return kit{
		a: a,
		s: s,
		r: r,
	}
}

func (c kit) A() *Logger {
	return c.a
}

func (c kit) S() *Logger {
	return c.s
}

func (c kit) R() *Logger {
	return c.r
}

func setKitLog(dir string) {
	var withFile bool
	if len(dir) > 0 {
		withFile = true
	}

	// 调用日志和请求日志
	accessLog := NewLogging(filepath.Join(dir, "access.log"), OtherLevelAccess, withFile)
	accessLog.SetRotateByDay()

	// sql日志
	sqlLog := NewLogging(filepath.Join(dir, "sql.log"), OtherLevelSql, withFile)
	sqlLog.SetRotateByDay()

	// 请求下游business日志
	requestLog := NewLogging(filepath.Join(dir, "request.log"), OtherLevelRequest, withFile)
	requestLog.SetRotateByDay()

	DefaultKit = NewKit(accessLog, sqlLog, requestLog)
}
