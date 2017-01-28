package cloudwatchlogs

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/najeira/gigo"
)

const (
	pluginName      = "cloudwatchlogs"
	batchSize       = 768 * 1024
	batchCount      = 10000
	rowOverhead     = 26
	rowMaxSize      = 256 * 1024
	defaultInterval = time.Second * 5
)

var (
	ErrClosed = errors.New("cloudwatchlogs: writer closed")
	ErrSize   = errors.New("cloudwatchlogs: too long")
)

type WriterConfig struct {
	Credentials *credentials.Credentials
	Region      string
	Group       string
	Stream      string
	Interval    time.Duration
	BatchSize   int
	BatchCount  int
}

type Writer struct {
	gigo.Mixin

	svc      writerService
	group    string
	stream   string
	sequence *string

	interval   time.Duration
	eventCh    chan *cloudwatchlogs.InputLogEvent
	events     []*cloudwatchlogs.InputLogEvent
	size       int
	batchSize  int
	batchCount int
	closed     chan struct{}
}

func NewWriter(config WriterConfig) *Writer {
	if config.Interval <= 0 {
		config.Interval = defaultInterval
	}
	w := &Writer{
		svc:      newClient(config.Region, config.Credentials),
		group:    config.Group,
		stream:   config.Stream,
		sequence: nil,
		interval: config.Interval,
		eventCh:  make(chan *cloudwatchlogs.InputLogEvent, 100),
		closed:   make(chan struct{}),
	}
	if config.BatchSize > 0 {
		w.batchSize = config.BatchSize
	} else {
		w.batchSize = batchSize
	}
	if config.BatchCount > 0 {
		w.batchCount = config.BatchCount
	} else {
		w.batchCount = batchCount
	}
	w.Name = pluginName
	go w.run()
	return w
}

func (w *Writer) Write(msg string) error {
	if w.eventCh == nil {
		w.Info(ErrClosed)
		return ErrClosed
	} else if len(msg) > rowMaxSize {
		return ErrSize
	}
	event := &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(msg),
		Timestamp: aws.Int64(timeToMilli(time.Now())),
	}
	w.eventCh <- event
	w.Debugf("write a row %d bytes", len(msg))
	return nil
}

func (w *Writer) run() {
	defer close(w.closed)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	eventCh := w.eventCh
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				w.flush() // flush remaining events
				return
			}

			if w.size >= w.batchSize {
				w.flush()
			} else if len(w.events) >= w.batchCount {
				w.flush()
			}
			w.events = append(w.events, event)
			w.size += (len(aws.StringValue(event.Message)) + rowOverhead)
		case <-ticker.C:
			w.flush()
		}
	}
}

func (w *Writer) flush() error {
	if len(w.events) <= 0 {
		return nil
	}

	events := w.events[:]
	size := w.size
	w.events = nil
	w.size = 0

	resp, err := w.svc.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  aws.String(w.group),
		LogStreamName: aws.String(w.stream),
		SequenceToken: w.sequence,
	})
	if err != nil {
		return err
	} else if resp.RejectedLogEventsInfo != nil {
		return errors.New(resp.RejectedLogEventsInfo.GoString())
	}

	w.sequence = resp.NextSequenceToken
	w.Infof("put %d events %d bytes sequence %s", len(events), size, aws.StringValue(w.sequence))
	return nil
}

func (w *Writer) Close() error {
	if w.eventCh != nil {
		close(w.eventCh)
		w.eventCh = nil
	}
	<-w.closed
	return nil
}

func newClient(region string, credentials *credentials.Credentials) *cloudwatchlogs.CloudWatchLogs {
	cfg := &aws.Config{Region: aws.String(region)}
	if credentials != nil {
		cfg.Credentials = credentials
	}
	sess := session.New(cfg)
	return cloudwatchlogs.New(sess)
}

type writerService interface {
	PutLogEvents(*cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error)
}
