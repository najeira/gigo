package in_tail

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
)

type testLogger struct {
	debug  bytes.Buffer
	info   bytes.Buffer
	notice bytes.Buffer
	warn   bytes.Buffer
	err    bytes.Buffer
	crit   bytes.Buffer
}

func (l *testLogger) Debugf(format string, args ...interface{}) {
	l.debug.WriteString(fmt.Sprintf(format, args...))
}

func (l *testLogger) Infof(format string, args ...interface{}) {
	l.info.WriteString(fmt.Sprintf(format, args...))
}

func (l *testLogger) Noticef(format string, args ...interface{}) {
	l.notice.WriteString(fmt.Sprintf(format, args...))
}

func (l *testLogger) Warnf(format string, args ...interface{}) {
	l.warn.WriteString(fmt.Sprintf(format, args...))
}

func (l *testLogger) Errorf(format string, args ...interface{}) {
	l.err.WriteString(fmt.Sprintf(format, args...))
}

func (l *testLogger) Criticalf(format string, args ...interface{}) {
	l.crit.WriteString(fmt.Sprintf(format, args...))
}

type testEmitter struct {
	buffer bytes.Buffer
}

func (e *testEmitter) Emit(msg interface{}) (err error) {
	if b, ok := msg.([]byte); ok {
		_, err = e.buffer.Write(b)
	} else if s, ok := msg.(string); ok {
		_, err = e.buffer.WriteString(s)
	} else {
		err = fmt.Errorf("unknown type")
	}
	return
}

func TestTrimCrLf(t *testing.T) {
	if trimCrLf("hoge") != "hoge" {
		t.Fail()
	}
	if trimCrLf("hoge\n") != "hoge" {
		t.Fail()
	}
	if trimCrLf("hoge\r") != "hoge" {
		t.Fail()
	}
	if trimCrLf("hoge\r\n") != "hoge" {
		t.Fail()
	}
	if trimCrLf("hoge\n\r") != "hoge" {
		t.Fail()
	}
}

func TestScan(t *testing.T) {
	p := New(Config{})

	var ret bytes.Buffer
	r, w := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		p.scan(r, func(line string) {
			ret.WriteString(line)
			ret.WriteString("\n")
		})
		wg.Done()
	}()

	w.Write([]byte("this\n"))
	w.Write([]byte("is\n"))
	w.Write([]byte("test\n"))
	w.Close()

	wg.Wait()

	lines := strings.Split(ret.String(), "\n")

	if len(lines) != 4 {
		t.Errorf("invalid lines")
	}
	if lines[0] != "this" {
		t.Errorf("invalid line")
	}
	if lines[1] != "is" {
		t.Errorf("invalid line")
	}
	if lines[2] != "test" {
		t.Errorf("invalid line")
	}
	if lines[3] != "" {
		t.Errorf("invalid line")
	}
}

func TestHandleErrPipe(t *testing.T) {
	l := testLogger{}
	p := New(Config{Logger: &l})

	r, w := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		p.handleErrPipe(r)
		wg.Done()
	}()

	w.Write([]byte("this\n"))
	w.Write([]byte("is\n"))
	w.Write([]byte("test\n"))
	w.Close()

	wg.Wait()

	rets := l.warn.String()

	if rets != "thisistest" {
		t.Errorf("invalid warn")
	}
}

func TestHandleLine(t *testing.T) {
	e := testEmitter{}
	p := New(Config{Emitter: &e})

	p.handleLine("this")
	p.handleLine("is")
	p.handleLine("test")

	rets := e.buffer.String()

	if rets != "thisistest" {
		t.Errorf("invalid warn")
	}
}
