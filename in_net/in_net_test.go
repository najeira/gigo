package in_net

import (
	"bufio"
	"net"
	"testing"

	"github.com/najeira/gigo/testutil"
)

func TestReader(t *testing.T) {
	addr := ":9753"

	l := testutil.Logger{}
	p, err := Open(Config{Logger: &l, Net: "tcp", Addr: addr})
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

	scanner := bufio.NewScanner(p)
	if scanner.Scan() {
		line := scanner.Text()
		if line != "this" {
			t.Errorf("expect 'this' got: '%s'", line)
		}
	} else {
		t.Errorf("scan failed")
	}

	if scanner.Scan() {
		line := scanner.Text()
		if line != "is" {
			t.Errorf("expect 'is' got: '%s'", line)
		}
	} else {
		t.Errorf("scan failed")
	}

	if scanner.Scan() {
		line := scanner.Text()
		if line != "test" {
			t.Errorf("expect 'test' got: '%s'", line)
		}
	} else {
		t.Errorf("scan failed")
	}

	if scanner.Scan() {
		line := scanner.Text()
		if line != "second" {
			t.Errorf("expect 'second' got: '%s'", line)
		}
	} else {
		t.Errorf("scan failed")
	}

	if err := scanner.Err(); err != nil {
		t.Error(err)
	}

	if err = p.Close(); err != nil {
		t.Error(err)
	}

	if warns := l.Warn.String(); warns != "" {
		t.Errorf("invalid warn: %s", warns)
	}
}
