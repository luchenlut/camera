package echo

import (
	"bytes"
	"fmt"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

const (
	colorNoColor = "\033[0m"
	colorRed     = "\033[91m"
	colorGreen   = "\033[92m"
	colorYellow  = "\033[93m"
	colorMagenta = "\033[95m"
	colorCyan    = "\033[96m"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

var (
	isTerminal bool
)

func init() {
	isTerminal = isatty.IsTerminal(os.Stdout.Fd())
	formatter = NewTextFormat(true)
}

type textFormat struct {
	forceColors bool
}

func NewTextFormat(forceColor ...bool) *textFormat {
	return &textFormat{
		forceColors: len(forceColor) == 1 && forceColor[0],
	}
}

func (f *textFormat) Format(entry *logrus.Entry) ([]byte, error) {
	levelText := strings.ToUpper(entry.Level.String())[0:4]
	buf := bytes.NewBuffer(make([]byte, 0, 32))
	if (f.forceColors || isTerminal) && runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		color := colorNoColor
		switch entry.Level {
		case logrus.DebugLevel:
			color = colorCyan
		case logrus.InfoLevel:
			color = colorGreen
		case logrus.WarnLevel:
			color = colorYellow
		case logrus.ErrorLevel:
			color = colorMagenta
		case logrus.PanicLevel, logrus.FatalLevel:
			color = colorRed
		}
		buf.WriteString(color)
	}
	buf.WriteString(fmt.Sprintf("[%s] ", entry.Time.Format(timeFormat)))
	buf.WriteString(fmt.Sprintf("[%s] ", levelText))

	if source, ok := entry.Data["_source"]; ok {
		buf.WriteString(fmt.Sprintf("[%s]", source))
		delete(entry.Data, "_source")
	}

	for k, v := range entry.Data {
		buf.WriteString(fmt.Sprintf("[%s=%v] ", k, v))
	}
	buf.WriteString(entry.Message)
	if (f.forceColors || isTerminal) && runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		buf.WriteString(colorNoColor)
	}
	buf.WriteString("\n")
	return buf.Bytes(), nil
}
