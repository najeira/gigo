package in_tail

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/najeira/gigo"
)

var (
	_ io.ReadCloser = (*Reader)(nil)
)

type Config struct {
	File   string
	Logger gigo.Logger
}

type Reader struct {
	cmd     *exec.Cmd
	outPipe io.ReadCloser
	logger  gigo.Logger
}

func Open(config Config) (*Reader, error) {
	r := &Reader{
		logger: gigo.EnsureLogger(config.Logger),
	}
	if err := r.open(config.File); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Reader) open(file string) error {
	r.cmd = exec.Command("tail", "-n", "0", "-F", file)

	outPipe, err := r.cmd.StdoutPipe()
	if err != nil {
		r.logger.Warnf("in_tail: stdout error %s", err)
		return err
	}
	r.outPipe = outPipe

	errPipe, err := r.cmd.StderrPipe()
	if err != nil {
		r.logger.Warnf("in_tail: stderr error %s", err)
		return err
	}

	if err := r.cmd.Start(); err != nil {
		r.logger.Warnf("in_tail: start error %s", err)
		return err
	}

	go r.scanErrPipe(errPipe)

	r.logger.Infof("in_tail: tail -n 0 -F %s", file)
	return nil
}

func (r *Reader) Read(buf []byte) (int, error) {
	return r.outPipe.Read(buf)
}

func (r *Reader) scanErrPipe(pipe io.Reader) error {
	r.logger.Debugf("in_tail: scan stderr")

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		r.logger.Warnf(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		r.logger.Warnf("in_tail: scan error %s", err)
		return err
	}

	r.logger.Debugf("in_tail: scan end")
	return nil
}

func (r *Reader) Close() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	r.logger.Debugf("in_tail: kill %d", r.cmd.Process.Pid)

	if err := r.cmd.Process.Kill(); err != nil {
		r.logger.Warnf("in_tail: kill error %s", err)
		return err
	}

	if err := r.cmd.Wait(); err != nil {
		// err will be "signal: killed"
		r.logger.Debugf("in_tail: wait %s", err)
	}
	r.cmd = nil

	r.logger.Infof("in_tail: stop")
	return nil
}
