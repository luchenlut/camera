package echo

import (
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var formatter logrus.Formatter

type wrapper struct {
	old  logrus.Formatter
	hook *LfsHook
}

func (w *wrapper) Format(entry *logrus.Entry) ([]byte, error) {
	modified := entry.WithField(w.hook.Field, w.hook.Formatter(w.hook.findCaller()))
	modified.Level = entry.Level
	modified.Message = entry.Message
	return w.old.Format(modified)
}

func newFormatter(old logrus.Formatter, hook *LfsHook) logrus.Formatter {
	return &wrapper{old: old, hook: hook}
}

func (hook *LfsHook) findCaller() (string, string, int) {
	var (
		pc       uintptr
		file     string
		function string
		line     int
	)
	for i := 0; i < 10; i++ {
		pc, file, line = getCaller(hook.Skip + i)
		if !hook.skipFile(file) {
			break
		}
	}
	if pc != 0 {
		frames := runtime.CallersFrames([]uintptr{pc})
		frame, _ := frames.Next()
		function = frame.Function
	}
	return file, function, line
}

func (hook *LfsHook) skipFile(file string) bool {
	for i := range hook.SkipPrefixes {
		if strings.HasPrefix(file, hook.SkipPrefixes[i]) {
			return true
		}
	}

	return false
}

func getCaller(skip int) (uintptr, string, int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return 0, "", 0
	}

	n := 0
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			n++
			if n >= 2 {
				file = file[i+1:]
				break
			}
		}
	}

	return pc, file, line
}
