package out_file

import (
	"io"
	"os"

	"github.com/najeira/gigo"
)

var (
	_ io.WriteCloser = (*Writer)(nil)
)

type Config struct {
	Name   string
	Flag   int
	Perm   os.FileMode
	Logger gigo.Logger
}

type Writer struct {
	file   *os.File
	logger gigo.Logger
}

func Open(config Config) (*Writer, error) {
	w := &Writer{
		logger: gigo.EnsureLogger(config.Logger),
	}
	if err := w.open(config.Name, config.Flag, config.Perm); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Writer) open(name string, flag int, perm os.FileMode) error {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		w.logger.Warnf("out_file: start error %s", err)
		return err
	}
	w.file = f
	w.logger.Infof("out_file: start")
	return nil
}

func (w *Writer) Write(msg []byte) (int, error) {
	n, err := w.file.Write(msg)
	if err != nil {
		w.logger.Warnf("out_file: write error %s", err)
	} else {
		w.logger.Debugf("out_file: write %d bytes", n)
	}
	return n, err
}

func (w *Writer) Close() error {
	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	if err != nil {
		w.logger.Warnf("out_file: stop error %s", err)
		return err
	}

	w.logger.Infof("out_file: stop")
	return nil
}
