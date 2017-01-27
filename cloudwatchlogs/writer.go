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
	overheadRow     = 26
	eventMaxSize    = 256 * 1024
	defaultInterval = time.Second * 5
)

var (
	ErrClosed = errors.New("cloudwatchlogs: writer closed")
	ErrSize   = errors.New("cloudwatchlogs: too long")
)

type Config struct {
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

	svc      service
	group    string
	stream   string
	sequence *string

	interval time.Duration
	eventCh  chan *cloudwatchlogs.InputLogEvent
	events   eventsBuffer

	closed chan struct{}
}

func NewWriter(config Config) *Writer {
	if config.Interval <= 0 {
		config.Interval = defaultInterval
	}
	w := &Writer{
		svc:      newClient(config),
		group:    config.Group,
		stream:   config.Stream,
		sequence: nil,
		interval: config.Interval,
		eventCh:  make(chan *cloudwatchlogs.InputLogEvent, 100),
		closed:   make(chan struct{}),
	}
	if config.BatchSize > 0 {
		w.events.batchSize = config.BatchSize
	} else {
		w.events.batchSize = batchSize
	}
	if config.BatchCount > 0 {
		w.events.batchCount = config.BatchCount
	} else {
		w.events.batchCount = batchCount
	}
	w.Name = pluginName
	go w.run()
	return w
}

func (w *Writer) Write(msg string) error {
	if w.eventCh == nil {
		w.Info(ErrClosed)
		return ErrClosed
	} else if len(msg) > eventMaxSize {
		return ErrSize
	}

	nowMilli := time.Now().UnixNano() / int64(time.Millisecond)
	event := &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(msg),
		Timestamp: aws.Int64(nowMilli),
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

			if w.events.ready() {
				w.flush()
			}
			w.events.add(event)
		case <-ticker.C:
			w.flush()
		}
	}
}

func (w *Writer) flush() error {
	events, size := w.events.drain()
	if len(events) <= 0 {
		return nil
	}

	resp, err := w.svc.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  aws.String(w.group),
		LogStreamName: aws.String(w.stream),
		SequenceToken: w.sequence,
	})
	if err != nil {
		return err
	}

	if resp.RejectedLogEventsInfo != nil {
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

func newClient(config Config) *cloudwatchlogs.CloudWatchLogs {
	cfg := &aws.Config{Region: aws.String(config.Region)}
	if config.Credentials != nil {
		cfg.Credentials = config.Credentials
	}
	sess := session.New(cfg)
	return cloudwatchlogs.New(sess)
}

type eventsBuffer struct {
	events     []*cloudwatchlogs.InputLogEvent
	size       int
	batchSize  int
	batchCount int
}

func (w *eventsBuffer) add(e *cloudwatchlogs.InputLogEvent) {
	w.events = append(w.events, e)
	w.size += (len(aws.StringValue(e.Message)) + overheadRow)
}

func (w *eventsBuffer) drain() ([]*cloudwatchlogs.InputLogEvent, int) {
	events := w.events[:]
	size := w.size
	w.events = nil
	w.size = 0
	return events, size
}

func (w *eventsBuffer) ready() bool {
	if w.size >= w.batchSize {
		return true
	} else if len(w.events) >= w.batchCount {
		return true
	}
	return false
}

type service interface {
	PutLogEvents(*cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error)
}
