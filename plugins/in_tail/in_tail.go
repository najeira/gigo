package in_tail

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/najeira/gigo"
)

type Config struct {
	File    string
	Emitter gigo.Emitter
	Logger  gigo.Logger
}

type Input struct {
	file    string
	emitter gigo.Emitter
	logger  gigo.Logger
	cmd     *exec.Cmd
}

var _ gigo.Input = (*Input)(nil)

func New(config Config) *Input {
	return &Input{
		emitter: config.Emitter,
		logger:  config.Logger,
		file:    config.File,
		cmd:     nil,
	}
}

func (p *Input) Start() error {
	gigo.Debugf(p.logger, "in_tail: start")
	if err := p.exec(p.file); err != nil {
		return err
	}
	return nil
}

func (p *Input) Stop() error {
	gigo.Debugf(p.logger, "in_tail: stop")
	if p.cmd != nil && p.cmd.Process != nil {
		gigo.Debugf(p.logger, "in_tail: killing %d", p.cmd.Process.Pid)
		return p.cmd.Process.Kill()
	}
	return nil
}

func (p *Input) exec(file string) error {
	p.cmd = exec.Command("tail", "-n", "0", "-F", file)

	outPipe, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	errPipe, err := p.cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := p.cmd.Start(); err != nil {
		return err
	}

	p.handlePipes(outPipe, errPipe)

	gigo.Debugf(p.logger, "in_tail: tail -n 0 -F %s", file)
	return nil
}

func (p *Input) handlePipes(outPipe, errPipe io.Reader) {
	p.handleOutPipe(outPipe)
	p.handleErrPipe(errPipe)
}

func (p *Input) handleOutPipe(outPipe io.Reader) {
	go p.scan(outPipe, func(line string) {
		p.handleLine(line)
	})
}

func (p *Input) handleErrPipe(errPipe io.Reader) {
	go p.scan(errPipe, func(line string) {
		p.logger.Warnf(line)
	})
}

func (p *Input) handleLine(line string) {
	if p.emitter != nil {
		trimmed := p.trimCrLf(line)
		p.emitter.Emit(trimmed)
	}
}

func (p *Input) trimCrLf(line string) string {
	for len(line) > 0 {
		ch := line[len(line)-1]
		if ch != '\n' && ch != '\r' {
			return line
		}
		line = line[:len(line)]
	}
	return line
}

func (p *Input) scan(reader io.Reader, handler func(line string)) {
	gigo.Debugf(p.logger, "in_tail: scan started")
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		handler(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		gigo.Warnf(p.logger, "in_tail: scan error %v", err)
	}
	gigo.Debugf(p.logger, "in_tail: scan finished")
}
