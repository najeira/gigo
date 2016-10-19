package gigo

import (
	"fmt"
	"strings"
)

const (
	LogError LogLevel = iota
	LogInfo
	LogDebug
	LogNo
)

type LogLevel uint8

func (lvl LogLevel) String() string {
	switch lvl {
	case LogDebug:
		return "debug"
	case LogInfo:
		return "info"
	case LogError:
		return "error"
	}
	return ""
}

func ParseLogLevel(level string) LogLevel {
	if len(level) > 0 {
		c := []rune(strings.ToLower(level))[0]
		switch c {
		case 'e':
			return LogError
		case 'i':
			return LogInfo
		case 'd':
			return LogDebug
		case 'n':
			return LogNo
		}
	}
	return LogNo
}

type Logger interface {
	Print(message string)
}

type Mixin struct {
	Logger
	Name     string
	LogLevel LogLevel
}

func (m *Mixin) Debug(args ...interface{}) {
	m.output(LogDebug, args...)
}

func (m *Mixin) Debugf(format string, args ...interface{}) {
	m.outputf(LogDebug, format, args...)
}

func (m *Mixin) Info(args ...interface{}) {
	m.output(LogInfo, args...)
}

func (m *Mixin) Infof(format string, args ...interface{}) {
	m.outputf(LogInfo, format, args...)
}

func (m *Mixin) Error(args ...interface{}) {
	m.output(LogError, args...)
}

func (m *Mixin) Errorf(format string, args ...interface{}) {
	m.outputf(LogError, format, args...)
}

func (m *Mixin) output(level LogLevel, args ...interface{}) {
	if name := m.logLevelName(level); len(name) > 0 {
		m.printToLogger(name, fmt.Sprint(args...))
	}
}

func (m *Mixin) outputf(level LogLevel, format string, args ...interface{}) {
	if name := m.logLevelName(level); len(name) > 0 {
		m.printToLogger(name, fmt.Sprintf(format, args...))
	}
}

func (m *Mixin) logLevelName(level LogLevel) string {
	if m.Logger == nil {
		return ""
	} else if m.LogLevel < level {
		return ""
	}
	return level.String()
}

func (m *Mixin) printToLogger(name string, msg string) {
	m.Logger.Print(fmt.Sprintf("[%s] %s: %s", name, m.Name, msg))
}
