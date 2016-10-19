package main

import (
	"bufio"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/najeira/gigo"
	"github.com/najeira/gigo/in_tail"
)

const (
	commandName = "tail_s3"
)

var (
	trapSignals = []os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT}
	lineEnd = []byte{'\n'}
)

type inTailOutS3 struct {
	gigo.Mixin

	config Config
	input  *in_tail.Reader
	output *Writer
}

func newInTailOutS3(config Config, logger *logger) *inTailOutS3 {
	p := inTailOutS3{config: config}
	p.Name = commandName
	p.LogLevel = gigo.ParseLogLevel(config.LogLevel)
	p.Logger = logger
	return &p
}

func (p *inTailOutS3) initInput() error {
	input := in_tail.New(in_tail.Config{
		File: p.config.File,
	})
	input.LogLevel = gigo.ParseLogLevel(p.config.LogLevelTail)
	input.Logger = p.Logger
	if err := input.Open(); err != nil {
		return err
	}
	p.input = input
	return nil
}

func (p *inTailOutS3) initOutput() error {
	output, err := NewWriter(WriterConfig{
		Key:               p.config.Key,
		Secret:            p.config.Secret,
		Region:            p.config.Region,
		Bucket:            p.config.Bucket,
		Path:              p.config.Path,
		Hostname:          p.config.Hostname,
		PublicRead:        p.config.PublicRead,
		ReducedRedundancy: p.config.ReducedRedundancy,
		TimeFormat:        p.config.TimeFormat,
		BufferSize:        p.config.BufferSize,
		FlushInterval:     p.config.FlushInterval,
	})
	if err != nil {
		return err
	}
	output.LogLevel = gigo.ParseLogLevel(p.config.LogLevelS3)
	output.Logger = p.Logger
	p.output = output
	return nil
}

func (p *inTailOutS3) run() error {
	if err := p.initInput(); err != nil {
		p.Error(err)
		return err
	}
	if err := p.initOutput(); err != nil {
		p.Error(err)
		return err
	}

	if pprofile := os.Getenv("PPROF"); pprofile != "" {
		f, err := os.Create(pprofile)
		if err != nil {
			p.Error(err)
			return err
		}
		p.Infof("profiling file %s", f.Name())
		pprof.StartCPUProfile(f)
	}

	go p.waitSignals()

	p.loop()

	pprof.StopCPUProfile()
	return nil
}

func (p *inTailOutS3) loop() {
	p.Info("start")
	scanner := bufio.NewScanner(p.input)
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		data := scanner.Bytes()
		if len(data) > 0 {
			p.output.Write(data)
			p.output.Write(lineEnd)
		}
	}
	if err := scanner.Err(); err != nil {
		p.Info(err)
	}
	p.output.Close()
	p.Info("end")
}

func (p *inTailOutS3) waitSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, trapSignals...)
	if sig, ok := <-sigCh; ok {
		p.Infof("signal %s", sig)
	}

	go func() {
		time.Sleep(10 * time.Second)
		if sig, ok := <-sigCh; ok {
			p.Errorf("signal %s before shutdown completed", sig)
			os.Exit(1)
		}
	}()

	if err := p.input.Close(); err != nil {
		p.Error(err)
	}
}
