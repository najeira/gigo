package main

import (
	"bytes"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/najeira/gigo"
	"github.com/najeira/gigo/out_s3"
)

const (
	DefaultTimeFormat    = "2006-01-02-15-04-05"
	DefaultBufferSize    = 10 * 1000 * 1000
	DefaultFlushInterval = 13 * 60 // seconds
)

var (
	ErrClosed = errors.New("out_s3: closed")
)

type WriterConfig struct {
	Key               string
	Secret            string
	Region            string
	Bucket            string
	Path              string
	Hostname          bool
	PublicRead        bool
	ReducedRedundancy bool
	TimeFormat        string
	BufferSize        int
	FlushInterval     int64
}

// Writer writes data to S3.
type Writer struct {
	gigo.Mixin

	// config
	config   WriterConfig
	cred     *credentials.Credentials
	hostname string

	// writer
	mu       sync.Mutex
	writer   *out_s3.Writer
	sequence int64

	// ready flush to S3
	ready chan *out_s3.Writer

	// wait for closing
	closed chan struct{}
}

// Creates a new Writer.
func NewWriter(config WriterConfig) (*Writer, error) {
	var cred *credentials.Credentials
	if config.Key != "" && config.Secret != "" {
		cred = credentials.NewStaticCredentials(config.Key, config.Secret, "")
	}

	var hostname string
	if config.Hostname {
		hostname_, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		hostname = hostname_
	}

	w := &Writer{
		config:   config,
		cred:     cred,
		hostname: hostname,
		ready:    make(chan *out_s3.Writer, 1),
		closed:   make(chan struct{}),
	}
	w.Name = "out_buf"
	return w, nil
}

func (w *Writer) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer == nil {
		if w.sequence != 0 {
			w.Info(ErrClosed)
			return 0, ErrClosed
		}

		// init first writer
		w.rotateImpl(true)
		go w.flush()
	}

	n, err := w.writer.Write(data)
	if err != nil {
		w.Error(err)
		return n, err
	}

	if w.writer.Len() > w.config.BufferSize {
		// current writer is full
		w.Infof("rotate by buffer size")
		w.rotateImpl(true)
	}

	//w.Debugf("write %d bytes", n)
	return n, err
}

func (w *Writer) rotate(next bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.rotateImpl(next)
}

func (w *Writer) rotateImpl(next bool) {
	if w.writer != nil {
		if w.writer.Len() > 0 {
			w.ready <- w.writer
			w.Debug("enqueue")
		}
		w.writer = nil
	}

	if !next {
		return
	}

	now := time.Now().Unix()
	if w.sequence < now {
		w.sequence = now
	} else {
		// same seconds, force forward
		w.sequence += 1
	}
	seqTime := time.Unix(w.sequence, 0)
	timeKey := seqTime.Format(w.config.TimeFormat)
	fileKey := getFileKey(w.config.Path, timeKey, w.hostname)

	output := out_s3.New(out_s3.Config{
		Credentials:       w.cred,
		Region:            w.config.Region,
		Bucket:            w.config.Bucket,
		Key:               fileKey,
		PublicRead:        w.config.PublicRead,
		ReducedRedundancy: w.config.ReducedRedundancy,
	})
	output.logLevel = w.logLevel
	output.logger = w.logger
	w.writer = output
	w.Debugf("new writer %s", fileKey)
}

func (w *Writer) flush() {
	defer close(w.closed)

	interval := time.Duration(w.config.FlushInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w.Debug("flush start")
	for {
		select {
		case chunk, ok := <-w.ready:
			if !ok {
				// chan closed
				w.Debug("flush end")
				return
			}

			// flush to S3
			if err := chunk.Flush(); err != nil {
				w.ready <- chunk // retry
			}

		case <-ticker.C:
			w.rotate(true)
		}
	}
}

func (w *Writer) Close() {
	w.rotate(false)
	close(w.ready)
	<-w.closed
	w.Debug("closed")
}

// %{path}%{time}_%{hostname}.log
func getFileKey(path string, timeKey string, hostname string) string {
	var buf bytes.Buffer
	if path != "" {
		buf.WriteString(path)
	}

	buf.WriteString(timeKey)

	if hostname != "" {
		buf.WriteString("_")
		buf.WriteString(hostname)
	}

	buf.WriteString(".log")
	return buf.String()
}
