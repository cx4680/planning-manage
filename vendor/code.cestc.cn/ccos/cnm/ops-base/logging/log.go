package logging

func NewLogging(path, otherLevel string, withFile bool) *Logger {
	//l, _ := NewJSON(path, rolling.HourlyRolling)
	l := New()
	//l.SetFlags(0)
	l.SetPrintLevel(false)
	l.SetHighlighting(false)
	if withFile {
		l.SetOutputByName(path, otherLevel)
	} else {
		l.SetOutputNotWithFile(otherLevel)
	}
	l.SetTimeFmt()
	return l
}
