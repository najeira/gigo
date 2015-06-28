package out_fluent

import (
	"fmt"
	"testing"

	"github.com/fluent/fluent-logger-golang/fluent"
	//"github.com/najeira/gigo/testutil"
)

type value struct {
	tag string
	msg interface{}
}

type testFluent struct {
	messages []value
}

func (f *testFluent) Post(tag string, msg interface{}) error {
	f.messages = append(f.messages, value{tag: tag, msg: msg})
	return nil
}

func (f *testFluent) Close() error {
	return nil
}

func checkValue(v value, tag, field, str string) error {
	if v.tag != tag {
		return fmt.Errorf("invalid tag")
	}
	m, ok := v.msg.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid type")
	}
	s, ok := m[field]
	if !ok {
		return fmt.Errorf("field not found")
	}
	if s != str {
		return fmt.Errorf("invalid string")
	}
	return nil
}

func TestEmit(t *testing.T) {
	f := testFluent{}
	o := New(Config{
		Config:    fluent.Config{},
		Tag:       "tag",
		FieldName: "message",
	})
	o.output = &f

	var err error

	err = o.Emit("hoge")
	if err != nil {
		t.Error(err)
	}

	err = o.Emit("fuga")
	if err != nil {
		t.Error(err)
	}

	err = o.Emit("piyo")
	if err != nil {
		t.Error(err)
	}

	if len(f.messages) != 3 {
		t.Errorf("invalid len")
	}

	if err = checkValue(f.messages[0], "tag", "message", "hoge"); err != nil {
		t.Error(err)
	}
}
