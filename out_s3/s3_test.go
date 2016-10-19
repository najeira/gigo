package out_s3

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type testS3Service struct {
	buf     bytes.Buffer
	written int64
	err     error
}

func (svc *testS3Service) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	svc.written, svc.err = io.Copy(&svc.buf, input.Body)
	return &s3.PutObjectOutput{
		ETag: aws.String("test"),
	}, nil
}

func TestNewWriteFlush(t *testing.T) {
	svc := &testS3Service{}

	p := New(Config{
		Region: "ap-northeast-1",
		Bucket: "test",
	})
	p.svc = svc

	_, err := io.WriteString(p, "this\n")
	if err != nil {
		t.Error(err)
	}

	_, err = io.WriteString(p, "is\n")
	if err != nil {
		t.Error(err)
	}

	_, err = io.WriteString(p, "test\n")
	if err != nil {
		t.Error(err)
	}

	err = p.Flush()
	if err != nil {
		t.Error(err)
	}

	if svc.err != nil {
		t.Error(svc.err)
	}

	br := bytes.NewReader(svc.buf.Bytes())
	gr, _ := gzip.NewReader(br)
	body, err := ioutil.ReadAll(gr)
	str := string(body)
	if str != "this\nis\ntest\n" {
		t.Errorf("invalid body: %s", str)
	}
}
