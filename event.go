package gigo

import (
	"fmt"
	"strings"
)

const (
	logError logLevel = iota
	logInfo
	logDebug
	logNo
)

type logLevel uint8

func (lvl logLevel) String() string {
	switch lvl {
	case logDebug:
		return "debug"
	case logInfo:
		return "info"
	case logError:
		return "error"
	}
	return ""
}

func parseLogLevel(level string) logLevel {
	if len(level) > 0 {
		c := []rune(strings.ToLower(level))[0]
		switch c {
		case 'e':
			return logError
		case 'w': // warning to error
			return logError
		case 'i':
			return logInfo
		case 'd':
			return logDebug
		case 't': // trace to debug
			return logDebug
		}
	}
	return logNo
}

type Logger func(calldepth int, s string) error

type Mixin struct {
	Name string

	logger   Logger
	logLevel logLevel
}

func (m *Mixin) SetLogging(fn Logger, lvl string) {
	m.logger = fn
	m.logLevel = parseLogLevel(lvl)
}

func (m *Mixin) Debug(args ...interface{}) {
	m.output(logDebug, args...)
}

func (m *Mixin) Debugf(format string, args ...interface{}) {
	m.outputf(logDebug, format, args...)
}

func (m *Mixin) Info(args ...interface{}) {
	m.output(logInfo, args...)
}

func (m *Mixin) Infof(format string, args ...interface{}) {
	m.outputf(logInfo, format, args...)
}

func (m *Mixin) Error(args ...interface{}) {
	m.output(logError, args...)
}

func (m *Mixin) Errorf(format string, args ...interface{}) {
	m.outputf(logError, format, args...)
}

func (m *Mixin) output(level logLevel, args ...interface{}) {
	if name := m.logLevelName(level); len(name) > 0 {
		m.printToLogger(name, fmt.Sprint(args...))
	}
}

func (m *Mixin) outputf(level logLevel, format string, args ...interface{}) {
	if name := m.logLevelName(level); len(name) > 0 {
		m.printToLogger(name, fmt.Sprintf(format, args...))
	}
}

func (m *Mixin) logLevelName(level logLevel) string {
	if m.logger == nil {
		return ""
	} else if m.logLevel < level {
		return ""
	}
	return level.String()
}

func (m *Mixin) printToLogger(name string, msg string) {
	m.logger(5, fmt.Sprintf("[%s] %s: %s", name, m.Name, msg))
}
