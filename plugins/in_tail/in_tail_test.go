package in_tail

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
)

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

type scanReceiver struct {
	buffer bytes.Buffer
	err    error
}

func (s *scanReceiver) handle(line string) {
	_, s.err = s.buffer.WriteString(line)
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
