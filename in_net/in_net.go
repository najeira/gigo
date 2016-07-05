package in_net

import (
	"io"
	"net"

	"github.com/najeira/gigo"
)

var (
	_ io.ReadCloser = (*Reader)(nil)
)

type Config struct {
	Net    string
	Addr   string
	Logger gigo.Logger
}

type Reader struct {
	listener net.Listener
	conn     net.Conn
	logger   gigo.Logger
}

func Open(config Config) (*Reader, error) {
	r := &Reader{
		logger: gigo.EnsureLogger(config.Logger),
	}
	if err := r.open(config.Net, config.Addr); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Reader) open(network, address string) error {
	ln, err := net.Listen(network, address)
	if err != nil {
		r.logger.Warnf("in_net: listen error %s", err)
		return err
	}
	r.listener = ln
	r.logger.Infof("in_net: listen %s %s", network, address)
	return nil
}

func (r *Reader) Read(buf []byte) (int, error) {
	if r.conn == nil {
		conn, err := r.listener.Accept()
		if err != nil {
			r.logger.Warnf("in_net: accept error %s", err)
			return 0, err
		}
		r.conn = conn
	}

	n, err := r.conn.Read(buf)
	if err != nil {
		r.conn.Close()
		r.conn = nil

		if err != io.EOF {
			r.logger.Warnf("in_net: read error %s", err)
			return n, err
		}
		r.logger.Debugf("in_net: read error %s", err)
	} else {
		r.logger.Tracef("in_net: read %d bytes", n)
	}
	return n, nil
}

func (r *Reader) Close() error {
	if r.conn == nil {
		return nil
	}

	err := r.conn.Close()
	r.conn = nil
	if err != nil {
		r.logger.Warnf("in_net: close error %s", err)
	} else {
		r.logger.Infof("in_net: close")
	}
	return err
}
