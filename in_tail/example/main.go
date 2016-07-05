package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/najeira/gigo/plugins/in_tail"
	"github.com/najeira/goutils/nlog"
)

type emitter struct {
}

func (e *emitter) Emit(msg interface{}) error {
	if b, ok := msg.([]byte); ok {
		fmt.Println(string(b))
	} else if s, ok := msg.(string); ok {
		fmt.Println(s)
	} else {
		return fmt.Errorf("unknown type")
	}
	return nil
}

func waitSignal(plugin *in_tail.Input) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	if err := plugin.Stop(); err != nil {
		fmt.Println(err)
	}
}

func main() {
	file := os.Args[1]

	logger := nlog.NewLogger(&nlog.Config{
		Level: nlog.Trace,
		Flag:  nlog.LstdFlags,
	})
	e := emitter{}

	plugin := in_tail.New(in_tail.Config{
		File:    file,
		Emitter: &e,
		Logger:  logger,
	})

	go waitSignal(plugin)

	if err := plugin.Start(); err != nil {
		fmt.Println(err)
	}
}
