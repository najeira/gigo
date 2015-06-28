package gigo

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Noticef(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Criticalf(format string, args ...interface{})
}

func Debugf(l Logger, format string, args ...interface{}) {
	if l != nil {
		l.Debugf(format, args...)
	}
}

func Infof(l Logger, format string, args ...interface{}) {
	if l != nil {
		l.Infof(format, args...)
	}
}

func Noticef(l Logger, format string, args ...interface{}) {
	if l != nil {
		l.Noticef(format, args...)
	}
}

func Warnf(l Logger, format string, args ...interface{}) {
	if l != nil {
		l.Warnf(format, args...)
	}
}

func Errorf(l Logger, format string, args ...interface{}) {
	if l != nil {
		l.Errorf(format, args...)
	}
}

func Criticalf(l Logger, format string, args ...interface{}) {
	if l != nil {
		l.Criticalf(format, args...)
	}
}
