package testutil

import (
	"bytes"
	"fmt"
	"log"
)

type Logger struct {
	Trace bytes.Buffer
	Debug bytes.Buffer
	Info  bytes.Buffer
	Warn  bytes.Buffer
	Error bytes.Buffer
}

func myPrintf(buf bytes.Buffer, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	log.Println(s)
	buf.WriteString(s + "\n")
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	myPrintf(l.Trace, format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	myPrintf(l.Debug, format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	myPrintf(l.Info, format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	myPrintf(l.Warn, format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	myPrintf(l.Error, format, args...)
}
