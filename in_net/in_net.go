package in_net

import (
	"io"
	"net"

	"github.com/najeira/gigo"
)

var (
	_ io.Closer = (*Reader)(nil)
)

type Handler func(net.Conn)

type Config struct {
	Net     string
	Addr    string
	Handler Handler
	Logger  gigo.Logger
}

type Reader struct {
	listener net.Listener
	handler  Handler
	logger   gigo.Logger
}

func Open(config Config) (*Reader, error) {
	r := &Reader{
		handler: config.Handler,
		logger:  gigo.EnsureLogger(config.Logger),
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

	go r.accept(ln)

	return nil
}

func (r *Reader) accept(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			r.logger.Warnf("in_net: accept error %s", err)
			return
		}
		r.logger.Debugf("in_net: accept %s->%s",
			conn.LocalAddr().String(), conn.RemoteAddr().String())
		go r.handleConn(conn)
	}
	panic("unreachable")
}

func (r *Reader) handleConn(conn net.Conn) {
	defer conn.Close()
	if r.handler != nil {
		r.handler(conn)
	}
}

func (r *Reader) Close() error {
	err := r.listener.Close()
	if err != nil {
		r.logger.Warnf("in_net: listener close error %s", err)
	} else {
		r.logger.Debugf("in_net: listener close")
	}
	return err
}
