package in_tail

import (
	"bufio"
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
	File string
}

type Reader struct {
	gigo.Mixin

	file    string
	cmd     *exec.Cmd
	outPipe io.ReadCloser
}

func New(config Config) *Reader {
	r := &Reader{}
	r.Name = pluginName
	r.file = config.File
	return r
}

func (r *Reader) Open() error {
	r.cmd = exec.Command("tail", "-n", "0", "-F", r.file)

	outPipe, err := r.cmd.StdoutPipe()
	if err != nil {
		r.Errorf("stdout error %s", err)
		return err
	}
	r.outPipe = outPipe

	errPipe, err := r.cmd.StderrPipe()
	if err != nil {
		r.Errorf("stderr error %s", err)
		return err
	}

	if err := r.cmd.Start(); err != nil {
		r.Errorf("start error %s", err)
		return err
	}

	go r.scanErrPipe(errPipe)

	r.Debugf("tail -n 0 -F %s", r.file)
	return nil
}

func (r *Reader) Read(buf []byte) (int, error) {
	n, err := r.outPipe.Read(buf)
	if err != nil {
		r.Infof("read %s", err)
	} else {
		r.Debugf("read %d bytes", n)
	}
	return n, err
}

func (r *Reader) scanErrPipe(pipe io.Reader) error {
	r.Debug("scan stderr")

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		r.Infof(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		r.Infof("scan error %s", err)
		return err
	}

	r.Debug("scan end")
	return nil
}

func (r *Reader) Close() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	if err := r.cmd.Process.Kill(); err != nil {
		r.Infof("kill error %s", err)
		return err
	}

	r.Debugf("kill %d", r.cmd.Process.Pid)

	if err := r.cmd.Wait(); err != nil {
		// err will be "signal: interrupt"
		r.Debugf("end %s", err)
	}
	r.cmd = nil
	return nil
}
