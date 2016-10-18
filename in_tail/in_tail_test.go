package in_tail

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

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

	p, err := Open(Config{File: path})
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
}
