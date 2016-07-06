package in_net

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/najeira/gigo/testutil"
)

func TestReader(t *testing.T) {
	addr := ":9753"

	l := testutil.Logger{}

	var buf bytes.Buffer
	handler := func(conn net.Conn) {
		for {
			_, err := io.Copy(&buf, conn)
			if err != nil {
				return
			}
		}
	}

	p, err := Open(Config{
		Logger:  &l,
		Net:     "tcp",
		Addr:    addr,
		Handler: handler,
	})
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Write([]byte("this\n"))
	if err != nil {
		t.Error(err)
	}

	_, err = conn.Write([]byte("is\n"))
	if err != nil {
		t.Error(err)
	}

	_, err = conn.Write([]byte("test\n"))
	if err != nil {
		t.Error(err)
	}

	if err := conn.Close(); err != nil {
		t.Error(err)
	}

	conn, err = net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Write([]byte("second\n"))
	if err != nil {
		t.Error(err)
	}

	if err := conn.Close(); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Millisecond * 10)

	ret, err := ioutil.ReadAll(&buf)
	rets := string(ret)
	if err != nil {
		t.Error(err)
	} else if rets != "this\nis\ntest\nsecond\n" {
		t.Error(rets)
	}

	if err = p.Close(); err != nil {
		t.Error(err)
	}

	if warns := l.Warn.String(); warns != "" {
		t.Errorf("invalid warn: %s", warns)
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
