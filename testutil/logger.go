package testutil

import (
	"bytes"
	"fmt"

	"github.com/najeira/gigo"
)

type EventLogger struct {
	buf bytes.Buffer
}

var _ gigo.Eventer = (*EventLogger)(nil)

func (e *EventLogger) Emit(tag string, level int, message string) {
	if name := levelName(level); len(name) > 0 {
		fmt.Fprintf("[%s] %s: %s\n", name, tag, message)
	}
}

func levelName(level int) string {
	switch level {
	case gigo.Debug:
		return "debug"
	case gigo.Info:
		return "info"
	case gigo.Err:
		return "error"
	}
	return ""
}
