package in_tail

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
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
	l.debug.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *testLogger) Infof(format string, args ...interface{}) {
	l.info.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *testLogger) Noticef(format string, args ...interface{}) {
	l.notice.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *testLogger) Warnf(format string, args ...interface{}) {
	l.warn.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *testLogger) Errorf(format string, args ...interface{}) {
	l.err.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (l *testLogger) Criticalf(format string, args ...interface{}) {
	l.crit.WriteString(fmt.Sprintf(format, args...) + "\n")
}

type testEmitter struct {
	buffer bytes.Buffer
}

func (e *testEmitter) Emit(msg interface{}) (err error) {
	if b, ok := msg.([]byte); ok {
		_, err = e.buffer.Write(b)
		e.buffer.WriteString("\n")
		//fmt.Println(string(b))
	} else if s, ok := msg.(string); ok {
		_, err = e.buffer.WriteString(s)
		e.buffer.WriteString("\n")
		//fmt.Println(s)
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
	if rets != "this\nis\ntest\n" {
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
	if rets != "this\nis\ntest\n" {
		t.Errorf("invalid emit")
	}
}

func TestHandleOutPipe(t *testing.T) {
	e := testEmitter{}
	p := New(Config{Emitter: &e})

	r, w := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		p.handleOutPipe(r)
		wg.Done()
	}()

	w.Write([]byte("this\n"))
	w.Write([]byte("is\n"))
	w.Write([]byte("test\n"))
	w.Close()

	wg.Wait()

	rets := e.buffer.String()
	if rets != "this\nis\ntest\n" {
		t.Errorf("invalid emit")
	}
}

func TestHandlePipes(t *testing.T) {
	e := testEmitter{}
	l := testLogger{}
	p := New(Config{Emitter: &e, Logger: &l})

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		p.handlePipes(outR, errR)
		wg.Done()
	}()

	outW.Write([]byte("this\n"))
	outW.Write([]byte("is\n"))
	outW.Write([]byte("test\n"))
	outW.Close()

	errW.Write([]byte("this\n"))
	errW.Write([]byte("is\n"))
	errW.Write([]byte("test\n"))
	errW.Close()

	wg.Wait()

	rets := e.buffer.String()
	if rets != "this\nis\ntest\n" {
		t.Errorf("invalid emit")
	}

	rets2 := l.warn.String()
	if rets2 != "this\nis\ntest\n" {
		t.Errorf("invalid warn")
	}
}

func TestTail(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		t.Error(err)
	}

	path := f.Name()

	defer func() {
		f.Close()
		os.Remove(path)
	}()

	e := testEmitter{}
	l := testLogger{}
	p := New(Config{Emitter: &e, Logger: &l, File: path})

	err = p.Start()
	if err != nil {
		t.Error(err)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = f.WriteString("this\n")
	if err != nil {
		t.Error(err)
	}

	_, err = f.WriteString("is\n")
	if err != nil {
		t.Error(err)
	}

	_, err = f.WriteString("test\n")
	if err != nil {
		t.Error(err)
	}

	err = f.Sync()
	if err != nil {
		t.Error(err)
	}

	time.Sleep(10 * time.Millisecond)

	err = p.Stop()
	if err != nil {
		t.Error(err)
	}

	rets := e.buffer.String()
	if rets != "this\nis\ntest\n" {
		t.Errorf("invalid emit: %s", rets)
	}

	rets2 := l.warn.String()
	if rets2 != "" {
		t.Errorf("invalid warn: %s", rets2)
	}

	//fmt.Println(l.debug.String())
}
