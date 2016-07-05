package gigo

type Logger interface {
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func EnsureLogger(l Logger) Logger {
	if l != nil {
		return l
	}
	return &blackholeLogger{}
}

type blackholeLogger struct {
}

func (l *blackholeLogger) Tracef(format string, args ...interface{}) {
}

func (l *blackholeLogger) Debugf(format string, args ...interface{}) {
}

func (l *blackholeLogger) Infof(format string, args ...interface{}) {
}

func (l *blackholeLogger) Warnf(format string, args ...interface{}) {
}

func (l *blackholeLogger) Errorf(format string, args ...interface{}) {
}
