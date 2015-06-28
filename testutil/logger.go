package testutil

import (
	"bytes"
	"fmt"
)

type Logger struct {
	Debug    bytes.Buffer
	Info     bytes.Buffer
	Notice   bytes.Buffer
	Warn     bytes.Buffer
	Error    bytes.Buffer
	Critical bytes.Buffer
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *Logger) Noticef(format string, args ...interface{}) {
	l.Notice.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *Logger) Criticalf(format string, args ...interface{}) {
	l.Critical.WriteString(fmt.Sprintf(format, args...) + "\n")
}
