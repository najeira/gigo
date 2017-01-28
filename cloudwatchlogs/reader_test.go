package cloudwatchlogs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type testReaderService struct {
	input  *cloudwatchlogs.GetLogEventsInput
	output *cloudwatchlogs.GetLogEventsOutput
}

func (s *testReaderService) GetLogEvents(input *cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
	s.input = input
	return s.output, nil
}

func TestReaderRead(t *testing.T) {
	output := &cloudwatchlogs.GetLogEventsOutput{
		Events: []*cloudwatchlogs.OutputLogEvent{
			{
				Message:       aws.String("hoge"),
				Timestamp:     aws.Int64(timeToMilli(time.Now())),
				IngestionTime: aws.Int64(timeToMilli(time.Now())),
			},
		},
		NextBackwardToken: aws.String("back"),
		NextForwardToken:  aws.String("forward"),
	}
	group := "test group"
	stream := "test stream"
	token := "test token"
	svc := &testReaderService{output: output}

	r := NewReader(ReaderConfig{
		Group:     group,
		Stream:    stream,
		NextToken: token,
	})
	if r == nil {
		t.FailNow()
	}
	r.svc = svc

	got, err := r.Read()
	if g := aws.StringValue(svc.input.LogGroupName); g != group {
		t.Errorf("invalid group %s expect %s", g, group)
	}
	if s := aws.StringValue(svc.input.LogStreamName); s != stream {
		t.Errorf("invalid stream %s expect %s", s, stream)
	}

	if err != nil {
		t.Error(err)
	} else if got == nil {
		t.Error("Read returns nil")
	} else {
		gotM := aws.StringValue(got.Message)
		expM := aws.StringValue(output.Events[0].Message)
		if gotM != expM {
			t.Errorf("invalid message %s expect %s", gotM, expM)
		}

		gotT := aws.Int64Value(got.Timestamp)
		expT := aws.Int64Value(output.Events[0].Timestamp)
		if gotT != expT {
			t.Errorf("invalid timestamp %d expect %d", gotT, expT)
		}

		gotI := aws.Int64Value(got.IngestionTime)
		expI := aws.Int64Value(output.Events[0].IngestionTime)
		if gotI != expI {
			t.Errorf("invalid timestamp %d expect %d", gotI, expI)
		}
	}
}
