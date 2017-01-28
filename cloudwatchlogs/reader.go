package cloudwatchlogs

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

const (
	pullInterval = time.Second * 3
)

type ReaderConfig struct {
	Credentials   *credentials.Credentials
	Region        string
	Group         string
	Stream        string
	NextToken     string
	StartTime     time.Time
	EndTime       time.Time
	StartFromHead bool
}

type Reader struct {
	svc           readerService
	group         *string
	stream        *string
	nextToken     *string
	startTime     *int64
	endTime       *int64
	startFromHead *bool
	nextPullTime  time.Time
	eventCh       chan *cloudwatchlogs.OutputLogEvent
}

func NewReader(config ReaderConfig) *Reader {
	r := &Reader{
		svc:     newClient(config.Region, config.Credentials),
		group:   aws.String(config.Group),
		stream:  aws.String(config.Stream),
		eventCh: make(chan *cloudwatchlogs.OutputLogEvent),
	}
	if config.NextToken != "" {
		r.nextToken = aws.String(config.NextToken)
	}
	if !config.StartTime.IsZero() {
		r.startTime = aws.Int64(timeToMilli(config.StartTime))
	}
	if !config.EndTime.IsZero() {
		r.endTime = aws.Int64(timeToMilli(config.EndTime))
	}
	if config.StartFromHead {
		r.startFromHead = aws.Bool(config.StartFromHead)
	}
	return r
}

func (r *Reader) Read() (*cloudwatchlogs.OutputLogEvent, error) {
	for {
		select {
		case event, ok := <-r.eventCh:
			if !ok {
				return nil, ErrClosed
			}
			return event, nil
		default:
			if err := r.pullEvents(); err != nil {
				return nil, err
			}
		}
	}
}

func (r *Reader) pullEvents() error {
	for {
		if remain := r.nextPullTime.Sub(time.Now()); remain > 0 {
			time.Sleep(remain)
		}
		events, err := r.getEvents()
		r.nextPullTime = time.Now().Add(pullInterval)
		if err != nil {
			return err
		} else if len(events) > 0 {
			go r.enqueEvents(events)
			return nil
		}
	}
}

func (r *Reader) enqueEvents(events []*cloudwatchlogs.OutputLogEvent) {
	for _, event := range events {
		r.eventCh <- event
	}
}

func (r *Reader) getEvents() ([]*cloudwatchlogs.OutputLogEvent, error) {
	params := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  r.group,
		LogStreamName: r.stream,
		NextToken:     r.nextToken,
		StartTime:     r.startTime,
		EndTime:       r.endTime,
		StartFromHead: r.startFromHead,
	}
	res, err := r.svc.GetLogEvents(params)
	if err != nil {
		return nil, err
	} else if res.NextForwardToken == nil {
		panic("nextForwardToken is nil")
	}
	r.nextToken = res.NextForwardToken
	return res.Events, nil
}

func (r *Reader) NextToken() string {
	return aws.StringValue(r.nextToken)
}

type readerService interface {
	GetLogEvents(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error)
}

func timeToMilli(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func milliToTime(ms int64) time.Time {
	return time.Unix(ms/1000, (ms%1000)*1000000)
}
