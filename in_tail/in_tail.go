package in_tail

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"github.com/najeira/gigo"
)

const (
	pluginName = "in_tail"
)

var (
	_ io.ReadCloser = (*Reader)(nil)
)

type Config struct {
	File    string
	Eventer gigo.Eventer
}

type Reader struct {
	cmd     *exec.Cmd
	outPipe io.ReadCloser
	eventer gigo.Eventer
}

func Open(config Config) (*Reader, error) {
	r := &Reader{eventer: config.Eventer}
	if err := r.open(config.File); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Reader) open(file string) error {
	r.cmd = exec.Command("tail", "-n", "0", "-F", file)

	outPipe, err := r.cmd.StdoutPipe()
	if err != nil {
		r.infof("stdout error %s", err)
		return err
	}
	r.outPipe = outPipe

	errPipe, err := r.cmd.StderrPipe()
	if err != nil {
		r.infof("stderr error %s", err)
		return err
	}

	if err := r.cmd.Start(); err != nil {
		r.infof("start error %s", err)
		return err
	}

	go r.scanErrPipe(errPipe)

	r.infof("tail -n 0 -F %s", file)
	return nil
}

func (r *Reader) Read(buf []byte) (int, error) {
	n, err := r.outPipe.Read(buf)
	if err != nil {
		r.debugf("read %s", err)
	} else {
		r.debugf("read %d bytes", n)
	}
	return n, err
}

func (r *Reader) scanErrPipe(pipe io.Reader) error {
	r.debugf("scan stderr")

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		r.infof(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		r.infof("scan error %s", err)
		return err
	}

	r.debugf("scan end")
	return nil
}

func (r *Reader) Close() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	if err := r.cmd.Process.Kill(); err != nil {
		r.infof("kill error %s", err)
		return err
	}

	r.debugf("kill %d", r.cmd.Process.Pid)

	if err := r.cmd.Wait(); err != nil {
		// err will be "signal: killed"
		r.debugf("end %s", err)
	}
	r.cmd = nil

	r.infof("close")
	return nil
}

func (r *Reader) debugf(msg string, args ...interface{}) {
	r.emitf(gigo.Debug, msg, args...)
}

func (r *Reader) infof(msg string, args ...interface{}) {
	r.emitf(gigo.Info, msg, args...)
}

func (r *Reader) errorf(msg string, args ...interface{}) {
	r.emitf(gigo.Err, msg, args...)
}

func (r *Reader) emitf(level int, msg string, args ...interface{}) {
	if r.eventer != nil {
		r.eventer.Emit(pluginName, level, fmt.Sprintf(msg, args...))
	}
}
