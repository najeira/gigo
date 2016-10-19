package out_s3

import (
	"bytes"
	"compress/gzip"
	"errors"

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

var (
	ErrClosed = errors.New("out_s3: writer closed")
)

type Config struct {
	Credentials       *credentials.Credentials
	Region            string
	Bucket            string
	Key               string
	PublicRead        bool
	ReducedRedundancy bool

	Logger   gigo.Logger
	LogLevel gigo.LogLevel
}

type Writer struct {
	gigo.Mixin

	bucket            string
	key               string
	publicRead        bool
	reducedRedundancy bool

	svc  *s3.S3
	buf  *bytes.Buffer
	gw   *gzip.Writer
	size int
}

func New(config Config) *Writer {
	buf := &bytes.Buffer{}
	w := &Writer{
		bucket:            config.Bucket,
		key:               config.Key,
		reducedRedundancy: config.ReducedRedundancy,
		publicRead:        config.PublicRead,
		svc:               newS3(config),
		buf:               buf,
		gw:                gzip.NewWriter(buf),
		size:              0,
	}
	w.Name = pluginName
	w.LogLevel = config.LogLevel
	w.Logger = config.Logger
	return w
}

func (w *Writer) Write(data []byte) (int, error) {
	if w.gw == nil {
		w.Info(ErrClosed)
		return 0, ErrClosed
	}

	n, err := w.gw.Write(data)
	if err != nil {
		w.Error(err)
		return n, err
	}

	w.size += n
	w.Debugf("write %d bytes", n)
	return n, nil
}

func (w *Writer) Flush() error {
	if w.buf == nil {
		// already flushed to S3
		return nil
	}

	// close gzip to flush
	if w.gw != nil {
		if err := w.gw.Close(); err != nil {
			w.Error(err)
			return err
		}
		w.gw = nil
	}

	n, err := w.put(w.buf.Bytes())
	if err != nil {
		w.Info(err)
		return err
	}

	w.Infof("put %d bytes (%d)", n, w.size)
	w.buf = nil
	w.size = 0
	return nil
}

func (w *Writer) Len() int {
	return w.size
}

func (w *Writer) put(data []byte) (int, error) {
	s3params := &s3.PutObjectInput{
		Bucket:          aws.String(w.bucket),
		Key:             aws.String(w.key),
		ContentType:     aws.String("text/plain"),
		ContentEncoding: aws.String("gzip"),
		Body:            bytes.NewReader(data),
	}
	if w.publicRead {
		s3params.ACL = aws.String("public-read")
	} else {
		s3params.ACL = aws.String("private")
	}
	if w.reducedRedundancy {
		s3params.StorageClass = aws.String("REDUCED_REDUNDANCY")
	}
	res, err := w.svc.PutObject(s3params)
	if err != nil {
		return 0, err
	}
	w.Debugf("s3 etag %s", aws.StringValue(res.ETag))
	return len(data), nil
}

func newS3(config Config) *s3.S3 {
	cfg := &aws.Config{Region: aws.String(config.Region)}
	if config.Credentials != nil {
		cfg.Credentials = config.Credentials
	}
	sess := session.New(cfg)
	return s3.New(sess)
}

func Exist(config Config) (bool, error) {
	svc := newS3(config)
	s3params := &s3.HeadObjectInput{
		Bucket: aws.String(config.Bucket),
		Key:    aws.String(config.Key),
	}
	_, err := svc.HeadObject(s3params)
	if err == nil {
		// no error means the object exists
		return true, nil
	}

	aerr, ok := err.(awserr.RequestFailure)
	if !ok {
		// unknown errors
		return false, err
	}

	code := aerr.StatusCode()
	if code != 403 && code != 404 {
		// other errors
		return false, err
	}

	// 403 and 404 means the object not exists
	return false, nil
}
