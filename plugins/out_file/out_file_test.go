package out_file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/najeira/gigo/testutil"
)

func TestReader(t *testing.T) {
	path := filepath.Join(os.TempDir(), "gigo_out_file_test")
	defer os.Remove(path)

	l := testutil.Logger{}
	p, err := Open(Config{
		Logger: &l,
		Name:   path,
		Flag:   os.O_RDWR | os.O_CREATE | os.O_TRUNC,
		Perm:   0666})
	if err != nil {
		t.Error(err)
	}

	_, err = p.Write([]byte("this\n"))
	if err != nil {
		t.Error(err)
	}

	_, err = p.Write([]byte("is\n"))
	if err != nil {
		t.Error(err)
	}

	_, err = p.Write([]byte("test\n"))
	if err != nil {
		t.Error(err)
	}

	if err := p.Close(); err != nil {
		t.Error(err)
	}

	ret, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}

	rets := string(ret)
	if rets != "this\nis\ntest\n" {
		t.Errorf("invalid emit: %s", rets)
	}

	if warns := l.Warn.String(); warns != "" {
		t.Errorf("invalid warn: %s", warns)
	}

	if errs := l.Error.String(); errs != "" {
		t.Errorf("invalid error: %s", errs)
	}
}
