package cloudwatchlogs

import (
	"errors"
	"strings"
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

func NewWriter(config WriterConfig) (*Writer, error) {
	svc := newClient(config.Region, config.Credentials)
	sequence, err := createStreamIfNotExists(svc, config.Group, config.Stream)
	if err != nil {
		return nil, err
	}
	w := newWriter(config)
	w.svc = svc
	w.sequence = sequence
	go w.run()
	return w, nil
}

func newWriter(config WriterConfig) *Writer {
	if config.Interval <= 0 {
		config.Interval = defaultInterval
	}
	w := &Writer{
		group:    config.Group,
		stream:   config.Stream,
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
	return w
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
	if w.closed == nil {
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
	timer := time.NewTimer(w.interval)
	for {
		if done := w.pull(w.eventCh, timer); done {
			return
		}
	}
	panic("fuga")
}

func (w *Writer) pull(eventCh <-chan *cloudwatchlogs.InputLogEvent, timer *time.Timer) bool {
	select {
	case event, ok := <-eventCh:
		if !ok {
			w.flush() // flush remaining events
			return true
		}
		if flushed := w.addEvent(event); flushed {
			// reset the timer
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(w.interval)
		}
	case <-timer.C:
		w.flush()
		timer.Reset(w.interval)
	}
	return false
}

func (w *Writer) addEvent(event *cloudwatchlogs.InputLogEvent) bool {
	flushed := false
	if w.size >= w.batchSize {
		w.flush()
		flushed = true
	} else if len(w.events) >= w.batchCount {
		w.flush()
		flushed = true
	}
	w.events = append(w.events, event)
	w.size += (len(aws.StringValue(event.Message)) + rowOverhead)
	return flushed
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
		w.Error(err)
		if sequenceToken := getNextSequenceTokenFromError(err); len(sequenceToken) > 0 {
			w.events = append(w.events, events...)
			w.size = size
			w.sequence = &sequenceToken
			w.Infof("retry %d events %d bytes with sequence %s", len(events), size, sequenceToken)
		}
		return err
	} else if resp.RejectedLogEventsInfo != nil {
		errstr := resp.RejectedLogEventsInfo.String()
		w.Error(errstr)
		return errors.New(errstr)
	}

	w.sequence = resp.NextSequenceToken
	w.Infof("put %d events %d bytes sequence %s", len(events), size, aws.StringValue(w.sequence))
	return nil
}

func (w *Writer) Close() error {
	if w.closed == nil {
		return ErrClosed
	}

	if w.eventCh != nil {
		close(w.eventCh)
	}
	<-w.closed
	w.closed = nil
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

func getNextSequenceTokenFromError(err error) string {
	lines := strings.Split(err.Error(), "\n")
	if len(lines) <= 0 {
		return ""
	}

	targetLine := lines[0]
	if !strings.Contains(targetLine, "InvalidSequenceTokenException") {
		return ""
	}

	parts := strings.Split(targetLine, "The next expected sequenceToken is: ")
	if len(parts) <= 1 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
