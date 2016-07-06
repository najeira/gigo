package in_tail

import (
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/najeira/gigo/testutil"
)

func TestScanErrPipe(t *testing.T) {
	l := testutil.Logger{}
	p := &Reader{logger: &l}

	r, w := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		p.scanErrPipe(r)
		wg.Done()
	}()

	w.Write([]byte("this\n"))
	w.Write([]byte("is\n"))
	w.Write([]byte("test\n"))
	w.Close()

	wg.Wait()

	rets := l.Warn.String()
	if rets != "this\nis\ntest\n" {
		t.Error(rets)
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

	l := testutil.Logger{}
	p, err := Open(Config{Logger: &l, File: path})
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

	if err = f.Sync(); err != nil {
		t.Error(err)
	}

	time.Sleep(10 * time.Millisecond)

	retCh := make(chan []byte, 1)
	errCh := make(chan error, 1)
	go func() {
		ret, err := ioutil.ReadAll(p)
		retCh <- ret
		errCh <- err
	}()

	if err = p.Close(); err != nil {
		t.Error(err)
	}

	if err := <-errCh; err != nil {
		t.Error(err)
	}

	rets := string(<-retCh)
	if rets != "this\nis\ntest\n" {
		t.Errorf("invalid emit: %s", rets)
	}

	checkNoWarnNoError(t, l)
}

func checkNoWarnNoError(t *testing.T, l testutil.Logger) {
	if warns := l.Warn.String(); warns != "" {
		t.Error(warns)
	}
	if errs := l.Error.String(); errs != "" {
		t.Error(errs)
	}
}
