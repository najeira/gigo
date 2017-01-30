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

func NewWriter(config Config) (*Writer, error) {
	svc := newClient(config)
	sequence, err := createStreamIfNotExists(svc, config.Group, config.Stream)
	if err != nil {
		return nil, err
	}

	if config.Interval <= 0 {
		config.Interval = defaultInterval
	}
	w := &Writer{
		svc:      svc,
		group:    config.Group,
		stream:   config.Stream,
		sequence: sequence,
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
	return w, nil
}

func createStreamIfNotExists(client *cloudwatchlogs.CloudWatchLogs, group, stream string) (*string, error) {
	streams, err := client.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group),
	})
	if err != nil {
		return nil, err
	}

	for _, logStream := range streams.LogStreams {
		if aws.StringValue(logStream.LogStreamName) == stream {
			return logStream.UploadSequenceToken, nil
		}
	}

	_, err = client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
	})
	return nil, err
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
				w.Debugf("closing")
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
		w.Error(err)
		w.events.add(events...)
		return err
	}

	if resp.RejectedLogEventsInfo != nil {
		errstr := resp.RejectedLogEventsInfo.String()
		w.Error(errstr)
		return errors.New(errstr)
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
	w.Debugf("closed")
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

func (w *eventsBuffer) add(events ...*cloudwatchlogs.InputLogEvent) {
	for _, event := range events {
		w.events = append(w.events, event)
		w.size += (len(aws.StringValue(event.Message)) + overheadRow)
	}
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
