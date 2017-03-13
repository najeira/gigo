package cloudwatchlogs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type testWriterService struct {
	events *cloudwatchlogs.PutLogEventsInput
}

func (w *testWriterService) PutLogEvents(events *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
	w.events = events
	return &cloudwatchlogs.PutLogEventsOutput{
		NextSequenceToken: aws.String("dummy sequence token"),
	}, nil
}

func TestWriterWriteFlush(t *testing.T) {
	msg := "this is test"
	group := "test group"
	stream := "test stream"

	svc := &testWriterService{}

	w := newWriter(WriterConfig{
		Group:    group,
		Stream:   stream,
		Interval: time.Hour,
	})
	if w == nil {
		t.FailNow()
	}
	w.svc = svc

	if err := w.Write(msg); err != nil {
		t.Error(err)
	}
	if done := w.pull(w.eventCh, time.NewTimer(time.Hour)); done {
		t.Fail()
	}
	if len(w.events) != 1 {
		t.Fail()
	}
	if w.size != (len(msg) + rowOverhead) {
		t.Fail()
	}

	if err := w.flush(); err != nil {
		t.Error(err)
	}
	if len(w.events) != 0 {
		t.Fail()
	}
	if w.size != 0 {
		t.Fail()
	}

	if svc.events == nil {
		t.Fail()
	} else if len(svc.events.LogEvents) != 1 {
		t.Fail()
	} else {
		if g := aws.StringValue(svc.events.LogGroupName); g != group {
			t.Errorf("invalid group %s expect %s", g, group)
		}
		if s := aws.StringValue(svc.events.LogStreamName); s != stream {
			t.Errorf("invalid stream %s expect %s", s, stream)
		}

		e := svc.events.LogEvents[0]
		if m := aws.StringValue(e.Message); m != msg {
			t.Errorf("invalid message %s expect %s", m, msg)
		}
		if e.Timestamp == nil {
			t.Fail()
		}
	}
}

func TestWriterRunWriteClose(t *testing.T) {
	msg := "this is test"
	group := "test group"
	stream := "test stream"

	svc := &testWriterService{}

	w := newWriter(WriterConfig{
		Group:    group,
		Stream:   stream,
		Interval: time.Hour,
	})
	if w == nil {
		t.FailNow()
	}
	w.svc = svc
	go w.run()

	if err := w.Write(msg); err != nil {
		t.Error(err)
	}
	if err := w.Close(); err != nil {
		t.Error(err)
	}

	if svc.events == nil {
		t.Fail()
	} else if len(svc.events.LogEvents) != 1 {
		t.Fail()
	} else {
		if g := aws.StringValue(svc.events.LogGroupName); g != group {
			t.Errorf("invalid group %s expect %s", g, group)
		}
		if s := aws.StringValue(svc.events.LogStreamName); s != stream {
			t.Errorf("invalid stream %s expect %s", s, stream)
		}

		e := svc.events.LogEvents[0]
		if m := aws.StringValue(e.Message); m != msg {
			t.Errorf("invalid message %s expect %s", m, msg)
		}
		if e.Timestamp == nil {
			t.Fail()
		}
	}
}
