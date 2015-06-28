package out_fluent

import (
	"fmt"

	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/najeira/gigo"
)

type Config struct {
	fluent.Config
	Tag       string
	FieldName string
	Logger    gigo.Logger
}

type Fluent interface {
	Post(string, interface{}) error
	Close() error
}

type Output struct {
	config    fluent.Config
	tag       string
	fieldName string
	logger    gigo.Logger
	output    Fluent
}

var _ gigo.Output = (*Output)(nil)

func New(config Config) *Output {
	return &Output{
		config:    config.Config,
		tag:       config.Tag,
		fieldName: config.FieldName,
		logger:    config.Logger,
	}
}

func (p *Output) Start() error {
	gigo.Debugf(p.logger, "out_fluent: start")
	if p.output != nil {
		return fmt.Errorf("already started")
	}
	output, err := fluent.New(p.config)
	if err != nil {
		return err
	}
	p.output = output
	return nil
}

func (p *Output) Stop() error {
	gigo.Debugf(p.logger, "out_fluent: stop")
	if p.output == nil {
		return fmt.Errorf("not started")
	}
	p.output.Close()
	return nil
}

func (p *Output) Emit(msg interface{}) error {
	if p.output == nil {
		return fmt.Errorf("not started")
	}
	v := map[string]interface{}{p.fieldName: msg}
	return p.output.Post(p.tag, v)
}
