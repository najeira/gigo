package out_s3

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/najeira/gigo"
)

const (
	pluginName = "out_s3"
)

type Config struct {
	Credentials       *credentials.Credentials
	Region            string
	Bucket            string
	PublicRead        bool
	ReducedRedundancy bool
	Eventer           gigo.Eventer
}

type Writer struct {
	bucket            string
	publicRead        bool
	reducedRedundancy bool
	eventer           gigo.Eventer

	svc  *s3.S3
	buf  bytes.Buffer
	gw   *gzip.Writer
	size int
}

func New(config Config) *Writer {
	cfg := &aws.Config{Region: aws.String(config.Region)}
	if config.Credentials != nil {
		cfg.Credentials = config.Credentials
	}
	sess := session.New(cfg)
	svc := s3.New(sess)
	return &Writer{
		svc:               svc,
		bucket:            config.Bucket,
		reducedRedundancy: config.ReducedRedundancy,
		publicRead:        config.PublicRead,
		eventer:           config.Eventer,
	}
}

func (w *Writer) Write(data []byte) (int, error) {
	if w.gw == nil {
		w.gw = gzip.NewWriter(&w.buf)
		w.size = 0
		w.debugf("new gzip writer")
	}

	n, err := w.gw.Write(data)
	if err != nil {
		w.errorf("gzip %s", err)
		return 0, err
	}

	w.size += n
	w.debugf("write %d bytes", n)
	return len(data), nil
}

func (w *Writer) Flush(key string) error {
	if err := w.gw.Close(); err != nil {
		w.errorf("gzip %s", err)
		return err
	}

	data := w.buf.Bytes()
	br := bytes.NewReader(data)
	if err := w.put(key, br); err != nil {
		w.errorf("s3 %s", err)
		return err
	}

	w.debugf("put %d bytes (%d)", len(data), w.size)
	w.gw = nil
	w.size = 0
	w.buf.Reset()
	return nil
}

func (w *Writer) Len() int {
	return w.size
}

func (w *Writer) Exist(key string) (bool, error) {
	s3params := &s3.HeadObjectInput{
		Bucket: aws.String(w.bucket),
		Key:    aws.String(key),
	}
	_, err := w.svc.HeadObject(s3params)
	if err != nil {
		if aerr, ok := err.(awserr.RequestFailure); ok {
			code := aerr.StatusCode()
			if code == 403 || code == 404 {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (w *Writer) put(key string, body io.ReadSeeker) error {
	s3params := &s3.PutObjectInput{
		Bucket:          aws.String(w.bucket),
		Key:             aws.String(key),
		ContentType:     aws.String("text/plain"),
		ContentEncoding: aws.String("gzip"),
		Body:            body,
	}
	if w.publicRead {
		s3params.ACL = aws.String("public-read")
	} else {
		s3params.ACL = aws.String("private")
	}
	if w.reducedRedundancy {
		s3params.StorageClass = aws.String("REDUCED_REDUNDANCY")
	}
	_, err := w.svc.PutObject(s3params)
	return err
}

func (w *Writer) debugf(msg string, args ...interface{}) {
	w.emitf(gigo.Debug, msg, args...)
}

func (w *Writer) infof(msg string, args ...interface{}) {
	w.emitf(gigo.Info, msg, args...)
}

func (w *Writer) errorf(msg string, args ...interface{}) {
	w.emitf(gigo.Err, msg, args...)
}

func (w *Writer) emitf(level int, msg string, args ...interface{}) {
	if w.eventer != nil {
		w.eventer.Emit(pluginName, level, fmt.Sprintf(msg, args...))
	}
}
