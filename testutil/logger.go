package testutil

import (
	"bytes"
	"fmt"
)

type Logger struct {
	Trace bytes.Buffer
	Debug bytes.Buffer
	Info  bytes.Buffer
	Warn  bytes.Buffer
	Error bytes.Buffer
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	l.Trace.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error.WriteString(fmt.Sprintf(format, args...) + "\n")
}
