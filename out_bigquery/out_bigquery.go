package out_bigquery

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/najeira/bigquery"
	"github.com/najeira/gigo"
)

type Config struct {
	Project string
	Dataset string
	Table   string
	Email   string
	Pem     []byte
	Logger  gigo.Logger
}

type Bigquery interface {
	Add(string, map[string]interface{}) error
	Close()
}

type Output struct {
	config Config
	output Bigquery
}

var _ gigo.Output = (*Output)(nil)

var (
	insertIdChars []byte     = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	random        *rand.Rand = nil
)

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func New(config Config) *Output {
	return &Output{
		config: config,
	}
}

func (p *Output) Start() error {
	gigo.Debugf(p.config.Logger, "out_bigquery: start")
	if p.output != nil {
		return fmt.Errorf("already started")
	}

	w := bigquery.NewWriter(p.config.Project, p.config.Dataset, p.config.Table)
	w.SetLogger(p.config.Logger)

	if err := w.Connect(p.config.Email, p.config.Pem); err != nil {
		return err
	}
	p.output = w
	return nil
}

func (p *Output) Stop() error {
	gigo.Debugf(p.config.Logger, "out_bigquery: stop")
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

	v, ok := msg.(map[string]interface{})
	if !ok {
		return fmt.Errorf("not started")
	}

	return p.output.Add(genInsertId(10), v)
}

func genInsertId(length int) string {
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = insertIdChars[random.Int()&len(insertIdChars)]
	}
	return string(buf)
}
